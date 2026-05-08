---
page_title: "airtelcloud_vpc Resource - Airtel Cloud"
subcategory: "Networking"
description: |-
  Manages an Airtel Cloud Virtual Private Cloud (VPC).
---

# airtelcloud_vpc (Resource)

Manages an Airtel Cloud Virtual Private Cloud (VPC).

## Example Usage

```terraform
resource "airtelcloud_vpc" "main" {
  name                 = "production-vpc"
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Environment = "production"
    Team        = "infrastructure"
  }
}
```

## Argument Reference

### Required

- `name` (String) - The name of the VPC.
- `cidr_block` (String) - The CIDR block for the VPC (e.g., `10.0.0.0/16`).

### Optional

- `enable_dns_hostnames` (Boolean) - Whether DNS hostnames are enabled. Default: `false`.
- `enable_dns_support` (Boolean) - Whether DNS support is enabled. Default: `true`.
- `tags` (Map of String) - A map of tags to assign to the VPC.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the VPC.
- `state` (String) - The current state of the VPC.
- `is_default` (Boolean) - Whether this is the default VPC.

## Import

VPCs can be imported using the `id`:

```shell
terraform import airtelcloud_vpc.main <vpc-id>
```
