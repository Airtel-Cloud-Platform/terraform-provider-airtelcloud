---
page_title: "airtelcloud_compute_snapshot Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud compute (VM) snapshot.
---

# airtelcloud_compute_snapshot (Resource)

Manages an immutable point-in-time snapshot of an Airtel Cloud compute instance.

Uses the v2.1 Compute API with domain/project URL paths. Snapshots are created under the compute
(`.../computes/{compute_id}/snapshot/`) and read or deleted through the project-level detail route
(`.../computes/snapshot/{uuid}/`).

## Example Usage

```terraform
# Reference the compute by ID
resource "airtelcloud_compute_snapshot" "web_snapshot" {
  compute_id    = airtelcloud_vm.web1.id
  snapshot_name = "web1-snapshot"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Or reference the compute by name (resolved to compute_id at apply)
resource "airtelcloud_compute_snapshot" "web_snapshot_by_name" {
  compute_name  = "web1"
  snapshot_name = "web1-snapshot"
}
```

## Argument Reference

### Required

- `snapshot_name` (String) - The name to give the snapshot. Forces replacement if changed.

Exactly one of the following must be specified to identify the compute instance to snapshot:

- `compute_id` (String) - The ID of the compute instance to snapshot. Forces replacement if changed. Computed when `compute_name` is used instead.
- `compute_name` (String) - The name of the compute instance to snapshot. When set, it is resolved to `compute_id` at apply time.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The UUID of the snapshot.
- `name` (String) - The snapshot name as reported by the API.
- `status` (String) - The current status of the snapshot.
- `is_active` (Boolean) - Whether the snapshot is active.
- `is_image` (Boolean) - Whether the snapshot has been converted to an image.
- `image_id` (String) - The ID of the image created from the snapshot.
- `created` (String) - The creation timestamp of the snapshot.
- `updated` (String) - The last update timestamp of the snapshot.

~> **Note:** Snapshots are immutable and cannot be updated. Any change to an argument forces a new resource.

## Import

Snapshots can be imported either by the snapshot UUID alone, or as `<compute_id>:<snapshot_uuid>`.
The composite form skips a project-wide lookup that the provider otherwise uses to recover the
owning compute instance:

```shell
terraform import airtelcloud_compute_snapshot.web_snapshot <snapshot_uuid>
terraform import airtelcloud_compute_snapshot.web_snapshot <compute_id>:<snapshot_uuid>
```
