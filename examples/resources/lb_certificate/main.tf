terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.1.9"
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

variable "ssl_cert_pem" {
  description = "SSL certificate in PEM format"
  type        = string
  sensitive   = true
}

variable "ssl_key_pem" {
  description = "SSL private key in PEM format"
  type        = string
  sensitive   = true
}

variable "ca_cert_pem" {
  description = "CA certificate in PEM format (optional)"
  type        = string
  sensitive   = true
  default     = ""
}

# Upload an SSL certificate to a Load Balancer Service
resource "airtelcloud_lb_certificate" "example" {
  lb_service_id  = "ac238c5b-6334-49b9-b9c0-decc0aaf63d6"
  name           = "my-ssl-cert"
  ssl_cert       = var.ssl_cert_pem
  ssl_private_key = var.ssl_key_pem
  ca_cert        = var.ca_cert_pem != "" ? var.ca_cert_pem : null
}

output "certificate_id" {
  description = "ID of the uploaded certificate"
  value       = airtelcloud_lb_certificate.example.id
}

output "certificate_status" {
  description = "Status of the certificate"
  value       = airtelcloud_lb_certificate.example.status
}
