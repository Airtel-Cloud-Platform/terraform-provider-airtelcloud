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

# Create NFS volume
resource "airtelcloud_file_storage" "basic" {
  name              = "${var.resource_prefix}-basic-nfs"
  size              = "300"
  availability_zone = "S2"
}

# NFS volume with description
#resource "airtelcloud_file_storage" "with_description" {
#  name              = "${var.resource_prefix}-app-data"
#  description       = "Shared storage for application data"
#  size              = "250"
#  availability_zone = "S2"
#}

# Output basic volume details
output "basic_volume_id" {
  description = "ID of the basic NFS volume"
  value       = airtelcloud_file_storage.basic.id
}

output "basic_volume_state" {
  description = "State of the basic NFS volume"
  value       = airtelcloud_file_storage.basic.state
}

#output "app_data_volume_id" {
#  description = "ID of the app data NFS volume"
#  value       = airtelcloud_file_storage.with_description.id
#}
