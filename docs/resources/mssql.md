---
page_title: "airtelcloud_mssql Resource - Airtel Cloud"
subcategory: "Database"
description: |-
  Manages an Airtel Cloud MSSQL instance.
---

# airtelcloud_mssql (Resource)

Manages an Airtel Cloud MSSQL instance.

Create, read, import, and delete are supported. Changing configuration forces a new instance. The API does not return the admin password. `vpc` and `subnet` are stored in state for later use and are not sent on create.

Create waits until status is `Active`. `Requested`, `Initiated`, and `Resizing` are still in progress. `Failed` fails the operation. Delete keeps polling through `Deleting` and finishes when GET returns 404.

## Example Usage

```terraform
variable "mssql_password" {
  type      = string
  sensitive = true
}

resource "airtelcloud_mssql" "app" {
  name              = "hello"
  database_edition  = "web"
  collation         = "SQL_Latin1_General_CP1_CI_AS"
  features          = ["SQLENGINE"]
  compute_size      = "db.mssqlweb.uhper.cci.xlarge"
  storage_size      = 200
  availability_zone = "S1"
  vpc               = "copper-vpc1"
  subnet            = "vlan-dbaas-91"
  db_name           = "AppDb"
  mssql_username    = "DbAdmin"
  password          = var.mssql_password

  backup = {
    protection_plan = "CH26_AZ1_PaaS_MSSQLIP_DAILY_FULL_2100_PP"
  }

  security_group = {
    allowed_ips = ["192.168.1.0/24"]
  }
}
```

## Schema

### Required

- `name` (String) Instance name. Lowercase letters, digits, and hyphens only.
- `database_edition` (String) Edition such as `web`, `standard`, or `enterprise`. Version comes from that edition's catalog entry.
- `collation` (String) Collation for the selected edition.
- `features` (List of String) Feature values, such as `SQLENGINE`, that the collation offers for the edition.
- `compute_size` (String) Flavor name for the selected edition.
- `storage_size` (Number) Size in GiB, from 200 to 2048 in steps of 10 starting at 200. Sent as both `database_size` and `volume_size`.
- `availability_zone` (String) Availability zone code, such as `S1`.
- `vpc` (String) VPC name. Stored only.
- `subnet` (String) Subnet name. Stored only.
- `db_name` (String) Initial database name.
- `mssql_username` (String) Admin username.
- `password` (String, Sensitive) Admin password. Also sent as `confirm_password`.

### Optional

- `description` (String)
- `storage_type` (String) Volume type label. Defaults to `High Performance`.
- `labels` (List of String) Sent as `add_labels`.
- `enable_tde` (Boolean) Defaults to `false`.
- `advanced_settings_enabled` (Boolean) Defaults to `false`. Maintenance windows are not set.
- `backup` (Attributes) Set `protection_plan` to a netbackup catalog name. Omit the block to disable backup.
- `security_group` (Attributes) `allowed_ips` is a list of CIDRs.
- `timeouts` (Block) `create` and `delete`. Default is 15 minutes.

### Read-Only

- `id` (String) Instance UUID.
- `version` (String) Version resolved from the edition.
- `status` (String)
- `connection_string` (String, Sensitive)
- `created_at` (String)
- `port` (Number)

## Import

Import is by instance UUID. Password, VPC, and subnet are not returned by the API, so a following plan will want to set them.
