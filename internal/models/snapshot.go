package models

import "encoding/json"

// ComputeSnapshot represents a VM snapshot (API response).
//
// Field types follow the wire format exactly: the API returns labels as an
// object and image_id as a number, and created/updated as epoch-second strings.
type ComputeSnapshot struct {
	ID                  int              `json:"id"`
	UUID                string           `json:"uuid"`
	SnapshotName        string           `json:"snapshot_name,omitempty"`
	SnapshotDescription string           `json:"snapshot_description,omitempty"`
	Status              string           `json:"status,omitempty"`
	Action              string           `json:"action,omitempty"`
	IsActive            bool             `json:"is_active,omitempty"`
	IsImage             bool             `json:"is_image,omitempty"`
	ImageID             interface{}      `json:"image_id,omitempty"`
	UserID              int              `json:"user_id,omitempty"`
	TaskID              string           `json:"task_id,omitempty"`
	Labels              json.RawMessage  `json:"labels,omitempty"`
	Created             string           `json:"created,omitempty"`
	Updated             string           `json:"updated,omitempty"`
	Compute             *SnapshotCompute `json:"compute,omitempty"`
	Image               json.RawMessage  `json:"image,omitempty"`
}

// SnapshotCompute is the compute instance embedded in a snapshot response.
type SnapshotCompute struct {
	ID           interface{} `json:"id"`
	InstanceName string      `json:"instance_name,omitempty"`
	NetworkID    string      `json:"network_id,omitempty"`
	ProjectID    interface{} `json:"project_id,omitempty"`
}

// ImageIDString returns image_id as a string, whether the API sent a number or a string.
func (s *ComputeSnapshot) ImageIDString() string {
	if s.ImageID == nil {
		return ""
	}
	return FlexString(s.ImageID)
}

// ComputeIDString returns the owning compute's ID, or "" if the API omitted the compute object.
func (s *ComputeSnapshot) ComputeIDString() string {
	if s.Compute == nil || s.Compute.ID == nil {
		return ""
	}
	return FlexString(s.Compute.ID)
}

// CreateComputeSnapshotRequest is the URL-encoded form body for creating a snapshot.
// snapshot_name is the only field the API requires.
type CreateComputeSnapshotRequest struct {
	SnapshotName string `form:"snapshot_name"`
	ComputeID    string `form:"compute_id,omitempty"`
	BillingUnit  string `form:"billing_unit,omitempty"`
}
