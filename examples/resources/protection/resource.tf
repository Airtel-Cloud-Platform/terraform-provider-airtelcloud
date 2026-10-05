terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.8"
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

variable "vm_name" {
  description = "Existing VM instance_name"
  type        = string
}

# Create a protection plan with daily schedule and 30-day retention
resource "airtelcloud_protection_plan" "daily" {
  name              = "${var.resource_prefix}-backup"
  description       = "Daily backup with 30-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "daily"
  retention         = 1
}

# Create a protection policy for a compute instance
resource "airtelcloud_protection" "web_server" {
  description     = "Backup policy for web server"
  vm_name         = var.vm_name
  protection_plan = airtelcloud_protection_plan.daily.id
  start_date      = "10/06/2026"
  start_time      = "02:00 AM"
}

# Weekly backup plan (alternative configuration)
resource "airtelcloud_protection_plan" "weekly" {
  name              = "${var.resource_prefix}-backup"
  description       = "Weekly backup with 84-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "weekly"
  retention         = 84
}

# Output protection details
output "protection_plan_id" {
  description = "ID of the daily protection plan"
  value       = airtelcloud_protection_plan.daily.id
}

output "protection_plan_name" {
  description = "Name of the daily protection plan"
  value       = airtelcloud_protection_plan.daily.name
}

output "protection_id" {
  description = "ID of the web server protection policy"
  value       = airtelcloud_protection.web_server.id
}

output "protection_status" {
  description = "Status of the web server protection policy"
  value       = airtelcloud_protection.web_server.status
}
