package models

// DNSRecord represents a DNS record in Airtel Cloud DNSaaS
type DNSRecord struct {
	UUID        string  `json:"uuid"`
	ZoneName    string  `json:"zone_name"`
	ZoneID      string  `json:"zone_id"`
	Owner       string  `json:"owner"`
	Data        string  `json:"data"`
	RecordType  string  `json:"record_type"`
	TTL         int     `json:"ttl"`
	OrgName     string  `json:"org_name"`
	OrgID       string  `json:"org_id"`
	Preference  *int    `json:"preference,omitempty"`
	CreatedBy   string  `json:"created_by"`
	Description *string `json:"description,omitempty"`
	CreatedAt   float64 `json:"created_at"`
	UpdatedAt   float64 `json:"updated_at"`
	// Additional optional fields for specific record types
	Tag            *string `json:"tag,omitempty"`             // For CAA records
	CAAFlag        *int    `json:"caa_flag,omitempty"`        // For CAA records
	ServiceSubtype *int    `json:"service_subtype,omitempty"` // For SRV records
	IsExternalRR   *int    `json:"is_external_rr,omitempty"`
}

// DNSRecordListResponse represents the response when listing DNS records
type DNSRecordListResponse struct {
	Items []DNSRecord `json:"items"`
	Count int         `json:"count"`
}

// CreateDNSRecordRequest represents the request to create a DNS record
type CreateDNSRecordRequest struct {
	Owner          *string `json:"owner,omitempty"`
	Data           *string `json:"data,omitempty"`
	RecordType     string  `json:"record_type"`
	TTL            *int    `json:"ttl,omitempty"`
	Description    *string `json:"description,omitempty"`
	Preference     *int    `json:"preference,omitempty"`      // For MX records
	Tag            *string `json:"tag,omitempty"`             // For CAA records
	CAAFlag        *int    `json:"caa_flag,omitempty"`        // For CAA records
	ServiceSubtype *int    `json:"service_subtype,omitempty"` // For SRV records
	IsExternalRR   *int    `json:"is_external_rr,omitempty"`
}

// UpdateDNSRecordRequest represents the request to update a DNS record
// Note: The API uses the same schema as CreateDNSRecordRequest for updates
type UpdateDNSRecordRequest struct {
	Owner          *string `json:"owner,omitempty"`
	Data           *string `json:"data,omitempty"`
	RecordType     string  `json:"record_type"`
	TTL            *int    `json:"ttl,omitempty"`
	Description    *string `json:"description,omitempty"`
	Preference     *int    `json:"preference,omitempty"`
	Tag            *string `json:"tag,omitempty"`
	CAAFlag        *int    `json:"caa_flag,omitempty"`
	ServiceSubtype *int    `json:"service_subtype,omitempty"`
	IsExternalRR   *int    `json:"is_external_rr,omitempty"`
}
