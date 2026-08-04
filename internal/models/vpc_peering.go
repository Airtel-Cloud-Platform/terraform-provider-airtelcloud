package models

import "encoding/json"

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

func (v *VPCPeering) UnmarshalJSON(data []byte) error {
	type rawVPCPeering struct {
		ID                string          `json:"id"`
		VPCPeeringID      string          `json:"vpcPeeringId"`
		Name              string          `json:"name"`
		Description       string          `json:"description,omitempty"`
		VPCSourceID       string          `json:"vpcSourceId"`
		VPCTargetID       string          `json:"vpcTargetId"`
		AZ                json.RawMessage `json:"az"`
		Region            string          `json:"region"`
		IsPclEnabled      bool            `json:"isPclEnabled"`
		AllowedSubnetList []string        `json:"allowedSubnetList,omitempty"`
		BlockedSubnetList []string        `json:"blockedSubnetList,omitempty"`
		State             string          `json:"state,omitempty"`
		CreatedBy         string          `json:"created_by,omitempty"`
		CreatedAt         string          `json:"created_at,omitempty"`
		UpdatedAt         string          `json:"updated_at,omitempty"`
		CreatedAtAlt      string          `json:"createdAt,omitempty"`
		UpdatedAtAlt      string          `json:"updatedAt,omitempty"`
	}

	var raw rawVPCPeering
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*v = VPCPeering{
		ID:                raw.ID,
		Name:              raw.Name,
		Description:       raw.Description,
		VPCSourceID:       raw.VPCSourceID,
		VPCTargetID:       raw.VPCTargetID,
		Region:            raw.Region,
		IsPclEnabled:      raw.IsPclEnabled,
		AllowedSubnetList: raw.AllowedSubnetList,
		BlockedSubnetList: raw.BlockedSubnetList,
		State:             raw.State,
		CreatedBy:         raw.CreatedBy,
		CreatedAt:         raw.CreatedAt,
		UpdatedAt:         raw.UpdatedAt,
	}

	if v.ID == "" {
		v.ID = raw.VPCPeeringID
	}
	if v.CreatedAt == "" {
		v.CreatedAt = raw.CreatedAtAlt
	}
	if v.UpdatedAt == "" {
		v.UpdatedAt = raw.UpdatedAtAlt
	}

	if len(raw.AZ) == 0 || string(raw.AZ) == "null" {
		return nil
	}

	if err := json.Unmarshal(raw.AZ, &v.AZ); err == nil {
		return nil
	}

	var azList []string
	if err := json.Unmarshal(raw.AZ, &azList); err != nil {
		return err
	}
	if len(azList) > 0 {
		v.AZ = azList[0]
	}

	return nil
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

func (r *VPCPeeringListResponse) UnmarshalJSON(data []byte) error {
	type rawListResponse struct {
		Count       int          `json:"count"`
		Items       []VPCPeering `json:"items"`
		VPCPeerings []VPCPeering `json:"vpcPeerings"`
	}

	var raw rawListResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Count = raw.Count
	r.Items = raw.Items
	if len(r.Items) == 0 && len(raw.VPCPeerings) > 0 {
		r.Items = raw.VPCPeerings
	}

	return nil
}
