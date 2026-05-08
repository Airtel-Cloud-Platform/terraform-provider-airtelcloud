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

# Create a file storage volume first
resource "airtelcloud_file_storage" "basic" {
  name              = "${var.resource_prefix}-basic-nfs"
  size              = "300"
  availability_zone = "S2"
}

# Create an NFS export path with NFSv4 protocol
resource "airtelcloud_file_storage_export_path" "nfsv4_export" {
  volume            = airtelcloud_file_storage.basic.name
  description       = "NFSv4 export for application servers"
  protocol          = "NFSv4"
  availability_zone = "S2"

  # NFS export settings
  default_access_type = "ReadWrite"
  default_user_squash = "RootSquash"
}

# Create an NFS export path with NFSv3 protocol for legacy systems
#resource "airtelcloud_file_storage_export_path" "nfsv3_export" {
#  volume            = airtelcloud_file_storage.basic.name
#  description       = "NFSv3 export for legacy systems"
#  protocol          = "NFSv3"
#  availability_zone = "S2"
#
#  default_access_type = "ReadWrite"
#  default_user_squash = "NoSquash"
#}

# Output export path details
output "nfsv4_export_id" {
  description = "ID of the NFSv4 export path"
  value       = airtelcloud_file_storage_export_path.nfsv4_export.id
}

output "nfsv4_export_path" {
  description = "NFS export path for NFSv4"
  value       = airtelcloud_file_storage_export_path.nfsv4_export.nfs_export_path
}

output "nfsv4_path_id" {
  description = "Path ID for NFSv4 export"
  value       = airtelcloud_file_storage_export_path.nfsv4_export.path_id
}

#output "nfsv3_export_id" {
#  description = "ID of the NFSv3 export path"
#  value       = airtelcloud_file_storage_export_path.nfsv3_export.id
#}
