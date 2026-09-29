terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.7"
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

# Step 3 of 3: add policy rules. The public IP must already be attached;
# target_vip, public_ip, and availability_zone are read from public_ip_name.

# Allow web traffic (HTTP + HTTPS) from any source
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_name = "my-vm-public-ip"
  rule_name   = "web-traffic"
  source         = "any"
  services       = ["HTTP", "HTTPS"]
  action         = "accept"
}

# Allow SSH access from a specific management IP only
resource "airtelcloud_public_ip_policy_rule" "ssh_mgmt" {
  public_ip_name = "my-vm-public-ip"
  rule_name   = "ssh-mgmt"
  source         = "192.168.100.5"
  services       = ["SSH"]
  action         = "accept"
}

# API-shaped example using source_config and service_config
resource "airtelcloud_public_ip_policy_rule" "temporal_rule" {
  public_ip_name = "my-vm-public-ip"
  rule_name   = "temporal-rule"
  action         = "accept"
  resource_type  = "ipam"
  revision_note  = "creating Policy"

  source_config = [
    {
      create_new  = false
      ip_cidr     = "182.77.78.18/32"
      source_type = "ip_cidr"
    },
    {
      source_type = "geographic"
      geographic = {
        country_code = "IN"
        country_name = "India"
      }
    }
  ]

  service_config = [
    {
      create_new = false
      name       = "tcp-443-443"
      is_default = false
    },
    {
      create_new = false
      name       = "tcp-5601-5601"
      is_default = false
    },
    {
      create_new = false
      name       = "SSH"
      is_default = false
    },
    {
      create_new = false
      name       = "DNS"
      is_default = false
    },
    {
      create_new = false
      name       = "HTTP"
      is_default = false
    },
    {
      create_new = false
      name       = "HTTPS"
      is_default = false
    }
  ]
}

output "web_rule_id" {
  description = "ID of the web traffic policy rule"
  value       = airtelcloud_public_ip_policy_rule.web_traffic.id
}

output "web_rule_state" {
  description = "State of the web traffic policy rule"
  value       = airtelcloud_public_ip_policy_rule.web_traffic.state
}

output "rdp_rule_id" {
  description = "ID of the CIDR + geographic policy rule"
  value       = airtelcloud_public_ip_policy_rule.temporal_rule.id
}
