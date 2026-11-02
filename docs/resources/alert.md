---
page_title: "airtelcloud_alert Resource - Airtel Cloud"
subcategory: "Alerting"
description: |-
  Manages an Airtel Cloud metric alert.
---

# airtelcloud_alert (Resource)

Manages an Airtel Cloud metric alert via Alert Manager API v2.1.

Create, read, update, import, and delete are supported (synchronous CRUD). Domain and project come from the provider `organization` and `project_name`.

Customer-facing arguments match the console **Create Alert** form. Monitored VM details are resolved automatically and are not Terraform arguments.

| Console field | Terraform argument |
|---|---|
| Name | `name` |
| Description | `description` (optional) |
| Email (max 6) | `emails` |
| Resource Type | `resource_type` |
| Severity | `severity` |
| Metric Type | `metric_type` |
| Function | `function` |
| Condition | `condition` |
| Value | `value` |

### What the provider resolves for you

On create and update the provider:

1. Validates `resource_type`, `metric_type`, `function`, and `condition` against the alert static catalogs (`/api/v1/alerts/static_values/...`)
2. Maps condition symbols such as `>=` to catalog labels (for example `Greater than, equal to (>=)`)
3. Lists Active virtual machines in the project and builds the API `resources_info` payload (id, name, IP, provider instance id)
4. Sends default evaluation settings expected by the API (`evaluation_periods=5`, `datapoints_to_alarm=3`, `enabled=true`)

Terraform state keeps your configured `condition` string so plans stay stable when the API normalizes labels.

## Valid values (console)

These match the Create Alert dropdowns. Catalog-backed fields are also re-checked live at apply time.

### Severity (`severity`)

Pass the integer; the console label is shown for reference.

| Value | Console label |
|---|---|
| `0` | 0-Default |
| `1` | 1-Information |
| `2` | 2-Warning |
| `3` | 3-Average |
| `4` | 4-High |
| `5` | 5-Disaster |

### Function (`function`)

- `Average`
- `Minimum`
- `Maximum`

### Condition (`condition`)

Catalog labels (or the symbol in parentheses):

- `Greater than, equal to (>=)` (or `>=`)
- `Equal to (==)` (or `==`)

### Metric Type (`metric_type`) for `Virtual Machine`

Validated against `/api/v1/alerts/static_values/metric-types?resource_type=Virtual+Machine`. Common values include:

- `Memory Total`
- `MemoryUsedValue`
- `CPU`
- `MemoryUsed`
- `CPUAverage`
- `VMSwapOut`
- `VMSwapIN`
- `DiskThroughPutWrite`
- `AverageDiskOperationRead`
- `DiskOperationRead`
- `AverageDiskOperationWrite`

Additional metric types may be returned by the catalog.

### Resource Type (`resource_type`)

Currently the catalog returns:

- `Virtual Machine`

## Example Usage

```terraform
resource "airtelcloud_alert" "memory" {
  name          = "high-memory"
  description   = "Memory above 80%"
  emails        = ["ops@example.com"]
  resource_type = "Virtual Machine"
  severity      = 0 # 0-Default
  metric_type   = "Memory Total"
  function      = "Average"
  condition     = "Greater than, equal to (>=)"
  value         = 80
}
```

You may also use a condition symbol:

```terraform
condition = ">="
```

## Argument Reference

### Required

- `name` (String) - Alert name (1–100 characters).
- `emails` (List of String) - Notification email addresses. At least 1 and at most 6.
- `resource_type` (String) - Resource type from the catalog (for example `Virtual Machine`).
- `severity` (Number) - Severity `0`–`5` (`0` Default, `1` Information, `2` Warning, `3` Average, `4` High, `5` Disaster).
- `metric_type` (String) - Metric type for the chosen resource type (for example `Memory Total` or `CPUAverage`).
- `function` (String) - Aggregation function: `Average`, `Minimum`, or `Maximum`.
- `condition` (String) - Comparison condition. Accepts catalog labels such as `Greater than, equal to (>=)` or symbols such as `>=` / `==`.
- `value` (Number) - Threshold between `0` and `100`.

### Optional

- `description` (String) - Optional alert description.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - Alert UUID used for GET/PUT/DELETE.
- `state` (String) - Alarm state: `OK`, `ALARM`, or `INSUFFICIENT_DATA`.
- `created_at` (String) - Creation timestamp.
- `updated_at` (String) - Last update timestamp.

## Import

Import by alert UUID (`uuid_id`):

```shell
terraform import airtelcloud_alert.memory <alert-uuid>
```

Import cannot reconstruct which VMs were attached; a following apply may recreate the alert if the resolved Active VM set differs from what the API stored.
