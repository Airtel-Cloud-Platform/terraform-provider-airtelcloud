---
page_title: "Airtel Cloud Terraform Provider - User Guide"
subcategory: ""
description: |-
  Complete user guide for managing Airtel Cloud infrastructure with Terraform.
  Covers all 22 resources with examples, argument references, and import instructions.
---

# Airtel Cloud Terraform Provider - User Guide

This guide is a complete reference for managing Airtel Cloud infrastructure using the Terraform provider. It covers all resources demonstrated in the [complete example](https://github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/tree/main/examples/complete), with argument references, working HCL examples, and import instructions.

-> **First time?** See the [Getting Started Guide](getting-started) for installation, build-from-source, and initial setup.

## Prerequisites

| Requirement | Details |
|---|---|
| Terraform | >= 1.0 |
| Airtel Cloud Account | With API key and secret |
| Region | `south` (availability zones: `S1`, `S2`) |

## Provider Configuration

```terraform
terraform {
  required_providers {
    airtelcloud = {
      source  = "terraform-providers/airtelcloud"
      version = "~> 0.2"
    }
  }
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

provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south"
  organization = var.organization
  project_name = var.project_name
}
```

### Provider Argument Reference

| Argument | Type | Required | Default | Description |
|---|---|---|---|---|
| `api_key` | String | Yes | — | API key for HMAC-SHA256 authentication. Sensitive. |
| `api_secret` | String | Yes | — | API secret for HMAC-SHA256 authentication. Sensitive. |
| `api_endpoint` | String | No | `https://api.south.cloud.airtel.in` | Airtel Cloud API endpoint URL. |
| `region` | String | No | `south` | Airtel Cloud region. |
| `organization` | String | No | — | Organization name or domain. |
| `project_name` | String | No | — | Project name for API calls. |
| `subnet_id` | String | No | — | Default subnet ID for volume API provider lookup. |

Credentials can also be provided via environment variables `AIRTEL_API_KEY` and `AIRTEL_API_SECRET`, or in a `terraform.tfvars` file.

---

## Networking & Security

### airtelcloud_security_group

Manages a security group (firewall group). Security groups are immutable after creation -- any change requires replacement.

#### Example Usage

```terraform
resource "airtelcloud_security_group" "web" {
  security_group_name = "sg-http-servers"
  availability_zone   = "S2"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `security_group_name` | String | Yes | Name of the security group. Changing this forces a new resource. |
| `availability_zone` | String | No | Availability zone (e.g., `S1`, `S2`). Changing this forces a new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | Int64 | Numeric ID of the security group. |
| `uuid` | String | UUID of the security group. |
| `status` | String | Current status. |
| `az_name` | String | Availability zone name. |
| `az_region` | String | Availability zone region. |
| `created_at` | String | Creation timestamp. |
| `updated_at` | String | Last update timestamp. |

#### Import

```bash
terraform import airtelcloud_security_group.web <numeric-id>
```

---

### airtelcloud_security_group_rule

Manages a security group rule. All attributes force replacement on change -- rules cannot be updated in-place.

#### Example Usage

```terraform
resource "airtelcloud_security_group_rule" "ssh" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "22"
  port_range_max    = "22"
  remote_ip_prefix  = "10.0.0.0/8"
  ethertype         = "IPv4"
  description       = "Allow SSH from internal network"
}

resource "airtelcloud_security_group_rule" "http" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "8080"
  port_range_max    = "8080"
  remote_ip_prefix  = "0.0.0.0/0"
  ethertype         = "IPv4"
  description       = "Allow HTTP"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `security_group_id` | Int64 | Yes | ID of the parent security group. Forces new resource. |
| `direction` | String | Yes | Traffic direction: `ingress` or `egress`. Forces new resource. |
| `protocol` | String | No | Protocol (e.g., `tcp`, `udp`, `icmp`). Forces new resource. |
| `port_range_min` | String | No | Minimum port number. Forces new resource. |
| `port_range_max` | String | No | Maximum port number. Forces new resource. |
| `remote_ip_prefix` | String | No | Remote CIDR (e.g., `0.0.0.0/0`). Forces new resource. |
| `remote_group_id` | String | No | Remote security group ID. Forces new resource. |
| `ethertype` | String | No | Ethertype: `IPv4` or `IPv6`. Forces new resource. |
| `description` | String | No | Rule description. Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | Int64 | Numeric ID of the rule. |
| `uuid` | String | UUID of the rule. |
| `security_group_uuid` | String | UUID of the parent security group. |
| `status` | String | Current status. |
| `provider_security_group_rule_id` | String | Provider-side rule ID. |

