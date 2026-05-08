package models

// DNSZone represents a DNS zone in Airtel Cloud DNSaaS
type DNSZone struct {
	UUID            string  `json:"uuid"`
	ZoneName        string  `json:"zone_name"`
	DNSZoneTemplate string  `json:"dns_zone_template"`
	Description     *string `json:"description"`
	OrgName         string  `json:"org_name"`
	OrgID           string  `json:"org_id"`
	ZoneType        string  `json:"zone_type"`
	CreatedBy       string  `json:"created_by"`
	CreatedAt       float64 `json:"created_at"`
	UpdatedAt       float64 `json:"updated_at"`
}

// DNSZoneListResponse represents the response when listing DNS zones
type DNSZoneListResponse struct {
	Items []DNSZone `json:"items"`
	Count int       `json:"count"`
}

// CreateDNSZoneRequest represents the request to create a DNS zone
type CreateDNSZoneRequest struct {
	ZoneName    string  `json:"zone_name"`
	ZoneType    string  `json:"zone_type"`
	Description *string `json:"description,omitempty"`
}

// UpdateDNSZoneRequest represents the request to update a DNS zone
type UpdateDNSZoneRequest struct {
	Description *string `json:"description,omitempty"`
}
