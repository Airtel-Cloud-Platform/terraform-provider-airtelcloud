---
page_title: "airtelcloud_file_storage Resource - Airtel Cloud"
subcategory: "File Storage"
description: |-
  Manages an Airtel Cloud NFS file storage volume.
---

# airtelcloud_file_storage (Resource)

Manages an Airtel Cloud NFS file storage volume.

After creating a file storage volume, use `airtelcloud_file_storage_export_path` to create NFS export paths for client access.

## Example Usage

```terraform
resource "airtelcloud_file_storage" "shared" {
  name              = "shared-nfs"
  description       = "Shared storage for application data"
  size              = "500"
  availability_zone = "S2"
}
```

## Argument Reference

### Required

- `name` (String) - The name of the file storage volume.
- `size` (String) - The size of the file storage volume in GB.

### Optional

- `description` (String) - Description of the file storage volume.
- `availability_zone` (String) - The availability zone (e.g., `S2`).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the file storage volume (same as `name`).
- `state` (String) - The current state of the volume.
- `failed_state_error` (String) - Error message in case of failed state.
- `created_at` (String) - The creation timestamp.
- `created_by` (String) - The user who created the volume.
- `uuid` (String) - The UUID of the volume.
- `provider_volume_id` (String) - The provider-specific volume identifier.

## Import

File storage volumes can be imported using the `id`:

```shell
terraform import airtelcloud_file_storage.shared <volume-id>
```
