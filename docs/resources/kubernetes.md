---
page_title: "airtelcloud_kubernetes Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud Kubernetes (CKP) cluster.
---

# airtelcloud_kubernetes (Resource)

Manages an Airtel Cloud Kubernetes cluster using the Cloud Compass Cluster Manager API.

The arguments are grouped in the same order as the **Create Cluster** form in the Airtel Cloud console:

1. **Basic Information** - name and description
2. **Cluster Information** - Kubernetes version and networking
3. **Node Configuration** - VPC, availability zone, OS distribution, and the node pools
4. **Advanced Options** (per node pool) - labels, annotations, and taints

> **Note:** This resource supports Bring-Your-Own-Host (BYOH) clusters. Create, read, import, and delete are supported. Changing any argument destroys and recreates the cluster.

## Quick Start (copy and paste)

Copy the block below into `main.tf`, replace the values marked `# <-- change`, and run `terraform init && terraform apply`.

```terraform
terraform {
  required_providers {
    airtelcloud = {
      source = "Airtel-Cloud-Platform/airtelcloud"
    }
  }
}

provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south"
  organization = var.organization
  project_name = var.project_name
}

variable "airtel_api_key" {
  type      = string
  sensitive = true
}

variable "airtel_api_secret" {
  type      = string
  sensitive = true
}

variable "organization" {
  type = string
}

variable "project_name" {
  type = string
}

resource "airtelcloud_kubernetes" "cluster" {
  # 1. Basic Information
  name        = "my-cluster"            # <-- change (lowercase letters, numbers, hyphens)
  description = "My first cluster"      # <-- change (optional)

  # 2. Cluster Information
  kubernetes_version = "v1.33.7"
  networking_name    = "calico"
  networking_version = "v3.30.6"

  # 3. Node Configuration
  vpc_name          = "my-vpc"          # <-- change
  availability_zone = "S1"              # <-- change
  os_distribution   = "Ubuntu"          # Ubuntu or Rhel

  node_pools = [
    {
      flavor_type   = "Compute dense"   # <-- change (optional)
      flavor        = "ccd.xLarge"      # <-- change
      subnet_name   = "my-subnet"       # <-- change
      desired_nodes = 2                 # <-- change
    }
  ]
}

output "cluster_id" {
  value = airtelcloud_kubernetes.cluster.id
}

output "cluster_state" {
  value = airtelcloud_kubernetes.cluster.state
}
```

Provide credentials with `terraform.tfvars` or environment variables (`TF_VAR_airtel_api_key`, `TF_VAR_airtel_api_secret`, and so on). Do not commit them.

## Console Field to Terraform Argument Map

| Console section | Console field | Terraform argument | Required |
|---|---|---|---|
| Basic Information | Name | `name` | Yes |
| Basic Information | Description | `description` | No |
| Cluster Information | Kubernetes Version | `kubernetes_version` | Yes |
| Cluster Information | Networking | `networking_name` | Yes |
| Cluster Information | Networking Version | `networking_version` | Yes |
| Node Configuration | VPC | `vpc_name` | No (needed when the subnet name is not unique) |
| Node Configuration | Availability Zone | `availability_zone` | Yes |
| Node Configuration | OS Distribution | `os_distribution` | Yes |
| Node Configuration | Number of node pools | One item in `node_pools` per pool | Yes (at least 1) |
| Node pool | Flavor Type | `node_pools[].flavor_type` | No |
| Node pool | Flavor | `node_pools[].flavor` | Yes |
| Node pool | Subnet | `node_pools[].subnet_name` | Yes |
| Node pool | Desired Nodes | `node_pools[].desired_nodes` (alias: `count`) | Yes unless autoscaling is enabled |
| Node pool | Enable Auto Scaling | `node_pools[].autoscaling.enabled` | No |
| Node pool | Max nodes (auto scaling) | `node_pools[].autoscaling.max_nodes` | Yes when `autoscaling` is set |
| Advanced Options | Labels | `node_pools[].labels` | No |
| Advanced Options | Annotations | `node_pools[].annotations` | No |
| Advanced Options | Taints | `node_pools[].taints` | No |

