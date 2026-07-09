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
  flavor_name       = "ccd.Large"
  image_name        = "Ubuntu22_04_Jul2026"
  os_type           = "linux"
  vpc_id            = "vpc-id-123"
  subnet_id         = "subnet-id-123"
  security_group_id = "sg-id-123"
  availability_zone = "S1"
  disk_size         = 100

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

You can attach the security group during provisioning by setting either `security_group_id` or `security_group_name`. If the security group is managed in the same Terraform configuration, prefer `security_group_id = airtelcloud_security_group.<name>.id`.

### Windows VM with Backup

```terraform
resource "airtelcloud_vm" "windows_server" {
  instance_name     = "win-server"
  flavor_name       = "ccd.Large"
  image_name        = "WIN2K19_PREACT_Jul2026"
  os_type           = "windows"
  vpc_id            = "vpc-id-123"
  subnet_id         = "subnet-id-123"
  security_group_id = "sg-id-123"
  availability_zone = "S1"
  disk_size         = 100
  boot_from_volume  = true
  enable_backup     = true
  protection_plan   = "daily"
  start_date        = "2025-07-15"
  start_time        = "02:00"

  tags = {
    Environment = "production"
    OS          = "windows"
  }
}
```

The VM resource currently does not expose dedicated username/password provisioning fields. Use image defaults, keypairs, or initialization scripts as supported by your image.

## Argument Reference

### Required

- `instance_name` (String) - The name of the compute instance.
- `flavor_id` or `flavor_name` (String) - Exactly one must be specified. Use `flavor_id` to pass the platform flavor ID, or `flavor_name` to pass the flavor name shown in the catalog. Forces replacement on change.
- `image_id` or `image_name` (String) - Exactly one must be specified. Use `image_id` to pass the platform image ID, or `image_name` to pass the image name shown in the catalog. Forces replacement on change.
- `vpc_id` (String) - The ID of the VPC. Forces replacement on change.
- `subnet_id` (String) - The ID of the subnet. Forces replacement on change.
- `os_type` (String) - The OS type: `"linux"` or `"windows"`. Forces replacement on change.

### Optional

- `flavor_name` (String) - The flavor name for the compute instance. Conflicts with `flavor_id`. Forces replacement on change.
- `image_name` (String) - The image name for the compute instance. Conflicts with `image_id`. Forces replacement on change.
- `security_group_id` (String) - The ID of the security group to attach during VM provisioning. One of `security_group_id` or `security_group_name` may be specified.
- `security_group_name` (String) - The name of the security group to attach during VM provisioning. One of `security_group_id` or `security_group_name` may be specified.
- `keypair_id` (String) - The ID of the key pair for SSH access. Forces replacement on change.
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
