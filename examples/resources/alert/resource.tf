terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.3.0"
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
  type      = string
  sensitive = true
}

variable "airtel_api_secret" {
  type      = string
  sensitive = true
}

variable "organization" { type = string }
variable "project_name" { type = string }

variable "alert_email" {
  type    = string
  default = "ops@example.com"
}

resource "airtelcloud_alert" "memory" {
  name          = "high-memory-vm1"
  description   = "Memory above 80%"
  emails        = [var.alert_email]
  resource_type = "Virtual Machine"
  severity      = 0
  metric_type   = "Memory Total"
  function      = "Average"
  condition     = "Greater than, equal to (>=)"
  value         = 80
}

output "alert_id" {
  value = airtelcloud_alert.memory.id
}

output "alert_state" {
  value = airtelcloud_alert.memory.state
}
