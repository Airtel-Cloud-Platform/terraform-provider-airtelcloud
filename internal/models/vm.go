package models

import (
	"encoding/json"
	"fmt"
)

// FlexString converts a flexible JSON field (string or number) to a string
func FlexString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%d", int(val))
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Port represents a network port attached to a compute instance
type Port struct {
	ID       int      `json:"id"`
	FixedIPs []string `json:"fixed_ips"`
}

// Compute represents a virtual machine in Airtel Cloud (matches API response)
type Compute struct {
	ID                 string    `json:"id"`
	ProviderInstanceID string    `json:"provider_instance_id"`
	InstanceName       string    `json:"instance_name,omitempty"`
	Description        string    `json:"description,omitempty"`
	InstanceType       string    `json:"instance_type,omitempty"`
	FlavorID           interface{} `json:"flavor_id"`
	ImageID            interface{} `json:"image_id,omitempty"`
	NetworkID          string    `json:"network_id"`
	SubnetID           string    `json:"subnet_id,omitempty"`
	Status             string    `json:"status,omitempty"`
	PublicIPs          interface{} `json:"public_ips,omitempty"`
	FloatingIP         string    `json:"floating_ip,omitempty"`
	VPCID              string    `json:"vpc_id,omitempty"`
	SecurityGroupID    int       `json:"sec_group_id,omitempty"`
	KeypairName        string    `json:"keypair_name,omitempty"`
	UserData           string    `json:"userdata,omitempty"`
	AvailabilityZone   string    `json:"availability_zone,omitempty"`
	AZName             string    `json:"az_name,omitempty"`
	Region             string    `json:"region,omitempty"`
	OSType             string    `json:"os_type,omitempty"`
	BootFromVolume     bool      `json:"boot_from_volume,omitempty"`
	VolumeSize         int       `json:"volume_size,omitempty"`
	VolumeTypeID       int       `json:"volume_type_id,omitempty"`
	EnableBackup       bool      `json:"enable_backup,omitempty"`
	EnableAntivirus    bool      `json:"enable_antivirus,omitempty"`
	BillingUnit        string    `json:"billing_unit,omitempty"`
	IsFirewall         bool      `json:"is_firewall,omitempty"`
	IsVPN              bool      `json:"is_vpn,omitempty"`
	ManagedBy          string    `json:"managed_by,omitempty"`
	CreatedAt          string `json:"created,omitempty"`
	UpdatedAt          string `json:"updated,omitempty"`
	DeletedAt          string `json:"deleted,omitempty"`
	Action             string          `json:"action,omitempty"`
	Labels             json.RawMessage `json:"labels,omitempty"`
	Ports              []Port          `json:"ports,omitempty"`
}

// PrivateIP extracts the private IP from ports[0].fixed_ips[0], falling back to FloatingIP.
func (c *Compute) PrivateIP() string {
	if len(c.Ports) > 0 && len(c.Ports[0].FixedIPs) > 0 {
		return c.Ports[0].FixedIPs[0]
	}
	return c.FloatingIP
}

// CreateComputeRequest represents the request to create a compute instance (v2.1 API)
type CreateComputeRequest struct {
	VPCID                string `form:"vpc_id,omitempty"`
	Description          string `form:"description,omitempty"`
	Region               string `form:"region,omitempty"`
	AZName               string `form:"az_name,omitempty"`
	InstanceName         string `form:"instance_name"`
	ImageID              string `form:"image_id,omitempty"`
	FlavorID             string `form:"flavor_id,omitempty"`
	SecurityGroupID      int    `form:"sec_group_id,omitempty"`
	BootFromVolume       bool   `form:"boot_from_volume,omitempty"`
	VolumeSize           int    `form:"volume_size,omitempty"`
	VolumeTypeID         int    `form:"volume_type_id,omitempty"`
	OSType               string `form:"os_type,omitempty"`
	VMCount              int    `form:"vm_count,omitempty"`
	KeypairID            string `form:"keypair_id,omitempty"`
	SubnetID             string `form:"subnetId,omitempty"`
	NetworkID            string `form:"network_id,omitempty"`
	UserCloudInitScripts string `form:"user_cloud_init_scripts,omitempty"`
	ProtectionPlan       string `form:"protection_plan,omitempty"`
	EnableBackup         bool   `form:"enable_backup,omitempty"`
	StartDate            string `form:"start_date,omitempty"`
	StartTime            string `form:"start_time,omitempty"`
}

// UpdateComputeRequest represents the request to update a compute instance (v2.1 API)
type UpdateComputeRequest struct {
	InstanceName    string `form:"instance_name,omitempty"`
	Description     string `form:"description,omitempty"`
	SecurityGroupID int    `form:"sec_group_id,omitempty"`
}

// ComputeAction represents actions that can be performed on a compute instance
type ComputeAction string

const (
	ComputeActionStart   ComputeAction = "start"
	ComputeActionStop    ComputeAction = "stop"
	ComputeActionReboot  ComputeAction = "reboot"
	ComputeActionSuspend ComputeAction = "suspend"
	ComputeActionResume  ComputeAction = "resume"
)

// Flavor represents a compute flavor (from list API)
type Flavor struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	DisplayName string      `json:"display_name,omitempty"`
	VCPU        interface{} `json:"vcpu,omitempty"`
	RAM         interface{} `json:"ram,omitempty"`
	Disk        interface{} `json:"disk,omitempty"`
}

// Image represents a compute image (from list API)
type Image struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	OSType      string `json:"os_type,omitempty"`
	Description string `json:"description,omitempty"`
}

// SecurityGroup represents a security group (from list API)
type SecurityGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Keypair represents an SSH keypair (from list API)
type Keypair struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint,omitempty"`
}
