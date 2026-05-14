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

# Create a load balancer service
resource "airtelcloud_lb_service" "example" {
  name        = "${var.resource_prefix}-lb-service"
  description = "Example load balancer service"
  network_id  = "35df162d-5211-4d58-84ed-6a499626949c"
  vpc_id      = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  vpc_name    = "perftest-cell1-vpc1"
  ha          = false

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Allocate a VIP for the load balancer
resource "airtelcloud_lb_vip" "example" {
  lb_service_id = airtelcloud_lb_service.example.id
}

# Upload an SSL certificate for HTTPS termination
# Uncomment and provide cert/key file paths to use
#resource "airtelcloud_lb_certificate" "example" {
#  lb_service_id   = airtelcloud_lb_service.example.id
#  name            = "${var.resource_prefix}-ssl-cert"
#  ssl_cert        = file("${path.module}/cert.pem")
#  ssl_private_key = file("${path.module}/key.pem")
#}

# Create an HTTP virtual server with two backend nodes
resource "airtelcloud_lb_virtual_server" "http" {
  lb_service_id     = airtelcloud_lb_service.example.id
  name              = "${var.resource_prefix}-http-vs"
  vip_port_id       = tonumber(airtelcloud_lb_vip.example.id)
  protocol          = "HTTP"
  port              = 80
  routing_algorithm = "ROUND_ROBIN"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  interval          = 30

  nodes = [
    {
      compute_id = 101
      compute_ip = "192.168.1.10"
      port       = 8080
      weight     = 50
    },
    {
      compute_id = 102
      compute_ip = "192.168.1.11"
      port       = 8080
      weight     = 50
    },
  ]

  persistence_enabled = true
  persistence_type    = "source_ip"
  x_forwarded_for     = true

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

# Output load balancer details
output "lb_service_id" {
  description = "ID of the load balancer service"
  value       = airtelcloud_lb_service.example.id
}

output "lb_service_status" {
  description = "Status of the load balancer service"
  value       = airtelcloud_lb_service.example.status
}

output "lb_vip_id" {
  description = "ID of the load balancer VIP"
  value       = airtelcloud_lb_vip.example.id
}

output "lb_vip_public_ip" {
  description = "Public IP of the load balancer VIP"
  value       = airtelcloud_lb_vip.example.public_ip
}

output "http_virtual_server_id" {
  description = "ID of the HTTP virtual server"
  value       = airtelcloud_lb_virtual_server.http.id
}

output "http_virtual_server_vip" {
  description = "VIP address of the HTTP virtual server"
  value       = airtelcloud_lb_virtual_server.http.vip
}
