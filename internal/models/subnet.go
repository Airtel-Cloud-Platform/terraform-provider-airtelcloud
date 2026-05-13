package models

// Subnet represents a subnet in Airtel Cloud network
type Subnet struct {
	Domain           string   `json:"domain,omitempty"`
	Project          string   `json:"project,omitempty"`
	NetworkID        string   `json:"networkId,omitempty"`
	SubnetID         string   `json:"subnetId,omitempty"`
	Name             string   `json:"name,omitempty"`
	Description      string   `json:"desc,omitempty"`
	Start            string   `json:"start,omitempty"`
	End              string   `json:"end,omitempty"`
	Gateway          string   `json:"gateway,omitempty"`
	VLAN             int      `json:"vlan,omitempty"`
	VNI              int      `json:"vni,omitempty"`
	MTU              int      `json:"mtu,omitempty"`
	DNS              []string `json:"dns,omitempty"`
	DHCPServer       []string `json:"dhcpServer,omitempty"`
	AvailabilityZone string   `json:"az,omitempty"`
	Region           string   `json:"region,omitempty"`
	State            string   `json:"state,omitempty"`
	IPv4AddressSpace string   `json:"ipv4AddressSpace,omitempty"`
	IPv6AddressSpace string   `json:"ipv6AddressSpace,omitempty"`
	EnableGateway    bool     `json:"enableGateway,omitempty"`
	EnableDHCP       bool     `json:"enableDhcp,omitempty"`
	PhysicalNetName  string   `json:"physicalNetName,omitempty"`
	PhysicalNetType  string   `json:"physicalNetType,omitempty"`
	SubnetRole       string   `json:"subnetRole,omitempty"`
	SubnetSubRole    string   `json:"subnetSubRole,omitempty"`
	DeploymentType   string   `json:"deploymentType,omitempty"`
	IsOnboarded      bool     `json:"isOnboarded,omitempty"`
	Labels           []string `json:"labels,omitempty"`
	CreatedBy        string   `json:"createdBy,omitempty"`
	CreateTime       string   `json:"createTime,omitempty"`
	ErrorMessage     string   `json:"errorMessage,omitempty"`
}

// SubnetListResponse represents the response for listing subnets
type SubnetListResponse struct {
	Count int      `json:"count"`
	Items []Subnet `json:"items"`
}

// CreateSubnetRequest represents the request to create a subnet
type CreateSubnetRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"desc,omitempty"`
	AvailabilityZone string   `json:"az,omitempty"`
	IPv4AddressSpace string   `json:"ipv4AddressSpace"`
	SubnetSubRole    string   `json:"subnetSubRole,omitempty"`
	Labels           []string `json:"labels,omitempty"`
}

// UpdateSubnetRequest represents the request to update a subnet
type UpdateSubnetRequest struct {
	Description string `json:"desc,omitempty"`
	AliasName   string `json:"aliasName,omitempty"`
	Summary     string `json:"summary,omitempty"`
}
