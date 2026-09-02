terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.3"
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

variable "subnet_id" {
  description = "Subnet ID where baremetal primary interface is attached"
  type        = string
}

variable "keypair_name" {
  description = "Existing SSH keypair name"
  type        = string
  default     = "keypair-1"
}

resource "airtelcloud_baremetal" "app" {
  name              = "${var.resource_prefix}-bm-app-01"
  flavor            = "metal-c56-m1024"
  os_image          = "Ubuntu22_04_Aug2026"
  subnet_id         = var.subnet_id
  availability_zone = "S2"
  network_name      = "eth0"

  keypair        = var.keypair_name
  is_reserved    = false
  policy_enabled = true
  tags           = ["terraform", "baremetal", "example"]

  # Destruction controls
  delete_disks = true
  secure_erase = false
}

output "baremetal_name" {
  description = "Baremetal server name"
  value       = airtelcloud_baremetal.app.name
}

output "baremetal_uuid" {
  description = "Baremetal server UUID"
  value       = airtelcloud_baremetal.app.uuid
}

output "baremetal_state" {
  description = "Baremetal server state"
  value       = airtelcloud_baremetal.app.state
}

output "baremetal_ip_addresses" {
  description = "Baremetal IP addresses"
  value       = airtelcloud_baremetal.app.ip_addresses
}

output "baremetal_backend_port_id" {
  description = "Baremetal backend port id"
  value       = airtelcloud_baremetal.app.backend_port_id
}
