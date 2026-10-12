---
page_title: "airtelcloud_mssql Resource - Airtel Cloud"
subcategory: "Database"
description: |-
  Manages an Airtel Cloud MSSQL instance.
---

# airtelcloud_mssql (Resource)

Manages an Airtel Cloud MSSQL instance.

Create, read, import, and delete are supported. Changing configuration forces a new instance. The API does not return the admin password. Set `custom_cidr`, or `vpc` and `subnet`, or both. Network names are not sent on create; subnet CIDR is resolved into API `security_group.allowed_ips`.

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
  custom_cidr       = ["192.168.1.0/24"]
  db_name           = "AppDb"
  mssql_username    = "DbAdmin"
  password          = var.mssql_password

  backup = {
    protection_plan = "CH26_AZ1_PaaS_MSSQLIP_DAILY_FULL_2100_PP"
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
- `db_name` (String) Initial database name.
- `mssql_username` (String) Admin username.
- `password` (String, Sensitive) Admin password. Also sent as `confirm_password`.

### Optional

- `description` (String) Optional in Terraform; when omitted, sent to the API as `""`.
- `vpc` (String) VPC name used to resolve `subnet`. Required with `subnet` when `custom_cidr` is omitted. Not sent on create.
- `subnet` (String) Subnet name within `vpc`. Required with `vpc` when `custom_cidr` is omitted. Resolved `ipv4AddressSpace` is sent as an allowed IP. Not sent on create.
- `custom_cidr` (List of String) Custom CIDRs sent as API `security_group.allowed_ips`. Required when `vpc`/`subnet` are omitted. Can be set together with `vpc`/`subnet`.
- `storage_type` (String) Volume type label. Defaults to `High Performance`.
- `labels` (List of String) Sent as `add_labels`.
- `enable_tde` (Boolean) Defaults to `false`.
- `advanced_settings_enabled` (Boolean) Defaults to `false`. Maintenance windows are not set.
- `backup` (Attributes) Set `protection_plan` to a netbackup catalog name. Omit the block to disable backup.
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
