---
page_title: "airtelcloud_storage_bucket Resource - Airtel Cloud"
subcategory: "Object Storage"
description: |-
  Manages an Airtel Cloud object storage bucket.
---

# airtelcloud_storage_bucket (Resource)

Manages an Airtel Cloud object storage bucket (S3-compatible).

## Example Usage

```terraform
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "tf-private-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"
  versioning        = true
  object_locking    = false

  tags = {
    Environment = "production"
    Type        = "private"
  }
}
```

## Argument Reference

### Required

- `name` (String) - The name of the bucket. Must be globally unique.
- `replication_type` (String) - The replication type. Valid values:
  - `Local` - Local replication only
  - `Replicated within region` - Replicated within the same region
  - `Replicated across region` - Replicated across regions
- `replication_tag` (String) - The replication tag. Valid values: `north_N1`, `north_N2`, `north_N1_N2`, `north_N2_N1`, `north_south_N1_S1`, `north_south_N2_S2`, `south_S1`, `south_S2`, `south_S1_S2`, `south_S2_S1`, `south_north_S1_N1`, `south_north_S2_N2`.
- `availability_zone` (String) - The availability zone (e.g., `S1`, `S2`).

### Optional

- `versioning` (Boolean) - Whether versioning is enabled. Default: `false`.
- `object_locking` (Boolean) - Whether object locking is enabled. Default: `false`.
- `tags` (Map of String) - A map of tags to assign to the bucket.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the bucket.
- `s3_endpoint` (String) - The S3 endpoint for the bucket.
- `public_endpoint` (String) - The public endpoint for the bucket.
- `state` (String) - The current state of the bucket.

## Import

Storage buckets can be imported using the `id`:

```shell
terraform import airtelcloud_storage_bucket.private_bucket <bucket-id>
```
