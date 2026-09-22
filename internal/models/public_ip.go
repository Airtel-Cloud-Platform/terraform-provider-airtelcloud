package models

import "strings"

// PublicIP represents a public IP allocation (API response)
type PublicIP struct {
	UUID            string `json:"uuid"`
	IP              string `json:"ip"`
	PublicIP        string `json:"public_ip"`
	Domain          string `json:"domain,omitempty"`
	ObjectName      string `json:"object_name,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	TargetVIP       string `json:"target_vip,omitempty"`
	PortID          int    `json:"port_id,omitempty"`
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

// CreatePublicIPRequest reserves a public IP. port_id must be JSON null
// until a later attach call binds the allocation to a VM or LB port.
type CreatePublicIPRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PortID      *int   `json:"port_id"`
}

// AttachPublicIPRequest binds a reserved public IP to a compute or LB port.
type AttachPublicIPRequest struct {
	PortID int `json:"port_id"`
}

// PublicIPListResponse represents the paginated list response for public IPs
type PublicIPListResponse struct {
	Items []PublicIP `json:"items"`
	Count int        `json:"count"`
}

// PublicIPDisplayName is the customer-facing name from list or get payloads.
func PublicIPDisplayName(ip PublicIP) string {
	if strings.TrimSpace(ip.ObjectName) != "" {
		return strings.TrimSpace(ip.ObjectName)
	}
	return strings.TrimSpace(ip.Name)
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

// PublicIPPolicyRuleGeographicInput is a country selector in a policy source.
type PublicIPPolicyRuleGeographicInput struct {
	CountryCode string `json:"country_code,omitempty"`
	CountryName string `json:"country_name,omitempty"`
}

// PublicIPPolicyRuleSourceInput represents one source selector in a
// source-of-truth public IP policy create payload.
type PublicIPPolicyRuleSourceInput struct {
	CreateNew  *bool                              `json:"create_new,omitempty"`
	IPCIDR     string                             `json:"ip_cidr,omitempty"`
	SourceType string                             `json:"source_type,omitempty"`
	Geographic *PublicIPPolicyRuleGeographicInput `json:"geographic,omitempty"`
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
