# Project Index: Airtel Cloud Terraform Provider

Generated: 2026-07-31

## Project Structure

```
terraform-provider/
├── main.go                    # Provider entry point
├── go.mod / go.sum            # Go module (1.24.0, plugin-framework v1.16.1)
├── Makefile                   # build, install, test, testacc, fmt, lint, docs-generate
├── docs/                      # Auto-generated terraform-plugin-docs
│   ├── index.md
│   ├── guides/getting-started.md
│   └── resources/             # 22 resource doc files
├── examples/
│   ├── complete/              # Full working example
│   ├── resources/             # Per-resource examples (21 dirs)
│   └── import/                # Import example
├── internal/
│   ├── client/                # HTTP client layer (18 files + 14 unit + 12 integration tests)
│   │   └── testutil/          # Mock server for unit tests
│   ├── models/                # Data structures (15 files)
│   └── provider/              # Terraform resources (22 files + provider.go)
└── tests/resources/           # 7 acceptance tests + helpers
```

## Stats

- **Go source files**: 57 (excl. tests)
- **Test files**: 31 (13 unit, 10 integration, 7 acceptance, 1 helper)
- **Total source lines**: ~13,200 (excl. tests)
- **Provider version**: 1.2.1

## Entry Points

- **Binary**: `main.go` - Provider server via `providerserver.Serve()`
- **Provider**: `internal/provider/provider.go` (190 lines) - Resource registration, schema, config
- **Client**: `internal/client/client.go` (961 lines) - Base HTTP client, HMAC auth, all HTTP methods
- **Tests**: `tests/resources/` (acceptance), `internal/client/*_test.go` (unit/integration)

## Resources (22 registered)

| Resource | File | Lines |
|----------|------|-------|
| `airtelcloud_vm` | `vm_resource.go` | 842 |
| `airtelcloud_volume` | `volume_resource.go` | 761 |
| `airtelcloud_lb_virtual_server` | `lb_virtual_server_resource.go` | 604 |
| `airtelcloud_object_storage_bucket` | `object_storage_bucket_resource.go` | 419 |
| `airtelcloud_compute_snapshot` | `compute_snapshot_resource.go` | 418 |
| `airtelcloud_protection` | `protection_resource.go` | 417 |
| `airtelcloud_dns_record` | `dns_record_resource.go` | 416 |
| `airtelcloud_vpc_peering` | `vpc_peering_resource.go` | 410 |
| `airtelcloud_security_group_rule` | `security_group_rule_resource.go` | 374 |
| `airtelcloud_lb_service` | `lb_service_resource.go` | 374 |
| `airtelcloud_public_ip_policy_rule` | `public_ip_policy_rule_resource.go` | 364 |
| `airtelcloud_subnet` | `subnet_resource.go` | 333 |
| `airtelcloud_file_storage_export_path` | `file_storage_export_path_resource.go` | 331 |
| `airtelcloud_dns_zone` | `dns_zone_resource.go` | 297 |
| `airtelcloud_file_storage` | `file_storage_resource.go` | 296 |
| `airtelcloud_protection_plan` | `protection_plan_resource.go` | 295 |
| `airtelcloud_lb_vip` | `lb_vip_resource.go` | 284 |
| `airtelcloud_public_ip` | `public_ip_resource.go` | 280 |
| `airtelcloud_lb_certificate` | `lb_certificate_resource.go` | 270 |
| `airtelcloud_object_storage_access_key` | `object_storage_access_key_resource.go` | 268 |
| `airtelcloud_vpc` | `vpc_resource.go` | 265 |
| `airtelcloud_security_group` | `security_group_resource.go` | 235 |

## Client Layer (`internal/client/`)

| File | Lines | Purpose |
|------|-------|---------|
| `client.go` | 961 | Base HTTP client, HMAC-SHA256 auth, JSON/form/URL-encoded methods |
| `snapshot.go` | 288 | Compute snapshot operations |
| `lb_service.go` | 237 | Load balancer service CRUD |
| `vm.go` | 236 | Compute CRUD + actions (start/stop/reboot) |
| `lb_virtual_server.go` | 231 | LB virtual server CRUD |
| `backup.go` | 229 | Backup/protection plan operations |
| `volume.go` | 209 | Block storage CRUD + attach/detach |
| `public_ip.go` | 203 | Public IP allocation + policy rules |
| `subnet.go` | 178 | Subnet operations |
| `security_group.go` | 169 | Security group + rule CRUD |
| `compute.go` | 159 | Additional compute helpers (snapshots, protection) |
| `object_storage.go` | 139 | Object storage buckets + access keys |
| `nfs.go` | 136 | File storage + NFS exports |
| `vpc_peering.go` | 112 | VPC peering create/get/accept/reject/delete |
| `dns_record.go` | 110 | DNS record CRUD |
| `vpc.go` | 108 | VPC operations |
| `baremetal.go` | 107 | Baremetal node listing + resolution |
| `dns_zone.go` | 95 | DNS zone CRUD |

