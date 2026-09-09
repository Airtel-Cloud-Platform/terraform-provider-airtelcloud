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
  description = "organization for the resources"
  type        = string
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
}

variable "postgres_password" {
  description = "Admin password for the PostgreSQL cluster"
  type        = string
  sensitive   = true
}

variable "resource_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "tft"
}

resource "airtelcloud_postgres" "app" {
  cluster_name      = "${var.resource_prefix}-pg"
  description       = "Application PostgreSQL cluster"
  version           = "17"
  database_name     = "app"
  postgres_username = "dbadmin"
  password          = var.postgres_password
  compute_size      = "db.postgres.uhper.ccs.xlarge"
  storage_size      = 200
  availability_zone = "S1"
  high_availability = true
  num_replicas      = 1
  pg_extensions     = ["pgvector"]
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

output "postgres_id" {
  description = "UUID of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.id
}

output "postgres_cluster_name" {
  description = "Cluster name of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.cluster_name
}

output "postgres_version" {
  description = "PostgreSQL major version"
  value       = airtelcloud_postgres.app.version
}

output "postgres_database_name" {
  description = "Initial database name"
  value       = airtelcloud_postgres.app.database_name
}

output "postgres_username" {
  description = "Admin username for the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.postgres_username
}

output "postgres_status" {
  description = "Status of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.status
}

output "postgres_topology" {
  description = "Resolved topology of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.topology
}

output "postgres_num_replicas" {
  description = "Number of standby replicas"
  value       = airtelcloud_postgres.app.num_replicas
}

output "postgres_availability_zone" {
  description = "Availability zone of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.availability_zone
}

output "postgres_created_at" {
  description = "Creation timestamp of the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.created_at
}

output "postgres_connection_string" {
  description = "Connection string for the PostgreSQL cluster"
  value       = airtelcloud_postgres.app.connection_string
}
