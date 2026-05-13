---
page_title: "airtelcloud_file_storage_export_path Resource - Airtel Cloud"
subcategory: "File Storage"
description: |-
  Manages an Airtel Cloud NFS export path on a file storage volume.
---

# airtelcloud_file_storage_export_path (Resource)

Manages an NFS export path on an Airtel Cloud file storage volume. An export path defines how NFS clients can access the file storage volume.

## Example Usage

```terraform
resource "airtelcloud_file_storage" "data" {
  name              = "app-data"
  size              = "300"
  availability_zone = "S2"
}

resource "airtelcloud_file_storage_export_path" "app_export" {
  volume              = airtelcloud_file_storage.data.name
  description         = "NFSv4 export for application servers"
  protocol            = "NFSv4"
  availability_zone   = "S2"
  nfs_export_path     = "/exports/app"
  default_access_type = "ReadWrite"
  default_user_squash = "RootSquash"
}
```

## Argument Reference

### Required

- `volume` (String) - The name of the file storage volume to create the export path for.

### Optional

- `description` (String) - Description of the NFS export path.
- `protocol` (String) - The NFS protocol version. Valid values: `NFSv3`, `NFSv4`. Default: `NFSv4`.
- `availability_zone` (String) - The availability zone for the export path.
- `nfs_export_path` (String) - The NFS export directory path name.
- `default_access_type` (String) - Default access type. Valid values: `NoAccess`, `ReadOnly`, `ReadWrite`. Default: `ReadWrite`.
- `default_user_squash` (String) - Default user squash setting. Valid values: `NoSquash`, `RootIdSquash`, `RootSquash`, `AllSquash`. Default: `NoSquash`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the export path (same as `path_id`).
- `path_id` (String) - The unique path ID assigned by the system.
- `created_at` (String) - The creation timestamp.
- `created_by` (String) - The user who created the export path.
- `provider_export_path_id` (String) - The provider-specific export path identifier.

## Import

Export paths can be imported using the `id`:

```shell
terraform import airtelcloud_file_storage_export_path.app_export <path-id>
```
