terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.0"
    }
  }
}

provider "airtelcloud" {
  api_endpoint = var.api_endpoint
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = var.region
  organization = var.organization
  project_name = var.project_name
}

variable "api_endpoint" {
  description = "Airtel Cloud API endpoint"
  type        = string
  default     = "https://south.cloud.airtel.in"
}

variable "region" {
  description = "Airtel Cloud region"
  type        = string
  default     = "south"
}

variable "airtel_api_key" {
  description = "Airtel Cloud API key"
  type        = string
  default     = "55e0500c-4e2b-4313-9580-30cdfb3713f9"
  sensitive   = true
}

variable "airtel_api_secret" {
  description = "Airtel Cloud API secret"
  type        = string
  default     = "RvmcpJ0mixdxpNjALLa2KYQ/x5SXK6GXKhGf47crcKb"
  sensitive   = true
}

variable "organization" {
  description = "Organization for the resources"
  type        = string
  default     = "coredemo"
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
  default     = "compass"
}

variable "resource_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "tfw"
}

variable "availability_zone" {
  description = "Availability zone where VM/Public IP should be created"
  type        = string
  default     = "S1"
}

variable "vpc_name" {
  description = "Existing VPC name where VM will be created"
  type        = string
  default     = "compass-vpc1"
}

variable "subnet_name" {
  description = "Existing subnet name where VM will be created"
  type        = string
  default     = "compass-vpc1-subnet1-az1"
}

variable "security_group_name" {
  description = "Existing security group name to attach to the VM"
  type        = string
  default     = "ss-ubuntu"

}

variable "protection_subnet_id" {
  description = "Subnet ID required by backup/protection plan API"
  type        = string
  default     = "682bac6e-9ef0-4c12-affb-d9a95a7621af"
}

variable "flavor_name" {
  description = "Windows VM flavor name"
  type        = string
  default     = "ccd.Large"
}

variable "image_name" {
  description = "Windows VM image name"
  type        = string
  default     = "WIN2K22_BYOL_Jul2026"
}

variable "rdp_source_cidr" {
  description = "Allowed source CIDR for RDP access through public IP policy"
  type        = string
  default     = "122.161.51.26/32"
}

variable "create_public_ip_and_policy" {
  description = "Create Public IP and policy along with the VM. Keep this value unchanged between apply and destroy runs."
  type        = bool
  default     = true
}

# Usage notes:
# - Keep create_public_ip_and_policy unchanged between apply and destroy.
# - Apply creates resources in this order: VM -> Public IP -> Public IP policy rule.
# - Destroy removes resources in reverse order: Public IP policy rule -> Public IP -> VM.
#
# Example:
# terraform apply -auto-approve
# terraform destroy -auto-approve

# Protection plan used by VM backup settings.
#resource "airtelcloud_protection_plan" "windows_daily" {
#  name           = "${var.resource_prefix}-windows-daily-backup"
#  description    = "Daily backup plan for Windows VM example"
#  retention      = 7
#  retention_unit = "DAYS"
#  recurrence     = 86400
#  selector_key   = "AZ"
#  selector_value = var.availability_zone
#  subnet_id      = var.protection_subnet_id
#}

resource "airtelcloud_vm" "windows_server" {
  instance_name       = "${var.resource_prefix}-win-server-12"
  os_type             = "windows"
  flavor_name         = var.flavor_name
  image_name          = var.image_name
  vpc_name            = var.vpc_name
  subnet_name         = var.subnet_name
  security_group_name = var.security_group_name
  availability_zone   = var.availability_zone
  boot_from_volume    = true
  disk_size           = 200
  description         = "Windows VM with backup enabled"

  enable_backup = true
  # Use the protection plan UUID/id, not the plan name.
  protection_plan = "S1-COREDEMO-COMPASS-TFW-WINDOWS-DAILY-BACKUP-BKP-PP"
  start_date      = "2026-08-14"
  start_time      = "11:20 AM"

  labels = ["example", "windows", "backup"]
}

resource "airtelcloud_public_ip" "windows_vm_nat" {
  count             = var.create_public_ip_and_policy ? 1 : 0
  object_name       = "${var.resource_prefix}-windows-public-ip-44"
  vip               = airtelcloud_vm.windows_server.private_ip
  availability_zone = var.availability_zone

  depends_on = [airtelcloud_vm.windows_server]

  timeouts {
    create = "10m"
    delete = "5m"
  }
}

resource "airtelcloud_public_ip_policy_rule" "allow_rdp" {
  count        = var.create_public_ip_and_policy ? 1 : 0
  public_ip_id = airtelcloud_public_ip.windows_vm_nat[0].id
  display_name = "allow-rdp"

  depends_on = [airtelcloud_public_ip.windows_vm_nat]

  source_config = [
    {
      create_new  = true
      ip_cidr     = var.rdp_source_cidr
      source_type = "ip_cidr"
    }
  ]

  services          = ["RDP"]
  action            = "accept"
  target_vip        = airtelcloud_public_ip.windows_vm_nat[0].vip
  public_ip         = airtelcloud_public_ip.windows_vm_nat[0].public_ip
  availability_zone = var.availability_zone
}

output "windows_vm_id" {
  description = "ID of the Windows VM"
  value       = airtelcloud_vm.windows_server.id
}

output "windows_vm_private_ip" {
  description = "Private IP of the Windows VM"
  value       = airtelcloud_vm.windows_server.private_ip
}

output "windows_public_ip_id" {
  description = "ID of the allocated public IP"
  value       = try(airtelcloud_public_ip.windows_vm_nat[0].id, null)
}

output "windows_public_ip_address" {
  description = "Public IP address allocated to the Windows VM"
  value       = try(airtelcloud_public_ip.windows_vm_nat[0].public_ip, null)
}

output "rdp_policy_rule_id" {
  description = "Policy rule ID for RDP access"
  value       = try(airtelcloud_public_ip_policy_rule.allow_rdp[0].id, null)
}