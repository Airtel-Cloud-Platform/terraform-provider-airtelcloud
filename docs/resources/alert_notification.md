---
page_title: "airtelcloud_alert_notification Resource - Airtel Cloud"
subcategory: "Alerting"
description: |-
  Manages an Airtel Cloud alert notification channel.
---

# airtelcloud_alert_notification (Resource)

Manages an email or SMS notification channel bound to an existing alert.

`alert_id` and `type` are force-new. `alert_id` is the string form of the integer alert id (for example from `airtelcloud_alert.id`).

## Example Usage

```terraform
resource "airtelcloud_alert_notification" "ops_email" {
  alert_id = airtelcloud_alert.cpu.id
  type     = "email"
  target   = "ops@example.com"
  enabled  = true
}
```

## Schema

### Required

- `alert_id` (String) Alert id (string form of integer Alert.id). Force-new.
- `type` (String) `email` or `sms`. Force-new.
- `target` (String) Destination address or phone number.

### Optional

- `enabled` (Boolean) Defaults to `true`.

### Read-Only

- `id` (String) Notification id as a string (API integer).
- `created_at` (String)

## Import

```shell
terraform import airtelcloud_alert_notification.ops_email 55
```
