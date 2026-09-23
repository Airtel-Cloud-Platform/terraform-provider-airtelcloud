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

variable "resource_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "tft"
}

resource "airtelcloud_asg" "example" {
  name              = "${var.resource_prefix}-asg"
  vpc_name          = "copper-vpc1"
  subnet            = "subnet1213"
  availability_zone = "S1"

  flavor_name          = "ccd.Large"
  image_name           = "Ubuntu22_04_Sep2026"
  security_group_names = ["networksg-1"]
  keypair_name         = "keypair-1"

  minimum_size  = 1
  maximum_size  = 3
  desired_count = 1

  scale_up_step_size   = 1
  scale_down_step_size = 1
  scaling_interval     = 2
  cooloff_period       = 10
  termination_policy   = "oldest"
  drain_period         = 0

  scaleup_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 60
  }]

  scaledown_rules = [{
    metric_type      = "cpu"
    aggregation_type = "avg"
    target_value     = 30
  }]

  labels = ["app", "production"]
}

# Alternative: attach the group to an existing load balancer. load_balancer_name
# enables vs_config on create; the VIP is resolved from the LB service network,
# not the ASG subnet. Uncomment this block (and comment the resource above) to
# create the LB-backed group instead.
# resource "airtelcloud_asg" "with_lb" {
#   name              = "${var.resource_prefix}-asg-lb"
#   vpc_name          = "copper-vpc1"
#   subnet            = "subnet1213"
#   availability_zone = "S1"
#
#   flavor_name          = "ccd.Large"
#   image_name           = "Ubuntu22_04_Sep2026"
#   security_group_names = ["networksg-1"]
#
#   minimum_size  = 1
#   maximum_size  = 3
#   desired_count = 1
#
#   scale_up_step_size   = 1
#   scale_down_step_size = 1
#   scaling_interval     = 2
#   cooloff_period       = 10
#   termination_policy   = "oldest"
#   drain_period         = 0
#
#   scaleup_rules = [{
#     metric_type      = "cpu"
#     aggregation_type = "avg"
#     target_value     = 60
#   }]
#
#   scaledown_rules = [{
#     metric_type      = "cpu"
#     aggregation_type = "avg"
#     target_value     = 30
#   }]
#
#   load_balancer_name    = "south-lb"
#   host_name             = "${var.resource_prefix}-asg-vs"
#   vip                   = "10.29.16.10"
#   protocol              = "TCP"
#   port                  = 1234
#   routing_algorithm     = "ROUND_ROBIN"
#   pool_name             = "ash-pool"
#   pool_port             = 12
#   max_connections       = 100
#   health_check_interval = 5
#   health_check_timeout  = 16
#   pool_monitor_protocol = "TCP"
#   enable_persistance    = false
# }

output "asg_id" {
  description = "ID of the autoscaling group"
  value       = airtelcloud_asg.example.id
}

output "asg_name" {
  description = "Name of the autoscaling group"
  value       = airtelcloud_asg.example.name
}

output "asg_vpc_id" {
  description = "Resolved VPC ID of the autoscaling group"
  value       = airtelcloud_asg.example.vpc_id
}

output "asg_network_id" {
  description = "Resolved subnet network ID of the autoscaling group"
  value       = airtelcloud_asg.example.network_id
}
