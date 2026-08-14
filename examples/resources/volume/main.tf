terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.1"
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

# Create a block storage volume with all common attributes
resource "airtelcloud_volume" "data_volume" {
  name              = "${var.resource_prefix}-storage-volume"
  size              = 50
  type              = "s1_wkld_ntp02_4iops_backend"
  availability_zone = "S1"
  compute_name        = "my-instance" # Name of the compute instance to attach the volume to
  # Or attach by name instead of id (mutually exclusive with compute_id):
  # compute_name    = "my-instance"

  # Networking
  vpc_name    = "copper-vpc01"
  subnet_name = "subnet-26jun"

  is_encrypted = false
  bootable     = false
}

# Create a secondary volume (alternative configuration)
#resource "airtelcloud_volume" "backup_volume" {
#  name              = "${var.resource_prefix}-backup-volume"
#  size              = 100
#  type              = "s1_wkld_ntp02_4iops_backend"
#  availability_zone = "S1"
#  compute_id        = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"
#
#  vpc_id    = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
#  subnet_id = "35df162d-5211-4d58-84ed-6a499626949c"
#
#  is_encrypted = true
#  bootable     = false
#}

# Output volume details
output "data_volume_uuid" {
  description = "UUID of the data volume"
  value       = airtelcloud_volume.data_volume.uuid
}

output "data_volume_status" {
  description = "Status of the data volume"
  value       = airtelcloud_volume.data_volume.status
}

#output "backup_volume_uuid" {
#  description = "UUID of the backup volume"
#  value       = airtelcloud_volume.backup_volume.uuid
#}
