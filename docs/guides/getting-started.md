---
page_title: "Getting Started with the Airtel Cloud Provider"
subcategory: ""
description: |-
  A step-by-step guide to building, configuring, and using the Airtel Cloud Terraform provider.
---

# Getting Started with the Airtel Cloud Provider

This guide walks you through building the Airtel Cloud Terraform provider from source, configuring authentication, and deploying your first resources.

## Overview

The Airtel Cloud Terraform provider lets you manage infrastructure on Airtel Cloud Platform using declarative HCL configuration. Supported resources include:

- **Virtual Machines** (`airtelcloud_vm`)
- **Block Storage Volumes** (`airtelcloud_volume`)
- **Virtual Private Clouds** (`airtelcloud_vpc`)
- **Subnets** (`airtelcloud_subnet`)
- **Object Storage Buckets** (`airtelcloud_storage_bucket`)
- **File Storage / NFS Volumes** (`airtelcloud_file_storage`)
- **File Storage Export Paths** (`airtelcloud_file_storage_export_path`)
- **DNS Zones** (`airtelcloud_dns_zone`)
- **DNS Records** (`airtelcloud_dns_record`)

## Prerequisites

| Requirement | Version |
|---|---|
| [Terraform](https://www.terraform.io/downloads.html) | >= 1.0 |
| [Go](https://golang.org/doc/install) | >= 1.24 |
| Airtel Cloud account | With API key and secret |

## Building from Source

Clone the repository and build using the Makefile:

```shell
git clone https://github.com/terraform-providers/terraform-provider-airtelcloud.git
cd acp-terraform

# Compile the provider binary
make build

# Install to the local Terraform plugin directory
make install
```

`make install` copies the binary to both locations:

- `~/.terraform.d/plugins/registry.terraform.io/terraform-providers/airtelcloud/0.2.0/linux_amd64/`
- `~/go/bin/`

-> **Note:** On macOS, replace `linux_amd64` with `darwin_amd64` in the Makefile `OS_ARCH` variable before running `make install`.

## Local Development Setup

For active development, use Terraform's `dev_overrides` so you can test changes without reinstalling into the plugin directory each time.

Create or update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "terraform-providers/airtelcloud" = "<GOPATH>/bin"
  }
  direct {}
}
```

Replace `<GOPATH>` with the output of `go env GOPATH` (typically `~/go`).

With `dev_overrides` active, `terraform init` is not required — Terraform will load the provider binary directly from the specified path.

## Provider Configuration

Add the provider block to your Terraform configuration:

```terraform
terraform {
  required_providers {
    airtelcloud = {
      source  = "terraform-providers/airtelcloud"
      version = "~> 0.2"
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

### Configuration Attributes

| Attribute | Type | Required | Default | Description |
|---|---|---|---|---|
| `api_endpoint` | String | No | `https://api.south.cloud.airtel.in` | Airtel Cloud API endpoint URL |
| `api_key` | String | Yes | — | API key for authentication (sensitive) |
| `api_secret` | String | Yes | — | API secret for HMAC authentication (sensitive) |
| `region` | String | No | `south-1` | Airtel Cloud region |
| `organization` | String | No | — | Organization name or domain |
| `project_name` | String | No | — | Project name for API calls |

## Authentication

The provider authenticates using an HMAC-SHA256 scheme that requires both an `api_key` and an `api_secret`.

### Option 1: Environment Variables (Recommended)

```bash
export AIRTEL_API_KEY="your-api-key"
export AIRTEL_API_SECRET="your-api-secret"
```

### Option 2: Variable Files

Create a `terraform.tfvars` file (never commit this to version control):

```hcl
airtel_api_key    = "your-api-key"
airtel_api_secret = "your-api-secret"
organization      = "your-organization"
project_name      = "your-project"
```

Declare the corresponding variables in your configuration:

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

variable "organization" {
  description = "Organization for the resources"
  type        = string
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
}
```

-> **Tip:** Add `*.tfvars` to your `.gitignore` to avoid accidentally committing credentials.

## Execution Steps

Run the standard Terraform workflow:

```shell
# Initialize the provider (downloads plugins, sets up backend)
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply

# Inspect current state
terraform show

# Tear down all managed resources
terraform destroy
```

## Complete Example Walkthrough

A working example is available at [`examples/complete/main.tf`](../../examples/complete/main.tf). It demonstrates four resources:

### Object Storage Bucket

Creates a private bucket with versioning enabled and local replication:

```terraform
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "${var.bucket_prefix}-private-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"
  versioning        = true
  object_locking    = false
}
```

### Subnet

Creates a private subnet within an existing VPC, with labels and custom timeouts:

```terraform
resource "airtelcloud_subnet" "private" {
  network_id         = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  name               = "ashok-d-subnet"
  ipv4_address_space = "10.1.14.0/24"
  subnet_sub_role    = "Private"
  availability_zone  = "S2"
  labels             = ["Provider Test"]

  timeouts {
    create = "5m"
    delete = "4m"
  }
}
```

### File Storage (NFS Volume)

Creates a 300 GB NFS volume:

```terraform
resource "airtelcloud_file_storage" "basic" {
  name              = "basic-nfs"
  size              = "300"
  availability_zone = "S2"
}
```

### File Storage Export Path

Creates an NFSv4 export path on the volume above:

```terraform
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

To run the complete example:

```shell
cd examples/complete
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init && terraform plan
```

## Debugging

Enable Terraform's built-in debug logging with environment variables:

```shell
# Set log level (TRACE, DEBUG, INFO, WARN, ERROR)
export TF_LOG=DEBUG

# Optionally write logs to a file
export TF_LOG_PATH=terraform-debug.log

terraform apply
```

## Troubleshooting

### 401 Unauthorized

- Verify your `api_key` and `api_secret` are correct and have not expired.
- Confirm the `api_endpoint` matches your account's region.

### Region Mismatch

- Resources may fail if the `region` or `availability_zone` in your configuration does not match a valid Airtel Cloud zone.
- Check the default region (`south-1`) and available zones (`S1`, `S2`) for your account.

### Provider Not Found After Build

- If you used `make install`, ensure the `OS_ARCH` in the Makefile matches your system (e.g., `linux_amd64`, `darwin_amd64`).
- If you are using `dev_overrides`, ensure the path in `~/.terraformrc` matches the output of `go env GOPATH`.
- Run `terraform providers` to verify Terraform detects the provider.
