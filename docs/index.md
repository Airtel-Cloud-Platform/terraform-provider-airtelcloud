---
page_title: "Airtel Cloud Provider"
subcategory: ""
description: |-
  The Airtel Cloud provider for Terraform allows you to manage Airtel Cloud resources.
---

# Airtel Cloud Provider

The Airtel Cloud provider allows Terraform to manage Airtel Cloud infrastructure resources such as virtual machines, volumes, VPCs, subnets, object storage buckets, file storage (NFS), and DNS records.

-> **New to this provider?** See the [Getting Started Guide](guides/getting-started) for a full walkthrough.

-> **Complete Resource Guide:** See the [User Guide](guides/user-guide) for all 22 resources with examples, argument references, and import instructions.

## Example Usage

```terraform
terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "~> 1.0.0"
    }
  }
}

provider "airtelcloud" {
  api_endpoint = "https://api.south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south-1"
  organization = var.organization
  project_name = var.project_name
}
```

## Authentication

The provider authenticates using an HMAC-SHA256 scheme that requires both an API key and an API secret. You can supply credentials in the following ways:

1. **Provider configuration** (recommended for CI/CD)
2. **Environment variables**: `AIRTEL_API_KEY`, `AIRTEL_API_SECRET`
3. **Terraform variable files** (see below)

### Provider Configuration

```terraform
provider "airtelcloud" {
  api_endpoint = "https://api.south.cloud.airtel.in"
  api_key      = "your-api-key-here"
  api_secret   = "your-api-secret-here"
  region       = "south-1"
}
```

### Environment Variables

```bash
export AIRTEL_API_KEY="your-api-key-here"
export AIRTEL_API_SECRET="your-api-secret-here"
```

### Using Variables

```terraform
variable "airtel_api_key" {
  description = "Airtel Cloud API key"
  type        = string
  sensitive   = true
}

variable "airtel_api_secret" {
  description = "Airtel Cloud API secret"
  type        = string
  sensitive   = true
}

provider "airtelcloud" {
  api_endpoint = "https://api.south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south-1"
}
```

## Configuration Reference

The following arguments are supported in the provider configuration:

- `api_endpoint` (Optional) - The Airtel Cloud API endpoint URL. Defaults to `https://api.south.cloud.airtel.in`.
- `api_key` (Required) - The API key for authenticating with Airtel Cloud API. Can also be set via the `AIRTEL_API_KEY` environment variable.
- `api_secret` (Required) - The API secret for HMAC authentication. Can also be set via the `AIRTEL_API_SECRET` environment variable.
- `region` (Optional) - The default region for resources. Defaults to `south-1`.
- `organization` (Optional) - Organization name or domain for Airtel Cloud.
- `project_name` (Optional) - Project name for Airtel Cloud API calls.

## Available Resources

The provider supports the following 22 resources:

**Compute:**
- [`airtelcloud_vm`](resources/vm) - Virtual Machine
- `airtelcloud_compute_snapshot` - Compute Snapshot

**Storage:**
- [`airtelcloud_volume`](resources/volume) - Block Storage Volume
- [`airtelcloud_storage_bucket`](resources/storage_bucket) - Object Storage Bucket
- `airtelcloud_object_storage_access_key` - Object Storage Access Key
- [`airtelcloud_file_storage`](resources/file_storage) - File Storage (NFS Volume)
- [`airtelcloud_file_storage_export_path`](resources/file_storage_export_path) - File Storage Export Path

**Networking:**
- [`airtelcloud_vpc`](resources/vpc) - Virtual Private Cloud
- [`airtelcloud_subnet`](resources/subnet) - VPC Subnet
- `airtelcloud_security_group` - Security Group
- `airtelcloud_security_group_rule` - Security Group Rule
- `airtelcloud_vpc_peering` - VPC Peering
- `airtelcloud_public_ip` - Public IP
- `airtelcloud_public_ip_policy_rule` - Public IP Policy Rule

**DNS:**
- [`airtelcloud_dns_zone`](resources/dns_zone) - DNS Zone
- [`airtelcloud_dns_record`](resources/dns_record) - DNS Record

**Load Balancing:**
- `airtelcloud_lb_service` - Load Balancer Service
- `airtelcloud_lb_vip` - Load Balancer VIP
- `airtelcloud_lb_certificate` - Load Balancer Certificate
- `airtelcloud_lb_virtual_server` - Load Balancer Virtual Server

**Backup:**
- `airtelcloud_protection_plan` - Protection Plan
- `airtelcloud_protection` - Protection Policy

## Example Infrastructure

Here's an example that creates an object storage bucket and an NFS volume with an export path:

```terraform
# Create a private bucket with versioning
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "tf-private-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"
  versioning        = true
  object_locking    = false
}

# Create an NFS volume
resource "airtelcloud_file_storage" "basic" {
  name              = "basic-nfs"
  size              = "300"
  availability_zone = "S2"
}

# Create an NFS export path
resource "airtelcloud_file_storage_export_path" "nfsv4_export" {
  volume              = airtelcloud_file_storage.basic.name
  description         = "NFSv4 export for application servers"
  protocol            = "NFSv4"
  availability_zone   = "S2"
  nfs_export_path     = "/exports/app"
  default_access_type = "ReadWrite"
  default_user_squash = "RootSquash"
}
```

For a more complete example, see [`examples/complete/main.tf`](https://github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/tree/main/examples/complete).

## Support and Contributing

For support, please refer to the [GitHub repository](https://github.com/terraform-providers/terraform-provider-airtelcloud) or open an issue.

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.
