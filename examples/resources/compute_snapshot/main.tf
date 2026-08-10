terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.1.5"
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

# Create a snapshot of a compute instance.
#
# Specify the target compute by ID or by name (exactly one of compute_id /
# compute_name). When compute_name is set it is resolved to compute_id at apply.
#
# Import an existing snapshot with either of:
#   terraform import airtelcloud_compute_snapshot.example <snapshot_uuid>
#   terraform import airtelcloud_compute_snapshot.example <compute_id>:<snapshot_uuid>
resource "airtelcloud_compute_snapshot" "example" {
  #compute_id = "aa8ad6fc-5400-452f-95ee-8ecb95a7f3d4"
  # Or reference the compute by name instead of compute_id:
  compute_name = "aakash"
  snapshot_name = "${var.resource_prefix}-snapshot"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

# Output snapshot details
output "snapshot_id" {
  description = "ID of the compute snapshot"
  value       = airtelcloud_compute_snapshot.example.id
}

output "snapshot_name" {
  description = "Name of the compute snapshot"
  value       = airtelcloud_compute_snapshot.example.name
}

output "snapshot_status" {
  description = "Status of the compute snapshot"
  value       = airtelcloud_compute_snapshot.example.status
}

output "snapshot_is_active" {
  description = "Whether the snapshot is active"
  value       = airtelcloud_compute_snapshot.example.is_active
}

output "snapshot_created" {
  description = "Creation timestamp of the snapshot"
  value       = airtelcloud_compute_snapshot.example.created
}
