---
page_title: "airtelcloud_dns_record Resource - Airtel Cloud"
subcategory: "DNS"
description: |-
  Manages an Airtel Cloud DNS record within a DNS zone.
---

# airtelcloud_dns_record (Resource)

Manages a DNS record within an Airtel Cloud DNS zone. Supports A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, PTR, and other record types.

## Example Usage

### A Record

```terraform
resource "airtelcloud_dns_record" "web" {
  zone_id     = airtelcloud_dns_zone.example.id
  owner       = "www"
  record_type = "A"
  data        = "192.168.1.10"
  ttl         = 300
  description = "Web server A record"
}
```

### AAAA Record

```terraform
resource "airtelcloud_dns_record" "web_ipv6" {
  zone_id     = airtelcloud_dns_zone.example.id
  owner       = "www"
  record_type = "AAAA"
  data        = "2001:db8::1"
  ttl         = 300
  description = "Web server IPv6 record"
}
```

### CNAME Record

```terraform
resource "airtelcloud_dns_record" "blog" {
  zone_id     = airtelcloud_dns_zone.example.id
  owner       = "blog"
  record_type = "CNAME"
  data        = "www.example.com."
  ttl         = 3600
  description = "Blog CNAME pointing to www"
}
```

### MX Record

```terraform
resource "airtelcloud_dns_record" "mail" {
  zone_id     = airtelcloud_dns_zone.example.id
  owner       = "@"
  record_type = "MX"
  data        = "mail.example.com."
  ttl         = 3600
  preference  = 10
  description = "Primary mail server"
}
```

### TXT Record (SPF)

```terraform
resource "airtelcloud_dns_record" "spf" {
  zone_id     = airtelcloud_dns_zone.example.id
  owner       = "@"
  record_type = "TXT"
  data        = "v=spf1 include:_spf.example.com ~all"
  ttl         = 3600
  description = "SPF record for email authentication"
}
```

### Using the exported attributes

```terraform
output "record_id" {
  value = airtelcloud_dns_record.web.id
}

output "record_details" {
  value = {
    zone_name  = airtelcloud_dns_record.web.zone_name
    org_name   = airtelcloud_dns_record.web.org_name
    org_id     = airtelcloud_dns_record.web.org_id
    created_by = airtelcloud_dns_record.web.created_by
    created_at = airtelcloud_dns_record.web.created_at
    updated_at = airtelcloud_dns_record.web.updated_at
  }
}
```

## Argument Reference

### Required

- `zone_id` (String) - The UUID of the DNS zone this record belongs to. Forces new resource if changed.
- `record_type` (String) - The type of DNS record. Valid values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `PTR`. Use upper case. Forces new resource if changed.

### Optional

- `owner` (String) - The owner/name of the DNS record (e.g., `www`, `@` for apex). Use the name relative to the zone (for example `www`, not `www.example.com.`). The provider stores the relative name, removing any trailing `.` and `.<zone_name>` that the API returns.
- `data` (String) - The data/value of the DNS record (e.g., IP address for A record, hostname for CNAME).
- `ttl` (Number) - The Time-To-Live in seconds. Default: `300`.
- `description` (String) - A description for the DNS record.
- `preference` (Number) - The preference/priority value (used for MX records).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The unique identifier (UUID) of the DNS record.
- `zone_name` (String) - The name of the DNS zone this record belongs to.
- `org_name` (String) - The organization name associated with the record.
- `org_id` (String) - The organization ID associated with the record.
- `created_by` (String) - The user who created the record.
- `created_at` (Number) - The creation timestamp (Unix timestamp).
- `updated_at` (Number) - The last updated timestamp (Unix timestamp).

## Import

DNS records are imported using the zone ID and record ID separated by a slash (`<zone-id>/<record-id>`):

```shell
terraform import airtelcloud_dns_record.web <zone-id>/<record-id>
```
