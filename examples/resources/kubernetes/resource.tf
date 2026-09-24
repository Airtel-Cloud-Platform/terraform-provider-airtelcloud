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
  description = "Organization (domain) for the resources"
  type        = string
}

variable "project_name" {
  description = "Project for the resources"
  type        = string
}

variable "subnet_name" {
  description = "Subnet name for the worker node pool"
  type        = string
}

variable "vpc_name" {
  description = "VPC that contains the subnet. Optional if the subnet name is unique."
  type        = string
  default     = ""
}

resource "airtelcloud_kubernetes" "cluster" {
  name                   = "test-kms"
  description            = "test"
  kubernetes_version     = "v1.33.7"
  networking_name        = "calico"
  networking_version     = "v3.30.6"
  availability_zone      = "S1"
  os_distribution        = "Ubuntu"
  control_plane_provider = "Kamaji"

  vpc_name = var.vpc_name

  node_pools = [
    {
      flavor      = "ccd.xLarge"
      flavor_type = "Compute dense"
      subnet_name = var.subnet_name
      count       = 0
      autoscaling = {
        enabled   = true
        max_nodes = 2
      }
      labels = [
        {
          key   = "test"
          value = "test"
        }
      ]
      annotations = [
        {
          key   = "new"
          value = "new"
        }
      ]
      taints = [
        {
          key    = "test"
          value  = "test"
          effect = "PreferNoSchedule"
        }
      ]
    }
  ]
}

output "kubernetes_id" {
  value = airtelcloud_kubernetes.cluster.id
}

output "kubernetes_state" {
  value = airtelcloud_kubernetes.cluster.state
}
