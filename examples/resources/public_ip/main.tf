# Allocate a public IP NATted against a VM's private IP
resource "airtelcloud_public_ip" "example" {
  object_name       = "my-vm-public-ip"
  vip               = "10.1.99.172" # Private IP of the VM or Load Balancer
  availability_zone = "S1"

  timeouts {
    create = "10m"
    delete = "5m"
  }
}

output "public_ip_id" {
  value = airtelcloud_public_ip.example.id
}

output "public_ip_address" {
  value = airtelcloud_public_ip.example.public_ip
}

output "public_ip_status" {
  value = airtelcloud_public_ip.example.status
}

# Add a policy rule to allow HTTP and HTTPS traffic
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_id      = airtelcloud_public_ip.example.id
  display_name      = "web-traffic"
  source            = "any"
  services          = ["HTTP", "HTTPS"]
  action            = "accept"
  target_vip        = airtelcloud_public_ip.example.vip
  public_ip         = airtelcloud_public_ip.example.public_ip
  availability_zone = "S1"
}
