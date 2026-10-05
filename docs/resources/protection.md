---
page_title: "airtelcloud_protection Resource - Airtel Cloud"
subcategory: "Backup"
description: |-
  Attaches an Airtel Cloud protection plan to a VM (backup policy). Policy name is the VM name; compute_id is looked up from that name.
---

# airtelcloud_protection (Resource)

Attaches a protection plan to a VM so backups actually run.

This matches the console create form (`POST .../backups/protections/`):

| Form field | Source |
| --- | --- |
| `name` | `vm_name` (the VM `instance_name`) |
| `compute_id` | VM UUID looked up from `vm_name` |
| `protection_plan` | Plan UUID (names are resolved) |
| `start_date` | `MM/DD/YYYY` |
| `start_time` | `H:MM AM/PM` |
| `description` | your description |

You do **not** set `name` or `compute_id`. Do not paste API keys or Bearer tokens in Terraform.

## Example Usage

```terraform
resource "airtelcloud_protection" "web" {
  description     = "Backup policy for web server"
  vm_name         = "dbcorptest"
  protection_plan = airtelcloud_protection_plan.weekly.id
  start_date      = "10/06/2026"
  start_time      = "02:00 AM"
}
```

`protection_plan` also accepts a console name (resolved at apply). Prefer `id`.

```terraform
resource "airtelcloud_protection" "web" {
  description     = "Backup policy for web server"
  vm_name         = "dbcorptest"
  protection_plan = "S1-ELEMENTS-COPPER-WEEKLYTEER-BKP-PP"
  start_date      = "2026-10-06"
  start_time      = "02:00"
}
```

ISO `YYYY-MM-DD` and 24-hour `HH:MM` are converted to the console formats.

### Weekday instead of a calendar date

```terraform
resource "airtelcloud_protection" "web" {
  description     = "Backup policy for web server"
  vm_name         = var.vm_name
  protection_plan = airtelcloud_protection_plan.weekly.id
  weekday         = "monday"
  start_time      = "02:00"
}
```

`weekday` is converted to the next matching date in IST. Do not set `start_date` and `weekday` together.

## Argument Reference

### Required

- `vm_name` (String) - Existing VM `instance_name`. Sent as form `name`. Forces replacement if changed.
- `protection_plan` (String) - Plan UUID or console name. Prefer `airtelcloud_protection_plan.example.id`.

### Optional

- `description` (String) - Policy description.
- `start_date` (String) - First run date: `YYYY-MM-DD` or `MM/DD/YYYY`. Mutually exclusive with `weekday`.
- `weekday` (String) - `monday`…`sunday` (or `mon`…`sun`). Mutually exclusive with `start_date`.
- `start_time` (String) - IST time: `HH:MM` or `H:MM AM/PM` (for example `02:00 AM`).
- `end_date` (String) - Optional end date, same formats as `start_date`.
- `policy_type_id` (String) - Optional policy type. Omitted unless set.
- `enable_scheduler` (String) - Kept for compatibility. The console create form does not send this field.

### You do not set

- `name` — always the VM name.
- `compute_id` — looked up from `vm_name`.

## Attribute Reference

- `id` (String) - Protection policy ID.
- `name` (String) - Same as `vm_name`.
- `compute_id` (String) - VM UUID sent to the API.
- `status` (String) - Policy status.
- `region` (String)
- `az_name` (String)
- `created` (String)

## Import

```shell
terraform import airtelcloud_protection.web <protection-id>
```

Import restores `id`. Set `vm_name`, `protection_plan`, and schedule in config to match.
