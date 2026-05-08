package models

// VPCPeering represents a VPC peering connection in Airtel Cloud
type VPCPeering struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	VPCSourceID       string   `json:"vpcSourceId"`
	VPCTargetID       string   `json:"vpcTargetId"`
	AZ                string   `json:"az"`
	Region            string   `json:"region"`
	IsPclEnabled      bool     `json:"isPclEnabled"`
	AllowedSubnetList []string `json:"allowedSubnetList,omitempty"`
	BlockedSubnetList []string `json:"blockedSubnetList,omitempty"`
	State             string   `json:"state,omitempty"`
	CreatedBy         string   `json:"created_by,omitempty"`
	CreatedAt         string   `json:"created_at,omitempty"`
	UpdatedAt         string   `json:"updated_at,omitempty"`
}

// CreateVPCPeeringRequest represents the request to create a VPC peering connection
type CreateVPCPeeringRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	VPCSourceID       string   `json:"vpcSourceId"`
	VPCTargetID       string   `json:"vpcTargetId"`
	AZ                string   `json:"az"`
	Region            string   `json:"region"`
	PeerVpcRegion     string   `json:"peerVpcRegion"`
	IsPclEnabled      bool     `json:"isPclEnabled"`
	AllowedSubnetList []string `json:"allowedSubnetList"`
	BlockedSubnetList []string `json:"blockedSubnetList"`
}

// VPCPeeringListResponse represents the response for listing VPC peerings
type VPCPeeringListResponse struct {
	Count int          `json:"count"`
	Items []VPCPeering `json:"items"`
}
