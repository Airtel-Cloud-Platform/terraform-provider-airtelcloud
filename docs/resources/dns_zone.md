---
page_title: "airtelcloud_dns_zone Resource - Airtel Cloud"
subcategory: "DNS"
description: |-
  Manages an Airtel Cloud DNS zone (DNSaaS).
---

# airtelcloud_dns_zone (Resource)

Manages an Airtel Cloud DNS zone. DNS zones are the foundation for managing DNS records.

## Example Usage

```terraform
resource "airtelcloud_dns_zone" "example" {
  zone_name   = "example.com"
  zone_type   = "forward"
  description = "Primary DNS zone for example.com"
}

resource "airtelcloud_dns_zone" "internal" {
  zone_name   = "internal.example.com"
  zone_type   = "forward"
  description = "Internal services DNS zone"
}
```

## Argument Reference

### Required

- `zone_name` (String) - The name of the DNS zone (e.g., `example.com`).

### Optional

- `zone_type` (String) - The type of DNS zone. Valid values: `forward`, `reverse`. Default: `forward`.
- `description` (String) - A description for the DNS zone.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier (UUID) of the DNS zone.
- `dns_zone_template` (String) - The DNS zone template.
- `org_name` (String) - The organization name associated with the zone.
- `org_id` (String) - The organization ID associated with the zone.
- `created_by` (String) - The user who created the zone.
- `created_at` (Number) - The creation timestamp (Unix timestamp).
- `updated_at` (Number) - The last updated timestamp (Unix timestamp).

## Import

DNS zones can be imported using the `id`:

```shell
terraform import airtelcloud_dns_zone.example <zone-id>
```
