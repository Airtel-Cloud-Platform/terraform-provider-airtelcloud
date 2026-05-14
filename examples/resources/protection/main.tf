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

# Create a protection plan with daily schedule and 30-day retention
resource "airtelcloud_protection_plan" "daily" {
  name           = "${var.resource_prefix}-daily-backup"
  description    = "Daily backup with 30-day retention"
  retention      = 1
  retention_unit = "DAYS"
  recurrence     = 86400
  selector_key   = "AZ"
  selector_value = "S1"
  subnet_id      = "35df162d-5211-4d58-84ed-6a499626949c"
}

# Create a protection policy for a compute instance
resource "airtelcloud_protection" "web_server" {
  name             = "${var.resource_prefix}-web-backup"
  description      = "Backup policy for web server"
  compute_id       = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"
  protection_plan  = airtelcloud_protection_plan.daily.name
  enable_scheduler = "true"
  start_date       = "2026-04-01"
  start_time       = "02:00"
}

# Weekly backup plan (alternative configuration)
#resource "airtelcloud_protection_plan" "weekly" {
#  name           = "${var.resource_prefix}-weekly-backup"
#  description    = "Weekly backup with 12-week retention"
#  retention      = 12
#  retention_unit = "WEEKS"
#  recurrence     = 604800
#  selector_key   = "AZ"
#  selector_value = "S1"
#}

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