## Example Usage

### Fixed-size node pool

```terraform
resource "airtelcloud_kubernetes" "fixed" {
  name        = "fixed-cluster"
  description = "Cluster with a fixed number of workers"

  kubernetes_version = "v1.33.7"
  networking_name    = "calico"
  networking_version = "v3.30.6"

  vpc_name          = "my-vpc"
  availability_zone = "S1"
  os_distribution   = "Ubuntu"

  node_pools = [
    {
      flavor_type   = "Compute dense"
      flavor        = "ccd.xLarge"
      subnet_name   = "my-subnet"
      desired_nodes = 3
    }
  ]
}
```

### Node pool with auto scaling

With auto scaling enabled, `desired_nodes` may be `0` and `max_nodes` is the upper limit.

```terraform
resource "airtelcloud_kubernetes" "autoscaled" {
  name        = "autoscaled-cluster"
  description = "Cluster with an auto scaled pool"

  kubernetes_version = "v1.33.7"
  networking_name    = "calico"
  networking_version = "v3.30.6"

  vpc_name          = "my-vpc"
  availability_zone = "S1"
  os_distribution   = "Ubuntu"

  node_pools = [
    {
      flavor_type   = "Compute dense"
      flavor        = "ccd.xLarge"
      subnet_name   = "my-subnet"
      desired_nodes = 0

      autoscaling = {
        enabled   = true
        max_nodes = 5
      }
    }
  ]
}
```

### Node pool with advanced options (labels, annotations, taints)

```terraform
resource "airtelcloud_kubernetes" "advanced" {
  name        = "advanced-cluster"
  description = "Cluster with labels, annotations and taints"

  kubernetes_version = "v1.33.7"
  networking_name    = "calico"
  networking_version = "v3.30.6"

  vpc_name          = "my-vpc"
  availability_zone = "S1"
  os_distribution   = "Ubuntu"

  node_pools = [
    {
      flavor_type   = "Compute dense"
      flavor        = "ccd.xLarge"
      subnet_name   = "my-subnet"
      desired_nodes = 2

      labels = [
        {
          key   = "env"
          value = "production"
        }
      ]

      annotations = [
        {
          key   = "owner"
          value = "platform-team"
        }
      ]

      taints = [
        {
          key    = "dedicated"
          value  = "database"
          effect = "NoSchedule"
        }
      ]
    }
  ]
}
```

### Multiple node pools

Add one object to `node_pools` for each pool. Each pool has its own flavor, subnet, size, and advanced options.

```terraform
resource "airtelcloud_kubernetes" "multi_pool" {
  name        = "multi-pool-cluster"
  description = "Cluster with two node pools"

  kubernetes_version = "v1.33.7"
  networking_name    = "calico"
  networking_version = "v3.30.6"

  vpc_name          = "my-vpc"
  availability_zone = "S1"
  os_distribution   = "Ubuntu"

  node_pools = [
    {
      flavor_type   = "Compute dense"
      flavor        = "ccd.xLarge"
      subnet_name   = "app-subnet"
      desired_nodes = 2

      labels = [{ key = "pool", value = "app" }]
    },
    {
      flavor_type   = "Compute dense"
      flavor        = "ccd.xLarge"
      subnet_name   = "batch-subnet"
      desired_nodes = 0

      autoscaling = {
        enabled   = true
        max_nodes = 10
      }

      labels = [{ key = "pool", value = "batch" }]
    }
  ]
}
```

## Argument Reference

### 1. Basic Information

- `name` (String, Required) - Cluster name. Lowercase letters, numbers, and hyphens only; 1 to 63 characters; cannot start or end with a hyphen. Forces new resource.
- `description` (String, Optional) - Cluster description. Forces new resource.

### 2. Cluster Information

- `kubernetes_version` (String, Required) - Kubernetes version as shown in the console (for example `v1.33.7`). Forces new resource.
- `networking_name` (String, Required) - Networking (CNI) name as shown in the console (for example `calico`). Forces new resource.
- `networking_version` (String, Required) - Networking version as shown in the console (for example `v3.30.6`). Forces new resource.

