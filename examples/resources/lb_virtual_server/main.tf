terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.1"
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

# Create a Layer-7 HTTP virtual server with two backend nodes.
# Each node can be referenced by compute_name (resolved at apply time) or
# compute_id (the numeric compute instance ID). Set exactly one per node.
resource "airtelcloud_lb_virtual_server" "http" {
  lb_service_id     = "ac238c5b-6334-49b9-b9c0-decc0aaf63d6"
  name              = "${var.resource_prefix}-http-vs"
  vip               = "10.1.1.100"
  protocol          = "HTTP"
  port              = 80
  routing_algorithm = "ROUND_ROBIN"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  interval          = 30
  monitor_name      = "http-monitor"
  monitor_protocol  = "HTTP"
  x_forwarded_for   = true

  nodes = [
    {
      compute_name = "web-server-10"
      compute_ip   = "10.1.1.10"
      port         = 8080
      weight       = 10
    },
    {
      compute_name = "web-server-2"
      compute_ip   = "10.1.1.11"
      port         = 8080
      weight       = 10
    },
  ]

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Create an HTTPS virtual server with SSL termination
# resource "airtelcloud_lb_virtual_server" "https" {
#   lb_service_id     = "35df162d-5211-4d58-84ed-6a499626949c"
#   name              = "${var.resource_prefix}-https-vs"
#   vip_port_id       = 12345
#   protocol          = "HTTPS"
#   port              = 443
#   routing_algorithm = "LEAST_CONNECTIONS"
#   vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
#   interval          = 30
#   certificate_id    = airtelcloud_lb_certificate.example.id
#   redirect_https    = false
#
#   persistence_enabled = true
#   persistence_type    = "cookie"
#
#   nodes = [
#     {
#       compute_name = "web-server-1"
#       compute_ip   = "10.1.1.10"
#       port         = 8443
#       weight       = 10
#     },
#   ]
# }

output "virtual_server_id" {
  description = "ID of the virtual server"
  value       = airtelcloud_lb_virtual_server.http.id
}

output "virtual_server_vip" {
  description = "VIP address of the virtual server"
  value       = airtelcloud_lb_virtual_server.http.vip
}

output "virtual_server_status" {
  description = "Status of the virtual server"
  value       = airtelcloud_lb_virtual_server.http.status
}
