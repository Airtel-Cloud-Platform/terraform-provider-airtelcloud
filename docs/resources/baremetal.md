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
- state refresh from list/detail APIs
- backup policy toggle updates (`policy_enabled`)
- server release on destroy

## Example Usage

```terraform
resource "airtelcloud_baremetal" "app" {
  name         = "tft-bm-app-01"
  flavor       = "ccd.large"
  os_image     = "Ubuntu22_04_Aug2026"
  subnet_id    = "8d5d63eb-f9a6-46ff-a5af-f7c05711391c"
  availability_zone = "N2"
  network_name = "33c9f6c7-ba79-41e0-9009-0ec94ab1cdf4"

  keypair      = "my-linux-keypair"
  keypair_id   = "f97f1f91-3d1a-4a7b-acf4-98df88f28945"
  public_key   = "ssh-rsa AAAA..."
  is_reserved  = false
  tags         = ["terraform", "app", "prod"]

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
- `subnet_id` (String) - Subnet ID for the primary network interface.
- `availability_zone` (String) - Availability zone used by baremetal APIs (for example `N1`, `N2`, `S1`, `S2`).

### Optional

- `network_name` (String) - Value sent as `networkInterface.name` (often a network UUID from UI requests).
- `cloud_init` (String) - Cloud-init script for first boot.
- `is_reserved` (Boolean) - Whether to allocate from reserved capacity. Defaults to `false`.
- `system_id` (String) - Optional system ID used by backend for reservation/release flows.
- `keypair` (String) - SSH keypair name to inject.
- `keypair_id` (String) - Optional keypair UUID sent as `keypairId`.
- `public_key` (String) - Optional SSH public key sent as `publicKey`.
- `tags` (List of String) - Tags to assign at allocation.
- `policy_enabled` (Boolean) - Mutable backup policy toggle sent to update API.
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
