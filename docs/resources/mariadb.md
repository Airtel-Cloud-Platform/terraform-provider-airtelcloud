---
page_title: "airtelcloud_mariadb Resource - Airtel Cloud"
subcategory: "Database"
description: |-
  Manages an Airtel Cloud MariaDB cluster.
---

# airtelcloud_mariadb (Resource)

Manages an Airtel Cloud MariaDB cluster.

v1 supports create, read, import, and delete. Changing configuration forces a new cluster. The API does not return the admin password; Terraform stores the configured value in state.

## Example Usage

```terraform
variable "mariadb_password" {
  type      = string
  sensitive = true
}

resource "airtelcloud_mariadb" "app" {
  cluster_name       = "app-db"
  version            = "10.6"
  database_name      = "app"
  mariadb_username   = "dbadmin"
  password           = var.mariadb_password
  compute_size       = "db.maria.uhper.ccs.xlarge"
  storage_size       = 200
  availability_zone  = "S1"
  high_availability  = true
  num_replicas       = 1

  backup = {
    enabled          = true
    protection_plan  = "weekly-full-daily-incr"
    schedule_time    = "02:00"
    schedule_day     = "Monday"
  }

  security_group = {
    allowed_ips = ["192.168.1.0/24"]
  }
}
```

## Argument Reference

### Required

- `cluster_name` (String) - Display name of the cluster. Forces new resource. The API may suffix a UUID onto the stored name; Terraform keeps this configured value.
- `version` (String) - MariaDB engine version (for example `10.6`, `10.11`, `11.4`, or `11.8`). Validated against the versions catalog. Forces new resource.
- `database_name` (String) - Name of the initial database. Forces new resource.
- `mariadb_username` (String) - Admin username. Forces new resource.
- `password` (String, Sensitive) - Admin password. Forces new resource. Not returned by the API.
- `compute_size` (String) - Flavor name from the mariadb flavors catalog (for example `db.maria.uhper.ccs.xlarge`). Resolved to flavor ID, vCPUs, and RAM internally. Forces new resource.
- `storage_size` (Number) - Data volume size in GB. Must be at least 200. Forces new resource.
- `availability_zone` (String) - Availability zone code (for example `S1`). Forces new resource.

### Optional

- `description` (String) - Cluster description. Forces new resource.
- `high_availability` (Boolean) - When true, creates master-slave topology. Defaults to `false`. Forces new resource.
- `num_replicas` (Number) - Slave replica count. Required and must be between `1` and `10` when `high_availability` is true. Defaults to `0`. Forces new resource.
- `is_superuser` (Boolean) - Whether the admin user is a superuser. Defaults to `false`. Forces new resource.
- `network_type` (String) - Network type. Defaults to `private`. Forces new resource.
- `storage_type` (String) - Storage class label from the volume-types catalog. Defaults to `High Performance`. Resolved internally. Forces new resource.
- `labels` (List of String) - Labels to assign. Forces new resource.
- `backup` (Attributes) - Backup configuration. Forces new resource.
  - `enabled` (Boolean) - Whether backup is enabled. Defaults to `false`.
  - `protection_plan` (String) - Plan value from the protection-plans catalog (for example `weekly-full-daily-incr`).
  - `compression_level` (Number) - Compression level. Defaults to `1`.
  - `retention` (Number) - Retention in days.
  - `schedule_time` (String) - Backup schedule time in `HH:MM` format. Required when `enabled` is `true`.
  - `schedule_day` (String) - Backup schedule day. Must be one of `Monday` through `Sunday`. Required when `enabled` is `true`.
- `security_group` (Attributes) - Allowed client CIDRs. Forces new resource.
  - `allowed_ips` (List of String) - CIDR blocks allowed to connect.
- `timeouts` (Block) - Create and delete timeouts. Defaults to 15 minutes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - Cluster UUID.
- `status` (String) - Current status (`Creating`, `Active`, `Failed`, and others).
- `topology` (String) - Resolved topology (`master` or `master-slave`).
- `connection_string` (String, Sensitive) - MariaDB connection string. Empty until the cluster is Active. The API masks the password.
- `port` (Number) - Database port (typically 3350).
- `users` (String) - Admin username returned by the API.
- `created_at` (String) - Creation timestamp.
- `backup.schedule_time` (String) - Backup schedule time returned by the API.
- `backup.schedule_day` (String) - Backup schedule day returned by the API.

## Import

Clusters can be imported using the cluster UUID. The admin password is not returned by the API, so import cannot populate `password`.

```shell
terraform import airtelcloud_mariadb.app <cluster-uuid>
```
