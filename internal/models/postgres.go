package models

// Topology values sent on create. high_availability=false maps to standalone;
// high_availability=true maps to primary-standby.
const (
	PostgresTopologyStandalone     = "standalone"
	PostgresTopologyPrimaryStandby = "primary-standby"
)

// Cluster lifecycle statuses from the DBaaS GET response.
const (
	PostgresStatusCreating     = "Creating"
	PostgresStatusActive       = "Active"
	PostgresStatusExpanding    = "Expanding"
	PostgresStatusResizing     = "Resizing"
	PostgresStatusPatching     = "Patching"
	PostgresStatusRestoring    = "Restoring"
	PostgresStatusDeleting     = "Deleting"
	PostgresStatusFailed       = "Failed"
	PostgresStatusDeleteFailed = "Delete_Failed"
	PostgresStatusDeleted      = "Deleted"
)

// CreatePostgresClusterRequest is the POST body for creating a PostgreSQL cluster.
type CreatePostgresClusterRequest struct {
	Name                 string                       `json:"name"`
	Description          string                       `json:"description,omitempty"`
	Version              string                       `json:"version"`
	Topology             string                       `json:"topology"`
	NumReplicas          int                          `json:"num_replicas,omitempty"`
	DatabaseName         string                       `json:"database_name"`
	PostgresUsername     string                       `json:"postgres_username"`
	Password             string                       `json:"password"`
	IsSuperuser          bool                         `json:"is_superuser"`
	NetworkType          string                       `json:"network_type,omitempty"`
	ComputeStorageConfig PostgresComputeStorageConfig `json:"compute_storage_config"`
	AZIDs                []string                     `json:"az_ids"`
	PGExtensions         []string                     `json:"pg_extensions,omitempty"`
	Labels               []string                     `json:"labels,omitempty"`
	Backup               *PostgresBackupConfig        `json:"backup,omitempty"`
	SecurityGroup        *PostgresSecurityGroup       `json:"security_group,omitempty"`
}

// PostgresComputeStorageConfig is the nested compute/storage object on create and get.
type PostgresComputeStorageConfig struct {
	Flavor           PostgresFlavorRef    `json:"flavor"`
	FlavorIDs        []PostgresFlavorAZID `json:"flavor_ids"`
	StorageSizeGB    int                  `json:"storage_size_gb"`
	StorageType      string               `json:"storage_type"`
	StorageName      string               `json:"storage_name"`
	DataVolumeTypeID int                  `json:"data_volume_type_id"`
}

// PostgresFlavorRef is the flavor name/RAM pair sent on create.
type PostgresFlavorRef struct {
	Name string `json:"name"`
	RAM  int    `json:"ram"`
}

// PostgresFlavorAZID binds a catalog flavor ID to an availability zone code.
type PostgresFlavorAZID struct {
	AZID string `json:"az_id"`
	ID   int    `json:"id"`
}

// PostgresBackupConfig is the backup object on create and get.
type PostgresBackupConfig struct {
	Enabled          bool   `json:"enabled"`
	ProtectionPlan   string `json:"protection_plan,omitempty"`
	CompressionLevel int    `json:"compression_level,omitempty"`
	Retention        int    `json:"retention,omitempty"`
	ScheduleTime     string `json:"schedule_time,omitempty"`
	ScheduleDay      string `json:"schedule_day,omitempty"`
}

// PostgresSecurityGroup is the allowed-IP object on create.
type PostgresSecurityGroup struct {
	AllowedIPs []string `json:"allowed_ips,omitempty"`
}

// PostgresCluster is the create/get response for a PostgreSQL cluster.
// Only fields needed for Terraform state mapping are included.
type PostgresCluster struct {
	UUID                 string                       `json:"uuid"`
	Name                 string                       `json:"name"`
	LevelName            string                       `json:"level_name,omitempty"`
	Version              string                       `json:"version"`
	Topology             string                       `json:"topology"`
	NumReplicas          int                          `json:"num_replicas,omitempty"`
	DatabaseName         string                       `json:"database_name"`
	PostgresUsername     string                       `json:"postgres_username"`
	IsSuperuser          bool                         `json:"is_superuser"`
	Description          string                       `json:"description,omitempty"`
	ComputeStorageConfig PostgresComputeStorageConfig `json:"compute_storage_config"`
	NetworkConfig        *PostgresNetworkConfig       `json:"network_config,omitempty"`
	AzNames              []string                     `json:"az_names,omitempty"`
	PGExtensions         []string                     `json:"pg_extensions,omitempty"`
	Labels               []string                     `json:"labels,omitempty"`
	Backup               *PostgresBackupConfig        `json:"backup,omitempty"`
	ConnectionString     *string                      `json:"connection_string"`
	Status               string                       `json:"status"`
	Message              string                       `json:"message,omitempty"`
	CreatedAt            string                       `json:"created_at,omitempty"`
}

// PostgresNetworkConfig is the nested network object returned on get.
type PostgresNetworkConfig struct {
	NetworkType string `json:"network_type,omitempty"`
}

// PostgresFlavor is a postgres flavors catalog item.
type PostgresFlavor struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	RAM         int    `json:"ram"`
	VCPUs       string `json:"vcpus"`
	Disk        string `json:"disk"`
	IsActive    bool   `json:"is_active"`
	Public      bool   `json:"public"`
}

// PostgresTopology is a postgres topologies catalog item.
type PostgresTopology struct {
	Topology    string `json:"topology"`
	Label       string `json:"label"`
	Description string `json:"description"`
	MinNodes    int    `json:"min_nodes"`
	MaxNodes    int    `json:"max_nodes"`
}

// PostgresVolumeType is a storage catalog item from the volume-types API.
type PostgresVolumeType struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Group    string `json:"group,omitempty"`
	IsActive bool   `json:"is_active"`
}

// PostgresZone is an availability-zone catalog item.
type PostgresZone struct {
	Name       string `json:"name"`
	AZCode     string `json:"azCode"`
	RegionCode string `json:"regionCode"`
	IsActive   bool   `json:"isActive"`
	IsDefault  bool   `json:"isDefault"`
}

// PostgresZoneListResponse wraps GET /api/auth-mgmt/v1/zones.
type PostgresZoneListResponse struct {
	Count int            `json:"count"`
	Items []PostgresZone `json:"items"`
}

// PostgresProtectionPlan is a DBaaS backup plan catalog item.
type PostgresProtectionPlan struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// PostgresProtectionPlanListResponse wraps GET /api/v1/protection-plans.
type PostgresProtectionPlanListResponse struct {
	ProtectionPlans []PostgresProtectionPlan `json:"protection_plans"`
}
