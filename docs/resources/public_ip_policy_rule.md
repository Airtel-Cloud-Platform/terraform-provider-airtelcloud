---
page_title: "airtelcloud_public_ip_policy_rule Resource - Airtel Cloud"
subcategory: "Security"
description: |-
  Manages a NAT Gateway policy rule on an attached Airtel Cloud Public IP.
---

# airtelcloud_public_ip_policy_rule (Resource)

Manages a policy rule on an Airtel Cloud Public IP (NAT Gateway). A rule decides which sources may reach which services through the public address.

This is step 3 of 3. The parent public IP must already be attached to a VM, load balancer, or baremetal server. Reserve it with [`airtelcloud_public_ip`](public_ip.md) and attach it with [`airtelcloud_public_ip_attachment`](public_ip_attachment.md) first.

Sources and services can be declared in two ways:

| Form | Arguments | Use when |
| --- | --- | --- |
| Simple | `source`, `services` | A single source and a plain list of service names. |
| Detailed | `source_config`, `service_config` | Multiple sources, CIDR plus country selectors, or explicit service flags. |

## Example Usage

### Allow Web Traffic from Anywhere

```terraform
resource "airtelcloud_public_ip_policy_rule" "web_traffic" {
  public_ip_name = airtelcloud_public_ip_attachment.vm.public_ip_name
  rule_name      = "web-traffic"
  source         = "any"
  services       = ["HTTP", "HTTPS"]
  action         = "accept"
}
```

### Allow SSH from a Single Management IP

```terraform
resource "airtelcloud_public_ip_policy_rule" "ssh_mgmt" {
  public_ip_name = airtelcloud_public_ip_attachment.vm.public_ip_name
  rule_name      = "ssh-mgmt"
  source         = "192.168.100.5/32"
  services       = ["SSH"]
  action         = "accept"
}
```

### Combine CIDR and Country Sources

Use `source_config` when a rule has more than one source, or when a source is a country rather than a CIDR.

```terraform
resource "airtelcloud_public_ip_policy_rule" "office_and_india" {
  public_ip_name = airtelcloud_public_ip_attachment.vm.public_ip_name
  rule_name      = "office-and-india"
  action         = "accept"

  source_config = [
    {
      create_new  = false
      ip_cidr     = "182.77.78.18/32"
      source_type = "ip_cidr"
    },
    {
      create_new  = false
      ip_cidr     = "122.161.51.26/32"
      source_type = "ip_cidr"
    },
    {
      source_type = "geographic"
      geographic = {
        country_code = "IN"
        country_name = "India"
      }
    }
  ]

  service_config = [
    {
      create_new = false
      name       = "HTTPS"
      is_default = false
    },
    {
      create_new = false
      name       = "tcp-5601-5601"
      is_default = false
    }
  ]
}
```

### Deny a Country

```terraform
resource "airtelcloud_public_ip_policy_rule" "block_country" {
  public_ip_name = airtelcloud_public_ip_attachment.vm.public_ip_name
  rule_name      = "block-wf"
  action         = "deny"

  source_config = [
    {
      source_type = "geographic"
      geographic = {
        country_code = "WF"
        country_name = "Wallis and Futuna"
      }
    }
  ]

  services = ["ALL"]
}
```

## Argument Reference

### Required

- `public_ip_name` (String) - The `object_name` of the parent public IP. It must already be attached to a workload. Forces replacement if changed.
- `rule_name` (String) - The name of the policy rule. Forces replacement if changed.
- `action` (String) - `accept` or `deny`. Forces replacement if changed.

One of `source` or `source_config` is required. One of `services` or `service_config` is required.

### Optional

