---
page_title: "airtelcloud_kubernetes Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud Kubernetes (CKP) cluster.
---

# airtelcloud_kubernetes (Resource)

Manages an Airtel Cloud Kubernetes cluster using the Cloud Compass Cluster Manager API.

v1 supports Bring-Your-Own-Host (BYOH) create, read, import, and delete. Changing configuration forces a new cluster. The create request matches the console payload (`POST /api/airtel/v1/domain/{organization}/project/{project}/cluster`).

## Example Usage

```terraform
resource "airtelcloud_kubernetes" "cluster" {
  name              = "test-kms"
  description       = "test"
  k8s_version       = "v1.33.7"
  cni_name          = "calico"
  cni_version       = "v3.30.6"
  master_nodes      = 3
  availability_zone = "S1"
  vpc_name          = var.vpc_name

  node_pools = [
    {
      name            = "md0"
      host_group      = "ccd.xLarge"
      group_name      = "Compute dense"
      subnet_name     = var.subnet_name
      os_distribution = "Ubuntu"
      count           = 1
    }
  ]
}
```

## Argument Reference

### Required

- `name` (String) - Cluster name. Sent as `cluster`. Forces new resource.
- `k8s_version` (String) - Kubernetes version (for example `v1.33.7`). Forces new resource.
- `cni_name` (String) - CNI plugin name (for example `calico`). Forces new resource.
- `cni_version` (String) - CNI plugin version (for example `v3.30.6`). Forces new resource.
- `master_nodes` (Number) - Number of control-plane nodes. Forces new resource.
- `availability_zone` (String) - Availability zone code (for example `S1`). Sent as `ce-availability-zone` and `requiredSchedulingTags.availabilityZone`. Forces new resource.
- `node_pools` (Attributes List) - BYOH worker node pools. Forces new resource.
  - `name` (String) - Pool key in the API map (for example `md0`).
  - `host_group` (String) - Host group identifier (for example `ccd.xLarge`).
  - `subnet_name` (String) - Subnet name. Resolved to the UUID the cluster API expects.
  - `os_distribution` (String) - `Ubuntu` or `Rhel`.
  - `count` (Number) - Machines in the pool.
  - `group_name` (String) - Optional display name (for example `Compute dense`).
  - `availability_zone` (String) - Optional pool AZ. Defaults to the cluster AZ.

### Optional

- `description` (String) - Cluster description. Forces new resource.
- `k8s_name` (String) - Distribution name. Defaults to `CKP`. Forces new resource.
- `worker_nodes` (Number) - Worker count. Defaults to the sum of node pool counts. Forces new resource.
- `provider_type` (String) - CKP provider. v1 supports `BringYourOwnHost` only. Defaults to `BringYourOwnHost`. Forces new resource.
- `control_plane_provider` (String) - `Kamaji` or `Kubeadm`. Defaults to `Kamaji`. Forces new resource.
- `master_host_group` (String) - Optional BYOH master host group. Forces new resource.
- `vpc_name` (String) - VPC used to resolve subnet names. Required if the same subnet name exists in more than one VPC. Forces new resource.
- `timeouts` (Block) - Create and delete timeouts. Defaults to 45 minutes.

## Attribute Reference

- `id` (String) - Cluster name.
- `state` (String) - Lifecycle state: `Unknown`, `Creating`, `Deleting`, `Ready`.
- `created_by` (String) - Creator, when returned by the API.

## Import

```shell
terraform import airtelcloud_kubernetes.cluster test-kms
```
