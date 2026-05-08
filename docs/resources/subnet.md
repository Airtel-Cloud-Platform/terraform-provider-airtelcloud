---
page_title: "airtelcloud_subnet Resource - Airtel Cloud"
subcategory: "Networking"
description: |-
  Manages an Airtel Cloud VPC subnet.
---

# airtelcloud_subnet (Resource)

Manages an Airtel Cloud VPC subnet.

Uses the v2.1 API with domain/project URL paths. The provider's `organization` and `project_name` settings are embedded in the API URL automatically.

## Example Usage

```terraform
resource "airtelcloud_subnet" "private" {
  network_id         = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  name               = "app-private-subnet"
  ipv4_address_space = "10.1.14.0/24"
  subnet_sub_role    = "Private"
  availability_zone  = "S2"
  labels             = ["Production"]

  timeouts {
    create = "5m"
    delete = "4m"
  }
}
```

## Argument Reference

### Required

- `network_id` (String) - The ID of the network (VPC) to create the subnet in.
- `name` (String) - The name of the subnet.
- `ipv4_address_space` (String) - The IPv4 CIDR block for the subnet (e.g., `10.1.14.0/24`).

### Optional

- `description` (String) - A description of the subnet.
- `availability_zone` (String) - The availability zone for the subnet (e.g., `S2`).
- `subnet_sub_role` (String) - The sub-role of the subnet. Valid values: `Private`, `VIP`.
- `region` (String) - The region of the subnet (e.g., `south`).
- `labels` (List of String) - List of labels for the subnet.

### Timeouts

The `timeouts` block supports:

- `create` (String) - Time to wait for subnet creation. Default: `10m`.
- `delete` (String) - Time to wait for subnet deletion. Default: `10m`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the subnet.
- `state` (String) - The current state of the subnet.
- `ipv6_address_space` (String) - The IPv6 address space.
- `created_by` (String) - The user who created the subnet.
- `create_time` (String) - The creation time of the subnet.

## Import

Subnets can be imported using the `id`:

```shell
terraform import airtelcloud_subnet.private <subnet-id>
```
