package models

// Topology values sent on create. high_availability=false maps to master;
// high_availability=true maps to master-slave.
const (
	MariaDBTopologyMaster      = "master"
	MariaDBTopologyMasterSlave = "master-slave"
)

// Cluster lifecycle statuses from the DBaaS GET response.
const (
	MariaDBStatusCreating     = "Creating"
	MariaDBStatusActive       = "Active"
	MariaDBStatusExpanding    = "Expanding"
	MariaDBStatusResizing     = "Resizing"
	MariaDBStatusPatching     = "Patching"
	MariaDBStatusRestoring    = "Restoring"
	MariaDBStatusDeleting     = "Deleting"
	MariaDBStatusFailed       = "Failed"
	MariaDBStatusDeleteFailed = "Delete_Failed"
	MariaDBStatusDeleted      = "Deleted"
)

// CreateMariaDBClusterRequest is the POST body for creating a MariaDB cluster.
type CreateMariaDBClusterRequest struct {
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	Version              string                      `json:"version"`
	Topology             string                      `json:"topology"`
	NumReplicas          int                         `json:"num_replicas"`
	DatabaseName         string                      `json:"database_name"`
	MariaDBUsername      string                      `json:"mariadb_username"`
	Password             string                      `json:"password"`
	IsSuperuser          bool                        `json:"is_superuser"`
	NetworkType          string                      `json:"network_type,omitempty"`
	ComputeStorageConfig MariaDBComputeStorageConfig `json:"compute_storage_config"`
	AZIDs                []string                    `json:"az_ids"`
	Labels               []string                    `json:"labels,omitempty"`
	Backup               *MariaDBBackupConfig        `json:"backup,omitempty"`
	SecurityGroup        *MariaDBSecurityGroup       `json:"security_group,omitempty"`
}

// MariaDBComputeStorageConfig is the nested compute/storage object on create and get.
type MariaDBComputeStorageConfig struct {
	Flavor           MariaDBFlavorRef    `json:"flavor"`
	FlavorIDs        []MariaDBFlavorAZID `json:"flavor_ids"`
	StorageSizeGB    int                 `json:"storage_size_gb"`
	StorageType      string              `json:"storage_type"`
	StorageName      string              `json:"storage_name"`
	DataVolumeTypeID int                 `json:"data_volume_type_id"`
}

// MariaDBFlavorRef is the flavor name/vcpus/RAM triple sent on create.
// Catalog RAM is already in GB (no /1024 conversion).
type MariaDBFlavorRef struct {
	Name  string `json:"name"`
	VCPUs string `json:"vcpus"`
	RAM   int    `json:"ram"`
}

// MariaDBFlavorAZID binds a catalog flavor ID to an availability zone code.
type MariaDBFlavorAZID struct {
	AZID string `json:"az_id"`
	ID   int    `json:"id"`
}

// MariaDBBackupConfig is the backup object on create and get.
type MariaDBBackupConfig struct {
	Enabled          bool   `json:"enabled"`
	ProtectionPlan   string `json:"protection_plan,omitempty"`
	CompressionLevel int    `json:"compression_level,omitempty"`
	Retention        int    `json:"retention,omitempty"`
	ScheduleTime     string `json:"schedule_time,omitempty"`
	ScheduleDay      string `json:"schedule_day,omitempty"`
}

// MariaDBSecurityGroup is the allowed-IP object on create.
type MariaDBSecurityGroup struct {
	AllowedIPs []string `json:"allowed_ips,omitempty"`
}

// MariaDBCluster is the create/get response for a MariaDB cluster.
// Only fields needed for Terraform state mapping are included.
type MariaDBCluster struct {
	UUID                 string                      `json:"uuid"`
	Name                 string                      `json:"name"`
	LevelName            string                      `json:"level_name,omitempty"`
	Version              string                      `json:"version"`
	Topology             string                      `json:"topology"`
	NumReplicas          int                         `json:"num_replicas,omitempty"`
	DatabaseName         string                      `json:"database_name"`
	MariaDBUsername      string                      `json:"mariadb_username"`
	IsSuperuser          bool                        `json:"is_superuser"`
	Description          string                      `json:"description,omitempty"`
	ComputeStorageConfig MariaDBComputeStorageConfig `json:"compute_storage_config"`
	NetworkConfig        *MariaDBNetworkConfig       `json:"network_config,omitempty"`
	AzNames              []string                    `json:"az_names,omitempty"`
	Labels               []string                    `json:"labels,omitempty"`
	Backup               *MariaDBBackupConfig        `json:"backup,omitempty"`
	ConnectionString     *string                     `json:"connection_string"`
	Port                 int                         `json:"port"`
	Status               string                      `json:"status"`
	Message              string                      `json:"message,omitempty"`
	CreatedAt            string                      `json:"created_at,omitempty"`
}

// MariaDBNetworkConfig is the nested network object returned on get.
type MariaDBNetworkConfig struct {
	NetworkType string `json:"network_type,omitempty"`
}

// MariaDBFlavor is a mariadb flavors catalog item. RAM is already in GB.
type MariaDBFlavor struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	RAM         int    `json:"ram"`
	VCPUs       string `json:"vcpus"`
	Disk        string `json:"disk"`
	IsActive    bool   `json:"is_active"`
	Public      bool   `json:"public"`
}

// MariaDBTopology is a mariadb topologies catalog item.
type MariaDBTopology struct {
	Topology    string `json:"topology"`
	Label       string `json:"label"`
	Description string `json:"description"`
	MinNodes    int    `json:"min_nodes"`
	MaxNodes    int    `json:"max_nodes"`
}

// MariaDBVolumeType is a storage catalog item from the volume-types API.
type MariaDBVolumeType struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Group    string `json:"group,omitempty"`
	IsActive bool   `json:"is_active"`
}

// MariaDBZone is an availability-zone catalog item.
type MariaDBZone struct {
	Name       string `json:"name"`
	AZCode     string `json:"azCode"`
	RegionCode string `json:"regionCode"`
	IsActive   bool   `json:"isActive"`
	IsDefault  bool   `json:"isDefault"`
}

// MariaDBZoneListResponse wraps GET /api/auth-mgmt/v1/zones.
type MariaDBZoneListResponse struct {
	Count int           `json:"count"`
	Items []MariaDBZone `json:"items"`
}

// MariaDBProtectionPlan is a DBaaS backup plan catalog item.
type MariaDBProtectionPlan struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// MariaDBProtectionPlanListResponse wraps GET .../mariadb/protection-plans.
type MariaDBProtectionPlanListResponse struct {
	ProtectionPlans []MariaDBProtectionPlan `json:"protection_plans"`
}

// MariaDBVersion is a versions catalog item.
type MariaDBVersion struct {
	Version   string `json:"version"`
	Supported bool   `json:"supported"`
}
