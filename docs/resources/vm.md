---
page_title: "airtelcloud_vm Resource - Airtel Cloud"
subcategory: "Compute"
description: |-
  Manages an Airtel Cloud virtual machine (compute instance).
---

# airtelcloud_vm (Resource)

Manages an Airtel Cloud virtual machine (compute instance).

Uses the v2.1 Compute API with domain/project URL paths. The provider's `organization` and `project_name` settings are embedded in the API URL automatically.

## Example Usage

### Basic Linux VM

```terraform
resource "airtelcloud_vm" "web_server" {
  instance_name     = "web-server"
  flavor_id         = "t2.micro"
  image_id          = "ubuntu-22.04"
  os_type           = "linux"
  vpc_id            = "vpc-abc123"
  subnet_id         = "subnet-def456"
  keypair_id        = "my-keypair"
  availability_zone = "S1"
  disk_size         = 40

  user_data = base64encode(<<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl start nginx
    EOF
  )

  tags = {
    Environment = "production"
    Purpose     = "web-server"
  }
}
```

### Windows VM with Backup

```terraform
resource "airtelcloud_vm" "windows_server" {
  instance_name     = "win-server"
  flavor_id         = "m5.large"
  image_id          = "windows-2022"
  os_type           = "windows"
  vpc_id            = "vpc-abc123"
  subnet_id         = "subnet-def456"
  admin_password    = var.windows_password
  availability_zone = "S2"
  disk_size         = 100
  boot_from_volume  = true
  enable_backup     = true
  protection_plan   = "daily"
  start_date        = "2025-01-15"
  start_time        = "02:00"

  tags = {
    Environment = "production"
    OS          = "windows"
  }
}
```

## Argument Reference

### Required

- `instance_name` (String) - The name of the compute instance.
- `flavor_id` (String) - The flavor ID for the compute instance (e.g., `t2.micro`, `m5.large`).
- `image_id` (String) - The ID of the image to use. Forces replacement on change.
- `vpc_id` (String) - The ID of the VPC. Forces replacement on change.
- `subnet_id` (String) - The ID of the subnet. Forces replacement on change.
- `os_type` (String) - The OS type: `"linux"` or `"windows"`. Forces replacement on change.

### Optional

- `security_group_id` (String) - The ID of the security group.
- `keypair_id` (String) - The ID of the key pair for SSH access. Forces replacement on change.
- `admin_password` (String, Sensitive) - Admin password for the instance. Forces replacement on change.
- `user_data` (String) - Cloud-init script to run on instance initialization. Forces replacement on change.
- `availability_zone` (String) - The availability zone (e.g., `S1`, `S2`). Forces replacement on change.
- `region` (String) - The region for the instance. Defaults to the provider's `region`.
- `disk_size` (Number) - The disk size in GB. Default: `20`.
- `boot_from_volume` (Boolean) - Whether to boot from volume. Default: `true`.
- `volume_type_id` (String) - The volume type ID.
- `description` (String) - A description of the compute instance.
- `enable_backup` (Boolean) - Whether backup is enabled. Default: `false`.
- `protection_plan` (String) - The protection plan for the instance.
- `start_date` (String) - The start date for backup scheduling (e.g., `"2025-01-15"`).
- `start_time` (String) - The start time for backup scheduling (e.g., `"02:00"`).
- `vm_count` (Number) - Number of VM instances to create. Must be between 1 and 10. Default: `1`.
- `tags` (Map of String) - A map of tags to assign to the instance.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier of the compute instance.
- `provider_instance_id` (String) - The provider-specific instance ID.
- `status` (String) - The current status of the instance (e.g., `ACTIVE`, `BUILD`, `SHUTOFF`).
- `public_ip` (String) - The public IP address of the instance.
- `private_ip` (String) - The private IP address of the instance.

## Import

VMs can be imported using the `id`:

```shell
terraform import airtelcloud_vm.web_server <compute-id>
```
