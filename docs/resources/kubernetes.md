---
page_title: "airtelcloud_kubernetes Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud Kubernetes (CKP) cluster.
---

# airtelcloud_kubernetes (Resource)

Manages an Airtel Cloud Kubernetes cluster using the Cloud Compass Cluster Manager API.

v1 supports Bring-Your-Own-Host (BYOH) create, read, import, and delete. Changing configuration forces a new cluster. The create request matches the console payload (`POST /api/airtel/v1/domain/{organization}/project/{project}/cluster`).

Terraform attribute names follow the Cloud Compass UI. JSON field names on the wire are unchanged (`k8sVersion`, `cniName`, `cniVersion`, `hostGroup`, `groupName`).

## Example Usage

```terraform
resource "airtelcloud_kubernetes" "cluster" {
  name                 = "test-kms"
  description          = "test"
  kubernetes_version   = "v1.33.7"
  networking_name      = "calico"
  networking_version   = "v3.30.6"
  availability_zone    = "S1"
  os_distribution      = "Ubuntu"
  vpc_name             = var.vpc_name

  node_pools = [
    {
      flavor          = "ccd.xLarge"
      flavor_type     = "Compute dense"
      subnet_name     = var.subnet_name
      count           = 0
      autoscaling = {
        enabled   = true
        max_nodes = 2
      }
      labels = [
        {
          key   = "test"
          value = "test"
        }
      ]
      annotations = [
        {
          key   = "new"
          value = "new"
        }
      ]
      taints = [
        {
          key    = "test"
          value  = "test"
          effect = "PreferNoSchedule"
        }
      ]
    }
  ]
}
```

## Argument Reference

### Basic Details

- `name` (String) - Cluster name. Sent as `cluster`. Forces new resource.
- `description` (String) - Optional cluster description. Forces new resource.

### Cluster Information

- `kubernetes_version` (String) - Kubernetes version (for example `v1.33.7`). Sent as `k8sVersion`. Forces new resource.
- `networking_name` (String) - Networking name (for example `calico`). Sent as `cniName`. Forces new resource.
- `networking_version` (String) - Networking version (for example `v3.30.6`). Sent as `cniVersion`. Forces new resource.
- `k8s_name` (String) - Distribution name. Defaults to `CKP`. Forces new resource.
- `worker_nodes` (Number) - Worker count. Defaults to the sum of node pool counts, or `1` when autoscaling from `0`. Forces new resource.

### Node Configuration

- `vpc_name` (String) - VPC used to resolve subnet names. Required if the same subnet name exists in more than one VPC. Forces new resource.
- `availability_zone` (String) - Availability zone code (for example `S1`). Sent as `ce-availability-zone` and `requiredSchedulingTags.availabilityZone`. Forces new resource.
- `os_distribution` (String) - `Ubuntu` or `Rhel`. Applied to every node pool. Forces new resource.
- `node_pools` (Attributes List) - BYOH worker node pools. Forces new resource. API keys are generated as `md0`, `md1`, and so on.
  - `flavor_type` (String) - Optional flavor type (for example `Compute dense`). Sent as `groupName`.
  - `flavor` (String) - Flavor (for example `ccd.xLarge`). Sent as `hostGroup`.
  - `subnet_name` (String) - Subnet name. Resolved to the UUID the cluster API expects.
  - `count` (Number) - Machines in the pool. Use `0` when autoscaling is enabled. Defaults to `0`.
  - `availability_zone` (String) - Optional pool AZ. Defaults to the cluster AZ.
  - `autoscaling` (Attributes) - Optional autoscaling. `enabled` defaults to `true` when the block is present. `max_nodes` is sent as `maxNodes`.

#### Advanced Options

  - `labels` / `annotations` (Attributes List) - Optional metadata. Default `op` is `MetaOpSet`.
  - `taints` (Attributes List) - Optional taints. `effect` accepts `NoSchedule`, `PreferNoSchedule`, or `NoExecute` and defaults to `PreferNoSchedule`. The provider maps these values to the Compass API enums. Default `op` is `MetaOpSet`.

### Provider Options

- `provider_type` (String) - CKP provider. v1 supports `BringYourOwnHost` only. Defaults to `BringYourOwnHost`. Forces new resource.
- `control_plane_provider` (String) - `Kamaji` or `Kubeadm`. Defaults to `Kamaji`. Forces new resource.
- `timeouts` (Block) - Create and delete timeouts. Defaults to 45 minutes. Delete uses Compass `DELETE /cluster/{name}?forceDelete=false`.

## Attribute Reference

- `id` (String) - Cluster name.
- `state` (String) - Compass cluster registration state. Creation waits while the state is `Not Registered` or `Not Connected` and completes when it becomes `Connected`.
- `created_by` (String) - Creator, when returned by the API.

## Import

```shell
terraform import airtelcloud_kubernetes.cluster test-kms
```
