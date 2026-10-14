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

resource "airtelcloud_monitoring_dashboard" "compute" {
  name          = "compute-default"
  resource_type = "compute"
  active        = true
  description   = "Default compute dashboard"

  panels = [
    {
      title       = "CPU Utilization"
      chart_type  = "line"
      description = "CPU over time"
      x           = 0
      y           = 0
      w           = 12
      h           = 4

      queries = [
        {
          query_template = "avg(rate(cpu_usage{instance=\"{{instance}}\"}[5m]))"
          query_type     = "prometheus"
          ref_id         = "A"
          placeholder_config = {
            instance = {
              source = "resource"
              field  = "name"
            }
          }
        }
      ]
    }
  ]
}

output "dashboard_id" {
  description = "ID of the monitoring dashboard"
  value       = airtelcloud_monitoring_dashboard.compute.id
}
