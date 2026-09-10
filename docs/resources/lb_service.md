---

page_title: "airtelcloud_lb_service Resource - Airtel Cloud"

subcategory: "Load Balancer"

description: |-
Manages an Airtel Cloud load balancer service.

---

**# airtelcloud_lb_service (Resource)**

Manages an Airtel Cloud load balancer service.

A load balancer service distributes incoming network traffic across backend resources and provides a single endpoint for applications and services.

**## Example Usage**

```terraform
resource "airtelcloud_lb_service" "example" {
  name        = "my-lb-service"
  description = "Production load balancer"

  vpc_name    = "my-vpc"
  subnet_name = "my-subnet"

  ha = false

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

**## Argument Reference**

**### Required**

* `name` (String) - The name of the load balancer service.

  Example:

  ```terraform
  name = "my-lb-service"
  ```

* `vpc_name` (String) - The name of the VPC in which the load balancer service is created.

  Example:

  ```terraform
  vpc_name = "my-vpc"
  ```

* `subnet_name` (String) - The name of the subnet associated with the load balancer service.

  Example:

  ```terraform
  subnet_name = "my-subnet"
  ```

**### Optional**

* `description` (String) - A description for the load balancer service.

  Example:

  ```terraform
  description = "Production load balancer"
  ```

* `ha` (Boolean) - Determines whether High Availability (HA) is enabled for the load balancer service.

  Set to `true` to enable HA:

  ```terraform
  ha = true
  ```

  Set to `false` to disable HA:

  ```terraform
  ha = false
  ```

  Enable HA when the workload requires higher availability and resilience.

**## Network Configuration**

The load balancer service requires a VPC and subnet.

The VPC is specified using `vpc_name`:

```terraform
vpc_name = "my-vpc"
```

The subnet is specified using `subnet_name`:

```terraform
subnet_name = "my-subnet"
```

Both values are required to identify where the load balancer service should be created.

> **Important:** `vpc_name` and `subnet_name` are required for creating a load balancer service.

**## High Availability**

The `ha` argument controls whether High Availability is enabled for the load balancer service.

For workloads that require higher availability and resilience:

```terraform
ha = true
```

For workloads that do not require HA:

```terraform
ha = false
```

**## Timeouts**

The `timeouts` block controls how long Terraform waits for load balancer operations to complete.

Supported operations are:

* `create` - Maximum time Terraform waits for the load balancer service to be created.
* `delete` - Maximum time Terraform waits for the load balancer service to be deleted.

Example:

```terraform
timeouts {
  create = "15m"
  delete = "10m"
}
```

The timeout values use Terraform duration syntax.

**## Complete Example**

```terraform
resource "airtelcloud_lb_service" "production" {
  name        = "production-lb"
  description = "Production application load balancer"

  vpc_name    = "production-vpc"
  subnet_name = "production-subnet"

  ha = true

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

**## Attribute Reference**

In addition to the arguments above, the following attributes are exported:

* `id` (String) - The unique identifier of the load balancer service.

* `name` (String) - The name of the load balancer service.

* `description` (String) - The description of the load balancer service.

* `vpc_name` (String) - The name of the associated VPC.

* `subnet_name` (String) - The name of the associated subnet.

* `ha` (Boolean) - Indicates whether High Availability is enabled.

**## Import**

An existing load balancer service can be imported using its resource ID:

```shell
terraform import airtelcloud_lb_service.example <load-balancer-service-id>
```

After importing the resource, run:

```shell
terraform plan
```

to verify that the Terraform configuration matches the existing load balancer service.

**## Important Notes**

* `name`, `vpc_name`, and `subnet_name` are required.
* `vpc_name` identifies the VPC in which the load balancer service is created.
* `subnet_name` identifies the subnet associated with the load balancer service.
* Enable `ha` for workloads that require higher availability and resilience.
* Use the `timeouts` block when load balancer creation or deletion may take longer than the default timeout.
