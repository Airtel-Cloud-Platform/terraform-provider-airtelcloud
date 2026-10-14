---
page_title: "airtelcloud_alert Resource - Airtel Cloud"
subcategory: "Alerting"
description: |-
  Manages an Airtel Cloud metric alert.
---

# airtelcloud_alert (Resource)

Manages an Airtel Cloud metric alert via Alert Manager API v2.1.

Create, read, update, import, and delete are supported (sync CRUD; no long poll). Domain and project path segments come from the provider `organization` and `project_name`. The API returns an integer alert id; Terraform stores it as a string.

`resources_info` is force-new. Changing it replaces the alert. On update, `evaluation_periods` and `datapoints_to_alarm` are required by the API.

## Example Usage

```terraform
resource "airtelcloud_alert" "cpu" {
  name                 = "high-cpu-vm1"
  description          = "CPU above 80%"
  resource_type_name   = "compute"
  metric_type          = "cpu"
  metric_name          = "cpu_utilization"
  threshold            = 80
  comparison_operator  = ">"
  evaluation_periods   = "5"
  datapoints_to_alarm  = "3"
  enabled              = true
  severity             = 2
  alert_emails         = ["ops@example.com"]

  resources_info = [
    {
      resource_id          = airtelcloud_vm.app.id
      resource_name        = airtelcloud_vm.app.name
      resource_ip          = "10.0.0.10"
      provider_instance_id = airtelcloud_vm.app.id
    }
  ]
}
```

## Schema

### Required

- `name` (String) Alert name.
- `resource_type_name` (String) Resource type name (for example `compute`).
- `metric_type` (String) Metric type (for example `cpu`).
- `metric_name` (String) Metric name (for example `cpu_utilization`).
- `threshold` (Number) Threshold between 0 and 100.
- `comparison_operator` (String) One of `>`, `>=`, `<`, `<=`, `=`, `!=`.
- `severity` (Number) Alert severity.
- `alert_emails` (List of String) Notification email addresses.
- `resources_info` (Attributes List) Resources to monitor. Force-new.
  - `resource_id` (String)
  - `resource_name` (String)
  - `resource_ip` (String)
  - `provider_instance_id` (String)

### Optional

- `description` (String)
- `evaluation_periods` (String) Required when updating.
- `datapoints_to_alarm` (String) Required when updating.
- `enabled` (Boolean) Defaults to `true`.

### Read-Only

- `id` (String) Alert id as a string (API integer).
- `state` (String) `OK`, `ALARM`, or `INSUFFICIENT_DATA`.
- `created_at` (String)
- `updated_at` (String)

## Import

Import is by alert id string (the integer id as text):

```shell
terraform import airtelcloud_alert.cpu 101
```

GET may return a list-show payload without `id`; Terraform keeps the imported id and refreshes returned fields. `resources_info` may not round-trip from the API.
