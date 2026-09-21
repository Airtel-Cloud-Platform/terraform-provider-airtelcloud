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

# Step 1 of 3: reserve a public IP.
# Create sends port_id as null, so no VM or port is needed yet.
resource "airtelcloud_public_ip" "example" {
  object_name       = "my-vm-public-ip"
  description       = "reserved public IP"
  availability_zone = "S1"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

output "public_ip_name" {
  description = "Name used by the attachment and policy resources"
  value       = airtelcloud_public_ip.example.object_name
}

output "public_ip_address" {
  description = "Allocated public IP address"
  value       = airtelcloud_public_ip.example.public_ip
}

output "public_ip_status" {
  description = "Status after reservation (reserved)"
  value       = airtelcloud_public_ip.example.status
}
