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

# Create a forward DNS zone
resource "airtelcloud_dns_zone" "tfas" {
  zone_name   = "${var.resource_prefix}-domain.com"
  zone_type   = "forward"
  description = "Example DNS zone for demonstration"
}

# Create another DNS zone for internal services
resource "airtelcloud_dns_zone" "internal" {
  zone_name   = "internal.${var.resource_prefix}-domain.com"
  zone_type   = "forward"
  description = "Internal services DNS zone"
}

# Output DNS zone details
output "example_zone_id" {
  description = "ID of the tfas DNS zone"
  value       = airtelcloud_dns_zone.tfas.id
}

output "example_zone_name" {
  description = "Name of the tfas DNS zone"
  value       = airtelcloud_dns_zone.tfas.zone_name
}

output "internal_zone_id" {
  description = "ID of the internal DNS zone"
  value       = airtelcloud_dns_zone.internal.id
}
