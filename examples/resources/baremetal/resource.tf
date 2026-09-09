terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.5"
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

variable "subnet_name" {
  description = "Primary subnet display name attached to the baremetal NIC"
  type        = string
}

variable "network_name" {
  description = "VPC name or UUID sent as networkInterface.name"
  type        = string
}

variable "additional_subnet_names" {
  description = "Optional extra subnet display names attached alongside the primary subnet"
  type        = list(string)
  default     = []
}

variable "keypair_name" {
  description = "Existing SSH keypair name"
  type        = string
}

variable "keypair_id" {
  description = "Existing SSH keypair UUID"
  type        = string
}

variable "public_key" {
  description = "SSH public key injected at allocate time"
  type        = string
}

variable "availability_zone" {
  description = "Availability zone matching the primary subnet"
  type        = string
  default     = "S1"
}

variable "os_image" {
  description = "OS image name, matching the UI osImage field"
  type        = string
  default     = "rhel/RHEL9-v1-Aug2026"
}

variable "tags" {
  description = "Allocation tags. The console sends a placement tag such as South-AZ1."
  type        = list(string)
  default     = ["South-AZ1"]
}

resource "airtelcloud_baremetal" "app" {
  name                    = "${var.resource_prefix}-bm-app-01"
  flavor                  = "metal-c56-m1024"
  os_image                = var.os_image
  subnet_name             = var.subnet_name
  additional_subnet_names = var.additional_subnet_names
  availability_zone       = var.availability_zone
  network_name            = var.network_name

  keypair    = var.keypair_name
  keypair_id = var.keypair_id
  public_key = var.public_key
  # cloud_init is generated from public_key in the same runcmd shape the console sends
  is_reserved    = false
  policy_enabled = true
  tags           = var.tags

  storage = [
    {
      name         = "baremetel"
      size         = "10"
      path         = "/test"
      type         = "BlockStorage"
      file_system  = "xfs"
      force_format = true
    }
  ]

  backup_config = {
    schedule_type       = "weekly_full"
    start_time          = "21:00"
    incr_days           = []
    full_days           = [7]
    full_retention      = 1
    full_retention_unit = "MONTHS"
    backup_selections   = ["/test"]
  }

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
