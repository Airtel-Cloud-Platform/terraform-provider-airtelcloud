package models

import "encoding/json"

// Volume represents a block storage volume in Airtel Cloud (matches API response)
type Volume struct {
	ID                  int                `json:"id"`
	ProviderVolumeID    string             `json:"provider_volume_id"`
	VolumeName          string             `json:"volume_name"`
	VolumeSize          int                `json:"volume_size,omitempty"`
	Description         string             `json:"description,omitempty"`
	Status              string             `json:"status,omitempty"`
	AvailabilityZone    string             `json:"availiability_zone,omitempty"`
	AZName              string             `json:"az_name,omitempty"`
	Region              string             `json:"region,omitempty"`
	VPCID               string             `json:"vpc_id,omitempty"`
	VolumeTypeID        json.RawMessage    `json:"volume_type_id"`
	VolumeSourceID      int                `json:"volume_source_id,omitempty"`
	Action              string             `json:"action,omitempty"`
	Host                string             `json:"host,omitempty"`
	Meta                json.RawMessage    `json:"meta,omitempty"`
	Labels              json.RawMessage    `json:"labels,omitempty"`
	VolumeImageMetadata json.RawMessage    `json:"volume_image_metadata,omitempty"`
	UserID              int                `json:"user_id"`
	ProjectID           int                `json:"project_id,omitempty"`
	Created             string             `json:"created"`
	Updated             string             `json:"updated,omitempty"`
	Deleted             *string            `json:"deleted,omitempty"`
	VolumeAttachments   []VolumeAttachment `json:"volume_attach_volume,omitempty"`
	VolumeType          VolumeType         `json:"volume_type,omitempty"`
	Bootable            bool               `json:"bootable,omitempty"`
	EnableBackup        bool               `json:"enable_backup,omitempty"`
	UUID                string             `json:"uuid,omitempty"`
}

// VolumeAttachment represents a volume attachment to a compute instance
type VolumeAttachment struct {
	ID                         int             `json:"id"`
	VolumeID                   json.RawMessage `json:"volume_id"`
	ComputeID                  string          `json:"compute_id"`
	Compute                    json.RawMessage `json:"compute,omitempty"`
	ProviderVolumeAttachmentID string          `json:"provider_volume_attachment_id"`
	VolumeAttachmentDeviceName string          `json:"volume_attachment_device_name,omitempty"`
}

// VolumeType represents a volume type configuration
type VolumeType struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description,omitempty"`
	Group             string          `json:"group,omitempty"`
	Label             string          `json:"label,omitempty"`
	MinSize           json.RawMessage `json:"min_size,omitempty"`
	MaxSize           json.RawMessage `json:"max_size,omitempty"`
	VolumeStepSize    json.RawMessage `json:"volume_step_size,omitempty"`
	Cost              int             `json:"cost,omitempty"`
	Unit              string          `json:"unit,omitempty"`
	IsActive          bool            `json:"is_active"`
	IsDefault         bool            `json:"is_default"`
	IsPublic          bool            `json:"is_public"`
	ExtraSpecs        json.RawMessage `json:"extra_specs,omitempty"`
	ProductAttributes json.RawMessage `json:"product_attributes,omitempty"`
	ProviderVTID      string          `json:"provider_vt_id,omitempty"`
	Prices            json.RawMessage `json:"prices"`
}

// CreateVolumeRequest represents the request to create a volume
type CreateVolumeRequest struct {
	VolumeName                 string `form:"volume_name"`
	VolumeSize                 int    `form:"volume_size"`
	Bootable                   bool   `form:"bootable"`
	AvailabilityZone           string `form:"availabilityZone,omitempty"`
	Network                    string `form:"network,omitempty"`
	VPCID                      string `form:"vpc_id,omitempty"`
	Subnet                     string `form:"subnet,omitempty"`
	SubnetID                   string `form:"subnetId,omitempty"`
	VolumeType                 string `form:"volume_type,omitempty"`
	VolumeTypeID               string `form:"volume_type_id,omitempty"`
	IsEncrypted                string `form:"is_encrypted,omitempty"`
	EnableBackup               bool   `form:"enable_backup"`
	VolumeAttachmentDeviceName string `form:"volume_attachment_device_name,omitempty"`
	ComputeID                  string `form:"compute_id,omitempty"`
	BillingUnit                string `form:"billing_unit,omitempty"`
	SnapshotID                 string `form:"snapshot_id,omitempty"`
	Products                   string `form:"products,omitempty"`
}

// UpdateVolumeRequest represents the request to update/extend a volume
type UpdateVolumeRequest struct {
	VolumeName   string `form:"volume_name"`
	VolumeSize   int    `form:"volume_size"`
	Bootable     bool   `form:"bootable"`
	EnableBackup bool   `form:"enable_backup"`
	VolumeType   string `form:"volume_type"`
	BillingUnit  string `form:"billing_unit"`
	VolumeRate   int    `form:"volume_rate"`
	ComputeID    string `form:"compute_id,omitempty"`
	IsEncrypted  string `form:"is_encrypted,omitempty"`
}

// VolumeAttachRequest represents the request to attach a volume to a compute instance
type VolumeAttachRequest struct {
	ComputeID                  string `form:"compute_id"`
	VolumeID                   int    `form:"volume_id"`
	VolumeAttachmentDeviceName string `form:"volume_attachment_device_name,omitempty"`
}

// VolumeDetachRequest represents the request to detach a volume from a compute instance
type VolumeDetachRequest struct {
	ComputeID string `form:"compute_id"`
	VolumeID  int    `form:"volume_id"`
}

// VolumeSnapshotRequest represents the request to create a volume snapshot
type VolumeSnapshotRequest struct {
	SnapshotName string `form:"snapshot_name"`
	BillingUnit  string `form:"billing_unit"`
	Products     string `form:"products,omitempty"`
	Description  string `form:"description,omitempty"`
}
