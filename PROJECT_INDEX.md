# Project Index: Airtel Cloud Terraform Provider

Generated: 2026-04-06

## Project Structure

```
terraform-provider/
├── main.go                    # Provider entry point
├── go.mod / go.sum            # Go module (1.24.0, plugin-framework v1.16.1)
├── Makefile                   # build, install, test, testacc, fmt, lint, docs-generate
├── docs/                      # Auto-generated terraform-plugin-docs
│   ├── index.md
│   ├── guides/getting-started.md
│   └── resources/             # 9 resource doc files
├── examples/
│   ├── complete/              # Full working example
│   ├── resources/             # Per-resource examples (17 dirs)
│   └── import/                # Import example
├── internal/
│   ├── client/                # HTTP client layer (17 files + 13 unit + 10 integration tests)
│   │   └── testutil/          # Mock server for unit tests
│   ├── models/                # Data structures (14 files)
│   └── provider/              # Terraform resources (22 files + provider.go)
└── tests/resources/           # 7 acceptance tests + helpers
```

## Stats

- **Go source files**: 57 (excl. tests)
- **Test files**: 31 (13 unit, 10 integration, 7 acceptance, 1 helper)
- **Total source lines**: ~13,200 (excl. tests)
- **Provider version**: 0.2.0

## Entry Points

- **Binary**: `main.go` - Provider server via `providerserver.Serve()`
- **Provider**: `internal/provider/provider.go` (203 lines) - Resource registration, schema, config
- **Client**: `internal/client/client.go` (907 lines) - Base HTTP client, HMAC auth, all HTTP methods
- **Tests**: `tests/resources/` (acceptance), `internal/client/*_test.go` (unit/integration)

## Resources (22 registered)

| Resource | File | Lines |
|----------|------|-------|
| `airtelcloud_vm` | `vm_resource.go` | 724 |
| `airtelcloud_volume` | `volume_resource.go` | 601 |
| `airtelcloud_lb_virtual_server` | `lb_virtual_server_resource.go` | 423 |
| `airtelcloud_object_storage_bucket` | `object_storage_bucket_resource.go` | 419 |
| `airtelcloud_dns_record` | `dns_record_resource.go` | 416 |
| `airtelcloud_vpc_peering` | `vpc_peering_resource.go` | 410 |
| `airtelcloud_subnet` | `subnet_resource.go` | 333 |
| `airtelcloud_file_storage_export_path` | `file_storage_export_path_resource.go` | 331 |
| `airtelcloud_public_ip_policy_rule` | `public_ip_policy_rule_resource.go` | 327 |
| `airtelcloud_security_group_rule` | `security_group_rule_resource.go` | 314 |
| `airtelcloud_lb_service` | `lb_service_resource.go` | 310 |
| `airtelcloud_protection` | `protection_resource.go` | 304 |
| `airtelcloud_dns_zone` | `dns_zone_resource.go` | 297 |
| `airtelcloud_file_storage` | `file_storage_resource.go` | 296 |
| `airtelcloud_object_storage_access_key` | `object_storage_access_key_resource.go` | 268 |
| `airtelcloud_vpc` | `vpc_resource.go` | 265 |
| `airtelcloud_public_ip` | `public_ip_resource.go` | 260 |
| `airtelcloud_protection_plan` | `protection_plan_resource.go` | 253 |
| `airtelcloud_security_group` | `security_group_resource.go` | 235 |
| `airtelcloud_lb_certificate` | `lb_certificate_resource.go` | 233 |
| `airtelcloud_lb_vip` | `lb_vip_resource.go` | 215 |
| `airtelcloud_compute_snapshot` | `compute_snapshot_resource.go` | 202 |

## Client Layer (`internal/client/`)

| File | Lines | Purpose |
|------|-------|---------|
| `client.go` | 907 | Base HTTP client, HMAC-SHA256 auth, JSON/form/URL-encoded methods |
| `volume.go` | 209 | Block storage CRUD + attach/detach |
| `vm.go` | 207 | Compute CRUD + actions (start/stop/reboot) |
| `subnet.go` | 178 | Subnet operations |
| `lb_service.go` | 177 | Load balancer service CRUD |
| `lb_virtual_server.go` | 173 | LB virtual server CRUD |
| `compute.go` | 145 | Additional compute helpers (snapshots, protection) |
| `public_ip.go` | 144 | Public IP allocation + policy rules |
| `object_storage.go` | 139 | Object storage buckets + access keys |
| `nfs.go` | 136 | File storage + NFS exports |
| `security_group.go` | 127 | Security group + rule CRUD |
| `backup.go` | 112 | Backup/protection plan operations |
| `vpc_peering.go` | 112 | VPC peering create/get/accept/reject/delete |
| `dns_record.go` | 110 | DNS record CRUD |
| `vpc.go` | 108 | VPC operations |
| `dns_zone.go` | 95 | DNS zone CRUD |
| `snapshot.go` | 90 | Compute snapshot operations |

## Models Layer (`internal/models/`)

| File | Lines | Key Types |
|------|-------|-----------|
| `object_storage.go` | 172 | Bucket, access key models |
| `vm.go` | 149 | `Compute`, `CreateComputeRequest` (form tags) |
| `nfs.go` | 142 | `FileStorageVolume`, `NFSExportInfo`, `NFSAccessRule` |
| `volume.go` | 120 | `Volume` (int ID), `VolumeAttachment` |
| `lb_service.go` | 94 | LB service, VIP, certificate, virtual server models |
| `public_ip.go` | 76 | Public IP + policy rule models |
| `vpc.go` | 70 | VPC structs with tag support |
| `backup.go` | 66 | Protection plan models |
| `subnet.go` | 59 | Subnet structs |
| `dns_record.go` | 59 | `DNSRecord` |
| `security_group.go` | 48 | Security group models |
| `vpc_peering.go` | 39 | VPC peering models |
| `dns_zone.go` | 33 | `DNSZone` |
| `snapshot.go` | 17 | Compute snapshot model |

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
```

## Test Coverage

| Category | Count | Location |
|----------|-------|----------|
| Unit tests | 13 | `internal/client/*_test.go` |
| Integration tests | 10 | `internal/client/*_integration_test.go` |
| Acceptance tests | 7 | `tests/resources/*_test.go` |
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
  api_endpoint = "https://api.south.cloud.airtel.in"  # optional
  api_key      = var.api_key                          # required
  api_secret   = var.api_secret                       # required
  region       = "south"                              # optional
  organization = "my-org"                             # optional
  project_name = "my-project"                         # optional
  subnet_id    = "subnet-abc"                         # optional
}
```
