---
page_title: "airtelcloud_baremetal Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Allocates and manages an Airtel Cloud baremetal server.
---

# airtelcloud_baremetal (Resource)

Allocates and manages an Airtel Cloud baremetal server.

The resource currently supports:

- server allocation via create API
- subnet names resolved to UUIDs before POST `/server`
- optional extra disks (`storage`) and backup schedule (`backup_config`)
- state refresh from list/detail APIs
- backup policy toggle updates (`policy_enabled`)
- server release on destroy

## Example Usage

```terraform
resource "airtelcloud_baremetal" "app" {
  name                    = "tft-bm-app-01"
  flavor                  = "metal-c56-m1024"
  os_image                = "ubuntu22_Aug2026"
  subnet_name             = "proxy-test-subnet"
  additional_subnet_names = ["temporal-subnet"]
  availability_zone       = "S1"
  network_name            = "copper-vpc1"

  keypair      = "Vinay"
  keypair_id   = "f97f1f91-3d1a-4a7b-acf4-98df88f28945"
  public_key   = "ssh-rsa AAAA..."
  is_reserved  = false
  tags         = ["South-AZ1"]

  storage = [
    {
      name         = "baremetel"
      size         = "10"
      path         = "/test"
      type         = "BlockStorage"
      file_system  = "xfs"
      force_format = true
    }
  ]

  backup_config = {
    schedule_type       = "weekly_full"
    start_time          = "21:00"
    incr_days           = []
    full_days           = [7]
    full_retention      = 1
    full_retention_unit = "MONTHS"
    backup_selections   = ["/test"]
  }

  # Mutable field (update API)
  policy_enabled = true

  # Destroy behavior (delete query options)
  delete_disks = true
  secure_erase = false
}
```

## Argument Reference

### Required

- `name` (String) - Baremetal server name.
- `flavor` (String) - Flavor profile to allocate.
- `os_image` (String) - OS image name to install.
- `subnet_name` (String) - Primary subnet display name. Resolved to `subnetId` before allocate.
- `availability_zone` (String) - Availability zone used by baremetal APIs (for example `N1`, `N2`, `S1`, `S2`).

### Optional

- `network_name` (String) - VPC name or UUID sent as `networkInterface.name`. Names are resolved to the VPC UUID. Required to resolve `subnet_name`.
- `additional_subnet_names` (List of String) - Extra subnet display names sent in `networkInterface.subnets` after the primary subnet.
- `storage` (List of Object) - Extra disks: `name`, `size`, `path`, `type`, `file_system`, `force_format`.
- `backup_config` (Object) - Backup schedule: `schedule_type`, `start_time`, `incr_days`, `full_days`, `full_retention`, `full_retention_unit`, `backup_selections`.
- `cloud_init` (String) - Cloud-init script for first boot.
- `is_reserved` (Boolean) - Whether to allocate from reserved capacity. Defaults to `false`.
- `system_id` (String) - Optional system ID used by backend for reservation/release flows.
- `keypair` (String) - SSH keypair name to inject.
- `keypair_id` (String) - Optional keypair UUID sent as `keypairId`.
- `public_key` (String) - Optional SSH public key sent as `publicKey`.
- `tags` (List of String) - Tags to assign at allocation. The console sends a placement tag such as `South-AZ1`; this is not the same as generic labels like `terraform`.
- `policy_enabled` (Boolean) - Mutable backup policy toggle sent to update API. When set at create, it is also included on `backupConfig.policyEnabled`.
- `delete_disks` (Boolean) - Whether to delete disks when resource is destroyed. Defaults to `false`.
- `secure_erase` (Boolean) - Whether to perform secure erase on destroy. Defaults to `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - Terraform resource ID (equal to `name`).
- `uuid` (String) - Baremetal server UUID.
- `state` (String) - Current server state.
- `hostname` (String) - Baremetal host name.
- `availability_zone` (String) - Server availability zone.
- `power` (String) - Current power state.
- `ip_addresses` (List of String) - Assigned IP addresses.
- `backend_port_id` (Number) - Backend port id from detail API (`networkInfo.portId`).

## Import

Import by baremetal server name:

```shell
terraform import airtelcloud_baremetal.app <baremetal-name>
```
