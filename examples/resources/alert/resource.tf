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
  description = "Organization for the resources"
  type        = string
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
}

resource "airtelcloud_alert" "cpu" {
  name                = "high-cpu-vm1"
  description         = "CPU above 80%"
  resource_type_name  = "compute"
  metric_type         = "cpu"
  metric_name         = "cpu_utilization"
  threshold           = 80
  comparison_operator = ">"
  evaluation_periods  = "5"
  datapoints_to_alarm = "3"
  enabled             = true
  severity            = 2
  alert_emails        = ["ops@example.com"]

  resources_info = [
    {
      resource_id          = "vm-uuid-1"
      resource_name        = "app-vm-1"
      resource_ip          = "10.0.0.10"
      provider_instance_id = "i-123"
    }
  ]
}
