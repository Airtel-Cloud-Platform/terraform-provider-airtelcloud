terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.1.1"
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

# Allow web traffic (HTTP + HTTPS) from any source through a public IP
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_id      = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"
  display_name      = "web-traffic"
  source            = "any"
  services          = ["HTTP", "HTTPS"]
  action            = "accept"
  target_vip        = "10.1.99.172"
  public_ip         = "203.0.113.10"
  availability_zone = "S1"
}

# Allow SSH access from a specific management IP only
resource "airtelcloud_public_ip_policy_rule" "ssh_mgmt" {
  public_ip_id      = "b603ccb5-fe35-4ddb-9a7c-2e966a9425c2"
  display_name      = "ssh-mgmt"
  source            = "192.168.100.5"
  services          = ["SSH"]
  action            = "accept"
  target_vip        = "10.1.99.172"
  public_ip         = "203.0.113.10"
  availability_zone = "S1"
}

output "web_rule_id" {
  description = "ID of the web traffic policy rule"
  value       = airtelcloud_public_ip_policy_rule.web_traffic.id
}

output "web_rule_state" {
  description = "State of the web traffic policy rule"
  value       = airtelcloud_public_ip_policy_rule.web_traffic.state
}
