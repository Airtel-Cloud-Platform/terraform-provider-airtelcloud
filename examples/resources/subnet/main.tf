terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "~> 0.2"
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
  description = "organization for the resources"
  type        = string
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
}

variable "resource_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "tft"
}

# Create a private subnet
resource "airtelcloud_subnet" "private" {
  network_id         = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  name               = "${var.resource_prefix}-tf1-subnet"
  ipv4_address_space = "10.1.19.0/24"
  subnet_sub_role    = "Private"
  availability_zone  = "S2"
  labels             = ["Provider Test"]

  timeouts {
    create = "5m"
    delete = "4m"
  }
}

# Create a VIP subnet
#resource "airtelcloud_subnet" "vip" {
#  network_id         = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
#  name               = "${var.resource_prefix}-vip-subnet"
#  ipv4_address_space = "10.1.12.0/24"
#  availability_zone  = "S2"
#  subnet_sub_role    = "Vip"
#  labels             = ["VIP Provider Test"]
#  description        = "VIP subnet for load balancers"
#
#  timeouts {
#    create = "5m"
#    delete = "4m"
#  }
#}

# Output subnet details
output "private_subnet_id" {
  description = "ID of the private subnet"
  value       = airtelcloud_subnet.private.id
}

#output "vip_subnet_id" {
#  description = "ID of the VIP subnet"
#  value       = airtelcloud_subnet.vip.id
#}
