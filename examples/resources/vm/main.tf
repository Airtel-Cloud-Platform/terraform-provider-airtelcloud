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

# Create a Linux virtual machine
resource "airtelcloud_vm" "web_server" {
  instance_name     = "${var.resource_prefix}-web-server"
  os_type           = "linux"
  flavor_name       = "ccd.Large"
  image_name        = "CentOS_Stream9_Mar2026"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 40
  availability_zone = "S2"
  description       = "Example web server instance"

  tags = {
    Environment = "example"
    Role        = "web-server"
  }
}

# Create a Windows VM with backup enabled (alternative configuration)
#resource "airtelcloud_vm" "windows_server" {
#  instance_name     = "${var.resource_prefix}-win-server"
#  os_type           = "windows"
#  flavor_name       = "ccd.Large"
#  image_name        = "Windows_Server_2022"
#  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
#  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
#  boot_from_volume  = true
#  disk_size         = 100
#  availability_zone = "S2"
#  description       = "Example Windows server with backup"
#
#  enable_backup    = true
#  protection_plan  = "daily-backup-plan"
#  start_date       = "2026-04-01"
#  start_time       = "02:00"
#
#  tags = {
#    Environment = "example"
#    Role        = "windows-server"
#  }
#}

# Output VM details
output "vm_id" {
  description = "ID of the web server VM"
  value       = airtelcloud_vm.web_server.id
}

output "vm_status" {
  description = "Status of the web server VM"
  value       = airtelcloud_vm.web_server.status
}

output "vm_private_ip" {
  description = "Private IP of the web server VM"
  value       = airtelcloud_vm.web_server.private_ip
}

output "vm_public_ip" {
  description = "Public IP of the web server VM"
  value       = airtelcloud_vm.web_server.public_ip
}
