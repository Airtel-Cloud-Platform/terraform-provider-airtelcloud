terraform {
  required_providers {
    airtelcloud = {
      source  = "terraform-providers/airtelcloud"
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
  default     = "tfa"
}

resource "airtelcloud_security_group" "web" {
  security_group_name = "${var.resource_prefix}-sg-http-servers"
  availability_zone = "S2"
}

# Create a private bucket with versioning
resource "airtelcloud_storage_bucket" "private_bucket" {
  name              = "${var.resource_prefix}-obj-bucket"
  replication_type  = "Local"
  replication_tag   = "south_S1"
  availability_zone = "S1"
  versioning        = true
  object_locking    = false
}

# Create a private subnet
resource "airtelcloud_subnet" "tf_private" {
  network_id        = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  name              = "${var.resource_prefix}-tf1-subnet"
  ipv4_address_space = "10.1.19.0/24"
  subnet_sub_role   = "Private"
  availability_zone = "S2"
  labels            = ["Provider Test"]

  timeouts {
      create = "5m"  # Override create timeout to 20 minutes
      delete = "4m"  # Override delete timeout to 15 minutes
    }
}

# Create NFS volume
resource "airtelcloud_file_storage" "basic" {
  name              = "${var.resource_prefix}-basic-nfs"
  size              = "300"
  availability_zone = "S2"
}

# Create an NFS export path with NFSv4 protocol
resource "airtelcloud_file_storage_export_path" "nfsv4_export" {
  volume            = airtelcloud_file_storage.basic.name
  description       = "NFSv4 export for application servers"
  protocol          = "NFSv4"
  availability_zone = "S2"

  # NFS export settings
  default_access_type = "ReadWrite"
  default_user_squash = "RootSquash"
}

# Create a block storage volume with all common attributes
resource "airtelcloud_volume" "data_volume" {
 name              = "${var.resource_prefix}-storage-volume"
 size              = 50
 type              = "s1_wkld_ntp02_4iops_backend"
 availability_zone = "S1"
 compute_id        = airtelcloud_vm.web1.id

  # Networking
 vpc_id    = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
 subnet_id = "35df162d-5211-4d58-84ed-6a499626949c"

 is_encrypted  = false
 bootable      = false
 depends_on = [airtelcloud_vm.web1]
}

# Create a forward DNS zone
resource "airtelcloud_dns_zone" "tfas" {
  zone_name   = "${var.resource_prefix}-domain.com"
  zone_type   = "forward"
  description = "Example DNS zone for demonstration"
}

# Create another DNS zone for internal services
resource "airtelcloud_dns_zone" "internal" {
  zone_name   = "internal.${var.resource_prefix}-domain.com"
  zone_type   = "forward"
  description = "Internal services DNS zone"
}

# Create an A record pointing to web server
#resource "airtelcloud_dns_record" "web" {
#  zone_id     = airtelcloud_dns_zone.tfas.id
#  owner       = "www"
# record_type = "A"
#  data        = "192.168.1.10"
#  ttl         = 300
#  description = "Web server A record"
#}

# Create an A record for the apex domain
resource "airtelcloud_dns_record" "apex" {
  zone_id     = airtelcloud_dns_zone.tfas.id
  owner       = "apex"
  record_type = "A"
  data        = "192.168.1.10"
  ttl         = 300
  description = "Apex domain A record"
}

#Create a CNAME record
# resource "airtelcloud_dns_record" "blog" {
#   zone_id     = airtelcloud_dns_zone.tfas.id
#   owner       = "blog"
#   record_type = "CNAME"
#   data        = "www.tfas.com."
#   ttl         = 3600#
#   description = "Blog CNAME pointing to www"
# }

# Create an MX record for email
# resource "airtelcloud_dns_record" "mail" {
#   zone_id     = airtelcloud_dns_zone.tfas.id
#   owner       = "@"
#   record_type = "MX"
#   data        = "mail.tfas.com."
#   ttl         = 3600
#   preference  = 10
#   description = "Primary mail server"
# }

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

resource "airtelcloud_security_group_rule" "ssh" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "22"
  port_range_max    = "22"
  remote_ip_prefix  = "10.0.0.0/8"
  ethertype         = "IPv4"
  description       = "Allow SSH from internal network"
}

# Create a VPC peering connection
# resource "airtelcloud_vpc_peering" "example" {
#   name          = "${var.resource_prefix}-vpc-peering"
#   description   = "VPC peering between two networks"
#   vpc_source_name = "perftest-cell1-vpc1"
#   vpc_target_name = "perftest-cell1-vpc2"
#   az            = "S2"
#   region        = "south"

#   is_pcl_enabled = false

#   # allowed_subnet_list = [
#   #   "35df162d-5211-4d58-84ed-6a499626949c",
#   # ]
#   # blocked_subnet_list = [
#   #   "35df162d-5211-4d58-84ed-6a499626949c",
#   # ]
#   timeouts {
#       create = "25m"  # Override create timeout to 25 minutes
#       delete = "15m"  # Override delete timeout to 15 minutes
#     }
# }

resource "airtelcloud_security_group_rule" "http" {
  security_group_id = airtelcloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "8080"
  port_range_max    = "8080"
  remote_ip_prefix  = "0.0.0.0/0"
  ethertype         = "IPv4"
  description       = "Allow HTTP"
}

# =============================================================================
# Virtual Machines
# =============================================================================

# Create web server 1 — flavor and image validated by API (ResolveFlavorID / ResolveImageID)
resource "airtelcloud_vm" "web1" {
  instance_name     = "${var.resource_prefix}-server-1"
  os_type           = "linux"
  flavor_name       = "ccd.Large"
  image_name        = "CentOS_Stream9_Apr2026"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 100
  availability_zone = "S1"
  description       = "Web server 1 behind load balancer"
}

# Create web server 2
resource "airtelcloud_vm" "web2" {
  instance_name     = "${var.resource_prefix}-server-2"
  os_type           = "linux"
  flavor_name       = "ccd.Large"
  image_name        = "CentOS_Stream9_Apr2026"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 100
  availability_zone = "S1"
  description       = "Web server 2 behind load balancer"
}

# =============================================================================
# Compute Snapshot
# =============================================================================

# Create a snapshot of web server 1
# resource "airtelcloud_compute_snapshot" "web_snapshot" {
#   compute_id = airtelcloud_vm.web1.id

#   timeouts {
#     create = "15m"
#     delete = "10m"
#   }
# }

# =============================================================================
# Backup Protection
# =============================================================================

# Create a protection plan with daily schedule and 30-day retention
resource "airtelcloud_protection_plan" "daily" {
  name           = "${var.resource_prefix}-daily-backup-plan"
  description    = "Daily backup with 30-day retention"
  retention      = 1
  retention_unit = "DAYS"
  recurrence     = 86400
  selector_key   = "AZ"
  selector_value = "S1"
  subnet_id      = "35df162d-5211-4d58-84ed-6a499626949c"
}

#Create a protection policy for web server 1
resource "airtelcloud_protection" "web_server" {
  name             = "${var.resource_prefix}-web-backup-schedule"
  description      = "Backup policy for web server"
  compute_id       = airtelcloud_vm.web1.id
  protection_plan  = airtelcloud_protection_plan.daily.name
  enable_scheduler = "true"
  start_date       = "2026-04-01"
  start_time       = "02:00"
}

# =============================================================================
# Load Balancer Service
# =============================================================================

# Create an LB service
resource "airtelcloud_lb_service" "web_lb" {
  name        = "${var.resource_prefix}-web-lb"
  description = "Load balancer for web tier"
  network_id  = "862f1f29-f4d2-4bcb-afdd-3bd96c0ef66e"
  vpc_id      = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  vpc_name    = "perftest-cell1-vpc2"
  ha          = false

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Allocate a VIP for the load balancer
resource "airtelcloud_lb_vip" "web_vip" {
  lb_service_id = airtelcloud_lb_service.web_lb.id
}

# Upload an SSL certificate for HTTPS termination
# resource "airtelcloud_lb_certificate" "web_cert" {
#   lb_service_id   = airtelcloud_lb_service.web_lb.id
#   name            = "${var.resource_prefix}-web-cert"
#   ssl_cert        = file("${path.module}/certs/server.pem")
#   ssl_private_key = file("${path.module}/certs/server-key.pem")
# }

# Create an HTTP virtual server with two backend nodes
resource "airtelcloud_lb_virtual_server" "http" {
  lb_service_id     = airtelcloud_lb_service.web_lb.id
  name              = "${var.resource_prefix}-http-vs"
  vip_port_id       = tonumber(airtelcloud_lb_vip.web_vip.id)
  protocol          = "HTTP"
  port              = 80
  routing_algorithm = "ROUND_ROBIN"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  interval          = 30

  # Backend nodes — compute_ip references VM private IPs dynamically.
  # compute_id is the platform's internal integer ID (not the VM UUID).
  nodes = [
    {
      compute_id = 1
      compute_ip = airtelcloud_vm.web1.private_ip
      port       = 8080
      weight     = 50
    },
    {
      compute_id = 2
      compute_ip = airtelcloud_vm.web2.private_ip
      port       = 8080
      weight     = 50
    },
  ]

  persistence_enabled = true
  persistence_type    = "source_ip"
  x_forwarded_for     = true

  timeouts {
    create = "5m"
    delete = "5m"
  }
  depends_on = [airtelcloud_vm.web1, airtelcloud_vm.web2]
}

# =============================================================================
# Public IP for Web Server 1
# =============================================================================

# Allocate a public IP NATted against web server 1's private IP
resource "airtelcloud_public_ip" "web1_public" {
  object_name       = "${var.resource_prefix}-web1-public-ip"
  vip               = airtelcloud_vm.web1.private_ip
  availability_zone = "S1"

  timeouts {
    create = "10m"
    delete = "5m"
  }
  depends_on = [airtelcloud_vm.web1]
}

# Allow HTTP and HTTPS traffic through the public IP
resource "airtelcloud_public_ip_policy_rule" "web1_http_https" {
  public_ip_id      = airtelcloud_public_ip.web1_public.id
  display_name      = "${var.resource_prefix}-web1-allow-http-https"
  source            = "any"
  services          = ["HTTP", "HTTPS"]
  action            = "accept"
  target_vip        = airtelcloud_public_ip.web1_public.vip
  public_ip         = airtelcloud_public_ip.web1_public.public_ip
  availability_zone = "S1"
}

# Output basic volume details
output "basic_volume_id" {
  description = "ID of the basic NFS volume"
  value       = airtelcloud_file_storage.basic.id
}

output "basic_volume_state" {
  description = "State of the basic NFS volume"
  value       = airtelcloud_file_storage.basic.state
}

output "private_subnet_id" {
  description = "ID of the private subnet"
  value       = airtelcloud_subnet.tf_private.id
}

# Output bucket details
output "private_bucket_id" {
  description = "ID of the private bucket"
  value       = airtelcloud_storage_bucket.private_bucket.id
}

# Output DNS zone details
output "example_zone_id" {
  description = "ID of the tfas DNS zone"
  value       = airtelcloud_dns_zone.tfas.id
}

output "example_zone_name" {
  description = "Name of the tfas DNS zone"
  value       = airtelcloud_dns_zone.tfas.zone_name
}

output "internal_zone_id" {
  description = "ID of the internal DNS zone"
  value       = airtelcloud_dns_zone.internal.id
}

# Output volume details
output "data_volume_uuid" {
  description = "UUID of the data volume"
  value       = airtelcloud_volume.data_volume.uuid
}

# Output compute snapshot details
# output "web_snapshot_id" {
#   description = "ID of the web server snapshot"
#   value       = airtelcloud_compute_snapshot.web_snapshot.id
# }

# output "web_snapshot_status" {
#   description = "Status of the web server snapshot"
#   value       = airtelcloud_compute_snapshot.web_snapshot.status
# }

# Output protection details
output "protection_plan_id" {
  description = "ID of the daily protection plan"
  value       = airtelcloud_protection_plan.daily.id
}

output "web_protection_id" {
  description = "ID of the web server protection policy" 
  value       = airtelcloud_protection.web_server.id
}

# Output load balancer details
output "web_lb_id" {
  description = "ID of the web load balancer service"
  value       = airtelcloud_lb_service.web_lb.id
}

output "web_lb_status" {
  description = "Status of the web load balancer"
  value       = airtelcloud_lb_service.web_lb.status
}

output "web_vip_id" {
  description = "ID of the web VIP"
  value       = airtelcloud_lb_vip.web_vip.id
}

output "web_vip_fixed_ips" {
  description = "Fixed IPs of the web VIP"
  value       = airtelcloud_lb_vip.web_vip.fixed_ips
}

output "http_virtual_server_id" {
  description = "ID of the HTTP virtual server"
  value       = airtelcloud_lb_virtual_server.http.id
}

output "http_virtual_server_vip" {
  description = "VIP address of the HTTP virtual server"
  value       = airtelcloud_lb_virtual_server.http.vip
}

# Output VM details
output "web1_id" {
  description = "ID of web server 1"
  value       = airtelcloud_vm.web1.id
}

output "web1_private_ip" {
  description = "Private IP of web server 1"
  value       = airtelcloud_vm.web1.private_ip
}

output "web2_id" {
  description = "ID of web server 2"
  value       = airtelcloud_vm.web2.id
}

output "web2_private_ip" {
  description = "Private IP of web server 2"
  value       = airtelcloud_vm.web2.private_ip
}

# Output public IP details
output "web1_public_ip_id" {
  description = "ID of the public IP allocated for web server 1"
  value       = airtelcloud_public_ip.web1_public.id
}

output "web1_public_ip_address" {
  description = "Public IP address of web server 1"
  value       = airtelcloud_public_ip.web1_public.public_ip
}
