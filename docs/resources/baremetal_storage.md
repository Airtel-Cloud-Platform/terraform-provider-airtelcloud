---
page_title: "airtelcloud_baremetal_storage Resource - Airtel Cloud"
subcategory: "Storage"
description: |-
  Manages an Airtel Cloud baremetal block storage volume.
---

# airtelcloud_baremetal_storage (Resource)

Manages an Airtel Cloud baremetal block storage volume via `POST /api/storage-plugin/v1/domain/{org}/project/{project}/block-storage/volume`.

## Example Usage

```terraform
resource "airtelcloud_baremetal_storage" "data" {
  name              = "ak-store"
  availability_zone = "S1"
  size              = 10
  description       = "Create new baremetal storage"
}
```

## Argument Reference

### Required

- `name` (String) - The name of the volume. Changing this forces a new resource.
- `availability_zone` (String) - Availability zone (e.g. `S1`). Changing this forces a new resource.
- `size` (Number) - Size in GB. Changing this forces a new resource.

### Optional

- `description` (String) - Description of the volume.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the volume (same as `name`).
- `state` (String) - The current state of the volume.
- `failed_state_error` (String) - Error message in case of failed state.
- `created_at` (String) - The creation timestamp.
- `created_by` (String) - The user who created the volume.
- `uuid` (String) - The UUID of the volume.
- `provider_volume_id` (String) - The provider-specific volume identifier.

## Import

```shell
terraform import airtelcloud_baremetal_storage.data <volume-name>
```
