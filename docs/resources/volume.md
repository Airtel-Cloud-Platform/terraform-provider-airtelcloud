---
page_title: "airtelcloud_volume Resource - Airtel Cloud"
subcategory: "Storage"
description: |-
  Manages an Airtel Cloud block storage volume.
---

# airtelcloud_volume (Resource)

Manages an Airtel Cloud block storage volume.

Uses the v2.1 Volume API with domain/project URL paths. Volumes are created using the `create-and-attach` endpoint.

## Example Usage

### Basic Volume

```terraform
resource "airtelcloud_volume" "data" {
  name              = "data-volume"
  size              = 100
  availability_zone = "S1"
  vpc_name          = "copper-vpc1"
  subnet_name       = "asg-vip-demo"
  type              = "Standard Performance"
}
```

### Volume with Encryption and Labels

```terraform
resource "airtelcloud_volume" "encrypted" {
  name              = "encrypted-volume"
  size              = 50
  availability_zone = "S1"
  vpc_name          = airtelcloud_vpc.example.name
  subnet_name       = airtelcloud_subnet.example.name
  type              = "Standard Performance"
  is_encrypted      = true
  enable_backup     = true
  labels            = ["app"]
}
```

### Volume Attached to Compute Instance

```terraform
resource "airtelcloud_volume" "attached" {
  name              = "attached-volume"
  size              = 50
  availability_zone = "S1"
  vpc_name          = airtelcloud_vpc.example.name
  subnet_name       = airtelcloud_subnet.example.name
  type              = "Standard Performance"
  compute_name      = airtelcloud_vm.example.instance_name
  is_encrypted      = true
  bootable          = false
}
```

### Re-attach Volume to a Different Compute Instance

Changing `compute_name` detaches the volume from the old instance and attaches it to the new one without destroying and recreating the volume.

```terraform
resource "airtelcloud_volume" "data" {
  name              = "data-volume"
  size              = 100
  availability_zone = "S1"
  vpc_name          = airtelcloud_vpc.example.name
  subnet_name       = airtelcloud_subnet.example.name
  type              = "Standard Performance"

  compute_name = airtelcloud_vm.new_instance.instance_name
}
```

## Argument Reference

### Required

- `name` (String) - The name of the volume.
- `size` (Number) - The size of the volume in GB. Must be between 50 and 4095.
- `availability_zone` (String) - Availability zone from the UI (for example `S1`).
- `type` (String) - Volume tier. API type name or UI label such as `Standard Performance`. Resolved via `GET .../volume_types?group=BLOCK_STORAGE` with the `subnet-id` header.
- `vpc_name` (String) - VPC name. Resolved to an ID for the API. Forces replacement if changed.
- `subnet_name` (String) - Subnet name. Resolved to an ID for the API and the `subnet-id` header. Forces replacement if changed.

### Optional

- `compute_name` (String) - VM instance name to attach to. Resolved internally to a compute ID. Changing this value detaches and re-attaches in place. Omit to leave the volume unattached.
- `labels` (List of String) - Optional labels. Set after create via `PATCH .../volumes/labels` using the numeric volume id (UUID is rejected by the API).
- `is_encrypted` (Boolean) - Whether the volume is encrypted. Default: `false`. Forces replacement if changed.
- `bootable` (Boolean) - Whether the volume is bootable. Default: `false`. Forces replacement if changed.
- `enable_backup` (Boolean) - Whether backup is enabled for the volume. Default: `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) - The unique identifier of the volume.
- `uuid` (String) - The UUID of the volume.
- `provider_volume_id` (String) - The provider-specific volume ID.
- `status` (String) - The current status of the volume (e.g., `available`, `in-use`).
- `attached_to` (String) - The ID of the compute instance the volume is attached to.
- `attachment_device` (String) - The device name when attached (e.g., `/dev/sdb`).

## Import

Volumes can be imported using the `uuid`:

```shell
terraform import airtelcloud_volume.data <volume-uuid>
```
