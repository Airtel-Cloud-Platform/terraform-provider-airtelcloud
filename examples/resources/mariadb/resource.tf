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

variable "mariadb_password" {
  description = "Admin password for the MariaDB cluster"
  type        = string
  sensitive   = true
}

variable "resource_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "tft"
}

resource "airtelcloud_mariadb" "app" {
  cluster_name      = "${var.resource_prefix}-maria"
  description       = "Application MariaDB cluster"
  version           = "10.6"
  database_name     = "app"
  mariadb_username  = "dbadmin"
  password          = var.mariadb_password
  compute_size      = "db.maria.uhper.ccs.xlarge"
  storage_size      = 200
  availability_zone = "S1"
  high_availability = true
  num_replicas      = 1
  labels            = ["terraform"]

  backup = {
    enabled         = true
    protection_plan = "weekly-full-daily-incr"
    schedule_time   = "02:00"
    schedule_day    = "Monday"
  }

  security_group = {
    allowed_ips = ["192.168.1.0/24"]
  }
}

output "mariadb_id" {
  description = "UUID of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.id
}

output "mariadb_cluster_name" {
  description = "Cluster name of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.cluster_name
}

output "mariadb_version" {
  description = "MariaDB engine version"
  value       = airtelcloud_mariadb.app.version
}

output "mariadb_database_name" {
  description = "Initial database name"
  value       = airtelcloud_mariadb.app.database_name
}

output "mariadb_username" {
  description = "Admin username for the MariaDB cluster"
  value       = airtelcloud_mariadb.app.mariadb_username
}

output "mariadb_status" {
  description = "Status of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.status
}

output "mariadb_topology" {
  description = "Resolved topology of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.topology
}

output "mariadb_num_replicas" {
  description = "Number of slave replicas"
  value       = airtelcloud_mariadb.app.num_replicas
}

output "mariadb_availability_zone" {
  description = "Availability zone of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.availability_zone
}

output "mariadb_port" {
  description = "Database port of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.port
}

output "mariadb_created_at" {
  description = "Creation timestamp of the MariaDB cluster"
  value       = airtelcloud_mariadb.app.created_at
}

output "mariadb_connection_string" {
  description = "Connection string for the MariaDB cluster"
  value       = airtelcloud_mariadb.app.connection_string
  sensitive   = true
}
