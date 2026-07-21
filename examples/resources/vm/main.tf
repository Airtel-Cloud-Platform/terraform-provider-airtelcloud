terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.1.1"
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

# Create a Linux virtual machine authenticated with username/password.
#
# A linux instance needs exactly one authentication method: either
# admin_username + admin_password, or keypair_id / keypair_name. Setting both
# is rejected at plan time.
resource "airtelcloud_vm" "web_server" {
  instance_name     = "${var.resource_prefix}-web-server"
  os_type           = "linux"
  flavor_name       = "ccd.Large"
  image_name        = "Ubuntu22_04_Jul2026"
  vpc_name          = "copper-vpc01"
  subnet_name       = "sub1"
  boot_from_volume  = true
  disk_size         = 100
  availability_zone = "N1"
  security_group_name = "all-open-az1"
  admin_username    = "clouduser"
  admin_password    = var.vm_admin_password
  #keypair_name = "ashok"
  description       = "Example web server instance"

  tags = {
    Environment = "example"
    Role        = "web-server"
  }
}

# The same instance authenticated with an SSH keypair instead:
#
# resource "airtelcloud_vm" "web_server_keypair" {
#   ...
#   keypair_name = "ashok"
# }

# Create a Windows VM with backup enabled (alternative configuration)
resource "airtelcloud_vm" "windows_server" {
  instance_name     = "${var.resource_prefix}-win-server"
  os_type           = "windows"
  flavor_name       = "ccd.Large"
  image_name        = "WIN2K22_BYOL_Jul2026"
  vpc_name          = "copper-vpc01"
  subnet_name       = "sub1"
  boot_from_volume  = true
  disk_size         = 200
  availability_zone = "N1"
  security_group_name = "all-open-az1"
  description       = "Example Windows server with backup"

#  enable_backup    = true
#  protection_plan  = "daily-backup-plan"
#  start_date       = "2026-04-01"
#  start_time       = "02:00"
#
  tags = {
    Environment = "example"
    Role        = "windows-server"
  }
}

resource "airtelcloud_volume" "data_volume" {
  name = "${var.resource_prefix}-storage-volume"
	size = 50
	type = "n1_wkld_ntp02_1iops_backend"
	availability_zone = "N1"
	#compute_id = airtelcloud_vm.web_server.id
	vpc_id = "66f0fd26-9362-4f52-8d41-10de059e88fb"
	subnet_id = "5d5d6aa1-5f24-41c7-bb5b-5c12f7f41c11"
	is_encrypted = false
	bootable = false
	depends_on = [airtelcloud_vm.web_server]
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
