package models

import "encoding/json"

// BlockStorageVolume represents a baremetal block storage volume (storage-plugin API).
type BlockStorageVolume struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description,omitempty"`
	Size             FlexInt64               `json:"size,omitempty"`
	AvailabilityZone string                  `json:"availabilityZone,omitempty"`
	State            BlockStorageVolumeState `json:"state,omitempty"`
	FailedStateError string                  `json:"failedStateError,omitempty"`
	CreatedAt        string                  `json:"createdAt,omitempty"`
	CreatedBy        string                  `json:"createdBy,omitempty"`
	UUID             string                  `json:"uuid,omitempty"`
	VolID            string                  `json:"volId,omitempty"`
	ProviderVolumeID string                  `json:"providerVolId,omitempty"`
	IsDeleted        bool                    `json:"isDeleted,omitempty"`
}

// UnmarshalJSON accepts both UI field names (description, availabilityZone)
// and storage-plugin aliases used by file storage (desc, az).
func (v *BlockStorageVolume) UnmarshalJSON(data []byte) error {
	type alias BlockStorageVolume
	aux := struct {
		alias
		Desc string `json:"desc"`
		AZ   string `json:"az"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*v = BlockStorageVolume(aux.alias)
	if v.Description == "" {
		v.Description = aux.Desc
	}
	if v.AvailabilityZone == "" {
		v.AvailabilityZone = aux.AZ
	}
	if v.UUID == "" {
		v.UUID = v.VolID
	}
	return nil
}

// BlockStorageVolumeState represents the state of a block storage volume.
type BlockStorageVolumeState string

const (
	BlockStorageStateCreating     BlockStorageVolumeState = "Creating"
	BlockStorageStateCreateFailed BlockStorageVolumeState = "CreateFailed"
	BlockStorageStateUpdating     BlockStorageVolumeState = "Updating"
	BlockStorageStateUpdateFailed BlockStorageVolumeState = "UpdateFailed"
	BlockStorageStateActive       BlockStorageVolumeState = "Active"
	BlockStorageStateDeleteFailed BlockStorageVolumeState = "DeleteFailed"
)

// CreateBlockStorageVolumeRequest is POST /block-storage/volume.
type CreateBlockStorageVolumeRequest struct {
	Name             string `json:"name"`
	AvailabilityZone string `json:"availabilityZone"`
	Size             int64  `json:"size"`
	Description      string `json:"description,omitempty"`
}

// UpdateBlockStorageVolumeRequest is PUT /block-storage/volume/{name}.
type UpdateBlockStorageVolumeRequest struct {
	Name             string `json:"name,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
	Size             int64  `json:"size,omitempty"`
	Description      string `json:"description,omitempty"`
}

// BlockStorageVolumeListResponse is GET /block-storage/volumes.
type BlockStorageVolumeListResponse struct {
	Count int                  `json:"count"`
	Items []BlockStorageVolume `json:"items"`
}
