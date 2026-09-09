---
page_title: "airtelcloud_postgres Resource - Airtel Cloud"
subcategory: "Database"
description: |-
  Manages an Airtel Cloud PostgreSQL cluster.
---

# airtelcloud_postgres (Resource)

Manages an Airtel Cloud PostgreSQL cluster.

v1 supports create, read, import, and delete. Changing configuration forces a new cluster. The API does not return the admin password; Terraform stores the configured value in state.

## Example Usage

```terraform
variable "postgres_password" {
  type      = string
  sensitive = true
}

resource "airtelcloud_postgres" "app" {
  cluster_name       = "app-db"
  version            = "17"
  database_name      = "app"
  postgres_username  = "dbadmin"
  password           = var.postgres_password
  compute_size       = "db.postgres.uhper.ccs.xlarge"
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
- `version` (String) - PostgreSQL major version (for example `17` or `18`). Forces new resource.
- `database_name` (String) - Name of the initial database. Forces new resource.
- `postgres_username` (String) - Admin username. Forces new resource.
- `password` (String, Sensitive) - Admin password. Forces new resource. Not returned by the API.
- `compute_size` (String) - Flavor name from the postgres flavors catalog (for example `db.postgres.uhper.ccs.xlarge`). Resolved to flavor ID and RAM internally. Forces new resource.
- `storage_size` (Number) - Data volume size in GB. Must be at least 200. Forces new resource.
- `availability_zone` (String) - Availability zone code (for example `S1`). Forces new resource.
- `security_group` (Attributes) - Allowed client CIDRs. Forces new resource.
  - `allowed_ips` (List of String) - CIDR blocks allowed to connect. Must contain at least one IP range.

### Optional

- `description` (String) - Cluster description. Forces new resource.
- `high_availability` (Boolean) - When true, creates primary-standby topology. Defaults to `false`. Forces new resource.
- `num_replicas` (Number) - Standby replica count. Required and must be between `1` and `10` when `high_availability` is true. Defaults to `0`. Forces new resource.
- `is_superuser` (Boolean) - Whether the admin user is a superuser. Defaults to `false`. Forces new resource.
- `network_type` (String) - Network type. Defaults to `private`. Forces new resource.
- `storage_type` (String) - Storage class label from the volume-types catalog. Defaults to `High Performance`. Resolved internally. Forces new resource.
- `pg_extensions` (List of String) - Extensions to enable (for example `pgvector`). Forces new resource.
- `labels` (List of String) - Labels to assign. Forces new resource.
- `backup` (Attributes) - Backup configuration. Forces new resource.
  - `enabled` (Boolean) - Whether backup is enabled. Defaults to `false`.
  - `protection_plan` (String) - Plan value from the protection-plans catalog (for example `weekly-full-daily-incr`).
  - `compression_level` (Number) - Compression level. Defaults to `6`.
  - `retention` (Number) - Retention in days. Defaults to `15`.
  - `schedule_time` (String) - Backup schedule time in `HH:MM` format. Required when `enabled` is `true`.
  - `schedule_day` (String) - Backup schedule day. Must be one of `Monday` through `Sunday`. Required when `enabled` is `true`.
- `timeouts` (Block) - Create and delete timeouts. Defaults to 30 minutes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - Cluster UUID.
- `status` (String) - Current status (`Creating`, `Active`, `Failed`, and others).
- `topology` (String) - Resolved topology (`standalone` or `primary-standby`).
- `connection_string` (String) - PostgreSQL connection string. Empty until the cluster is Active. The API masks the password.
- `created_at` (String) - Creation timestamp.
- `backup.schedule_time` (String) - Backup schedule time returned by the API.
- `backup.schedule_day` (String) - Backup schedule day returned by the API.

## Import

Clusters can be imported using the cluster UUID. The admin password is not returned by the API, so import cannot populate `password`.

```shell
terraform import airtelcloud_postgres.app <cluster-uuid>
```
