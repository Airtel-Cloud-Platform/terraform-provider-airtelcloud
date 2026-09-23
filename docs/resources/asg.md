---
page_title: "airtelcloud_asg Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud autoscaling group.
---

# airtelcloud_asg (Resource)

Manages an autoscaling group.

Create, read, import, and delete are supported. Changing configuration forces a new group. The group is placed in the availability zone of the chosen subnet; Terraform rejects the plan when `availability_zone` does not match that subnet.

## Example Usage

```terraform
resource "airtelcloud_asg" "example" {
  name              = "${var.resource_prefix}-asg"
  vpc_name          = "copper-vpc1"
  subnet            = "subnet1213"
  availability_zone = "S1"

  flavor_name          = "ccd.Large"
  image_name           = "Ubuntu22_04_Sep2026"
  security_group_names = ["networksg-1"]
  keypair_name         = "keypair-1"

  minimum_size  = 1
  maximum_size  = 3
  desired_count = 1

  scale_up_step_size   = 1
  scale_down_step_size = 1
  scaling_interval     = 2
  cooloff_period       = 10
  termination_policy   = "oldest"
  drain_period         = 0

  scaleup_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 60
  }]
  scaledown_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 30
  }]
}
```

### With a load balancer

Set `load_balancer_name` to attach a virtual server. The VIP is resolved from the load balancer service network, not the autoscaling-group subnet.

```terraform
resource "airtelcloud_asg" "with_lb" {
  name              = "${var.resource_prefix}-asg-lb"
  vpc_name          = "copper-vpc1"
  subnet            = "subnet1213"
  availability_zone = "S1"

  flavor_name          = "ccd.Large"
  image_name           = "Ubuntu22_04_Sep2026"
  security_group_names = ["networksg-1"]

  minimum_size  = 1
  maximum_size  = 3
  desired_count = 1

  scale_up_step_size   = 1
  scale_down_step_size = 1
  scaling_interval     = 2
  cooloff_period       = 10
  termination_policy   = "oldest"
  drain_period         = 0

  scaleup_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 60
  }]
  scaledown_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 30
  }]

  load_balancer_name    = "south-lb"
  host_name             = "${var.resource_prefix}-asg-vs"
  vip                   = "10.29.16.10"
  protocol              = "TCP"
  port                  = 1234
  routing_algorithm     = "ROUND_ROBIN"
  pool_name             = "ash-pool"
  pool_port             = 12
  max_connections       = 100
  health_check_interval = 5
  health_check_timeout  = 16
  pool_monitor_protocol = "TCP"
  enable_persistance    = false
}
```

## Argument Reference

### Required

- `name` (String) - Group name. Minimum length 3. Forces new resource.
- `vpc_name` (String) - VPC display name. Lowercase letters, digits, and hyphens only. Resolved to `vpc_id`. Forces new resource.
- `subnet` (String) - Subnet display name in that VPC. Resolved to `network_id`. Must be in the same zone as `availability_zone`. Forces new resource.
- `availability_zone` (String) - Zone code (for example `S1`). Forces new resource.
- `minimum_size` (Number) - Minimum group size. Must be at least `1` and less than `maximum_size`. Forces new resource.
- `maximum_size` (Number) - Maximum group size. Must be at least `1`. Forces new resource.
- `desired_count` (Number) - Desired instance count. Must equal `minimum_size`. Forces new resource.
- `scale_up_step_size` (Number) - Instances to add per scale-up. Between `1` and `3`. Forces new resource.
- `scale_down_step_size` (Number) - Instances to remove per scale-down. Between `1` and `3`. Forces new resource.
- `scaling_interval` (Number) - Scaling interval in minutes. Between `2` and `30`. Forces new resource.
- `cooloff_period` (Number) - Cool-off period in seconds. Between `0` and `300`. Forces new resource.
- `termination_policy` (String) - Instance termination policy. One of `oldest`, `newest`, `random`. Forces new resource.
- `drain_period` (Number) - Drain period in seconds. Between `0` and `300`. Forces new resource.
- `scaleup_rules` (List of Object) - Scale-up metric rules. At least one rule. Each `metric_type` must also appear in `scaledown_rules` with the same `aggregation_type` and a lower `target_value`. Forces new resource.
  - `metric_type` (String) - Metric name (for example `cpu`).
  - `aggregation_type` (String) - Aggregation (for example `avg`).
  - `target_value` (Number) - Threshold.
