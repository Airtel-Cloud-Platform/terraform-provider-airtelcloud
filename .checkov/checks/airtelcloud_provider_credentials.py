"""CKV_ACP_1: no hard-coded credentials in the airtelcloud provider block.

Mirrors Checkov's built-in CKV_AWS_41, but for our provider. Catches the
exact pattern that leaks keys through examples/ and docs:

    provider "airtelcloud" {
      api_key    = "<real key pasted here>"   # FAILS
      api_secret = var.airtel_api_secret      # passes (unresolved variable)
    }

Note: Checkov resolves variable defaults, so `variable "airtel_api_secret" { default = "<real key>" }`
also FAILS, which is what we want.

CREDENTIAL_ATTRIBUTES matches the Sensitive arguments in internal/provider/provider.go.
"""
import re
from typing import Any, Dict, List

from checkov.common.models.enums import CheckCategories, CheckResult
from checkov.terraform.checks.provider.base_check import BaseProviderCheck

CREDENTIAL_ATTRIBUTES = (
    "api_key",     # provider.go: Sensitive
    "api_secret",  # provider.go: Sensitive
)

# Values that are clearly not literals: unresolved references or env-var style placeholders.
_REFERENCE_PREFIXES = ("var.", "${var.", "local.", "${local.", "data.", "${data.")
# Doc placeholders such as "your-api-key", "<API_KEY>", "..." are fine.
_PLACEHOLDER_RE = re.compile(r"(?i)^(|\.\.\.|your[-_].*|<[^>]+>|replace[-_]?me|changeme|x{4,})$")


class AirtelCloudProviderCredentials(BaseProviderCheck):
    def __init__(self) -> None:
        super().__init__(
            name="Ensure no hard-coded credentials exist in the airtelcloud provider block",
            id="CKV_ACP_1",
            categories=[CheckCategories.SECRETS],
            supported_provider=["airtelcloud"],
        )

    def scan_provider_conf(self, conf: Dict[str, List[Any]]) -> CheckResult:
        for attr in CREDENTIAL_ATTRIBUTES:
            if attr not in conf:
                continue
            raw = conf[attr]
            value = raw[0] if isinstance(raw, list) and raw else raw
            if not isinstance(value, str):
                continue
            value = value.strip()
            if _PLACEHOLDER_RE.match(value) or value.startswith(_REFERENCE_PREFIXES):
                continue
            # Record the offending field so it appears in the report (value is masked by Checkov).
            conf[f"{self.id}_secret_{attr}"] = value
            return CheckResult.FAILED
        return CheckResult.PASSED


check = AirtelCloudProviderCredentials()
