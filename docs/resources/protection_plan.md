---
page_title: "airtelcloud_protection_plan Resource - Airtel Cloud"
subcategory: "Backup"
description: |-
  Manages an Airtel Cloud Protection Plan for scheduling VM backups and defining backup retention.

# airtelcloud_protection_plan (Resource)

Manages an Airtel Cloud Protection Plan in Airtel Cloud.

A protection plan defines:

- How frequently VM backups are created.
- How long backups are retained.
- Which existing VM the protection plan is associated with.

The VM specified by `vm_name` must already exist in the same Airtel Cloud project.

> **Destroy:** Terraform calls `DELETE` on the protection plan. If the plan is still in the console afterward, the next `apply` adopts it instead of failing with “already exists”.

## Example Usage

### Weekly Backup

```terraform
resource "airtelcloud_protection_plan" "weekly" {
  name              = "app-backup"
  description       = "Weekly backup with 30-day retention"
  vm_name           = "my-app-vm"
  recurrence        = 1
  recurrence_period = "weekly"
  retention         = 30
}
```

### Daily Backup

```terraform
resource "airtelcloud_protection_plan" "daily" {
  name              = "app-backup"
  description       = "Daily backup with 30-day retention"
  vm_name           = "my-app-vm"
  recurrence        = 1
  recurrence_period = "daily"
  retention         = 30
}
```

In the examples above, `my-app-vm` must be the `instance_name` of an existing VM in the same project.

## Arguments Reference

The following arguments are supported.

### `name`

**Required** — String

Name of the protection plan.

### `description`

**Required** — String

Description of the protection plan.

### `vm_name`

**Required** — String

Name of an existing VM in the same Airtel Cloud project.

The VM is used to associate the protection plan with the appropriate VM configuration.

### `recurrence`

**Required** — Number

Number of recurrence periods between backups.

Valid values are `1` to `30`.

For example:

```terraform
recurrence        = 1
recurrence_period = "daily"
```

creates a backup every day.

```terraform
recurrence        = 2
recurrence_period = "weekly"
```

creates a backup every two weeks.

### `recurrence_period`

**Required** — String

Defines the period used by `recurrence`.

Supported values:

- `daily`
- `weekly`

### `retention`

**Required** — Number

Number of days that backups are retained.

Valid values are `1` to `365`.

For example:

```terraform
retention = 30
```

retains backups for 30 days.

## Recurrence Examples

| `recurrence_period` | `recurrence` | Backup Frequency |
|---|---:|---|
| `daily` | `1` | Every day |
| `daily` | `2` | Every 2 days |
| `daily` | `7` | Every 7 days |
| `weekly` | `1` | Every week |
| `weekly` | `2` | Every 2 weeks |

The effective backup interval is calculated from `recurrence` and `recurrence_period`.

## Attribute Reference

In addition to the arguments above, the following attributes are exported.

### `id`

The unique ID (UUID) of the protection plan.

### `selector_value`

The availability zone associated with the VM used to create the protection plan.

### `subnet_id`

The network ID associated with the VM.

## Replacement Behavior

Changing any protection plan argument causes Terraform to replace the resource by creating a new protection plan.

The following arguments trigger replacement when changed:

- `name`
- `description`
- `vm_name`
- `recurrence`
- `recurrence_period`
- `retention`

Because the Airtel Cloud backup API does not support deleting protection plans, the previous plan remains in Airtel Cloud after replacement.

> **Note:** Avoid making unnecessary changes to protection plan configuration because old plans cannot be removed automatically by Terraform.

## Import

An existing protection plan can be imported using its UUID.

```shell
terraform import airtelcloud_protection_plan.weekly 898a41cb-8ce0-48e2-861a-8370f6799819
```

The UUID can be obtained from Airtel Cloud.

After importing the resource, configure the resource arguments to match the existing protection plan.

## Notes

- `vm_name` must refer to an existing VM.
- The VM must belong to the same Airtel Cloud project.
- `recurrence` must be between `1` and `30`.
- `recurrence_period` must be either `daily` or `weekly`.
- `retention` must be between `1` and `365` days.
- Changes to resource arguments create a new protection plan.
- Destroy calls DELETE. If the plan remains in the console, re-apply adopts it.
- Destroying the Terraform resource removes it from Terraform state only; the protection plan remains in Airtel Cloud.