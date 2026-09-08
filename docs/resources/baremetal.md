---

page_title: "airtelcloud_baremetal Resource - Airtel Cloud"

subcategory: "Compute"

description: |-
Allocates and manages an Airtel Cloud baremetal server.

---

**# airtelcloud_baremetal (Resource)**

Allocates and manages an Airtel Cloud baremetal server.

The resource supports:

* Baremetal server allocation.
* Primary and additional subnet configuration.
* OS image and flavor selection.
* Optional additional block storage.
* Optional backup schedule configuration.
* Cloud-init configuration for first boot.
* SSH keypair configuration.
* Reserved capacity allocation.
* Server state refresh.
* Backup policy updates.
* Server release and optional disk deletion on destroy.

**## Example Usage**

```terraform
resource "airtelcloud_baremetal" "app" {
  name              = "tft-bm-app-01"
  flavor            = "metal-c56-m1024"
  os_image          = "ubuntu22_Aug2026"

  subnet_name = "proxy-test-subnet"

  additional_subnet_names = [
    "temporal-subnet"
  ]

  availability_zone = "S1"
  network_name      = "copper-vpc1"

  keypair     = "Vinay"
  is_reserved = false

  tags = [
    "South-AZ1"
  ]

  cloud_init = <<-EOF
    #cloud-config
    package_update: true
    packages:
      - nginx
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
  EOF

  storage = [
    {
      name         = "baremetel"
      size         = "10"
      path         = "/test"
      type         = "BlockStorage"
      file_system  = "xfs"
      force_format = true
    }
  ]

  backup_config = {
    schedule_type       = "weekly_full"
    start_time          = "21:00"
    incr_days           = []
    full_days            = [7]
    full_retention      = 1
    full_retention_unit = "MONTHS"
    backup_selections   = ["/test"]
  }

  policy_enabled = true

  delete_disks = true
  secure_erase = false
}
```

**## Argument Reference**

**### Required**

* `name` (String) - The name of the baremetal server.

* `flavor` (String) - The flavor profile used to allocate the baremetal server.

  Example:

  ```terraform
  flavor = "metal-c56-m1024"
  ```

* `os_image` (String) - The OS image name used to install the operating system on the baremetal server.

  Example:

  ```terraform
  os_image = "ubuntu22_Aug2026"
  ```

* `subnet_name` (String) - The display name of the primary subnet to attach to the baremetal server.

  The provider resolves the subnet name to its UUID before sending the allocation request.

* `availability_zone` (String) - The availability zone in which the baremetal server is allocated.

  Valid availability zones include:

  * `N1`
  * `N2`
  * `S1`
  * `S2`

  Example:

  ```terraform
  availability_zone = "S1"
  ```

**### Optional**

* `network_name` (String) - The VPC name or UUID to which the baremetal server is connected.

  If a VPC name is provided, the provider resolves it to the VPC UUID before allocation.

  The VPC is required when resolving `subnet_name`.

  Example:

  ```terraform
  network_name = "copper-vpc1"
  ```

* `additional_subnet_names` (List of String) - A list of additional subnet display names to attach to the server.

  The primary subnet specified by `subnet_name` is attached first, followed by the additional subnets.

  Example:

  ```terraform
  additional_subnet_names = [
    "temporal-subnet",
    "backup-subnet"
  ]
  ```

* `storage` (List of Object) - Configures additional disks to be attached to the baremetal server.

  Each storage object supports:

  * `name` (String) - Name of the disk.
  * `size` (String) - Disk size.
  * `path` (String) - Mount path for the disk.
  * `type` (String) - Storage type, for example `BlockStorage`.
  * `file_system` (String) - File system to use for the disk, for example `xfs`.
  * `force_format` (Boolean) - Whether the disk should be formatted.

  Example:

  ```terraform
  storage = [
    {
      name         = "application-data"
      size         = "100"
      path         = "/data"
      type         = "BlockStorage"
      file_system  = "xfs"
      force_format = true
    }
  ]
  ```

* `backup_config` (Object) - Configures the backup schedule for the baremetal server.

  The object supports:

  * `schedule_type` (String) - Backup schedule type.
  * `start_time` (String) - Time at which the backup schedule starts, for example `21:00`.
  * `incr_days` (List of Number) - Days on which incremental backups are performed.
  * `full_days` (List of Number) - Days on which full backups are performed.
  * `full_retention` (Number) - Number of retention units for full backups.
  * `full_retention_unit` (String) - Retention unit, for example `MONTHS`.
  * `backup_selections` (List of String) - Paths to include in the backup.

  Example:

  ```terraform
  backup_config = {
    schedule_type       = "weekly_full"
    start_time          = "21:00"
    incr_days           = []
    full_days           = [7]
    full_retention      = 1
    full_retention_unit = "MONTHS"
    backup_selections   = ["/data"]
  }
  ```

* `cloud_init` (String) - Cloud-init configuration that is executed during the server's first boot.

  Cloud-init can be used to automate initial server configuration, such as:

  * Installing packages.
  * Creating users.
  * Configuring services.
  * Writing configuration files.
  * Running initialization commands.

  Example:

  ```terraform
  cloud_init = <<-EOF
    #cloud-config

    package_update: true

    packages:
      - nginx

    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
  EOF
  ```

  You can also provide a shell script:

  ```terraform
  cloud_init = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl enable nginx
    systemctl start nginx
  EOF
  ```

  > **Note:** Cloud-init is intended for first-boot initialization of the baremetal server.