- `scaledown_rules` (List of Object) - Scale-down metric rules. Same nested fields as `scaleup_rules`. Forces new resource.

Exactly one of:

- `flavor_id` (Number) - Flavor ID. Computed when `flavor_name` is used. Forces new resource if configured.
- `flavor_name` (String) - Flavor name. Resolved to `flavor_id`. Forces new resource if configured.

Exactly one image source:

- `image_id` (Number) - Image ID. Mutually exclusive with `image_name` and `snapshot_name`. Computed when `image_name` is used. Forces new resource if configured.
- `image_name` (String) - Image name. Mutually exclusive with `image_id` and `snapshot_name`. Forces new resource if configured.
- `snapshot_name` (String) - Compute snapshot name. Mutually exclusive with `image_id` and `image_name`. Requires `keypair_id` or `keypair_name`. Forces new resource.

Exactly one security group:

- `security_group_ids` (List of Number) - One security group ID. Mutually exclusive with `security_group_names`. Computed when names are used. Forces new resource if configured.
- `security_group_names` (List of String) - One security group name. Forces new resource if configured.

### Optional

- `keypair_id` (String) - Keypair UUID. Mutually exclusive with `keypair_name`. Required when `snapshot_name` is set. Forces new resource if configured.
- `keypair_name` (String) - Keypair name. Resolved to `keypair_id`. Forces new resource if configured.
- `disk_size` (Number) - Boot volume size in GB. Minimum `20`. Defaults to `100`, or `200` for Windows images. Forces new resource.
- `labels` (List of String) - Labels to assign. Forces new resource.
- `load_balancer_name` (String) - Existing load balancer service name. When set, the remaining load-balancer arguments below are required. Forces new resource.
- `host_name` (String) - Virtual server name. Requires `load_balancer_name`. Forces new resource.
- `vip` (String) - VIP address on the load balancer service network. Requires `load_balancer_name`. Forces new resource.
- `protocol` (String) - Listener protocol. Requires `load_balancer_name`. Forces new resource.
- `port` (Number) - Listener port. Requires `load_balancer_name`. Forces new resource.
- `routing_algorithm` (String) - Pool algorithm (for example `ROUND_ROBIN`). Requires `load_balancer_name`. Forces new resource.
- `pool_name` (String) - Pool name. Requires `load_balancer_name`. Forces new resource.
- `pool_port` (Number) - Member port. Requires `load_balancer_name`. Forces new resource.
- `max_connections` (Number) - Maximum connections per member. Requires `load_balancer_name`. Forces new resource.
- `health_check_interval` (Number) - Health-check interval in seconds. Must be at least `1`. Requires `load_balancer_name`. Forces new resource.
- `health_check_timeout` (Number) - Health-check timeout in seconds. Must be at least `1`. Requires `load_balancer_name`. Forces new resource.
- `pool_monitor_protocol` (String) - Monitor protocol. Requires `load_balancer_name`. Forces new resource.
- `enable_persistance` (Boolean) - Session persistence. Defaults to `false`. The attribute name matches the API spelling. Forces new resource.
- `timeouts` (Block) - Create and delete timeouts. Defaults to 15 minutes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - Autoscaling group UUID.
- `vpc_id` (String) - Resolved VPC ID.
- `network_id` (String) - Resolved subnet ID.
- `os_type` (String) - Image OS family. Empty for snapshot-based groups.

## Availability zone

A group is placed in the zone of the subnet given in `subnet`. Terraform rejects the plan when that zone differs from `availability_zone`. If the API still reports a different zone after create, Terraform keeps the configured value in state and emits a warning.

## Import

```shell
terraform import airtelcloud_asg.example <asg-uuid>
```

Name-only values that the GET API cannot return (for example `keypair_name` and snapshot-based image fields) remain unknown after import until supplied in configuration.
