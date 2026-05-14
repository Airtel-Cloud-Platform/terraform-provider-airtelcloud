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

# Create a VPC peering connection
resource "airtelcloud_vpc_peering" "example" {
  name            = "${var.resource_prefix}-vpc-peering"
  description     = "VPC peering between two networks"
  vpc_source_name = "perftest-cell1-vpc1"
  vpc_target_name = "perftest-cell1-vpc2"
  az              = "S2"
  region          = "south"

  is_pcl_enabled = false

  # allowed_subnet_list = [
  #   "35df162d-5211-4d58-84ed-6a499626949c",
  # ]
  # blocked_subnet_list = [
  #   "35df162d-5211-4d58-84ed-6a499626949c",
  # ]
  timeouts {
    create = "25m"
    delete = "15m"
  }
}

# VPC peering using IDs instead of names
#resource "airtelcloud_vpc_peering" "by_id" {
#  name          = "${var.resource_prefix}-vpc-peering-alt"
#  description   = "Peering between production and staging VPCs"
#  vpc_source_id = "source-vpc-id"
#  vpc_target_id = "target-vpc-id"
#  az            = "S2"
#  region        = "south"
#
#  is_pcl_enabled = false
#
#  timeouts {
#    create = "25m"
#    delete = "15m"
#  }
#}