* `is_reserved` (Boolean) - Specifies whether the server should be allocated from reserved capacity.

  Default:

  ```terraform
  is_reserved = false
  ```

* `system_id` (String) - Optional system identifier used by the backend during reservation and release operations.

* `keypair` (String) - The name of the SSH keypair to inject into the baremetal server.

  Example:

  ```terraform
  keypair = "Vinay"
  ```

* `tags` (List of String) - Tags used during server allocation.

  Example:

  ```terraform
  tags = [
    "South-AZ1"
  ]
  ```

  The tag represents the placement/allocation tag used by the Airtel Cloud baremetal service. It is not intended to be used as a generic Terraform label such as `terraform`.

* `policy_enabled` (Boolean) - Enables or disables the backup policy for the server.

  This is a mutable field and is updated using the backup policy update API.

  Example:

  ```terraform
  policy_enabled = true
  ```

  When specified during creation, the value is also included in the `backupConfig.policyEnabled` configuration.

* `delete_disks` (Boolean) - Determines whether attached disks should be deleted when the baremetal server is destroyed.

  Default:

  ```terraform
  delete_disks = false
  ```

  Set this to `true` when attached disks should also be deleted during Terraform destroy.

* `secure_erase` (Boolean) - Determines whether secure erase should be performed when the server is destroyed.

  Default:

  ```terraform
  secure_erase = false
  ```

**## Network Configuration**

A baremetal server can have one primary subnet and additional subnets.

The primary subnet is specified using `subnet_name`:

```terraform
subnet_name = "proxy-test-subnet"
```

Additional subnets can be specified using `additional_subnet_names`:

```terraform
additional_subnet_names = [
  "temporal-subnet",
  "backup-subnet"
]
```

When a subnet name is provided, the provider resolves the subnet display name to its UUID before sending the server allocation request.

If `network_name` is a VPC name rather than a UUID, the provider resolves the VPC name to its UUID.

**## Storage Configuration**

Additional disks can be configured using the `storage` argument.

Example:

```terraform
storage = [
  {
    name         = "application-data"
    size         = "100"
    path         = "/data"
    type         = "BlockStorage"
    file_system  = "xfs"
    force_format = true
  }
]
```

The `path` specifies where the disk is mounted on the server.

For example:

```text
/data
/test
/backup
```

**## Backup Configuration**

Backups can be configured using `backup_config`.

Example:

```terraform
backup_config = {
  schedule_type       = "weekly_full"
  start_time          = "21:00"
  incr_days           = []
  full_days           = [7]
  full_retention      = 1
  full_retention_unit = "MONTHS"
  backup_selections   = ["/data"]
}
```

The `backup_selections` field specifies the paths that should be included in the backup.

For example:

```terraform
backup_selections = [
  "/data",
  "/backup"
]
```

The `policy_enabled` argument can be used to enable or disable the backup policy after the server has been created.

**## Cloud-Init**

The `cloud_init` argument allows you to automatically configure the server during its first boot.

For example, the following configuration installs and starts Nginx:

```terraform
cloud_init = <<-EOF
  #cloud-config

  package_update: true

  packages:
    - nginx

  runcmd:
    - systemctl enable nginx
    - systemctl start nginx
EOF
```

Cloud-init can be used to automate initial server configuration, including package installation, service configuration, user creation, and startup commands.

**## Destroy Behavior**

When the Terraform resource is destroyed, the baremetal server is released.

The following arguments control what happens to attached disks:

```terraform
delete_disks = true
secure_erase = false
```

To delete attached disks when the server is destroyed:

```terraform
delete_disks = true
```

To perform secure erase:

```terraform
secure_erase = true
```

By default, both options are disabled:

```terraform
delete_disks = false
secure_erase = false
```

**## Attribute Reference**

In addition to all arguments above, the following attributes are exported:

* `id` (String) - The Terraform resource ID. This is equal to the baremetal server name.

* `uuid` (String) - The unique UUID of the baremetal server.

* `state` (String) - The current state of the baremetal server.

* `hostname` (String) - The hostname assigned to the baremetal server.

* `availability_zone` (String) - The availability zone in which the baremetal server is allocated.

* `power` (String) - The current power state of the baremetal server.

* `ip_addresses` (List of String) - The IP addresses assigned to the baremetal server.

* `backend_port_id` (Number) - The backend network port identifier returned by the detail API (`networkInfo.portId`).

**## Import**

An existing baremetal server can be imported using its server name:

```shell
terraform import airtelcloud_baremetal.app <baremetal-name>
```

After importing the resource, run:

```shell
terraform plan
```

to verify that the Terraform configuration matches the existing baremetal server.

**## Important Notes**

* `availability_zone` must be a valid availability zone supported by the Airtel Cloud baremetal service.
* `subnet_name` refers to the subnet display name and is resolved to a subnet UUID by the provider.
* `network_name` can be a VPC name or UUID.
* `additional_subnet_names` can be used when the server requires connectivity to multiple subnets.
* `cloud_init` is used for first-boot server initialization.
* `policy_enabled` is a mutable backup-policy setting and can be changed after the server is created.
* `delete_disks` controls whether attached disks are removed when the server is destroyed.
* `secure_erase` controls whether secure erase is requested during server destruction.
* `is_reserved` controls whether the allocation uses reserved capacity.
