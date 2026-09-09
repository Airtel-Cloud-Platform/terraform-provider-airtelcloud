terraform {
  required_providers {
    airtelcloud = {
      source = "Airtel-Cloud-Platform/airtelcloud"
    }
  }
}

variable "airtel_api_key" {
  type      = string
  sensitive = true
}

variable "airtel_api_secret" {
  type      = string
  sensitive = true
}

variable "organization" {
  type = string
}

variable "project_name" {
  type = string
}

provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south"
  organization = var.organization
  project_name = var.project_name
}