## Models Layer (`internal/models/`)

| File | Lines | Key Types |
|------|-------|-----------|
| `object_storage.go` | 172 | Bucket, access key models |
| `vm.go` | 160 | `Compute`, `CreateComputeRequest` (form tags) |
| `nfs.go` | 142 | `FileStorageVolume`, `NFSExportInfo`, `NFSAccessRule` |
| `volume.go` | 120 | `Volume` (int ID), `VolumeAttachment` |
| `lb_service.go` | 103 | LB service, VIP, certificate, virtual server models |
| `public_ip.go` | 84 | Public IP + policy rule models |
| `backup.go` | 80 | Protection plan models |
| `vpc.go` | 70 | VPC structs with tag support |
| `subnet.go` | 59 | Subnet structs |
| `dns_record.go` | 59 | `DNSRecord` |
| `snapshot.go` | 58 | Compute snapshot model |
| `security_group.go` | 48 | Security group models |
| `baremetal.go` | 40 | Baremetal node models |
| `vpc_peering.go` | 39 | VPC peering models |
| `dns_zone.go` | 33 | `DNSZone` |

## Architecture

```
Provider Layer (internal/provider/)       <- Terraform CRUD + Schema
    ↓ uses client methods
Client Layer (internal/client/)           <- HTTP + HMAC-SHA256 auth
    ↓ sends/receives
Models Layer (internal/models/)           <- Request/response structs
    ↓ HTTP requests
Airtel Cloud API
```

**Key Patterns**:
- HMAC-SHA256 auth via `Ce-Auth` header (apiKey.expiry.signature)
- Form-data encoding for VM/Volume APIs; JSON for network/DNS/storage
- URL-encoded form for some LB/compute operations
- Query-param POST/PATCH for snapshot/protection operations
- Integer IDs for volumes, string IDs for compute/other resources
- `WithSubnetID()` / `WithAvailabilityZone()` client modifiers
- 404 handling: graceful state removal for already-deleted resources
- Timeout support on subnet and VPC peering resources

## Resource Dependencies

```
VPC -> Subnet -> VM <- Volume
                  VM <- Compute Snapshot
                  VM <- Protection (Plan)
                  VM <- Public IP (Policy Rule)
File Storage -> NFS Export Path
DNS Zone -> DNS Record
Security Group -> Security Group Rule
VPC Peering (cross-VPC)
LB Service -> LB VIP
LB Service -> LB Certificate
LB Service -> LB Virtual Server
Object Storage (independent)
Baremetal (independent, read-only helper)
```

## Test Coverage

| Category | Count | Location |
|----------|-------|----------|
| Unit tests | 14 | `internal/client/*_test.go` |
| Integration tests | 12 | `internal/client/*_integration_test.go` |
| Acceptance tests | 7 | `tests/resources/*_test.go` |
| Provider tests | 2 | `internal/provider/*_test.go` |
| Mock server | 1 | `internal/client/testutil/mock_server.go` |

## Configuration

| File | Purpose |
|------|---------|
| `go.mod` | terraform-plugin-framework v1.16.1, Go 1.24.0 |
| `Makefile` | build, install, test, testacc, fmt, lint, docs-generate |
| `.env.example` | API credentials and region config template |

## Quick Start

```bash
make build                     # Compile provider
make install                   # Install to ~/.terraform.d/plugins/
go test ./internal/client/ -v  # Unit tests
TF_ACC=1 make testacc          # Acceptance tests (needs credentials)
make docs-generate             # Generate resource docs
```

## Provider Configuration

```hcl
provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"  # optional
  api_key      = var.api_key                       # required
  api_secret   = var.api_secret                    # required
  region       = "south"                           # optional
  organization = "my-org"                          # optional
  project_name = "my-project"                      # optional
  subnet_id    = "subnet-abc"                      # optional
}
```