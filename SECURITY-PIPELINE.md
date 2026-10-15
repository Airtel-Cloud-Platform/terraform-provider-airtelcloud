# Security pipeline: terraform-provider-airtelcloud

Every change goes through four layers before it can reach registry.terraform.io:

| Layer | Where | Can be bypassed? | What it stops |
|---|---|---|---|
| 1. pre-commit / pre-push hooks | developer laptop | yes (`--no-verify`) | most leaks, before they ever leave the laptop |
| 2. GitHub push protection | github.com, on `git push` | only with a logged, reviewable bypass | known token formats (AWS, GCP, GitHub, ...) |
| 3. `security` workflow, required check | every PR and push to `main` | no (branch ruleset) | secrets in any commit of the PR, IaC/SAST/dependency issues |
| 4. Release gate | `release.yml` on `v*` tags | no | releasing a tag not on `main`, or anything that fails a full-history scan |

## What's in this folder

| File | Purpose |
|---|---|
| `.github/workflows/security.yml` | The gate. Jobs: `secrets` (TruffleHog), `iac` (Checkov), `sast` (Semgrep), `deps` (govulncheck + Trivy), `workflows` (zizmor), and `gate`, which is the one required check. |
| `.github/workflows/release.yml` | Release flow: `verify-tag` -> `security` -> `release` (GoReleaser sign + publish). Signing key is read from the `release` environment. |
| `.pre-commit-config.yaml` | Local hooks: TruffleHog on staged files and on commits being pushed, Trivy Airtel key rules, Checkov, terraform fmt, golangci-lint. |
| `.checkov.yaml`, `.checkov/checks/` | Checkov config plus custom check `CKV_ACP_1`: no literal credentials in `provider "airtelcloud"` blocks. |
| `.semgrep/provider-rules.yml` | Provider rules: credential-like schema attributes must be `Sensitive: true` (framework and SDKv2), no credentials in `tflog`/`log`/`fmt`, no `InsecureSkipVerify`, no literal credentials in Go. |
| `trivy.yaml`, `trivy-secret.yaml`, `.trivyignore` | Trivy config, plus custom rules for Airtel Cloud keys (TruffleHog has no detector for them). |
| `.trufflehog-exclude.txt` | Paths TruffleHog skips. Keep it short. |
| `.github/dependabot.yml` | Weekly bumps for pinned action SHAs and Go modules. |
| `.github/CODEOWNERS` | DevOps review required on any change to the gate. |

## Two things that would have let the recent leak through

1. **TruffleHog cannot see Airtel Cloud keys.** It verifies 800+ third-party key types, but has no detector for our own cloud's keys. `trivy-secret.yaml` adds rules keyed on variable names (`airtelcloud_api_key = "..."`, `AIRTELCLOUD_SECRET_KEY=...`) and on literal values inside `provider "airtelcloud" {}`. Once the exact key format (prefix/length) is confirmed, add a format-only rule as noted in the file's TODO. That catches keys even when they aren't next to a recognisable name.
2. **Trivy skips examples, tests and Markdown by default.** Its built-in allow rules ignore any path containing `example`, any `*_test.go` and every `*.md` file, which covers `examples/`, acceptance tests and `docs/`. `trivy-secret.yaml` turns those allow rules off (`disable-allow-rules`).

## Rollout (in this order)

### Step 0: Contain the leaked keys (today, before anything else)

Scanning prevents the next leak. It does nothing about the current one: the repo is public, so assume the key was harvested within minutes of the push.

1. **Revoke and rotate** the leaked key in Airtel Cloud IAM. This is the only step that actually closes the exposure.
2. **Check what the key did**: audit logs for that key from the time of the commit until revocation.
3. **Find every place it's still published**:
   ```bash
   git log --all -p -S '<FIRST_8_CHARS_OF_KEY>' --format='%H %an %ad'
   git tag --contains <LEAK_COMMIT_SHA>   # these tags' auto-generated source zips/tarballs on GitHub Releases contain the key
   ```
4. **Rewriting history is optional and secondary.** It cleans the default view, but forks, clones, GitHub's cached views and released source archives keep the key. If you do it, use `git filter-repo --replace-text`, force-push, and ask GitHub Support to purge cached views. Rotation already makes the key worthless.
5. Once rotated, the key comes back as *unverified* in TruffleHog, so the weekly full-history scan passes without needing an exclusion.

