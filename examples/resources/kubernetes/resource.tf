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
  name                = "test-kms"
  description         = "test"
  k8s_version         = "v1.33.7"
  cni_name            = "calico"
  cni_version         = "v3.30.6"
  master_nodes        = 3
  availability_zone   = "S1"
  control_plane_provider = "Kamaji"

  vpc_name            = var.vpc_name

  node_pools = [
    {
      name              = "md0"
      host_group        = "ccd.xLarge"
      group_name        = "Compute dense"
      subnet_name       = var.subnet_name
      os_distribution   = "Ubuntu"
      count             = 1
    }
  ]
}

output "kubernetes_id" {
  value = airtelcloud_kubernetes.cluster.id
}

output "kubernetes_state" {
  value = airtelcloud_kubernetes.cluster.state
}
