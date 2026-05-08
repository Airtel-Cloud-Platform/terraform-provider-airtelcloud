package models

// ComputeSnapshot represents a VM snapshot (API response)
type ComputeSnapshot struct {
	ID           int    `json:"id"`
	UUID         string `json:"uuid"`
	SnapshotName string `json:"snapshot_name,omitempty"`
	Status       string `json:"status,omitempty"`
	Action       string `json:"action,omitempty"`
	IsActive     bool   `json:"is_active,omitempty"`
	IsImage      bool   `json:"is_image,omitempty"`
	ImageID      string `json:"image_id,omitempty"`
	Labels       string `json:"labels,omitempty"`
	TaskID       string `json:"task_id,omitempty"`
	Created      string `json:"created,omitempty"`
	Updated      string `json:"updated,omitempty"`
}