### Step 1: Turn on GitHub's native protections (free for public repos)

Settings -> Code security:
- **Secret scanning: on**
- **Push protection: on.** Blocks known token formats at `git push`, server-side. Strongest single control here.
- **Private vulnerability reporting: on**
- **Dependabot alerts and security updates: on**

Settings -> Actions -> General:
- Workflow permissions: **Read repository contents** (default token read-only)
- **Require approval for all outside collaborators** for fork PR workflows
- Allowed actions: if org policy allows, restrict to verified creators plus the actions used here, and **require actions to be pinned to a full-length commit SHA**

### Step 2: Merge the files in a PR

Open the PR from branch `security/scanning-pipeline`. Before merging:
- `.github/CODEOWNERS`: replace `@Airtel-Cloud-Platform/devops` with your real team or usernames.
- Create the `release` environment and move `GPG_PRIVATE_KEY` into it as an environment secret (see the integration guide).

The `security` workflow runs on that PR itself. Expect a first round of findings to triage (see Tuning).

### Step 3: Make it required

Settings -> Rules -> Rulesets -> New branch ruleset, target `main`:
- Require a pull request before merging (1 approval, dismiss stale approvals, require review from Code Owners)
- Require status checks: **`security / gate`** (plus your existing CI checks)
- Block force pushes; restrict deletions

New tag ruleset, target `v*`:
- Restrict creations/updates/deletions to the release managers (or a bypass list), so not everyone with write access can trigger a public release.

### Step 4: Developer setup (each engineer, once)

```bash
# macOS
brew install pre-commit trufflehog trivy golangci-lint
# Linux: pipx install pre-commit; install trufflehog and trivy from their GitHub releases

cd terraform-provider-airtelcloud
pre-commit install          # installs both pre-commit and pre-push hooks
pre-commit run --all-files  # first run downloads hook environments
```

Checkov runs via pre-commit's own Python environment. TruffleHog and Trivy must be on `PATH`.

## How each scanner is gated

| Scanner | Blocks on | Reported only |
|---|---|---|
| TruffleHog | `verified` (live key) and `unknown` (couldn't verify: treated as live) | `unverified` candidates in the weekly artifact |
| Checkov | any failed check (incl. `CKV_ACP_1`, secrets, GitHub Actions checks) | n/a |
| Semgrep | `ERROR` severity | `WARNING`/`INFO` in Security -> Code scanning |
| govulncheck | vulnerable functions the provider actually calls | n/a |
| Trivy | `HIGH`/`CRITICAL` vulns with a fix available; any secret match | everything else, in code scanning |
| zizmor | `medium`+ (template injection, `pull_request_target` misuse, unpinned actions, credential persistence) | `low` |

TruffleHog scan range: PRs scan `base..head`; pushes to `main` scan `before..after`; the weekly schedule, manual runs and releases scan the full history of all branches and tags, plus issue and PR comments.

## Tuning false positives

Every suppression needs a reason. CODEOWNERS makes them visible in review.

- **Checkov:** inline `# checkov:skip=CKV_ID:reason` on the resource, or `skip-check` in `.checkov.yaml`.
- **Semgrep:** `// nosemgrep: rule-id` on the line, with a reason.
- **Trivy:** CVE ID in `.trivyignore` with `exp:YYYY-MM-DD`; it starts failing again after expiry. For secrets, add to `allow-rules` in `trivy-secret.yaml`.
- **TruffleHog:** fake test credentials should use obvious placeholders. Path exclusions go in `.trufflehog-exclude.txt` as a last resort.

## Supply-chain notes

- Every third-party action is pinned to a full commit SHA, and tool versions are pinned (TruffleHog 3.99.2, Trivy v0.76.0, Semgrep 1.181.0, govulncheck v1.9.0, zizmor 1.31.0, GoReleaser v2.18.3). Tags are mutable. In March 2026 attackers force-pushed 76 of 77 `aquasecurity/trivy-action` tags to credential-stealing code (CVE-2026-33634). A SHA pin is not affected by that.
- The TruffleHog binary is downloaded and checked against the release checksum file before it runs.
- `persist-credentials: false` on every checkout, so the token isn't left in `.git/config` for later steps.
- The signing job doesn't use the Go build cache, and its secrets live in a `release` environment the other jobs can't read.
