---
page_title: "airtelcloud_protection_plan Resource - Airtel Cloud"
subcategory: "Backup"
description: |-
  Manages an Airtel Cloud Protection Plan (backup schedule and retention). Destroying this resource only removes it from Terraform state; the API does not delete plans.
---

# airtelcloud_protection_plan (Resource)

Creates a backup **protection plan**: how often backups run and how long they are kept.

You only set a name, description, schedule, retention, and an existing VM name. The provider looks up that VM, uses its availability zone (`S1`, `S2`, …) as the plan selector, and uses its subnet on create. You do not set AZ or subnet yourself.

Use the exported `id` when you attach the plan to a VM with [`airtelcloud_protection`](protection.md) or `airtelcloud_vm.protection_plan`.

| You set | Provider fills |
| --- | --- |
| `name`, `description` | API name `{name}-daily` or `{name}-weekly` |
| `recurrence` + `recurrence_period` | Interval in seconds (`daily` × 86400, `weekly` × 604800) |
| `retention` (days) | `retention_unit` = `DAYS` |
| `vm_name` | `selector_value` (AZ), `subnet_id` (create header only) |

~> **Destroy does not delete the plan.** The backup API has no delete. `terraform destroy` only removes the resource from state. The plan stays in the console. Change any argument and Terraform **replaces** the resource (creates another plan; the old one remains).

## Example Usage

Copy this into a working directory. Put secrets in `terraform.tfvars` or environment variables — never in the `.tf` file.

```terraform
terraform {
  required_providers {
    airtelcloud = {
      source = "Airtel-Cloud-Platform/airtelcloud"
    }
  }
}

provider "airtelcloud" {
  api_endpoint = "https://south.cloud.airtel.in"
  api_key      = var.airtel_api_key
  api_secret   = var.airtel_api_secret
  region       = "south"
  organization = var.organization
  project_name = var.project_name
}

variable "airtel_api_key" {
  type      = string
  sensitive = true
}

variable "airtel_api_secret" {
  type      = string
  sensitive = true
}

variable "organization" {
  type = string
}

variable "project_name" {
  type = string
}

# Existing VM instance_name (must already exist in this project).
variable "vm_name" {
  type = string
}

resource "airtelcloud_protection_plan" "weekly" {
  name              = "app-backup"
  description       = "Weekly backup, 30-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "weekly"
  retention         = 30
}

output "protection_plan_id" {
  description = "UUID to pass to airtelcloud_protection or a VM backup block"
  value       = airtelcloud_protection_plan.weekly.id
}
```

`terraform.tfvars` (local only; do not commit):

```hcl
organization  = "your-org"
project_name  = "your-project"
vm_name       = "your-existing-vm"
# airtel_api_key    = "..."
# airtel_api_secret = "..."
```

### Daily plan

```terraform
resource "airtelcloud_protection_plan" "daily" {
  name              = "app-backup"
  description       = "Daily backup, 30-day retention"
  vm_name           = var.vm_name
  recurrence        = 1
  recurrence_period = "daily"
  retention         = 30
}
```

With `name = "app-backup"`, the console/API name is `app-backup-daily` or `app-backup-weekly`. Terraform state keeps `app-backup`.

### Use the plan on a protection policy

```terraform
resource "airtelcloud_protection" "web" {
  name             = "web-backup"
  description      = "Attach weekly plan to the VM"
  compute_name     = var.vm_name
  protection_plan  = airtelcloud_protection_plan.weekly.id
  enable_scheduler = "true"
  weekday          = "monday"
  start_time       = "02:00"
}
```

Pass **`id`**, not `name`. See [`airtelcloud_protection`](protection.md).

## Recurrence

`recurrence` is how many periods between runs. Allowed range: **1–30**. Retention: **1–365** days. Invalid values fail at plan time; the backup API is not called.

| `recurrence_period` | `recurrence` | Meaning | Seconds sent to the API |
| --- | --- | --- | --- |
| `daily` | `1` | Every day | 86400 |
| `daily` | `7` | Every 7 days | 604800 |
| `weekly` | `1` | Every week | 604800 |
| `weekly` | `2` | Every 2 weeks | 1209600 |

## Argument Reference

All arguments force replacement if changed.

### Required

- `name` (String) - Plan name in Terraform. The API receives `{name}-daily` or `{name}-weekly`. Do not add that suffix unless you want it left as-is.
- `description` (String) - Description stored on the plan.
- `vm_name` (String) - Existing VM `instance_name` in the same project. AZ and subnet are read from this VM. The VM is not modified.
- `recurrence` (Number) - Count of `recurrence_period` units, 1–30.
- `recurrence_period` (String) - `daily` or `weekly`.
- `retention` (Number) - How many days to keep backups, 1–365.

### You do not set

- `selector_key` — always `AZ`.
- `retention_unit` — always `DAYS`.
- Availability zone and subnet — taken from `vm_name`.

## Attribute Reference

- `id` (String) - Plan UUID. Use this when attaching the plan to a VM.
- `selector_value` (String) - Availability zone from the VM (for example `S1`).
- `subnet_id` (String) - Subnet from the VM (`network_id`). Used on create; list/read use the AZ header.

## Import

Import by the plan UUID from the console (not the display name):

```shell
terraform import airtelcloud_protection_plan.weekly 898a41cb-8ce0-48e2-861a-8370f6799819
```

Import only restores `id`. Put the same `name`, `description`, `vm_name`, schedule, and retention in config so the next plan is empty.
