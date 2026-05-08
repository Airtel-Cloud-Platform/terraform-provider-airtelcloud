package models

// PublicIP represents a public IP allocation (API response)
type PublicIP struct {
	UUID            string `json:"uuid"`
	IP              string `json:"ip"`
	PublicIP        string `json:"public_ip"`
	Domain          string `json:"domain,omitempty"`
	ObjectName      string `json:"object_name,omitempty"`
	TargetVIP       string `json:"target_vip,omitempty"`
	Username        string `json:"username,omitempty"`
	OrgID           string `json:"org_id,omitempty"`
	OrgName         string `json:"org_name,omitempty"`
	AllocatedTime   string `json:"allocated_time,omitempty"`
	DeallocatedTime string `json:"deallocated_time,omitempty"`
	AZName          string `json:"az_name,omitempty"`
	ProjectName     string `json:"project_name,omitempty"`
	Region          string `json:"region,omitempty"`
	Status          string `json:"status,omitempty"`
}

// CreatePublicIPRequest represents the request to allocate a public IP
type CreatePublicIPRequest struct {
	ObjectName string `json:"object_name"`
	VIP        string `json:"vip"`
}

// MapPublicIPRequest represents the request to map a public IP with an internal VIP
type MapPublicIPRequest struct {
	TargetVIP string `json:"target_vip"`
	PublicIP  string `json:"public_ip"`
	UUID      string `json:"uuid"`
	PortID    int    `json:"port_id"`
}

// PublicIPListResponse represents the paginated list response for public IPs
type PublicIPListResponse struct {
	Items []PublicIP `json:"items"`
	Count int        `json:"count"`
}

// PublicIPPolicyRule represents a NAT policy rule on a public IP (API response)
type PublicIPPolicyRule struct {
	DisplayName string   `json:"display_name,omitempty"`
	UUID        string   `json:"uuid,omitempty"`
	OrgID       string   `json:"org_id,omitempty"`
	OrgName     string   `json:"org_name,omitempty"`
	AZName      string   `json:"az_name,omitempty"`
	SourceIP    string   `json:"source_ip,omitempty"`
	TargetVIP   string   `json:"target_vip,omitempty"`
	State       string   `json:"state,omitempty"`
	Services    []string `json:"services,omitempty"`
	Action      string   `json:"action,omitempty"`
}

// CreatePublicIPPolicyRuleRequest represents the request to create a NAT policy rule
type CreatePublicIPPolicyRuleRequest struct {
	DisplayName string   `json:"display_name"`
	Source      string   `json:"source"`
	ServiceList []string `json:"service_list"`
	Action      string   `json:"action"`
	TargetVIP   string   `json:"target_vip"`
	PublicIP    string   `json:"public_ip"`
	UUID        string   `json:"uuid"`
}

// PublicIPPolicyRuleListResponse represents the paginated list response for policy rules
type PublicIPPolicyRuleListResponse struct {
	Items []PublicIPPolicyRule `json:"items"`
	Count int                  `json:"count"`
}

// IPAMService represents a service/port available for policy rules
type IPAMService struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	PortRange   string  `json:"port_range,omitempty"`
	ProtoType   *string `json:"proto_type"`
	OrgName     string  `json:"org_name,omitempty"`
	ProjectName string  `json:"project_name,omitempty"`
	AZName      string  `json:"az_name,omitempty"`
	IsDefault   bool    `json:"is_default"`
	CreatedAt   *string `json:"created_at"`
}
