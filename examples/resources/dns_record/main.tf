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

# Create a DNS zone first
resource "airtelcloud_dns_zone" "tfas" {
  zone_name   = "${var.resource_prefix}-domain.com"
  zone_type   = "forward"
  description = "Example DNS zone"
}

# Create an A record for the apex domain
resource "airtelcloud_dns_record" "apex" {
  zone_id     = airtelcloud_dns_zone.tfas.id
  owner       = "apex"
  record_type = "A"
  data        = "192.168.1.10"
  ttl         = 300
  description = "Apex domain A record"
}

# Create a TXT record for SPF
resource "airtelcloud_dns_record" "spf" {
  zone_id     = airtelcloud_dns_zone.tfas.id
  owner       = "spf"
  record_type = "TXT"
  data        = "v=spf1 include:_spf.tfas.com ~all"
  ttl         = 3600
  description = "SPF record for email authentication"
}

# Create an AAAA record for IPv6
resource "airtelcloud_dns_record" "ipv6" {
  zone_id     = airtelcloud_dns_zone.tfas.id
  owner       = "www"
  record_type = "AAAA"
  data        = "2001:db8::1"
  ttl         = 300
  description = "IPv6 address for www"
}

# Create an A record pointing to web server
#resource "airtelcloud_dns_record" "web" {
#  zone_id     = airtelcloud_dns_zone.tfas.id
#  owner       = "www"
#  record_type = "A"
#  data        = "192.168.1.10"
#  ttl         = 300
#  description = "Web server A record"
#}

# Create a CNAME record
#resource "airtelcloud_dns_record" "blog" {
#  zone_id     = airtelcloud_dns_zone.tfas.id
#  owner       = "blog"
#  record_type = "CNAME"
#  data        = "www.tfas.com."
#  ttl         = 3600
#  description = "Blog CNAME pointing to www"
#}

# Create an MX record for email
#resource "airtelcloud_dns_record" "mail" {
#  zone_id     = airtelcloud_dns_zone.tfas.id
#  owner       = "@"
#  record_type = "MX"
#  data        = "mail.tfas.com."
#  ttl         = 3600
#  preference  = 10
#  description = "Primary mail server"
#}

# Output DNS record details
output "apex_record_id" {
  description = "ID of the apex DNS record"
  value       = airtelcloud_dns_record.apex.id
}

output "spf_record_id" {
  description = "ID of the SPF DNS record"
  value       = airtelcloud_dns_record.spf.id
}

output "ipv6_record_id" {
  description = "ID of the IPv6 DNS record"
  value       = airtelcloud_dns_record.ipv6.id
}

#output "web_record_id" {
#  description = "ID of the web server DNS record"
#  value       = airtelcloud_dns_record.web.id
#}

#output "mail_record_id" {
#  description = "ID of the mail MX record"
#  value       = airtelcloud_dns_record.mail.id
#}
