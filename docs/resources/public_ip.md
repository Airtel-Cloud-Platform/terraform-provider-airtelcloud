---
page_title: "airtelcloud_public_ip Resource - Airtel Cloud"
subcategory: "Security"
description: |-
  Reserves an Airtel Cloud Public IP in a given availability zone.
---

# airtelcloud_public_ip (Resource)

Reserves an Airtel Cloud Public IP (NAT Gateway address) in a single availability zone.

This resource only reserves the address. Create calls the IPAM API with `port_id: null` and waits until the public IP reports `reserved`. Binding the address to a workload and allowing traffic through it are separate resources:

| Step | Resource | Purpose |
| --- | --- | --- |
| 1 | `airtelcloud_public_ip` | Reserve the address. |
| 2 | [`airtelcloud_public_ip_attachment`](public_ip_attachment.md) | Bind it to a VM, load balancer, or baremetal server. |
| 3 | [`airtelcloud_public_ip_policy_rule`](public_ip_policy_rule.md) | Allow or deny traffic on the attached address. |

## Example Usage

### Reserve a Public IP

```terraform
resource "airtelcloud_public_ip" "web" {
  object_name       = "web-public-ip"
  description       = "public IP for the web tier"
  availability_zone = "S1"
}
```

### Reserve with Custom Timeouts

Reservation is asynchronous and commonly takes a few minutes.

```terraform
resource "airtelcloud_public_ip" "web" {
  object_name       = "web-public-ip"
  description       = "public IP for the web tier"
  availability_zone = "S1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

### Reference the Allocated Address

```terraform
output "web_public_ip" {
  description = "The address allocated by Airtel Cloud"
  value       = airtelcloud_public_ip.web.public_ip
}
```

## Argument Reference

### Required

- `object_name` (String) - The name of the public IP reservation. Used by `airtelcloud_public_ip_attachment` and `airtelcloud_public_ip_policy_rule` to reference this address. Forces replacement if changed.
- `availability_zone` (String) - The availability zone the address is reserved in (for example `S1`, `S2`). Forces replacement if changed.

### Optional

- `description` (String) - Free-form description stored with the reservation. Forces replacement if changed.
- `timeouts` (Block) - Operation timeouts. See [Timeouts](#timeouts).

### Timeouts

- `create` (String) - How long to wait for the address to reach `reserved`. Defaults to the provider's standard create timeout.
- `delete` (String) - How long to wait for the address to be released.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The UUID of the public IP.
- `public_ip` (String) - The allocated public IP address.
- `status` (String) - The current status (`reserved` after create, `attached` once bound to a resource).
- `vip` (String) - The target private IP the address is NATed to. Empty while the address is only reserved.
- `az_name` (String) - The availability zone name reported by the API.
- `region` (String) - The region of the public IP.
- `domain` (String) - The domain (organization) that owns the public IP.
- `allocated_time` (String) - The timestamp when the address was allocated.

## Import

Public IPs can be imported using the UUID:

```shell
terraform import airtelcloud_public_ip.web <public-ip-uuid>
```

## Important Notes

- Create never takes a private IP or port ID. The reservation is unbound until you add an `airtelcloud_public_ip_attachment`.
- `object_name`, `description`, and `availability_zone` all force replacement, because the API has no update endpoint for a reservation.
- Destroying a reserved address releases it. If the address is still attached, the provider detaches it first, then releases it.
- Policy rules cannot be created against a reservation. Attach the address first, otherwise `airtelcloud_public_ip_policy_rule` fails with a "Public IP Not Attached" error.
