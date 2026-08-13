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
	CreatedAt       string `json:"created_at,omitempty"`
	DeallocatedTime string `json:"deallocated_time,omitempty"`
	AZName          string `json:"az_name,omitempty"`
	AZ              string `json:"az,omitempty"`
	ProjectName     string `json:"project_name,omitempty"`
	Region          string `json:"region,omitempty"`
	Status          string `json:"status,omitempty"`
}

// CreatePublicIPRequest represents the request to allocate a public IP
// using the new TCPWave flow: allocation is done by port id, and the backend
// handles VIP object creation and static route setup.
type CreatePublicIPRequest struct {
	Name   string `json:"name"`
	PortID int    `json:"port_id"`
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

// PublicIPPolicyRuleSourceInput represents one source selector in a
// source-of-truth public IP policy create payload.
type PublicIPPolicyRuleSourceInput struct {
	CreateNew  *bool  `json:"create_new,omitempty"`
	IPCIDR     string `json:"ip_cidr,omitempty"`
	SourceType string `json:"source_type,omitempty"`
}

// PublicIPPolicyRuleServiceInput represents one service selector in a
// source-of-truth public IP policy create payload.
type PublicIPPolicyRuleServiceInput struct {
	CreateNew *bool  `json:"create_new,omitempty"`
	Name      string `json:"name,omitempty"`
	IsDefault *bool  `json:"is_default,omitempty"`
}

// CreatePublicIPPolicyRuleRequest represents the request to create a NAT policy rule
type CreatePublicIPPolicyRuleRequest struct {
	DisplayName   string                           `json:"display_name"`
	Source        string                           `json:"source"`
	SourceConfig  []PublicIPPolicyRuleSourceInput  `json:"source_config,omitempty"`
	ServiceList   []string                         `json:"service_list"`
	ServiceConfig []PublicIPPolicyRuleServiceInput `json:"service_config,omitempty"`
	Action        string                           `json:"action"`
	ResourceType  string                           `json:"resource_type,omitempty"`
	RevisionNote  string                           `json:"revision_note,omitempty"`
	TargetVIP     string                           `json:"target_vip"`
	PublicIP      string                           `json:"public_ip"`
	UUID          string                           `json:"uuid"`
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

// NetworkVIPPort represents LB VIP-to-port mapping from the networks VIPs API.
type NetworkVIPPort struct {
	LBName           string `json:"lb_name,omitempty"`
	VSName           string `json:"vs_name,omitempty"`
	PortID           int    `json:"port_id"`
	AllowedIPAddress string `json:"allowed_ip_address,omitempty"`
}
