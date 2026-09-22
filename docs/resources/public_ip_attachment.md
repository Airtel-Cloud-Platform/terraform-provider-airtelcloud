---
page_title: "airtelcloud_public_ip_attachment Resource - Airtel Cloud"
subcategory: "Security"
description: |-
  Attaches a reserved Airtel Cloud Public IP to a VM, load balancer, or baremetal server by name.
---

# airtelcloud_public_ip_attachment (Resource)

Attaches a reserved Airtel Cloud Public IP to a virtual machine, load balancer, or baremetal server.

The configuration is name-only. You name the reserved address and the workload, and the provider resolves everything else before calling the attach API:

- The availability zone is read from the public IP record.
- The target private IP (VIP) is looked up from the named VM, load balancer, or baremetal server.
- That private IP is resolved to the port ID sent to `POST .../public-ip/{uuid}/attach`.

This is step 2 of 3. Reserve the address with [`airtelcloud_public_ip`](public_ip.md) first, then allow traffic with [`airtelcloud_public_ip_policy_rule`](public_ip_policy_rule.md).

## Example Usage

### Attach to a Virtual Machine

```terraform
resource "airtelcloud_public_ip_attachment" "vm" {
  public_ip_name = airtelcloud_public_ip.web.object_name
  resource_type  = "vm"
  resource_name  = airtelcloud_vm.web.instance_name
}
```

### Attach to a Load Balancer

```terraform
resource "airtelcloud_public_ip_attachment" "lb" {
  public_ip_name = "lb-public-ip"
  resource_type  = "lb"
  resource_name  = "edge-lb"
}
```

### Attach to a Baremetal Server

```terraform
resource "airtelcloud_public_ip_attachment" "baremetal" {
  public_ip_name = "bm-public-ip"
  resource_type  = "baremetal"
  resource_name  = "db-node-1"
}
```

### Attach with Custom Timeouts

```terraform
resource "airtelcloud_public_ip_attachment" "vm" {
  public_ip_name = airtelcloud_public_ip.web.object_name
  resource_type  = "vm"
  resource_name  = airtelcloud_vm.web.instance_name

  timeouts {
    create = "15m"
    delete = "15m"
  }
}
```

### Read the Resolved Private IP

```terraform
output "nat_mapping" {
  description = "Public address and the private IP it is NATed to"
  value = {
    public_ip  = airtelcloud_public_ip_attachment.vm.public_ip
    target_vip = airtelcloud_public_ip_attachment.vm.target_vip
  }
}
```

## Argument Reference

### Required

- `public_ip_name` (String) - The `object_name` of a reserved public IP. Forces replacement if changed.
- `resource_type` (String) - The kind of workload to attach to. One of `vm`, `lb`, or `baremetal`. Forces replacement if changed.
- `resource_name` (String) - The name of the VM, load balancer, or baremetal server. Forces replacement if changed.

### Optional

- `timeouts` (Block) - Operation timeouts. See [Timeouts](#timeouts).

### Timeouts

- `create` (String) - How long to wait for the public IP to reach `attached`. Defaults to 20 minutes.
- `delete` (String) - How long to wait for the detach to complete. Defaults to 20 minutes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The UUID of the attached public IP.
- `target_vip` (String) - The private IP looked up from the named workload and used for the NAT mapping.
- `availability_zone` (String) - The availability zone, read from the public IP.
- `public_ip` (String) - The allocated public IP address.
- `status` (String) - The status after attach (`attached`).

## How the Private IP Is Resolved

You never supply the private IP. The provider selects the workload's primary address:

| `resource_type` | Source of `target_vip` |
| --- | --- |
| `vm` | The first NIC's fixed IP on the named compute instance. |
| `lb` | The first VIP on the named load balancer. |
| `baremetal` | The first `ipAddr` on the named baremetal server. |

If the named workload has no private IP, the apply fails with a "Missing Target VIP" error instead of attaching to the wrong port.

## Import

Attachments are identified by the public IP UUID:

```shell
terraform import airtelcloud_public_ip_attachment.vm <public-ip-uuid>
```

## Important Notes

- The public IP must already exist and be in `reserved` state. Reserve it with `airtelcloud_public_ip`.
- `availability_zone` and `target_vip` are never set in configuration. Both are looked up and exported as read-only attributes.
- There is no in-place update. Changing any argument destroys the attachment and creates a new one.
- Destroying this resource detaches the address and returns it to `reserved`. The reservation itself is not released.
- On multi-NIC VMs and multi-VIP load balancers the primary address is always used, matching the portal's default behavior.
