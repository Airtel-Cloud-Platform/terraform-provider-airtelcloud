package models

// FileStorageVolume represents a file storage volume in Airtel Cloud (matches API response)
type FileStorageVolume struct {
	Domain           string                 `json:"domain,omitempty"`
	Project          string                 `json:"project,omitempty"`
	Name             string                 `json:"name"`
	Description      string                 `json:"desc,omitempty"`
	Size             string                 `json:"size,omitempty"`
	AvailabilityZone string                 `json:"az,omitempty"`
	State            FileStorageVolumeState `json:"state,omitempty"`
	FailedStateError string                 `json:"failedStateError,omitempty"`
	CreatedAt        string                 `json:"createdAt,omitempty"`
	CreatedBy        string                 `json:"createdBy,omitempty"`
	UUID             string                 `json:"uuid,omitempty"`
	ProviderVolumeID string                 `json:"providerVolId,omitempty"`
}

// FileStorageVolumeState represents the state of a file storage volume
type FileStorageVolumeState string

const (
	FileStorageStateCreating     FileStorageVolumeState = "Creating"
	FileStorageStateCreateFailed FileStorageVolumeState = "CreateFailed"
	FileStorageStateUpdating     FileStorageVolumeState = "Updating"
	FileStorageStateUpdateFailed FileStorageVolumeState = "UpdateFailed"
	FileStorageStateActive       FileStorageVolumeState = "Active"
	FileStorageStateDeleteFailed FileStorageVolumeState = "DeleteFailed"
)

// CreateFileStorageVolumeRequest represents the request to create a file storage volume
type CreateFileStorageVolumeRequest struct {
	Name             string `json:"name"`
	Description      string `json:"desc,omitempty"`
	Size             string `json:"size"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

// UpdateFileStorageVolumeRequest represents the request to update a file storage volume
type UpdateFileStorageVolumeRequest struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"desc,omitempty"`
	Size             string `json:"size,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

// FileStorageVolumeListResponse represents the response for listing file storage volumes
type FileStorageVolumeListResponse struct {
	Count int                 `json:"count"`
	Items []FileStorageVolume `json:"items"`
}

// NFSProtocolType represents the NFS protocol type
type NFSProtocolType string

const (
	NFSProtocolUnknown NFSProtocolType = "Unknown"
	NFSProtocolV4      NFSProtocolType = "NFSv4"
	NFSProtocolV3      NFSProtocolType = "NFSv3"
)

// NFSAccessType represents the access type for NFS exports
type NFSAccessType string

const (
	NFSAccessNoAccess  NFSAccessType = "NoAccess"
	NFSAccessReadOnly  NFSAccessType = "ReadOnly"
	NFSAccessReadWrite NFSAccessType = "ReadWrite"
)

// NFSSquashType represents the user squash type for NFS exports
type NFSSquashType string

const (
	NFSSquashNone   NFSSquashType = "NoSquash"
	NFSSquashRootID NFSSquashType = "RootIdSquash"
	NFSSquashRoot   NFSSquashType = "RootSquash"
	NFSSquashAll    NFSSquashType = "AllSquash"
)

// NFSAccessRule represents a host-specific access rule for NFS exports
type NFSAccessRule struct {
	Host       string        `json:"host,omitempty"`
	AccessType NFSAccessType `json:"accessType,omitempty"`
	UserSquash NFSSquashType `json:"userSquash,omitempty"`
}

// NFSExportInfo represents NFS export path information
type NFSExportInfo struct {
	NFSExportPath     string          `json:"nfsExportPath,omitempty"`
	DefaultAccessType NFSAccessType   `json:"defaultAccessType,omitempty"`
	DefaultUserSquash NFSSquashType   `json:"defaultUserSquash,omitempty"`
	Hosts             []NFSAccessRule `json:"hosts,omitempty"`
}

// FileStorageExportPath represents an NFS export path for a file storage volume
type FileStorageExportPath struct {
	Domain            string          `json:"domain,omitempty"`
	Project           string          `json:"project,omitempty"`
	PathID            string          `json:"pathId,omitempty"`
	Volume            string          `json:"volume,omitempty"`
	Description       string          `json:"desc,omitempty"`
	CreatedAt         string          `json:"createdAt,omitempty"`
	CreatedBy         string          `json:"createdBy,omitempty"`
	Protocol          NFSProtocolType `json:"Proto,omitempty"`
	NFSInfo           *NFSExportInfo  `json:"nfsInfo,omitempty"`
	ProviderExpPathID string          `json:"providerExpPathId,omitempty"`
	AvailabilityZone  string          `json:"availabilityZone,omitempty"`
}

// FileStorageExportPathKey represents the response when creating an export path
type FileStorageExportPathKey struct {
	Domain           string `json:"domain,omitempty"`
	Project          string `json:"project,omitempty"`
	PathID           string `json:"pathId,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

// FileStorageExportPathListResponse represents the response for listing export paths
type FileStorageExportPathListResponse struct {
	Count int                     `json:"count"`
	Items []FileStorageExportPath `json:"items"`
}

// CreateFileStorageExportPathRequest represents the request to create an NFS export path
type CreateFileStorageExportPathRequest struct {
	Volume           string          `json:"volume"`
	Description      string          `json:"desc,omitempty"`
	Protocol         NFSProtocolType `json:"Proto"`
	NFSInfo          *NFSExportInfo  `json:"nfsInfo,omitempty"`
	AvailabilityZone string          `json:"availabilityZone,omitempty"`
}

// UpdateFileStorageExportPathRequest represents the request to update an NFS export path
type UpdateFileStorageExportPathRequest struct {
	PathID           string          `json:"pathId"`
	Volume           string          `json:"volume"`
	Description      string          `json:"desc,omitempty"`
	Protocol         NFSProtocolType `json:"Proto"`
	NFSInfo          *NFSExportInfo  `json:"nfsInfo,omitempty"`
	AvailabilityZone string          `json:"availabilityZone,omitempty"`
}
