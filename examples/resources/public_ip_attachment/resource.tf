terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.6"
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

# Step 2 of 3: attach the reserved public IP to a resource.
# For vm and baremetal, the private IP is looked up from the named resource.
# For lb, set target_vip to the VIP to attach to.
resource "airtelcloud_public_ip_attachment" "vm" {
  public_ip_name = "my-vm-public-ip"
  resource_type  = "vm"
  resource_name  = "my-vm"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

# Attach to a load balancer VIP
resource "airtelcloud_public_ip_attachment" "lb" {
  public_ip_name = "my-lb-public-ip"
  resource_type  = "lb"
  resource_name  = "my-lb"
  target_vip     = "10.101.21.35"
}

# Attach to a baremetal server
resource "airtelcloud_public_ip_attachment" "baremetal" {
  public_ip_name = "my-bm-public-ip"
  resource_type  = "baremetal"
  resource_name  = "my-baremetal"
}

output "attachment_status" {
  description = "Status after attach (attached)"
  value       = airtelcloud_public_ip_attachment.vm.status
}

output "attachment_target_vip" {
  description = "Private IP the public IP is NATted to"
  value       = airtelcloud_public_ip_attachment.vm.target_vip
}

output "attachment_availability_zone" {
  description = "Availability zone read from the public IP"
  value       = airtelcloud_public_ip_attachment.vm.availability_zone
}
