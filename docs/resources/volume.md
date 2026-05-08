---
page_title: "airtelcloud_volume Resource - Airtel Cloud"
subcategory: "Block Storage"
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
  name       = "data-volume"
  size       = 100
  type       = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
}
```

### Volume with VPC and Encryption

```terraform
resource "airtelcloud_volume" "encrypted" {
  name       = "encrypted-volume"
  size       = 50
  type       = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
  vpc_id            = airtelcloud_vpc.example.id
  subnet_id         = airtelcloud_subnet.example.id
  is_encrypted      = true
  enable_backup     = true
}
```

### Volume Attached to Compute Instance

```terraform
resource "airtelcloud_volume" "attached" {
  name       = "attached-volume"
  size       = 50
  type       = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
  vpc_id            = airtelcloud_vpc.example.id
  subnet_id         = airtelcloud_subnet.example.id
  compute_id        = airtelcloud_vm.example.id
  is_encrypted      = true
  bootable          = false
}
```

### Re-attach Volume to a Different Compute Instance

Changing `compute_id` detaches the volume from the old instance and attaches it to the new one without destroying and recreating the volume.

```terraform
resource "airtelcloud_volume" "data" {
  name       = "data-volume"
  size       = 100
  type       = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
  vpc_id            = airtelcloud_vpc.example.id
  subnet_id         = airtelcloud_subnet.example.id

  # Change this value to move the volume to a different instance
  compute_id = airtelcloud_vm.new_instance.id
}
```

## Argument Reference

### Required

- `volume_name` (String) - The name of the volume.
- `volume_size` (Number) - The size of the volume in GB.
- `type` (String) - The type of the volume (validated against active block storage volume types).
- `availability_zone` (String) - The availability zone where the volume is placed.
- `vpc_id` (String) - The VPC network ID for the volume. Forces replacement if changed.
- `subnet_id` (String) - The subnet ID for the volume. Forces replacement if changed.
- `is_encrypted` (Boolean) - Whether the volume is encrypted. Default: `false`. Forces replacement if changed.
- `bootable` (Boolean) - Whether the volume is bootable. Default: `false`. Forces replacement if changed.


### Optional

- `compute_id` (String) - The compute instance ID to attach the volume to. Changing this value will detach from the old instance and attach to the new one in-place (no replacement).
- `enable_backup` (Boolean) - Whether backup is enabled for the volume. Default: `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) - The unique identifier of the volume.
- `provider_volume_id` (String) - The provider-specific volume ID.
- `status` (String) - The current status of the volume (e.g., `available`, `in-use`).
- `attached_to` (String) - The ID of the compute instance the volume is attached to.
- `attachment_device` (String) - The device name when attached (e.g., `/dev/sdb`).

## Import

Volumes can be imported using the `id`:

```shell
terraform import airtelcloud_volume.data <volume-id>
```
