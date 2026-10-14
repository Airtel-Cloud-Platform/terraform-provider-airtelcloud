---
page_title: "airtelcloud_monitoring_provider Resource - Airtel Cloud"
subcategory: "Monitoring"
description: |-
  Manages an Airtel Cloud monitoring metrics provider.
---

# airtelcloud_monitoring_provider (Resource)

Manages an Airtel Cloud monitoring metrics provider (Prometheus, Zabbix, or Elasticsearch) for a resource type.

`api_token` is sensitive, required on create, and never returned by the API. On update, leave it unchanged or empty in the request path so the existing token is kept. Terraform state preserves the configured token across reads.

## Example Usage

```terraform
variable "monitoring_api_token" {
  type      = string
  sensitive = true
}

resource "airtelcloud_monitoring_provider" "compute" {
  name               = "south-prometheus-compute"
  resource_type      = "compute"
  provider_type      = "prometheus"
  dashboard_auth_url = "https://grafana.example.com"
  api_key_name       = "terraform-provider"
  api_token          = var.monitoring_api_token
  data_source_uid    = "prom-uid-1"
  data_source_id     = "1"
  az_name            = "S1"
  active             = true
  description        = "Prometheus for compute in S1"
}
```

## Schema

### Required

- `name` (String) Provider name.
- `resource_type` (String) Resource type this provider serves (for example `compute`, `mssql`).
- `dashboard_auth_url` (String) Dashboard authentication URL.
- `api_key_name` (String) API key name.
- `api_token` (String, Sensitive) API token. Required on create. Not returned on read.
- `data_source_uid` (String) Data source UID.
- `data_source_id` (String) Data source ID.

### Optional

- `provider_type` (String) One of `prometheus`, `zabbix`, `elasticsearch`.
- `az_name` (String) Optional availability zone name.
- `active` (Boolean) Whether the provider is active. Defaults to `true`.
- `description` (String) Optional provider description.

### Read-Only

- `id` (String) Provider ID.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

## Import

Import is by provider ID. `api_token` is not returned by the API, so a following plan will want to set it.

```shell
terraform import airtelcloud_monitoring_provider.example <provider-id>
```
