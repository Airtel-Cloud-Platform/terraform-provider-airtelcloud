terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.1.9"
    }
  }
}

provider "airtelcloud" {
  api_endpoint = "https://north.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "north"
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

# # Allocate a public IP NATted against a VM's private IP
resource "airtelcloud_public_ip" "example" {
  object_name       = "${var.resource_prefix}-my-vm-public-ip-1"
  # VIP must already exist on a VM NIC or LB VIP in this project and AZ.
  vip               = "10.10.3.237"
  availability_zone = "N1"

  timeouts {
    create = "10m"
    delete = "5m"
  }
}

output "public_ip_id" {
  value = airtelcloud_public_ip.example.id
}

output "public_ip_address" {
  value = airtelcloud_public_ip.example.public_ip
}

output "public_ip_status" {
  value = airtelcloud_public_ip.example.status
}

# Add a policy rule to allow HTTP and HTTPS traffic
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_id      = airtelcloud_public_ip.example.id
  display_name      = "web-traffic"
  source            = "any"
  services          = ["HTTP", "HTTPS"]
  action            = "accept"
  target_vip        = airtelcloud_public_ip.example.vip
  public_ip         = airtelcloud_public_ip.example.public_ip
  availability_zone = "N1"
}
