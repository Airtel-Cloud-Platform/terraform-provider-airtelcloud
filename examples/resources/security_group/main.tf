terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.2"
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

# Create a security group
# resource "airtelcloud_security_group" "web" {
#   security_group_name = "${var.resource_prefix}-sg-http-servers-2"
#   availability_zone   = "N2"
# }

resource "airtelcloud_security_group" "web" {
  security_group_name = "${var.resource_prefix}-sg-http-servers"
  availability_zone   = "N1"
}

resource "airtelcloud_security_group_rule" "http" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "80"
  port_range_max    = "80"
  remote_ip_prefix  = "0.0.0.0/0"
  ethertype         = "IPv4"
}

# Output security group details
output "security_group_id" {
  description = "ID of the security group"
  value       = airtelcloud_security_group.web.id
}

output "security_group_uuid" {
  description = "UUID of the security group"
  value       = airtelcloud_security_group.web.uuid
}

output "security_group_status" {
  description = "Status of the security group"
  value       = airtelcloud_security_group.web.status
}
