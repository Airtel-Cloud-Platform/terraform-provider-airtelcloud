---

page_title: "airtelcloud_baremetal_storage Resource - Airtel Cloud"

subcategory: "Storage"

description: |-
Manages an Airtel Cloud baremetal block storage volume.

---

**# airtelcloud_baremetal_storage (Resource)**

Manages an Airtel Cloud baremetal block storage volume.

The resource creates a block storage volume that can be used with Airtel Cloud baremetal workloads.

**## Example Usage**

```terraform
resource "airtelcloud_baremetal_storage" "data" {
  name              = "ak-store"
  availability_zone = "S1"
  size              = 10

  description = "Create new baremetal storage"
}
```

**## Argument Reference**

**### Required**

* `name` (String) - The name of the block storage volume. The name must be unique within the project. Changing this value forces a new resource.

* `availability_zone` (String) - The availability zone where the volume is created.

  Example:

  ```terraform
  availability_zone = "S1"
  ```

  Changing the availability zone forces a new resource.

* `size` (Number) - The size of the block storage volume in GB.

  Example:

  ```terraform
  size = 10
  ```

  This creates a 10 GB block storage volume. Changing the size forces a new resource.

**### Optional**

* `description` (String) - A description for the block storage volume.

  Example:

  ```terraform
  description = "Application data volume"
  ```

**## Availability Zone**

The `availability_zone` determines the location in which the block storage volume is provisioned.

For example:

```terraform
availability_zone = "S1"
```

Use the availability zone that corresponds to the baremetal workload with which the volume will be used.

Changing the availability zone forces a new volume to be created.

**## Size**

The `size` argument specifies the capacity of the block storage volume in **GB**.

For example:

```terraform
size = 100
```

creates a 100 GB volume.

> **Important:** Changing `size` forces a new resource. Terraform will create a new volume instead of modifying the existing volume.

**## Attribute Reference**

In addition to all arguments above, the following attributes are exported:

* `id` (String) - The unique identifier of the volume. This is the same as the volume `name`.

* `state` (String) - The current state of the volume.

* `failed_state_error` (String) - The error message associated with the volume when it enters a failed state. This attribute is empty when there is no failure.

* `created_at` (String) - The timestamp when the volume was created.

* `created_by` (String) - The user who created the volume.

* `uuid` (String) - The UUID assigned to the volume.

* `provider_volume_id` (String) - The provider-specific identifier of the underlying storage volume.

**## Complete Example**

```terraform
resource "airtelcloud_baremetal_storage" "application_data" {
  name              = "application-data"
  availability_zone = "S1"
  size              = 100

  description = "100 GB storage volume for application data"
}
```

**## Import**

An existing baremetal storage volume can be imported using its volume name:

```shell
terraform import airtelcloud_baremetal_storage.data <volume-name>
```

After importing the resource, run:

```shell
terraform plan
```

to verify that the Terraform configuration matches the existing volume.

**## Important Notes**

* The volume is created in the specified `availability_zone`.
* `name`, `availability_zone`, and `size` are immutable. Changing any of these values forces Terraform to create a new volume.
* `size` is specified in GB.
* `description` can be used to provide additional information about the volume.
* The `state` and `failed_state_error` attributes can be used to monitor the current status of the volume.
