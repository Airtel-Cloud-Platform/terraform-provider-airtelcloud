package models

import "encoding/json"

// LBService represents a load balancer service instance (API response)
type LBService struct {
	ID              string          `json:"id"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	FlavorID        int             `json:"flavor_id"`
	NetworkID       string          `json:"network_id,omitempty"`
	VPCID           string          `json:"vpc_id,omitempty"`
	VPCName         string          `json:"vpc_name,omitempty"`
	AZName          string          `json:"az_name,omitempty"`
	Status          string          `json:"status,omitempty"`
	OperatingStatus string          `json:"operating_status,omitempty"`
	Labels          json.RawMessage `json:"labels,omitempty"`
	Created         string          `json:"created,omitempty"`
	Updated         string          `json:"updated,omitempty"`
}

// CreateLBServiceRequest represents the request to create an LB service
type CreateLBServiceRequest struct {
	Name        string `form:"name,omitempty"`
	Description string `form:"description,omitempty"`
	FlavorID    int    `form:"flavor_id"`
	NetworkID   string `form:"network_id"`
	VPCID       string `form:"vpc_id"`
	VPCName     string `form:"vpc_name"`
	HA          bool   `form:"ha,omitempty"`
}

// LBVip represents a VIP port allocated for an LB service (API response)
type LBVip struct {
	ID             int    `json:"id"`
	Name           string `json:"name,omitempty"`
	Status         string `json:"status,omitempty"`
	FixedIPs       []string `json:"fixed_ips,omitempty"`
	PublicIP       string `json:"public_ip,omitempty"`
	ProviderPortID string `json:"provider_port_id,omitempty"`
	NetworkID      string `json:"network_id,omitempty"`
	DeviceID       string `json:"device_id,omitempty"`
	Created        string `json:"created,omitempty"`
	Updated        string `json:"updated,omitempty"`
}

// LBCertificate represents an SSL certificate attached to an LB service (API response)
type LBCertificate struct {
	ID      int    `json:"id"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// CreateLBCertificateRequest represents the request to create an SSL certificate
type CreateLBCertificateRequest struct {
	Name      string `form:"name"`
	SSLCert   string `form:"sslCert"`
	SSLPvtKey string `form:"sslPvtKey"`
	CACert    string `form:"caCert,omitempty"`
}

// LBVirtualServer represents a virtual server on an LB service (API response)
type LBVirtualServer struct {
	ID                 string `json:"id"`
	Name               string `json:"name,omitempty"`
	Protocol           string `json:"protocol,omitempty"`
	Port               int    `json:"port,omitempty"`
	RoutingAlgorithm   string `json:"routing_algorithm,omitempty"`
	VIP                string `json:"vip,omitempty"`
	Status             string `json:"status,omitempty"`
	PersistenceEnabled bool   `json:"persistence_enabled,omitempty"`
	PersistenceType    string `json:"persistence_type,omitempty"`
	PersistenceName    string `json:"persistence_name,omitempty"`
	PersistenceExpiry  int    `json:"persistence_expiry,omitempty"`
	XForwardedFor      bool   `json:"x_forwarded_for,omitempty"`
	RedirectHTTPS      bool   `json:"redirect_https,omitempty"`
	Created            string `json:"created,omitempty"`
	Updated            string `json:"updated,omitempty"`
}

// VirtualServerNode represents a backend node in a virtual server
type VirtualServerNode struct {
	ComputeID int    `json:"compute_id"`
	ComputeIP string `json:"compute_ip"`
	Port      int    `json:"port"`
	Weight    int    `json:"weight,omitempty"`
	MaxConn   int    `json:"max_conn,omitempty"`
}

// LBFlavor represents a load balancer flavor
type LBFlavor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
