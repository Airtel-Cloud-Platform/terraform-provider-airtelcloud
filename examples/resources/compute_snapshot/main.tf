terraform {
  required_providers {
    airtelcloud = {
      source  = "terraform-providers/airtelcloud"
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

# Create a snapshot of a compute instance
resource "airtelcloud_compute_snapshot" "example" {
  compute_id = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"

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
