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

variable "monitoring_api_token" {
  description = "Monitoring provider API token"
  type        = string
  sensitive   = true
}

resource "airtelcloud_monitoring_provider" "compute" {
  name               = "south-prometheus-compute"
  resource_type      = "compute"
  provider_type      = "prometheus"
  dashboard_auth_url = "https://grafana.example.com"
  api_key_name       = "terraform-provider"
  api_token          = var.monitoring_api_token
  data_source_uid    = "prom-uid-1"
  data_source_id     = "1"
  az_name            = "S1"
  active             = true
  description        = "Prometheus for compute in S1"
}

output "provider_id" {
  description = "ID of the monitoring provider"
  value       = airtelcloud_monitoring_provider.compute.id
}