#### Import

```bash
terraform import airtelcloud_security_group_rule.ssh <security_group_id>/<rule_id>
```

---

### airtelcloud_subnet

Manages a VPC subnet with configurable timeouts.

#### Example Usage

```terraform
resource "airtelcloud_subnet" "private" {
  network_id         = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  name               = "private-subnet"
  ipv4_address_space = "10.1.19.0/24"
  subnet_sub_role    = "Private"
  availability_zone  = "S2"
  labels             = ["Provider Test"]

  timeouts {
    create = "5m"
    delete = "4m"
  }
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `network_id` | String | Yes | VPC (network) ID. Forces new resource. |
| `name` | String | Yes | Subnet name. |
| `ipv4_address_space` | String | Yes | IPv4 CIDR block (e.g., `10.1.19.0/24`). Forces new resource. |
| `description` | String | No | Subnet description. |
| `availability_zone` | String | No | Availability zone. |
| `subnet_sub_role` | String | No | Subnet role (e.g., `Private`, `VIP`). |
| `region` | String | No | Region name. |
| `labels` | List of String | No | Labels for the subnet. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Subnet ID. |
| `state` | String | Current state. |
| `ipv6_address_space` | String | IPv6 CIDR block (if assigned). |
| `created_by` | String | Creator. |
| `create_time` | String | Creation timestamp. |

#### Timeouts

| Operation | Default |
|---|---|
| `create` | 10 minutes |
| `delete` | 10 minutes |

#### Import

```bash
terraform import airtelcloud_subnet.private <network_id>/<subnet_id>
```

---

### airtelcloud_public_ip

Allocates a public IP via NAT against a VM or Load Balancer private IP. Public IPs are availability-zone-specific and immutable -- any change requires replacement.

#### Example Usage

```terraform
resource "airtelcloud_public_ip" "web1_public" {
  object_name       = "web1-public-ip"
  vip               = airtelcloud_vm.web1.private_ip
  availability_zone = "S1"

  timeouts {
    create = "10m"
    delete = "5m"
  }
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `object_name` | String | Yes | Name for the public IP allocation. Forces new resource. |
| `vip` | String | Yes | Target private IP (VM or LB) to NAT against. Forces new resource. |
| `availability_zone` | String | Yes | Availability zone (e.g., `S1`, `S2`). Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | UUID of the public IP. |
| `public_ip` | String | The allocated public IP address. |
| `domain` | String | Domain of the public IP. |
| `status` | String | Current status. |
| `allocated_time` | String | Allocation timestamp. |
| `az_name` | String | Availability zone name. |
| `region` | String | Region. |

#### Timeouts

| Operation | Default |
|---|---|
| `create` | 10 minutes |
| `delete` | 10 minutes |

#### Import

```bash
terraform import airtelcloud_public_ip.web1_public <uuid>
```

---

### airtelcloud_public_ip_policy_rule

Manages a NAT policy rule on a public IP to control allowed or denied traffic. Service names (e.g., `HTTP`, `HTTPS`) are automatically resolved to their UUIDs by the provider.

#### Example Usage

```terraform
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_id      = airtelcloud_public_ip.web1_public.id
  display_name      = "allow-http-https"
  source            = "any"
  services          = ["HTTP", "HTTPS"]
  action            = "accept"
  target_vip        = airtelcloud_public_ip.web1_public.vip
  public_ip         = airtelcloud_public_ip.web1_public.public_ip
  availability_zone = "S1"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `public_ip_id` | String | Yes | UUID of the parent public IP. Forces new resource. |
| `display_name` | String | Yes | Display name for the rule. Forces new resource. |
| `source` | String | Yes | Source IP or `any`. Forces new resource. |
| `services` | List of String | Yes | Service names (e.g., `HTTP`, `HTTPS`, `SSH`). |
| `action` | String | Yes | Action: `accept` or `deny`. Forces new resource. |
| `target_vip` | String | Yes | Target private IP (VIP). Forces new resource. |
| `public_ip` | String | Yes | Public IP address. Forces new resource. |
| `availability_zone` | String | Yes | Availability zone. Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | UUID of the policy rule. |
| `state` | String | Current state. |

#### Import

```bash
terraform import airtelcloud_public_ip_policy_rule.web_traffic <public_ip_id>/<target_vip>/<public_ip>/<rule_id>
```

---

## Compute

### airtelcloud_vm

Manages a virtual machine (compute instance). Supports name-based or ID-based lookup for flavors, images, VPCs, subnets, security groups, and keypairs.

#### Example Usage

```terraform
resource "airtelcloud_vm" "web1" {
  instance_name     = "web-server-1"
  os_type           = "linux"
  flavor_name       = "ccd.Large"
  image_name        = "CentOS_Stream9_Mar2026"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 100
  availability_zone = "S1"
  description       = "Web server behind load balancer"
}
```

-> **Note:** Provide exactly one of `flavor_id` / `flavor_name`, `image_id` / `image_name`, `vpc_id` / `vpc_name`, and `subnet_id` / `subnet_name`. The provider resolves names to IDs automatically.

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `instance_name` | String | Yes | Name of the VM. |
| `os_type` | String | Yes | Operating system type (`linux` or `windows`). Forces new resource. |
| `flavor_id` | String | No | Flavor ID. Exactly one of `flavor_id` or `flavor_name` is required. |
| `flavor_name` | String | No | Flavor name (e.g., `ccd.Large`). Resolved to ID by the provider. |
| `image_id` | String | No | Image ID. Exactly one of `image_id` or `image_name` is required. Forces new resource. |
| `image_name` | String | No | Image name. Resolved to ID by the provider. Forces new resource. |
| `vpc_id` | String | No | VPC ID. Exactly one of `vpc_id` or `vpc_name` is required. Forces new resource. |
| `vpc_name` | String | No | VPC name. Resolved to ID. |
| `subnet_id` | String | No | Subnet ID. Exactly one of `subnet_id` or `subnet_name` is required. Forces new resource. |
| `subnet_name` | String | No | Subnet name. Resolved to ID. |
| `security_group_id` | String | No | Security group ID. Mutually exclusive with `security_group_name`. |
| `security_group_name` | String | No | Security group name. |
| `keypair_id` | String | No | Keypair ID. Mutually exclusive with `keypair_name`. Forces new resource. |
| `keypair_name` | String | No | Keypair name. Forces new resource. |
| `user_data` | String | No | Cloud-init user data script. Forces new resource. |
| `availability_zone` | String | No | Availability zone (e.g., `S1`). Forces new resource. |
| `disk_size` | Int64 | No | Root disk size in GB. Default: `20`. Forces new resource. |
| `boot_from_volume` | Bool | No | Boot from volume. Default: `true`. Forces new resource. |
| `volume_type_id` | String | No | Volume type ID for the boot disk. |
| `description` | String | No | Instance description. |
| `enable_backup` | Bool | No | Enable backup. Default: `false`. |
| `protection_plan` | String | No | Protection plan name (if backup enabled). |
| `start_date` | String | No | Backup start date. |
| `start_time` | String | No | Backup start time. |
| `vm_count` | Int64 | No | Number of VMs to create (1-10). Default: `1`. |
| `tags` | Map of String | No | Key-value tags. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | UUID of the VM. |
| `provider_instance_id` | String | Provider-side instance ID. |
| `public_ip` | String | Public IP address (if assigned). |
| `private_ip` | String | Private IP address. |
| `status` | String | Current status. |
| `region` | String | Region. |

#### Import

```bash
terraform import airtelcloud_vm.web1 <compute-id>
```

---

## Storage

### airtelcloud_storage_bucket

Manages an S3-compatible object storage bucket.

#### Example Usage

```terraform
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "my-private-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"
  versioning        = true
  object_locking    = false
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Bucket name. Forces new resource. |
| `replication_type` | String | Yes | One of: `Local`, `Replicated within region`, `Replicated across region`. Forces new resource. |
| `replication_tag` | String | Yes | Replication tag (e.g., `south_S1`, `south_S2`, `south_S1_S2`). Forces new resource. |
| `availability_zone` | String | Yes | Availability zone. Forces new resource. |
| `versioning` | Bool | No | Enable versioning. Default: `false`. |
| `object_locking` | Bool | No | Enable object locking. Default: `false`. |
| `tags` | Map of String | No | Key-value tags. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Bucket ID. |
| `s3_endpoint` | String | S3-compatible endpoint URL. |
| `public_endpoint` | String | Public endpoint URL. |
| `state` | String | Current state. |

#### Import

```bash
terraform import airtelcloud_storage_bucket.private_bucket <bucket-name>
```

---

### airtelcloud_volume

Manages a block storage volume. Volumes can be attached to a compute instance via `compute_id`, and the attachment can be changed in-place without replacement.

#### Example Usage

```terraform
resource "airtelcloud_volume" "data_volume" {
  name              = "storage-volume"
  size              = 50
  type              = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
  compute_id        = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  is_encrypted      = false
  bootable          = false
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Volume name. |
| `size` | Int64 | Yes | Volume size in GB. |
| `type` | String | No | Volume type identifier. |
| `availability_zone` | String | No | Availability zone. |
| `vpc_id` | String | No | VPC ID. Forces new resource. |
| `subnet_id` | String | No | Subnet ID. Forces new resource. |
| `compute_id` | String | No | Compute instance ID to attach to. Can be changed in-place (triggers detach/re-attach). |
| `is_encrypted` | Bool | No | Enable encryption. Default: `false`. Forces new resource. |
| `bootable` | Bool | No | Mark as bootable. Default: `false`. Forces new resource. |
| `enable_backup` | Bool | No | Enable backup. Default: `false`. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | Int64 | Numeric volume ID. |
| `uuid` | String | Volume UUID. |
| `provider_volume_id` | String | Provider-side volume ID. |
| `status` | String | Current status. |
| `attached_to` | String | Compute instance the volume is attached to. |
| `attachment_device` | String | Device path of the attachment. |

#### Import

```bash
terraform import airtelcloud_volume.data_volume <volume-id>
```

-> **Note:** The volume ID for import is the numeric `id`, not the `uuid`.

---

### airtelcloud_file_storage

Manages an NFS file storage volume.

#### Example Usage

```terraform
resource "airtelcloud_file_storage" "basic" {
  name              = "basic-nfs"
  size              = "300"
  availability_zone = "S2"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Volume name. Forces new resource. |
| `size` | String | Yes | Size in GB (as a string, e.g., `"300"`). |
| `description` | String | No | Volume description. |
| `availability_zone` | String | No | Availability zone. Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Volume ID. |
| `uuid` | String | Volume UUID. |
| `state` | String | Current state. |
| `failed_state_error` | String | Error message if in failed state. |
| `created_at` | String | Creation timestamp. |
| `created_by` | String | Creator. |
| `provider_volume_id` | String | Provider-side volume ID. |

#### Import

```bash
terraform import airtelcloud_file_storage.basic <id>
```

---

### airtelcloud_file_storage_export_path

Manages an NFS export path (mount point) for a file storage volume.

#### Example Usage

```terraform
resource "airtelcloud_file_storage_export_path" "nfsv4_export" {
  volume              = airtelcloud_file_storage.basic.name
  description         = "NFSv4 export for application servers"
  protocol            = "NFSv4"
  availability_zone   = "S2"
  default_access_type = "ReadWrite"
  default_user_squash = "RootSquash"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `volume` | String | Yes | File storage volume name. Forces new resource. |
| `description` | String | No | Export path description. |
| `protocol` | String | No | NFS protocol: `NFSv3` or `NFSv4`. Default: `NFSv4`. |
| `availability_zone` | String | No | Availability zone. Forces new resource. |
| `nfs_export_path` | String | No | Custom NFS export path. |
| `default_access_type` | String | No | Access type: `NoAccess`, `ReadOnly`, or `ReadWrite`. Default: `ReadWrite`. |
| `default_user_squash` | String | No | Squash mode: `NoSquash`, `RootIdSquash`, `RootSquash`, or `AllSquash`. Default: `NoSquash`. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Export path ID. |
| `path_id` | String | Path identifier. |
| `created_at` | String | Creation timestamp. |
| `created_by` | String | Creator. |
| `provider_export_path_id` | String | Provider-side export path ID. |

#### Import

```bash
terraform import airtelcloud_file_storage_export_path.nfsv4_export <path-id>
```

---

## DNS

### airtelcloud_dns_zone

Manages a DNS zone (DNSaaS). Zone names and types are immutable after creation.

#### Example Usage

```terraform
resource "airtelcloud_dns_zone" "main" {
  zone_name   = "example-domain.com"
  zone_type   = "forward"
  description = "Primary DNS zone"
}

resource "airtelcloud_dns_zone" "internal" {
  zone_name   = "internal.example-domain.com"
  zone_type   = "forward"
  description = "Internal services DNS zone"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `zone_name` | String | Yes | DNS zone name. Forces new resource. |
| `zone_type` | String | No | Zone type. Default: `forward`. Forces new resource. |
| `description` | String | No | Zone description. Updatable. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Zone UUID. |
| `dns_zone_template` | String | Zone template. |
| `org_name` | String | Organization name. |
| `org_id` | String | Organization ID. |
| `created_by` | String | Creator. |
| `created_at` | Float64 | Creation timestamp (Unix). |
| `updated_at` | Float64 | Last update timestamp (Unix). |

#### Import

```bash
terraform import airtelcloud_dns_zone.main <zone-uuid>
```

---

### airtelcloud_dns_record

Manages a DNS record within a zone. Supports A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, and PTR record types.

#### Example Usage

```terraform
# A record
resource "airtelcloud_dns_record" "apex" {
  zone_id     = airtelcloud_dns_zone.main.id
  owner       = "apex"
  record_type = "A"
  data        = "192.168.1.10"
  ttl         = 300
  description = "Apex domain A record"
}

# TXT record (SPF)
resource "airtelcloud_dns_record" "spf" {
  zone_id     = airtelcloud_dns_zone.main.id
  owner       = "spf"
  record_type = "TXT"
  data        = "v=spf1 include:_spf.example.com ~all"
  ttl         = 3600
  description = "SPF record for email authentication"
}

# AAAA record (IPv6)
resource "airtelcloud_dns_record" "ipv6" {
  zone_id     = airtelcloud_dns_zone.main.id
  owner       = "www"
  record_type = "AAAA"
  data        = "2001:db8::1"
  ttl         = 300
  description = "IPv6 address for www"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `zone_id` | String | Yes | Parent zone UUID. Forces new resource. |
| `record_type` | String | Yes | Record type (e.g., `A`, `AAAA`, `CNAME`, `MX`, `TXT`). Forces new resource. |
| `owner` | String | No | Record owner / hostname. |
| `data` | String | No | Record data (IP address, CNAME target, etc.). |
| `ttl` | Int64 | No | Time to live in seconds. Default: `300`. |
| `description` | String | No | Record description. |
| `preference` | Int64 | No | MX record preference (priority). Only used for MX records. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Record UUID. |
| `zone_name` | String | Parent zone name. |
| `org_name` | String | Organization name. |
| `org_id` | String | Organization ID. |
| `created_by` | String | Creator. |
| `created_at` | Float64 | Creation timestamp (Unix). |
| `updated_at` | Float64 | Last update timestamp (Unix). |

#### Import

```bash
terraform import airtelcloud_dns_record.apex <zone_id>/<record_id>
```

---

## Load Balancing

Load balancer resources form a dependency chain: **LB Service** -> **LB VIP** -> **LB Virtual Server**.

### airtelcloud_lb_service

Manages a load balancer service. All user-settable attributes are immutable -- any change requires replacement.

#### Example Usage

```terraform
resource "airtelcloud_lb_service" "web_lb" {
  name        = "web-lb"
  description = "Load balancer for web tier"
  network_id  = "862f1f29-f4d2-4bcb-afdd-3bd96c0ef66e"
  vpc_id      = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  vpc_name    = "my-vpc"
  ha          = false

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `network_id` | String | Yes | Network (subnet) ID. Forces new resource. |
| `vpc_id` | String | Yes | VPC ID. Forces new resource. |
| `vpc_name` | String | Yes | VPC name. Forces new resource. |
| `name` | String | No | Service name. Forces new resource. |
| `description` | String | No | Service description. Forces new resource. |
| `ha` | Bool | No | Enable high availability. Default: `false`. Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | LB service ID. |
| `flavor_id` | Int64 | Auto-resolved flavor ID. |
| `status` | String | Provisioning status. |
| `operating_status` | String | Operating status. |
| `az_name` | String | Availability zone. |
| `created` | String | Creation timestamp. |

#### Timeouts

| Operation | Default |
|---|---|
| `create` | 10 minutes |
| `delete` | 10 minutes |

#### Import

```bash
terraform import airtelcloud_lb_service.web_lb <lb-service-id>
```

---

### airtelcloud_lb_vip

Manages a VIP (Virtual IP) port for a load balancer service. Only requires the parent LB service ID -- the VIP is allocated automatically.

#### Example Usage

```terraform
resource "airtelcloud_lb_vip" "web_vip" {
  lb_service_id = airtelcloud_lb_service.web_lb.id
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `lb_service_id` | String | Yes | Parent LB service ID. Forces new resource. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | VIP port ID. |
| `name` | String | VIP name. |
| `status` | String | Current status. |
| `fixed_ips` | String | Assigned fixed IPs. |
| `public_ip` | String | Public IP (if assigned). |
| `provider_port_id` | String | Provider-side port ID. |

#### Import

```bash
terraform import airtelcloud_lb_vip.web_vip <lb_service_id>/<vip_id>
```

---

### airtelcloud_lb_virtual_server

Manages a virtual server on a load balancer service. Defines L4/L7 load balancing rules with backend nodes. Supports in-place updates for routing algorithm, persistence settings, and forwarding options.

#### Example Usage

```terraform
resource "airtelcloud_lb_virtual_server" "http" {
  lb_service_id     = airtelcloud_lb_service.web_lb.id
  name              = "http-vs"
  vip_port_id       = tonumber(airtelcloud_lb_vip.web_vip.id)
  protocol          = "HTTP"
  port              = 80
  routing_algorithm = "ROUND_ROBIN"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  interval          = 30

  nodes = [
    {
      compute_id = 1
      compute_ip = airtelcloud_vm.web1.private_ip
      port       = 8080
      weight     = 50
    },
    {
      compute_id = 2
      compute_ip = airtelcloud_vm.web2.private_ip
      port       = 8080
      weight     = 50
    },
  ]

  persistence_enabled = true
  persistence_type    = "source_ip"
  x_forwarded_for     = true

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
```

-> **Note:** The `vip_port_id` expects an Int64. Since `airtelcloud_lb_vip` exports `id` as a String, use `tonumber()` to convert it.

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `lb_service_id` | String | Yes | Parent LB service ID. Forces new resource. |
| `vip_port_id` | Int64 | Yes | VIP port ID (from `airtelcloud_lb_vip`). Forces new resource. |
| `protocol` | String | Yes | Protocol (e.g., `HTTP`, `HTTPS`, `TCP`). Forces new resource. |
| `port` | Int64 | Yes | Listening port. Forces new resource. |
| `routing_algorithm` | String | Yes | Algorithm (e.g., `ROUND_ROBIN`, `LEAST_CONNECTION`). Updatable. |
| `vpc_id` | String | Yes | VPC ID. Forces new resource. |
| `interval` | Int64 | Yes | Health check interval in seconds. |
| `nodes` | List of Object | Yes | Backend nodes (see below). |
| `name` | String | No | Virtual server name. |
| `persistence_enabled` | Bool | No | Enable session persistence. Default: `false`. Updatable. |
| `persistence_type` | String | No | Persistence type (e.g., `source_ip`). Updatable. |
| `x_forwarded_for` | Bool | No | Add X-Forwarded-For header. Default: `false`. Updatable. |
| `redirect_https` | Bool | No | Redirect HTTP to HTTPS. Default: `false`. |
| `certificate_id` | String | No | SSL certificate ID (for HTTPS). |
| `monitor_protocol` | String | No | Health check monitor protocol. |

**Node Object Attributes:**

| Argument | Type | Required | Description |
|---|---|---|---|
| `compute_id` | Int64 | Yes | Compute instance numeric ID. |
| `compute_ip` | String | Yes | Compute instance private IP. |
| `port` | Int64 | Yes | Backend port. |
| `weight` | Int64 | No | Load balancing weight. |
| `max_conn` | Int64 | No | Maximum connections. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Virtual server ID. |
| `status` | String | Current status. |
| `vip` | String | VIP address. |

#### Timeouts

| Operation | Default |
|---|---|
| `create` | 5 minutes |
| `delete` | 5 minutes |

#### Import

```bash
terraform import airtelcloud_lb_virtual_server.http <lb_service_id>/<virtual_server_id>
```

---

## Backup & Protection

### airtelcloud_protection_plan

Manages a backup protection plan that defines schedule and retention policies.

~> **Warning:** The Airtel Cloud API does not support deletion of protection plans. Running `terraform destroy` will only remove the plan from Terraform state -- the plan will continue to exist in the Airtel Cloud platform.

#### Example Usage

```terraform
resource "airtelcloud_protection_plan" "daily" {
  name           = "daily-backup"
  description    = "Daily backup with 30-day retention"
  schedule_type  = "daily"
  retention      = 30
  retention_unit = "days"
  recurrence     = 1
}
```

#### Argument Reference

All arguments force replacement on change (protection plans are immutable).

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Plan name. |
| `description` | String | No | Plan description. |
| `schedule_type` | String | No | Schedule type (e.g., `daily`, `weekly`). |
| `selector_key` | String | No | Selector key for filtering. |
| `selector_value` | String | No | Selector value. |
| `retention` | Int64 | No | Number of backups to retain. |
| `retention_unit` | String | No | Retention unit (e.g., `days`). |
| `recurrence` | Int64 | No | Recurrence interval. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Plan ID. |

#### Import

```bash
terraform import airtelcloud_protection_plan.daily <plan-id>
```

---

### airtelcloud_protection

Manages a backup protection policy for a compute instance. Links a VM to a protection plan with scheduling options. Supports in-place updates for name, description, schedule, and plan assignment.

#### Example Usage

```terraform
resource "airtelcloud_protection" "web_server" {
  name             = "web-backup"
  description      = "Backup policy for web server"
  compute_id       = airtelcloud_vm.web1.id
  protection_plan  = airtelcloud_protection_plan.daily.name
  enable_scheduler = "true"
  start_date       = "2026-04-01"
  start_time       = "02:00"
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Protection policy name. Updatable. |
| `compute_id` | String | Yes | Compute instance ID. Forces new resource. |
| `protection_plan` | String | Yes | Protection plan name. Updatable. |
| `description` | String | No | Policy description. Updatable. |
| `policy_type_id` | String | No | Policy type ID. |
| `enable_scheduler` | String | No | Enable scheduled backups. Default: `"true"`. Updatable. |
| `start_date` | String | No | Schedule start date (e.g., `2026-04-01`). Updatable. |
| `end_date` | String | No | Schedule end date. Updatable. |
| `start_time` | String | No | Schedule start time (e.g., `02:00`). Updatable. |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | String | Protection policy ID. |
| `status` | String | Current status. |
| `region` | String | Region. |
| `az_name` | String | Availability zone. |
| `created` | String | Creation timestamp. |

#### Import

```bash
terraform import airtelcloud_protection.web_server <protection-id>
```

---

## Additional Resources

The following resources are available in the provider but were not included in the main walkthrough because they require specific prerequisites.

### airtelcloud_vpc_peering

Creates a peering connection between two VPCs. Requires two existing VPCs.

```terraform
resource "airtelcloud_vpc_peering" "example" {
  name            = "vpc-peering"
  description     = "VPC peering between two networks"
  vpc_source_name = "vpc-1"
  vpc_target_name = "vpc-2"
  az              = "S2"
  region          = "south"
  is_pcl_enabled  = false

  timeouts {
    create = "25m"
    delete = "15m"
  }
}
```

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | String | Yes | Peering name. Forces new resource. |
| `az` | String | Yes | Availability zone. Forces new resource. |
| `region` | String | Yes | Region. Forces new resource. |
| `description` | String | No | Description. Forces new resource. |
| `vpc_source_id` | String | No | Source VPC ID. One of `vpc_source_id` or `vpc_source_name` required. |
| `vpc_source_name` | String | No | Source VPC name. |
| `vpc_target_id` | String | No | Target VPC ID. One of `vpc_target_id` or `vpc_target_name` required. |
| `vpc_target_name` | String | No | Target VPC name. |
| `is_pcl_enabled` | Bool | No | Enable peering control list. |
| `allowed_subnet_list` | List of String | No | Allowed subnet IDs (only when `is_pcl_enabled = true`). |
| `blocked_subnet_list` | List of String | No | Blocked subnet IDs (only when `is_pcl_enabled = true`). |

Timeouts: create (5m default), delete (5m default). Import: `terraform import airtelcloud_vpc_peering.example <id>`

---

### airtelcloud_compute_snapshot

Creates an immutable point-in-time snapshot of a VM. Requires an existing compute instance.

```terraform
resource "airtelcloud_compute_snapshot" "web_snapshot" {
  compute_id = airtelcloud_vm.web1.id

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

| Argument | Type | Required | Description |
|---|---|---|---|
| `compute_id` | String | Yes | VM ID to snapshot. Forces new resource. |

Computed: `id` (UUID), `name`, `status`, `is_active`, `is_image`, `created`. Timeouts: create (10m default), delete (10m default). Import: `terraform import airtelcloud_compute_snapshot.web_snapshot <uuid>`

-> **Note:** Snapshots are immutable and cannot be updated.

---

### airtelcloud_lb_certificate

Uploads an SSL certificate to a load balancer service for HTTPS termination. Requires PEM-encoded certificate files.

```terraform
resource "airtelcloud_lb_certificate" "web_cert" {
  lb_service_id   = airtelcloud_lb_service.web_lb.id
  name            = "web-cert"
  ssl_cert        = file("${path.module}/certs/server.pem")
  ssl_private_key = file("${path.module}/certs/server-key.pem")
}
```

| Argument | Type | Required | Description |
|---|---|---|---|
| `lb_service_id` | String | Yes | Parent LB service ID. Forces new resource. |
| `name` | String | Yes | Certificate name. Forces new resource. |
| `ssl_cert` | String | Yes | PEM-encoded certificate. Sensitive. Forces new resource. |
| `ssl_private_key` | String | Yes | PEM-encoded private key. Sensitive. Forces new resource. |
| `ca_cert` | String | No | PEM-encoded CA certificate. Sensitive. Forces new resource. |

Computed: `id`, `status`. Import: `terraform import airtelcloud_lb_certificate.web_cert <lb_service_id>/<certificate_id>`

-> **Note:** Certificate content is write-only and cannot be read back from the API.

---

## Resource Dependency Graph

The following diagram shows the provisioning order used in the complete example:

```
Security Group ──> Security Group Rules
                                                   ┌──> Compute Snapshot
VPC ──> Subnet ──> VM ──────────────────────┤
                    │                               └──> Protection ──> Protection Plan
                    │
                    ├──> Volume (attach)
                    │
                    └──> Public IP ──> Policy Rule

File Storage ──> NFS Export Path

DNS Zone ──> DNS Record

LB Service ──> LB VIP ──────────────────┐
               LB Certificate ──────────┤
                                        └──> LB Virtual Server (references VM IPs)

Object Storage Bucket (independent)
```

---

## Importing Existing Resources

To import an existing Airtel Cloud resource into Terraform state:

```bash
terraform import <resource_type>.<name> <import_id>
```

### Import ID Formats

| Resource | Import ID Format | Example |
|---|---|---|
| `airtelcloud_security_group` | `<numeric-id>` | `42` |
| `airtelcloud_security_group_rule` | `<sg_id>/<rule_id>` | `42/10` |
| `airtelcloud_subnet` | `<network_id>/<subnet_id>` | `029a.../35df...` |
| `airtelcloud_public_ip` | `<uuid>` | `a1b2c3d4-...` |
| `airtelcloud_public_ip_policy_rule` | `<public_ip_id>/<target_vip>/<public_ip>/<rule_id>` | `uuid/10.1.1.5/203.0.113.5/uuid` |
| `airtelcloud_vm` | `<compute-id>` | `b603ccb5-...` |
| `airtelcloud_storage_bucket` | `<bucket-name>` | `my-bucket` |
| `airtelcloud_volume` | `<numeric-id>` | `123` |
| `airtelcloud_file_storage` | `<id>` | `fs-abc123` |
| `airtelcloud_file_storage_export_path` | `<path-id>` | `exp-xyz789` |
| `airtelcloud_dns_zone` | `<zone-uuid>` | `a1b2c3d4-...` |
| `airtelcloud_dns_record` | `<zone_id>/<record_id>` | `zone-uuid/record-uuid` |
| `airtelcloud_lb_service` | `<lb-service-id>` | `lb-svc-1` |
| `airtelcloud_lb_vip` | `<lb_service_id>/<vip_id>` | `lb-svc-1/vip-1` |
| `airtelcloud_lb_virtual_server` | `<lb_service_id>/<vs_id>` | `lb-svc-1/vs-1` |
| `airtelcloud_lb_certificate` | `<lb_service_id>/<cert_id>` | `lb-svc-1/cert-1` |
| `airtelcloud_protection_plan` | `<plan-id>` | `1` |
| `airtelcloud_protection` | `<protection-id>` | `5` |
| `airtelcloud_vpc_peering` | `<peering-id>` | `peer-abc123` |
| `airtelcloud_compute_snapshot` | `<snapshot-uuid>` | `snap-uuid-1234` |

---

## Troubleshooting

### 401 Unauthorized

Verify your `api_key` and `api_secret` are correct. The provider uses HMAC-SHA256 authentication with a 120-second token expiry -- ensure your system clock is synchronized.

### Region Mismatch

Ensure the `region` in your provider configuration matches the region where your resources exist. The default is `south`.

### Timeout Errors

For resources that support timeouts (`subnet`, `lb_service`, `lb_virtual_server`, `public_ip`, `vpc_peering`, `compute_snapshot`), increase the timeout value:

```terraform
timeouts {
  create = "30m"
  delete = "15m"
}
```

### Protection Plan Cannot Be Deleted

The Airtel Cloud API does not support deleting protection plans. Running `terraform destroy` removes the plan from Terraform state only. The plan continues to exist on the platform. This is expected behavior.

### Provider Not Found

If Terraform cannot find the provider, ensure the `required_providers` block is configured correctly and you have run `terraform init`. For local development builds, use `dev_overrides` in your Terraform CLI configuration -- see the [Getting Started Guide](getting-started) for details.
