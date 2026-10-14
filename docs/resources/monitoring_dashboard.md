---
page_title: "airtelcloud_monitoring_dashboard Resource - Airtel Cloud"
subcategory: "Monitoring"
description: |-
  Manages an Airtel Cloud monitoring dashboard definition.
---

# airtelcloud_monitoring_dashboard (Resource)

Manages an Airtel Cloud monitoring dashboard. Create, read, update, and delete are synchronous admin CRUD operations.

Panels are nested objects with optional nested queries. Query `query_type` selects the metrics backend (for example `prometheus`).

## Example Usage

```terraform
resource "airtelcloud_monitoring_dashboard" "compute" {
  name          = "compute-default"
  resource_type = "compute"
  active        = true
  description   = "Default compute dashboard"

  panels = [
    {
      title      = "CPU Utilization"
      chart_type = "line"
      description = "CPU over time"
      x          = 0
      y          = 0
      w          = 12
      h          = 4

      queries = [
        {
          query_template = "avg(rate(cpu_usage{instance=\"{{instance}}\"}[5m]))"
          query_type     = "prometheus"
          ref_id         = "A"
          placeholder_config = {
            instance = {
              source = "resource"
              field  = "name"
            }
          }
        }
      ]
    }
  ]
}
```

## Schema

### Required

- `name` (String) Dashboard name.
- `resource_type` (String) Resource type this dashboard applies to (for example `compute`, `mssql`).

### Optional

- `active` (Boolean) Whether the dashboard is active. Defaults to `true`.
- `description` (String) Optional dashboard description.
- `panels` (Attributes List) Dashboard panels.
  - `title` (String) Panel title. Required.
  - `chart_type` (String) One of `line`, `area`, `gauge`, `stats`, `table`. Required.
  - `description` (String) Panel description.
  - `x` (Number) Grid x position. Defaults to `0`.
  - `y` (Number) Grid y position. Defaults to `0`.
  - `w` (Number) Grid width (1-12). Defaults to `12`.
  - `h` (Number) Grid height. Defaults to `4`.
  - `min_w` (Number) Minimum panel width.
  - `min_h` (Number) Minimum panel height.
  - `display_order` (Number) Panel display order. Defaults to `0`.
  - `display_config` (Dynamic) Arbitrary display configuration object.
  - `id` (String) Panel ID from the API.
  - `queries` (Attributes List) Queries for this panel.
    - `query_template` (String) Query template with placeholders. Required.
    - `query_type` (String) Backend type such as `prometheus`, `zabbix`, or `elasticsearch`. Required.
    - `ref_id` (String) Reference ID. Defaults to `A`.
    - `display_order` (Number) Order within the panel.
    - `placeholder_config` (Dynamic) Arbitrary placeholder configuration object.
    - `series_config` (Dynamic) Arbitrary series configuration object.
    - `id` (String) Query ID from the API.

### Read-Only

- `id` (String) Dashboard ID.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

```shell
terraform import airtelcloud_monitoring_dashboard.example <dashboard-id>
```
