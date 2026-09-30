terraform {
  required_providers {
    airtelcloud = {
      source  = "Airtel-Cloud-Platform/airtelcloud"
      version = "1.2.6"
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

variable "mssql_password" {
  description = "Admin password for the MSSQL instance"
  type        = string
  sensitive   = true
}

resource "airtelcloud_mssql" "app" {
  name              = "hello"
  description       = "Application MSSQL instance"
  database_edition  = "web"
  collation         = "SQL_Latin1_General_CP1_CI_AS"
  features          = ["SQLENGINE"]
  compute_size      = "db.mssqlweb.uhper.cci.xlarge"
  storage_size      = 200
  storage_type      = "High Performance"
  availability_zone = "S1"
  vpc               = "copper-vpc1"
  subnet            = "vlan-dbaas-91"
  db_name           = "AppDb"
  mssql_username    = "DbAdmin"
  password          = var.mssql_password

  backup = {
    protection_plan = "CH26_AZ1_PaaS_MSSQLIP_DAILY_FULL_2100_PP"
  }

  security_group = {
    allowed_ips = ["192.168.1.0/24"]
  }
}

output "mssql_id" {
  description = "UUID of the MSSQL instance"
  value       = airtelcloud_mssql.app.id
}

output "mssql_status" {
  description = "Status of the MSSQL instance"
  value       = airtelcloud_mssql.app.status
}

output "mssql_version" {
  description = "Version resolved from the selected edition"
  value       = airtelcloud_mssql.app.version
}
