---
page_title: "airtelcloud_protection Resource - Airtel Cloud"
subcategory: "Backup"
description: |-
  Manages an Airtel Cloud backup protection policy for a VM.

# airtelcloud_protection (Resource)

Manages a backup protection policy for an existing Airtel Cloud VM.

A protection policy associates a VM with a protection plan and defines when the backup policy starts running.

The VM and protection plan must already exist before creating the protection policy.

## Example Usage

### Using a Start Date

```terraform
resource "airtelcloud_protection" "web" {
  vm_name         = "my-app-vm"
  protection_plan = "my-weekly-backup"
  start_date      = "10/06/2026"
  start_time      = "02:00 AM"
}
```

### Using a Weekday

```terraform
resource "airtelcloud_protection" "web" {
  vm_name         = "my-app-vm"
  protection_plan = "my-weekly-backup"
  weekday         = "monday"
  start_time      = "02:00 AM"
}
```

In both examples:

- `my-app-vm` must be an existing VM.
- `my-weekly-backup` must be an existing protection plan.
- Either `start_date` or `weekday` must be specified.
- `start_date` and `weekday` cannot be specified together.
- `start_time` must be specified.

## Arguments Reference

### Required Arguments

### `vm_name`

**Required** — String

Name of an existing Airtel Cloud VM.

The VM must exist in the same project.

### `protection_plan`

**Required** — String

Name of an existing Airtel Cloud protection plan.

Example:

```terraform
protection_plan = "my-weekly-backup"
```

### `start_time`

**Required** — String

Time at which the protection policy should start.

Supported formats:

- `HH:MM`
- `H:MM AM/PM`

Example:

```terraform
start_time = "02:00 AM"
```

The configured time is interpreted as India Standard Time (IST).

### `start_date`

**Required when `weekday` is not specified** — String

Date on which the protection policy should start.

The date must be specified in `MM/DD/YYYY` format.

Example:

```terraform
start_date = "10/06/2026"
```

`start_date` and `weekday` are mutually exclusive. You must specify exactly one of them.

### `weekday`

**Required when `start_date` is not specified** — String

Day of the week on which the protection policy should start.

Supported values:

- `monday`
- `tuesday`
- `wednesday`
- `thursday`
- `friday`
- `saturday`
- `sunday`

Short forms such as `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, and `sun` are also supported.

Example:

```terraform
weekday = "monday"
```

`weekday` and `start_date` are mutually exclusive. You must specify exactly one of them.

## Optional Arguments

### `description`

**Optional** — String

Description of the backup protection policy.

Example:

```terraform
description = "Backup policy for application VM"
```

### `policy_type_id`

**Optional** — String

Identifier of the policy type.

If not specified, the default policy type is used.

### `enable_scheduler`

**Optional** — String

Scheduler configuration maintained for compatibility with existing configurations.

Example:

```terraform
enable_scheduler = "true"
```

## Scheduling

A protection policy must have a start schedule.

You can configure the schedule using either a specific date or a weekday.

### Start Date

Use `start_date` when the policy should begin on a specific date.

```terraform
resource "airtelcloud_protection" "web" {
  vm_name         = "my-app-vm"
  protection_plan = "my-weekly-backup"
  start_date      = "10/06/2026"
  start_time      = "02:00 AM"
}
```

### Weekday

Use `weekday` when the policy should begin on a specific day of the week.

```terraform
resource "airtelcloud_protection" "web" {
  vm_name         = "my-app-vm"
  protection_plan = "my-weekly-backup"
  weekday         = "monday"
  start_time      = "02:00 AM"
}
```

> **Note:** `start_date` and `weekday` cannot be used together. Exactly one of them must be specified.

## Attribute Reference

In addition to the arguments above, the following attributes are exported.

### `id`

Unique ID of the protection policy.

### `name`

Name of the protection policy.

The policy name corresponds to the VM name.

### `compute_id`

Unique ID of the VM associated with the protection policy.

### `status`

Current status of the protection policy.

### `region`

Region associated with the protection policy.

### `az_name`

Availability zone associated with the protection policy.

### `created`

Creation timestamp of the protection policy.

## Import

An existing protection policy can be imported using its ID.

```shell
terraform import airtelcloud_protection.web <protection-id>
```

After importing the resource, configure `vm_name`, `protection_plan`, and the scheduling arguments to match the existing protection policy.

## Notes

- `vm_name` is required and must refer to an existing VM.
- `protection_plan` is required and must refer to an existing protection plan.
- `start_time` is required.
- Either `start_date` or `weekday` must be specified.
- `start_date` and `weekday` cannot be specified together.
- `start_date` must use the `MM/DD/YYYY` format.
- `start_time` is interpreted in IST.
- The protection plan is specified using its name.