### 3. Node Configuration

- `vpc_name` (String, Optional) - VPC that contains the node pool subnets. Used to resolve `subnet_name`. Set it when the same subnet name exists in more than one VPC. Forces new resource.
- `availability_zone` (String, Required) - Availability zone code (for example `S1`). Forces new resource.
- `os_distribution` (String, Required) - `Ubuntu` or `Rhel`. Applied to every node pool. Forces new resource.
- `node_pools` (List of Object, Required) - One object per node pool. At least one is required. Forces new resource.

#### `node_pools` items

- `flavor` (String, Required) - Flavor (for example `ccd.xLarge`).
- `flavor_type` (String, Optional) - Flavor type (for example `Compute dense`).
- `subnet_name` (String, Required) - Subnet name. Resolved to the subnet ID automatically.
- `desired_nodes` (Number, Optional) - Desired Nodes. Number of machines in the pool. Must be at least `1` unless auto scaling is enabled. Defaults to `0`.
- `count` (Number, Optional, Deprecated) - Older name for `desired_nodes`. Use `desired_nodes` in new configurations.
- `availability_zone` (String, Optional) - Availability zone for this pool. Defaults to the cluster `availability_zone`.
- `autoscaling` (Object, Optional) - Enable Auto Scaling.
  - `enabled` (Boolean, Optional) - Defaults to `true` when the `autoscaling` object is present.
  - `max_nodes` (Number, Required) - Maximum number of nodes in the pool. Must be at least `1`.

### 4. Advanced Options (per node pool)

- `labels` (List of Object, Optional) - Kubernetes node labels.
  - `key` (String, Required)
  - `value` (String, Optional)
- `annotations` (List of Object, Optional) - Kubernetes node annotations.
  - `key` (String, Required)
  - `value` (String, Optional)
- `taints` (List of Object, Optional) - Kubernetes node taints.
  - `key` (String, Required)
  - `value` (String, Optional)
  - `effect` (String, Optional) - `NoSchedule`, `PreferNoSchedule`, or `NoExecute`. Defaults to `PreferNoSchedule`.

### Settings Not Shown in the Console Form

These have working defaults. You normally do not need to set them.

- `k8s_name` (String, Optional) - Kubernetes distribution. Defaults to `CKP`.
- `provider_type` (String, Optional) - Only `BringYourOwnHost` is supported. Defaults to `BringYourOwnHost`.
- `control_plane_provider` (String, Optional) - `Kamaji` or `Kubeadm`. Defaults to `Kamaji`.
- `worker_nodes` (Number, Optional) - Total worker count. Defaults to the sum of `desired_nodes` across node pools, or `1` when every pool starts at `0` with auto scaling.

### Timeouts

```terraform
timeouts {
  create = "45m"
  delete = "45m"
}
```

Both default to 45 minutes. Creation waits until the cluster state becomes `Connected`.

## Attribute Reference

In addition to the arguments above, these attributes are exported:

- `id` (String) - Cluster name.
- `state` (String) - Cluster registration state. `Connected` means the cluster is ready.
- `created_by` (String) - User that created the cluster, when the API returns it.

## Common Errors

| Message | Cause | Fix |
|---|---|---|
| `Invalid node pool count` | A pool has `desired_nodes = 0` and auto scaling is disabled. | Set `desired_nodes` to at least `1`, or add an `autoscaling` object. |
| `Subnet Resolution Error` | The subnet name was not found, or it exists in more than one VPC. | Check `subnet_name`, and set `vpc_name`. |
| Name rejected | `name` has uppercase letters, underscores, or starts or ends with `-`. | Use lowercase letters, numbers, and hyphens. |
| Terraform wants to recreate the cluster | Any argument was changed. | Expected behavior. Review the plan before applying. |

## Import

Import an existing cluster by its name:

```shell
terraform import airtelcloud_kubernetes.cluster my-cluster
```
