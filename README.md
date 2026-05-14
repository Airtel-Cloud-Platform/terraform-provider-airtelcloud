# Terraform Provider for Airtel Public Cloud

This repository contains a Terraform provider for managing Airtel Public Cloud resources.

## Supported Resources
- **Virtual Machine (VM)**: Create and manage Virtual machines
- **Volumes**: Manage block storage volumes and attachments
- **Virtual Private Cloud (VPC)**: Create and manage VPCs and VPC Peering
- **Subnets**: Manage VPC subnets
- **Object Storage**: Manage storage buckets
- **File Storage (NFS)**: Create and manage NFS volumes and export paths
- **DNS Zones**: Manage DNS zones
- **DNS Records**: Manage DNS records within zones
- **Security Groups & Rules**: Manage Security Groups and Associated Security Rules.
- **Load Balancer**: Create and manage Load Balancers
- **Backup**: Create and manage virtual machine backups and protection plans.
- **Public IP**: Create and manage Public IPs and related policies.



## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Building the Provider

Clone the repository and use the Makefile:

```shell
git clone https://github.com/terraform-providers/terraform-provider-airtelcloud.git
cd acp-terraform

# Compile the provider binary
make build

# Build and install to the local Terraform plugin directory
make install
```

`make install` places the binary at:

```
~/.terraform.d/plugins/registry.terraform.io/terraform-providers/airtelcloud/0.2.0/<OS_ARCH>/
```

where `<OS_ARCH>` is `linux_amd64` by default (configurable in the Makefile).

### Local Development with dev_overrides

For rapid iteration, configure Terraform to load the provider from your `GOPATH` instead of the plugin directory.

Create or update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "Airtel-Cloud-Platform/airtelcloud" = "<GOPATH>/bin"
  }
  direct {}
}
```

Replace `<GOPATH>` with the output of `go env GOPATH` (typically `~/go`). With `dev_overrides` active, run `make build && make install` and Terraform will pick up the new binary without needing `terraform init`.

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the Provider

> See the full [Getting Started Guide](docs/guides/getting-started.md) for a detailed walkthrough.

### Provider Configuration

```hcl
terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
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

### Example Usage

A complete working example is available at [`examples/complete/main.tf`](examples/complete/main.tf). It demonstrates creating an object storage bucket, a subnet, an NFS volume, and an NFS export path.

```shell
cd examples/complete
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init && terraform plan
```

## Development

### Running the tests

```shell
make test
```

### Running acceptance tests

```shell
make testacc
```

### Generating documentation

```shell
make docs-generate
```

