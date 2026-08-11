terraform {
  required_providers {
    airtelcloud = {
      source = "Airtel-Cloud-Platform/airtelcloud"
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

variable "vm_admin_password" {
  description = "Login password for the VM admin user"
  type        = string
  sensitive   = true
}

variable "vm_keypair_name" {
  description = "Existing SSH keypair name for Linux keypair-auth example"
  type        = string
  default     = "keypair-1"
}

# Create a Linux virtual machine authenticated with username/password.
#
# A linux instance needs exactly one authentication method: either
# admin_username + admin_password, or keypair_id / keypair_name. Setting both
# is rejected at plan time.
#
# Labels are applied with a follow-up PATCH after the VM is created because the
# compute create endpoint ignores labels. To patch labels onto an already
# created VM, change this labels list and run `terraform apply` again.
resource "airtelcloud_vm" "web_server" {
  instance_name       = "${var.resource_prefix}-web-server-3"
  os_type             = "linux"
  flavor_name         = "ccd.Large"
  image_name          = "Ubuntu22_04_Aug2026"
  vpc_name            = "copper-vpc01"
  subnet_name         = "sub1"
  boot_from_volume    = true
  disk_size           = 100
  availability_zone   = "N1"
  security_group_name = "all-open-az1"
  admin_username      = "clouduser"
  admin_password      = var.vm_admin_password
  description         = "Example web server instance"
  enable_backup       = true
  protection_plan     = "daily"
  start_date          = "2025-07-15"
  start_time          = "02:00"

  labels = ["example", "web-server"]
}

# Example update for an already created VM:
# 1. Apply once with the labels above.
# 2. Change the labels list below into the resource block.
# 3. Run `terraform apply` again to send the labels PATCH for the existing VM.
#
# labels = ["example", "web-server", "backend-app"]


# # Create a Windows VM with backup enabled (alternative configuration)
resource "airtelcloud_vm" "windows_server" {
  instance_name       = "${var.resource_prefix}-win-server"
  os_type             = "windows"
  flavor_name         = "ccd.Large"
  image_name          = "WIN2K22_BYOL_Jul2026"
  vpc_name            = "copper-vpc01"
  subnet_name         = "subnet-az1"
  boot_from_volume    = true
  disk_size           = 200
  availability_zone   = "N1"
  security_group_name = "all-open-az1"
  description         = "Example Windows server with backup"

  #  enable_backup    = true
  #  protection_plan  = "daily-backup-plan"
  #  start_date       = "2026-04-01"
  #  start_time       = "02:00"
  #
  labels = ["backup", "example", "windows-server"]
}

resource "airtelcloud_volume" "data_volume" {
  name              = "${var.resource_prefix}-storage-volume"
  size              = 50
  type              = "n2_wkld_ntp02_1iops_backend"
  availability_zone = "N1"
  #compute_id = airtelcloud_vm.web_server.id
  vpc_name       = "copper-vpc01"
  subnet_name    = "subnet-az2"
  is_encrypted = false
  bootable     = false
  depends_on   = [airtelcloud_vm.web_server]
}

resource "airtelcloud_compute_snapshot" "example" {
  compute_id    = airtelcloud_vm.web_server.id
  snapshot_name = "tft-snap-as"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

# Output VM details
output "web_vm_id" {
  description = "ID of the web server VM"
  value       = airtelcloud_vm.web_server.id
}

output "web_vm_status" {
  description = "Status of the web server VM"
  value       = airtelcloud_vm.web_server.status
}

output "web_vm_private_ip" {
  description = "Private IP of the web server VM"
  value       = airtelcloud_vm.web_server.private_ip
}

output "windows_vm_id" {
  description = "ID of the windows server VM"
  value       = airtelcloud_vm.windows_server.id
}
