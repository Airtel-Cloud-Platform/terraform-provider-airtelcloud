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

# Create a security group to attach rules to
resource "airtelcloud_security_group" "web" {
  security_group_name = "${var.resource_prefix}-sg-http-servers"
  availability_zone   = "S2"
}

# Allow SSH from internal network
resource "airtelcloud_security_group_rule" "ssh" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "22"
  port_range_max    = "22"
  remote_ip_prefix  = "10.0.0.0/8"
  ethertype         = "IPv4"
  description       = "Allow SSH from internal network"
}

# Allow HTTP traffic
#resource "airtelcloud_security_group_rule" "http" {
#  security_group_id = airtelcloud_security_group.web.id
#  direction         = "ingress"
#  protocol          = "tcp"
#  port_range_min    = "8080"
#  port_range_max    = "8080"
#  remote_ip_prefix  = "0.0.0.0/0"
#  ethertype         = "IPv4"
#  description       = "Allow HTTP"
#}

# Output rule details
output "ssh_rule_id" {
  description = "ID of the SSH security group rule"
  value       = airtelcloud_security_group_rule.ssh.id
}

output "ssh_rule_uuid" {
  description = "UUID of the SSH security group rule"
  value       = airtelcloud_security_group_rule.ssh.uuid
}

#output "http_rule_id" {
#  description = "ID of the HTTP security group rule"
#  value       = airtelcloud_security_group_rule.http.id
#}
