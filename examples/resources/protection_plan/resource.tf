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
  description = "Existing VM instance_name used to read AZ and subnet"
  type        = string
}

# Create a daily backup protection plan with 30-day retention.
# Note: the API does not support deletion of protection plans.
# Destroying this resource only removes it from Terraform state.
resource "airtelcloud_protection_plan" "daily" {
  name              = "${var.resource_prefix}-backup"
  description       = "Daily backup with 30-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "daily"
  retention         = 30
}

# Weekly backup plan
resource "airtelcloud_protection_plan" "weekly" {
  name              = "${var.resource_prefix}-backup"
  description       = "Weekly backup with 84-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "weekly"
  retention         = 84
}

output "daily_plan_id" {
  description = "ID of the daily protection plan"
  value       = airtelcloud_protection_plan.daily.id
}

output "weekly_plan_id" {
  description = "ID of the weekly protection plan"
  value       = airtelcloud_protection_plan.weekly.id
}
