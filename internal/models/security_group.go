package models

// SecurityGroupDetail represents the full security group response from the network API
type SecurityGroupDetail struct {
	ID                int                       `json:"id"`
	UUID              string                    `json:"uuid"`
	SecurityGroupName string                    `json:"security_group_name"`
	Status            string                    `json:"status"`
	ProjectID         int                       `json:"project_id"`
	AZName            string                    `json:"az_name"`
	AZRegion          string                    `json:"az_region"`
	Created           string                    `json:"created"`
	Updated           string                    `json:"updated"`
	Rules             []SecurityGroupRuleDetail `json:"security_group_rules"`
}

// SecurityGroupRuleDetail represents a security group rule from the network API
type SecurityGroupRuleDetail struct {
	ID                          int    `json:"id"`
	UUID                        string `json:"uuid"`
	Direction                   string `json:"direction"`
	Protocol                    string `json:"protocol"`
	PortRangeMin                string `json:"port_range_min"`
	PortRangeMax                string `json:"port_range_max"`
	RemoteIPPrefix              string `json:"remote_ip_prefix"`
	RemoteGroupID               string `json:"remote_group_id"`
	Ethertype                   string `json:"ethertype"`
	Status                      string `json:"status"`
	Description                 string `json:"description"`
	ProviderSecurityGroupRuleID string `json:"provider_security_group_rule_id"`
}

// CreateSecurityGroupRequest represents the request to create a security group
type CreateSecurityGroupRequest struct {
	SecurityGroupName string `form:"security_group_name"`
}

// CreateSecurityGroupRuleRequest represents the request to create a security group rule
type CreateSecurityGroupRuleRequest struct {
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol,omitempty"`
	PortRangeMin   string `json:"port_range_min,omitempty"`
	PortRangeMax   string `json:"port_range_max,omitempty"`
	RemoteIPPrefix string `json:"remote_ip_prefix,omitempty"`
	RemoteGroupID  string `json:"remote_group_id,omitempty"`
	Ethertype      string `json:"ethertype,omitempty"`
	Description    string `json:"description,omitempty"`
}
