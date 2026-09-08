---

page_title: "airtelcloud_storage_bucket Resource - Airtel Cloud"

subcategory: "Object Storage"

description: |-
Manages an Airtel Cloud object storage bucket.

---

**# airtelcloud_storage_bucket (Resource)**

Manages an Airtel Cloud object storage bucket (S3-compatible).

**## Example Usage**

```terraform
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "tf-private-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"

  versioning   = true
  object_locking = false

  tags = {
    Environment = "production"
    Type        = "private"
  }
}
```

**## Replication Configuration**

The `replication_type`, `replication_tag`, and `availability_zone` fields work together.

You must select a `replication_tag` that is compatible with the selected `replication_type`.

**### Replication Types**

| Replication Type           | Description                                                        |
| -------------------------- | ------------------------------------------------------------------ |
| `Local`                    | Stores data locally within a single availability zone.             |
| `Replicated within region` | Replicates data between availability zones within the same region. |
| `Replicated across region` | Replicates data between availability zones in different regions.   |

**### Replication Type and Replication Tag Mapping**

| `replication_type`         | Allowed `replication_tag` | Replication         |
| -------------------------- | ------------------------- | ------------------- |
| `Local`                    | `north_N1`                | North N1            |
| `Local`                    | `north_N2`                | North N2            |
| `Local`                    | `south_S1`                | South S1            |
| `Local`                    | `south_S2`                | South S2            |
| `Replicated within region` | `north_N1_N2`             | North N1 → North N2 |
| `Replicated within region` | `north_N2_N1`             | North N2 → North N1 |
| `Replicated within region` | `south_S1_S2`             | South S1 → South S2 |
| `Replicated within region` | `south_S2_S1`             | South S2 → South S1 |
| `Replicated across region` | `north_south_N1_S1`       | North N1 → South S1 |
| `Replicated across region` | `north_south_N2_S2`       | North N2 → South S2 |
| `Replicated across region` | `south_north_S1_N1`       | South S1 → North N1 |
| `Replicated across region` | `south_north_S2_N2`       | South S2 → North N2 |

**### How to Choose Replication Configuration**

Use the following guidelines:

* For **local storage in a single availability zone**, use `Local`.
* For **replication between availability zones in the same region**, use `Replicated within region`.
* For **replication between availability zones across different regions**, use `Replicated across region`.

For example, if you want replication from `S1` to `S2` within the South region:

```terraform
replication_type  = "Replicated within region"
replication_tag   = "south_S1_S2"
availability_zone = "S1"
```

If you want replication from North `N1` to South `S1`:

```terraform
replication_type  = "Replicated across region"
replication_tag   = "north_south_N1_S1"
availability_zone = "N1"
```

> **Important:** The `replication_tag` must match the selected `replication_type`. The availability zone should correspond to the source availability zone represented by the selected replication tag.

**## Argument Reference**

**### Required**

* `name` (String) - The name of the bucket. Must be globally unique.

* `replication_type` (String) - The replication type. Valid values:

  * `Local` - Local replication only.
  * `Replicated within region` - Replicated within the same region.
  * `Replicated across region` - Replicated across regions.

  See [Replication Configuration](#replication-configuration) for the valid `replication_tag` values for each replication type.

* `replication_tag` (String) - The replication tag. The available values depend on the selected `replication_type`.

  Valid values:

  * `north_N1`
  * `north_N2`
  * `north_N1_N2`
  * `north_N2_N1`
  * `north_south_N1_S1`
  * `north_south_N2_S2`
  * `south_S1`
  * `south_S2`
  * `south_S1_S2`
  * `south_S2_S1`
  * `south_north_S1_N1`
  * `south_north_S2_N2`

  Refer to the [Replication Type and Replication Tag Mapping](#replication-type-and-replication-tag-mapping) table to select the appropriate value.

* `availability_zone` (String) - The availability zone associated with the bucket. Valid values include `N1`, `N2`, `S1`, and `S2`.

**### Optional**

* `versioning` (Boolean) - Whether versioning is enabled. Default: `false`.

* `object_locking` (Boolean) - Whether object locking is enabled. Default: `false`.

* `object_lock_validity_days` (Number) - Object lock retention in days (`config.objLockValidityDays`). Defaults to `30` when object locking is enabled.

* `tags` (Map of String) - A map of tags to assign to the bucket.

**## Attribute Reference**

In addition to all arguments above, the following attributes are exported:

* `id` (String) - The unique identifier of the bucket.

* `s3_endpoint` (String) - The S3 endpoint for the bucket.

* `public_endpoint` (String) - The public endpoint for the bucket.

* `state` (String) - The current state of the bucket.

**## Import**

Storage buckets can be imported using the `id`:

```shell
terraform import airtelcloud_storage_bucket.private_bucket <bucket-id>
```
