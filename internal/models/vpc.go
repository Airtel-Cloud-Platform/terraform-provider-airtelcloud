package models

import "time"

// Tag represents a key-value tag
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagsToMap converts a slice of Tags to a map
func TagsToMap(tags []Tag) map[string]string {
	if tags == nil {
		return nil
	}
	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		result[tag.Key] = tag.Value
	}
	return result
}

// MapToTags converts a map to a slice of Tags
func MapToTags(m map[string]string) []Tag {
	if m == nil {
		return nil
	}
	result := make([]Tag, 0, len(m))
	for k, v := range m {
		result = append(result, Tag{Key: k, Value: v})
	}
	return result
}

// VPC represents a Virtual Private Cloud in Airtel Cloud
type VPC struct {
	ID                 string    `json:"networkId"`
	Name               string    `json:"name"`
	CIDRBlock          string    `json:"cidr_block"`
	State              string    `json:"state"`
	EnableDNSHostnames bool      `json:"enable_dns_hostnames"`
	EnableDNSSupport   bool      `json:"enable_dns_support"`
	IsDefault          bool      `json:"is_default"`
	Tags               []Tag     `json:"tags,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateVPCRequest represents the request to create a VPC
type CreateVPCRequest struct {
	Name               string `json:"name"`
	CIDRBlock          string `json:"cidr_block"`
	EnableDNSHostnames bool   `json:"enable_dns_hostnames,omitempty"`
	EnableDNSSupport   bool   `json:"enable_dns_support,omitempty"`
	Tags               []Tag  `json:"tags,omitempty"`
}

// UpdateVPCRequest represents the request to update a VPC
type UpdateVPCRequest struct {
	Name               string `json:"name,omitempty"`
	EnableDNSHostnames *bool  `json:"enable_dns_hostnames,omitempty"`
	EnableDNSSupport   *bool  `json:"enable_dns_support,omitempty"`
	Tags               []Tag  `json:"tags,omitempty"`
}

// VPCListResponse represents the response for listing VPCs
type VPCListResponse struct {
	Count int   `json:"count"`
	Items []VPC `json:"items"`
}