- `source` (String) - A single source: `any`, `all`, or a CIDR such as `182.77.78.18/32`. Forces replacement if changed.
- `source_config` (List of Object) - Detailed source entries. Takes precedence over `source`. See [source_config](#source_config).
- `services` (List of String) - Service names to match, such as `["HTTP", "HTTPS"]`. Use `["ALL"]` for every service.
- `service_config` (List of Object) - Detailed service entries. Takes precedence over `services`. See [service_config](#service_config).
- `resource_type` (String) - Resource type sent to the policy API. Defaults to `ipam`.
- `revision_note` (String) - Revision note stored with the rule. Defaults to `creating Policy`.
- `target_vip` (String) - The private IP the rule applies to. Read from the parent public IP when omitted. Forces replacement if changed.
- `public_ip` (String) - The public address the rule applies to. Read from the parent public IP when omitted. Forces replacement if changed.
- `availability_zone` (String) - The availability zone (for example `S1`). Read from the parent public IP when omitted. Forces replacement if changed.

### source_config

Each element supports:

- `source_type` (String) - `ip_cidr`, `geographic`, or `all`.
- `ip_cidr` (String) - The source CIDR. Required when `source_type` is `ip_cidr`.
- `geographic` (Object) - Country selector. Required when `source_type` is `geographic`.
  - `country_code` (String) - ISO 3166-1 alpha-2 code, such as `IN`. See [Country Codes](#country-codes).
  - `country_name` (String) - Country name, such as `India`.
- `create_new` (Boolean) - Whether the API should create a new source selector. Defaults to `false`. Sent for `ip_cidr` and `all` sources only; it is omitted for `geographic` sources.

When `source_type` is omitted, the provider infers it: `geographic` if a `geographic` block is present, `ip_cidr` if `ip_cidr` is set, otherwise `all`.

### service_config

Each element supports:

- `name` (String, Required) - The service name, such as `HTTPS`, `SSH`, or a port-range service like `tcp-5601-5601`.
- `create_new` (Boolean) - Whether the API should create a new service selector. Defaults to `false`.
- `is_default` (Boolean) - Whether the service is marked as default. Defaults to `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` (String) - The UUID of the policy rule.
- `public_ip_id` (String) - The UUID of the parent public IP, resolved from `public_ip_name`.
- `state` (String) - The current state of the rule, for example `Applied`.

## Import

Policy rules are imported using `public_ip_id/target_vip/public_ip/rule_id`:

```shell
terraform import airtelcloud_public_ip_policy_rule.web_traffic <public_ip_id>/<target_vip>/<public_ip>/<rule_id>
```

## Important Notes

- The parent public IP must be attached to a VM, load balancer, or baremetal server. Creating a rule against a `reserved` address fails with a "Public IP Not Attached" error. Point `public_ip_name` at `airtelcloud_public_ip_attachment.*.public_ip_name` so Terraform waits for attach without `depends_on`.
- This resource has no in-place update. Any change to a rule destroys and recreates it.
- `public_ip`, `target_vip`, and `availability_zone` default to the values on the parent public IP, so you rarely set them.
- When both forms are supplied, `source_config` and `service_config` win over `source` and `services`.
- `country_code` is what the API matches on. `country_name` is the label shown alongside it and should match the portal's spelling.

## Country Codes

Values for `source_config.geographic`. `country_code` is the ISO 3166-1 alpha-2 code; `country_name` is the matching ISO English short name.

| Country name | `country_code` |
| --- | --- |
| Afghanistan | `AF` |
| Åland Islands | `AX` |
| Albania | `AL` |
| Algeria | `DZ` |
| American Samoa | `AS` |
| Andorra | `AD` |
| Angola | `AO` |
| Anguilla | `AI` |
| Antarctica | `AQ` |
| Antigua and Barbuda | `AG` |
| Argentina | `AR` |
| Armenia | `AM` |
| Aruba | `AW` |
| Australia | `AU` |
| Austria | `AT` |
| Azerbaijan | `AZ` |
| Bahamas | `BS` |
| Bahrain | `BH` |
| Bangladesh | `BD` |
| Barbados | `BB` |
| Belarus | `BY` |
| Belgium | `BE` |
| Belize | `BZ` |
| Benin | `BJ` |
| Bermuda | `BM` |
| Bhutan | `BT` |
| Bolivia, Plurinational State of | `BO` |
| Bonaire, Sint Eustatius and Saba | `BQ` |
| Bosnia and Herzegovina | `BA` |
| Botswana | `BW` |
| Bouvet Island | `BV` |
| Brazil | `BR` |
| British Indian Ocean Territory | `IO` |
| Brunei Darussalam | `BN` |
| Bulgaria | `BG` |
| Burkina Faso | `BF` |
| Burundi | `BI` |
| Cabo Verde | `CV` |
| Cambodia | `KH` |
| Cameroon | `CM` |
| Canada | `CA` |
| Cayman Islands | `KY` |
| Central African Republic | `CF` |
| Chad | `TD` |
| Chile | `CL` |
| China | `CN` |
| Christmas Island | `CX` |
| Cocos (Keeling) Islands | `CC` |
| Colombia | `CO` |
| Comoros | `KM` |
| Congo | `CG` |
| Congo, The Democratic Republic of the | `CD` |
| Cook Islands | `CK` |
| Costa Rica | `CR` |
| Côte d'Ivoire | `CI` |
| Croatia | `HR` |
| Cuba | `CU` |
| Curaçao | `CW` |
| Cyprus | `CY` |
| Czechia | `CZ` |
| Denmark | `DK` |
| Djibouti | `DJ` |
| Dominica | `DM` |
| Dominican Republic | `DO` |
| Ecuador | `EC` |
| Egypt | `EG` |
| El Salvador | `SV` |
| Equatorial Guinea | `GQ` |
| Eritrea | `ER` |
| Estonia | `EE` |
| Eswatini | `SZ` |
| Ethiopia | `ET` |
| Falkland Islands (Malvinas) | `FK` |
| Faroe Islands | `FO` |
| Fiji | `FJ` |
| Finland | `FI` |
| France | `FR` |
| French Guiana | `GF` |
| French Polynesia | `PF` |
| French Southern Territories | `TF` |
| Gabon | `GA` |
| Gambia | `GM` |
| Georgia | `GE` |
| Germany | `DE` |
| Ghana | `GH` |
| Gibraltar | `GI` |
| Greece | `GR` |
| Greenland | `GL` |
| Grenada | `GD` |
| Guadeloupe | `GP` |
| Guam | `GU` |
| Guatemala | `GT` |
| Guernsey | `GG` |
| Guinea | `GN` |
| Guinea-Bissau | `GW` |
| Guyana | `GY` |
| Haiti | `HT` |
| Heard Island and McDonald Islands | `HM` |
| Holy See (Vatican City State) | `VA` |
| Honduras | `HN` |
| Hong Kong | `HK` |
| Hungary | `HU` |
| Iceland | `IS` |
| India | `IN` |
| Indonesia | `ID` |
| Iran, Islamic Republic of | `IR` |
| Iraq | `IQ` |
| Ireland | `IE` |
| Isle of Man | `IM` |
| Israel | `IL` |
| Italy | `IT` |
| Jamaica | `JM` |
| Japan | `JP` |
| Jersey | `JE` |
| Jordan | `JO` |
| Kazakhstan | `KZ` |
| Kenya | `KE` |
| Kiribati | `KI` |
| Korea, Democratic People's Republic of | `KP` |
| Korea, Republic of | `KR` |
| Kuwait | `KW` |
| Kyrgyzstan | `KG` |
| Lao People's Democratic Republic | `LA` |
| Latvia | `LV` |
| Lebanon | `LB` |
| Lesotho | `LS` |
| Liberia | `LR` |
| Libya | `LY` |
| Liechtenstein | `LI` |
| Lithuania | `LT` |
| Luxembourg | `LU` |
| Macao | `MO` |
| Madagascar | `MG` |
| Malawi | `MW` |
| Malaysia | `MY` |
| Maldives | `MV` |
| Mali | `ML` |
| Malta | `MT` |
| Marshall Islands | `MH` |
| Martinique | `MQ` |
| Mauritania | `MR` |
| Mauritius | `MU` |
| Mayotte | `YT` |
| Mexico | `MX` |
| Micronesia, Federated States of | `FM` |
| Moldova, Republic of | `MD` |
| Monaco | `MC` |
| Mongolia | `MN` |
| Montenegro | `ME` |
| Montserrat | `MS` |
| Morocco | `MA` |
| Mozambique | `MZ` |
| Myanmar | `MM` |
| Namibia | `NA` |
| Nauru | `NR` |
| Nepal | `NP` |
| Netherlands | `NL` |
| New Caledonia | `NC` |
| New Zealand | `NZ` |
| Nicaragua | `NI` |
| Niger | `NE` |
| Nigeria | `NG` |
| Niue | `NU` |
| Norfolk Island | `NF` |
| North Macedonia | `MK` |
| Northern Mariana Islands | `MP` |
| Norway | `NO` |
| Oman | `OM` |
| Pakistan | `PK` |
| Palau | `PW` |
| Palestine, State of | `PS` |
| Panama | `PA` |
| Papua New Guinea | `PG` |
| Paraguay | `PY` |
| Peru | `PE` |
| Philippines | `PH` |
| Pitcairn | `PN` |
| Poland | `PL` |
| Portugal | `PT` |
| Puerto Rico | `PR` |
| Qatar | `QA` |
| Réunion | `RE` |
| Romania | `RO` |
| Russian Federation | `RU` |
| Rwanda | `RW` |
| Saint Barthélemy | `BL` |
| Saint Helena, Ascension and Tristan da Cunha | `SH` |
| Saint Kitts and Nevis | `KN` |
| Saint Lucia | `LC` |
| Saint Martin (French part) | `MF` |
| Saint Pierre and Miquelon | `PM` |
| Saint Vincent and the Grenadines | `VC` |
| Samoa | `WS` |
| San Marino | `SM` |
| Sao Tome and Principe | `ST` |
| Saudi Arabia | `SA` |
| Senegal | `SN` |
| Serbia | `RS` |
| Seychelles | `SC` |
| Sierra Leone | `SL` |
| Singapore | `SG` |
| Sint Maarten (Dutch part) | `SX` |
| Slovakia | `SK` |
| Slovenia | `SI` |
| Solomon Islands | `SB` |
| Somalia | `SO` |
| South Africa | `ZA` |
| South Georgia and the South Sandwich Islands | `GS` |
| South Sudan | `SS` |
| Spain | `ES` |
| Sri Lanka | `LK` |
| Sudan | `SD` |
| Suriname | `SR` |
| Svalbard and Jan Mayen | `SJ` |
| Sweden | `SE` |
| Switzerland | `CH` |
| Syrian Arab Republic | `SY` |
| Taiwan, Province of China | `TW` |
| Tajikistan | `TJ` |
| Tanzania, United Republic of | `TZ` |
| Thailand | `TH` |
| Timor-Leste | `TL` |
| Togo | `TG` |
| Tokelau | `TK` |
| Tonga | `TO` |
| Trinidad and Tobago | `TT` |
| Tunisia | `TN` |
| Türkiye | `TR` |
| Turkmenistan | `TM` |
| Turks and Caicos Islands | `TC` |
| Tuvalu | `TV` |
| Uganda | `UG` |
| Ukraine | `UA` |
| United Arab Emirates | `AE` |
| United Kingdom | `GB` |
| United States | `US` |
| United States Minor Outlying Islands | `UM` |
| Uruguay | `UY` |
| Uzbekistan | `UZ` |
| Vanuatu | `VU` |
| Venezuela, Bolivarian Republic of | `VE` |
| Viet Nam | `VN` |
| Virgin Islands, British | `VG` |
| Virgin Islands, U.S. | `VI` |
| Wallis and Futuna | `WF` |
| Western Sahara | `EH` |
| Yemen | `YE` |
| Zambia | `ZM` |
| Zimbabwe | `ZW` |
