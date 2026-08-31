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
variable "vm_admin_password" {
  description = "Login password for the VM admin user"
  type        = string
  sensitive   = true
}


resource "airtelcloud_vm" "web_server" {
  instance_name     = "web-server"
  flavor_name       = "ccd.Large"
  image_name        = "Ubuntu22_04_Jul2026"
  os_type           = "linux"
  vpc_name          = "vpc01"
  subnet_name       = "sub1"
  security_group_names = ["all-open-az1"]
  availability_zone = "N1"
  disk_size         = 100
  admin_username    = "clouduser"
  admin_password    = var.vm_admin_password

  user_data = base64encode(<<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl start nginx
    EOF
  )

  labels = ["production", "web-server", "frontend"]
}
```

The provider sends VM labels in a follow-up compute labels PATCH request after
the instance becomes ready. Use `labels` when you want the dashboard to show
plain labels exactly as written.

### Linux VM with SSH Keypair

```terraform
resource "airtelcloud_vm" "web_server_keypair" {
  instance_name       = "web-server-keypair"
  flavor_name         = "ccd.Large"
  image_name          = "Ubuntu22_04_Jul2026"
  os_type             = "linux"
  vpc_name            = "vpc01"
  subnet_name         = "sub1"
  security_group_names = ["all-open-az1"]
  availability_zone   = "S1"
  boot_from_volume    = true
  disk_size           = 120
  volume_type_id      = "31"
  keypair_name        = "my-linux-keypair"

  labels = ["production", "keypair-auth", "api"]
}
```

### Linux VM Using image_name With Multiple Security Groups

```terraform
resource "airtelcloud_vm" "web_server_multi_sg" {
  instance_name       = "web-server-multi-sg"
  flavor_name         = "ccd.Large"
  image_name          = "Ubuntu22_04_Jul2026"
  os_type             = "linux"
  vpc_name            = "vpc01"
  subnet_name         = "sub1"
  security_group_names = ["all-open-az1", "ssh-open", "app-allow", "db-allow"]
  availability_zone   = "S1"
  boot_from_volume    = true
  disk_size           = 100
  volume_type_id      = "31"
  admin_username      = "clouduser"
  admin_password      = var.vm_admin_password
}
```

### Linux VM Using snapshot_name

```terraform
resource "airtelcloud_vm" "web_server_from_snapshot" {
  instance_name         = "web-server-from-snapshot"
  flavor_name           = "ccd.Large"
  snapshot_name         = "snap-test-snap"
  os_type               = "linux"
  vpc_name              = "copper-vpc1"
  subnet_name           = "proxy-test-subnet"
  security_group_names  = ["proxy-test-security-group", "testaz-2"]
  availability_zone     = "S1"
  boot_from_volume      = true
  disk_size             = 200
  admin_username        = "clouduser"
  admin_password        = var.vm_admin_password
  description           = "Linux VM created from snapshot image"
}
```

Replace the name placeholders with real values from your environment. Prefer name-based fields (`vpc_name`, `subnet_name`, `security_group_names`) for readability. If needed, you can still use IDs via references, for example: `vpc_id = airtelcloud_vpc.main.id`, `subnet_id = airtelcloud_subnet.main.id`, and `security_group_ids = [airtelcloud_security_group.main.id]`.

You can attach one or more security groups during provisioning by setting either `security_group_ids` or `security_group_names`. These inputs are mutually exclusive. If security groups are managed in the same Terraform configuration, prefer IDs.

For Linux VMs, configure exactly one authentication method:
- Username/password: set both `admin_username` and `admin_password`
- SSH keypair: set one of `keypair_id` or `keypair_name`

Do not combine username/password with keypair fields.

### Windows VM with Backup

```terraform
resource "airtelcloud_vm" "windows_server" {
  instance_name       = "win-server"
  flavor_name         = "ccd.Large"
  image_name          = "WIN2K19_PREACT_Jul2026"
  os_type             = "windows"
  vpc_name            = "vpc01"
  subnet_name         = "sub1"
  security_group_names = ["all-open-az1"]
  availability_zone   = "N1"
  disk_size           = 100
  boot_from_volume    = true
  enable_backup       = true
  protection_plan     = "<protection-plan-id>"
  weekday             = "monday"
  start_time          = "02:00"

  labels = ["production", "windows", "backup"]
}
```

## Argument Reference

### Required

- `instance_name` (String) - The name of the compute instance.
- `flavor_id` or `flavor_name` (String) - Exactly one must be specified. Use `flavor_id` to pass the platform flavor ID, or `flavor_name` to pass the flavor name shown in the catalog. Forces replacement on change.
- `image_id`, `image_name`, or `snapshot_name` (String) - Exactly one must be specified. Use `image_id` to pass the platform image ID, `image_name` to pass the image name shown in the catalog, or `snapshot_name` to resolve a compute snapshot to its backing image ID. Forces replacement on change.
- `vpc_id` or `vpc_name` (String) - Exactly one must be specified. The VPC ID or VPC name. Forces replacement on change.
- `subnet_id` or `subnet_name` (String) - Exactly one must be specified. The subnet ID or subnet name. Forces replacement on change.
- `os_type` (String) - The OS type: `"linux"`, `"ubuntu"`, `"rhel"`, `"suse"`, `"centos"`, or `"windows"`. Forces replacement on change.

### Optional

- `flavor_name` (String) - The flavor name for the compute instance. Conflicts with `flavor_id`. Forces replacement on change.
- `image_name` (String) - The image name for the compute instance. Conflicts with `image_id`. Forces replacement on change.
- `snapshot_name` (String) - The compute snapshot name to use for VM creation. The provider resolves this through the snapshot list API and uses the snapshot `image_id`. Conflicts with `image_id` and `image_name`. Forces replacement on change.
- `vpc_name` (String) - The VPC name. Conflicts with `vpc_id`.
- `subnet_name` (String) - The subnet name. Conflicts with `subnet_id`.
- `security_group_ids` (List of String) - List of security group IDs to attach during VM provisioning. Mutually exclusive with `security_group_names`.
- `security_group_names` (List of String) - List of security group names to attach during VM provisioning. Each name is resolved to an ID. Mutually exclusive with `security_group_ids`.
- `keypair_id` (String) - The ID of the key pair for SSH access. Forces replacement on change.
- `keypair_name` (String) - The name of the key pair for SSH access. Conflicts with `keypair_id`. Forces replacement on change.
- `admin_username` (String) - Login username to create on the instance. Supported for `os_type`: `linux`, `ubuntu`, `rhel`, `suse`, `centos`. Must be set together with `admin_password`, and cannot be combined with `keypair_id` or `keypair_name`. Forces replacement on change.
- `admin_password` (String, Sensitive) - Login password for `admin_username`. Supported for `os_type`: `linux`, `ubuntu`, `rhel`, `suse`, `centos`. Must be set together with `admin_username`, and cannot be combined with `keypair_id` or `keypair_name`. Minimum 8 characters with at least one uppercase letter, one lowercase letter, and one special character. Stored in plaintext in Terraform state. Forces replacement on change.
- `user_data` (String) - Cloud-init script to run on instance initialization. Forces replacement on change.
- `availability_zone` (String) - The availability zone (e.g., `S1`, `S2`). Forces replacement on change.
- `region` (String) - The region for the instance. Defaults to the provider's `region`.
- `disk_size` (Number) - The disk size in GB. Default: `20`.
- `boot_from_volume` (Boolean) - Whether to boot from volume. Default: `true`.
- `volume_type_id` (String) - The volume type ID.
- `description` (String) - A description of the compute instance.
- `enable_backup` (Boolean) - Whether backup is enabled. Default: `false`.
- `protection_plan` (String) - Protection plan UUID or name for the instance. The provider accepts either the UUID (e.g. `4cb5b1b6-f62f-4fea-b348-d17aa407d64d`) or the plan name (e.g. `daily-backup`). Names are resolved to UUIDs at apply time.
- `start_date` (String) - The start date for backup scheduling (e.g., `"2025-01-15"`). Mutually exclusive with `weekday`.
- `weekday` (String) - Optional weekday convenience input for backup scheduling (`monday`..`sunday` or `mon`..`sun`). Mutually exclusive with `start_date`. Converted internally to the next matching `start_date` in IST (`Asia/Kolkata`).
- `start_time` (String) - The start time for backup scheduling (e.g., `"02:00"`).
- `vm_count` (Number) - Number of VM instances to create. Must be between 1 and 10. Default: `1`.
- `labels` (List of String) - Plain labels to assign to the instance. These are sent exactly as provided in the follow-up compute labels PATCH request. Maximum 5 labels. Each label must be between 3 and 15 characters long. Use simple label values such as `web-server`.


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

The API does not return credentials, so `admin_username` and `admin_password` are
empty after an import. If your configuration sets them, the next plan plans a
replacement — the same limitation applies to `keypair_id`.
