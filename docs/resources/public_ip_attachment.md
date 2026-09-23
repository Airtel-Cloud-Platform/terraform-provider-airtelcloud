---
page_title: "airtelcloud_public_ip_attachment Resource - Airtel Cloud"
subcategory: "Security"
description: |-
  Attaches a reserved Airtel Cloud Public IP to a VM, load balancer, or baremetal server by name.
---

# airtelcloud_public_ip_attachment (Resource)

Attaches a reserved Airtel Cloud Public IP to a virtual machine, load balancer, or baremetal server.

You name the reserved address and the workload, and the provider resolves the rest before calling the attach API:

- The availability zone is read from the public IP record.
- For `vm` and `baremetal`, the private IP is looked up from the named resource. Do not set `target_vip`.
- For `lb`, set `target_vip` to the load balancer VIP to attach to. Load balancers can have more than one VIP, so this must be explicit.
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
  target_vip     = "10.101.21.35"
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

- `target_vip` (String) - Required when `resource_type` is `lb`: the load balancer VIP to attach the public IP to. Do not set this for `vm` or `baremetal`. Forces replacement if changed.
- `timeouts` (Block) - Operation timeouts. See [Timeouts](#timeouts).

### Timeouts

- `create` (String) - How long to wait for the public IP to reach `attached`. Defaults to 20 minutes.
- `delete` (String) - How long to wait for the detach to complete. Defaults to 20 minutes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The UUID of the attached public IP.
- `target_vip` (String) - The private IP used for the NAT mapping. For `lb` this is the VIP you set. For `vm` and `baremetal` this is looked up from the named resource.
- `availability_zone` (String) - The availability zone, read from the public IP.
- `public_ip` (String) - The allocated public IP address.
- `status` (String) - The status after attach (`attached`).

## How the Private IP Is Resolved

| `resource_type` | Source of `target_vip` |
| --- | --- |
| `vm` | Looked up: the first NIC's fixed IP on the named compute instance. Do not set `target_vip`. |
| `lb` | Required in configuration: the VIP on the named load balancer to attach to. |
| `baremetal` | Looked up: the first `ipAddr` on the named baremetal server. Do not set `target_vip`. |

If the named workload has no matching private IP, the apply fails with a "Missing Target VIP" error instead of attaching to the wrong port.

## Import

Attachments are identified by the public IP UUID:

```shell
terraform import airtelcloud_public_ip_attachment.vm <public-ip-uuid>
```

## Important Notes

- The public IP must already exist and be in `reserved` state. Reserve it with `airtelcloud_public_ip`.
- `availability_zone` is never set in configuration. It is looked up from the public IP and exported as a read-only attribute.
- `target_vip` is required only for `lb`. For `vm` and `baremetal` it is looked up and exported after attach.
- There is no in-place update. Changing any argument destroys the attachment and creates a new one.
- Destroying this resource detaches the address and returns it to `reserved`. The reservation itself is not released.
- On multi-NIC VMs and multi-address baremetal servers the primary address is always used. On load balancers you choose the VIP with `target_vip`.
