package models

// BaremetalAllocateStorageMap models an extra storage mapping in baremetal
// allocate/update API requests.
type BaremetalAllocateStorageMap struct {
	Name        string `json:"name,omitempty"`
	Size        string `json:"size,omitempty"`
	Path        string `json:"path,omitempty"`
	Type        string `json:"type,omitempty"`
	FileSystem  string `json:"fileSystem,omitempty"`
	ForceFormat bool   `json:"forceFormat,omitempty"`
}

// BaremetalSubnetConfig models a subnet entry inside networkInterface.subnets.
type BaremetalSubnetConfig struct {
	SubnetID  string `json:"subnetId,omitempty"`
	IsPrimary bool   `json:"isPrimary,omitempty"`
}

// BaremetalNetworkInterface models the networkInterface field used by allocate.
type BaremetalNetworkInterface struct {
	Name     string                  `json:"name,omitempty"`
	SubnetID string                  `json:"subnetId,omitempty"`
	Subnets  []BaremetalSubnetConfig `json:"subnets,omitempty"`
}

// BaremetalMetadata models optional metadata passed during allocation.
type BaremetalMetadata struct {
	Keypair string `json:"keypair,omitempty"`
}

// BaremetalBackupConfig models optional baremetal backup configuration.
type BaremetalBackupConfig struct {
	PolicyName           string   `json:"policyName,omitempty"`
	ScheduleType         string   `json:"scheduleType,omitempty"`
	StartTime            string   `json:"startTime,omitempty"`
	IncrDays             []int    `json:"incrDays,omitempty"`
	FullDays             []int    `json:"fullDays,omitempty"`
	FullRetention        int      `json:"fullRetention,omitempty"`
	FullRetentionUnit    string   `json:"fullRetentionUnit,omitempty"`
	IncrRetention        int      `json:"incrRetention,omitempty"`
	IncrRetentionUnit    string   `json:"incrRetentionUnit,omitempty"`
	BackupSelections     []string `json:"backupSelections,omitempty"`
	PolicyEnabled        bool     `json:"policyEnabled,omitempty"`
	FullDate             int      `json:"fullDate,omitempty"`
	MonthlyRetention     int      `json:"monthlyRetention,omitempty"`
	MonthlyRetentionUnit string   `json:"monthlyRetentionUnit,omitempty"`
}

// AllocateBaremetalRequest models POST /server allocate payload.
type AllocateBaremetalRequest struct {
	Name             string                        `json:"name,omitempty"`
	Flavor           string                        `json:"flavor,omitempty"`
	OSImage          string                        `json:"osImage,omitempty"`
	CloudInit        string                        `json:"cloudInit,omitempty"`
	KeypairID        string                        `json:"keypairId,omitempty"`
	PublicKey        string                        `json:"publicKey,omitempty"`
	Storage          []BaremetalAllocateStorageMap `json:"storage,omitempty"`
	IsReserved       bool                          `json:"isReserved,omitempty"`
	SystemID         string                        `json:"systemId,omitempty"`
	NetworkInterface *BaremetalNetworkInterface    `json:"networkInterface,omitempty"`
	Tags             []string                      `json:"tags,omitempty"`
	Metadata         *BaremetalMetadata            `json:"metadata,omitempty"`
	BackupConfig     *BaremetalBackupConfig        `json:"backupConfig,omitempty"`
}

// UpdateBaremetalRequest models PUT /server/{name} payload.
type UpdateBaremetalRequest struct {
	Storage       []BaremetalAllocateStorageMap `json:"storage,omitempty"`
	PolicyEnabled *bool                         `json:"policyEnabled,omitempty"`
}

// ReleaseBaremetalOptions models optional query parameters for DELETE /server/{name}.
type ReleaseBaremetalOptions struct {
	SystemID    string
	DeleteDisks *bool
	SecureErase *bool
}

// BaremetalListResponse wraps the baremetal server list API response.
type BaremetalListResponse struct {
	Count int         `json:"count"`
	Items []Baremetal `json:"items"`
}

// BaremetalDetailResponse wraps GET /server/{name} detail payload.
type BaremetalDetailResponse struct {
	ServerDetails Baremetal   `json:"serverDetails"`
	NetworkInfo   NetworkInfo `json:"networkInfo"`
}

// NetworkInfo is the nested network block of the /server/{name} detail response.
type NetworkInfo struct {
	// PortID -> virtual server node backend_port_id. The API sends it as either a
	// JSON string ("17963") or a number, so use FlexInt64 to accept both.
	PortID FlexInt64 `json:"portId"`
}

// Baremetal represents a baremetal server (from the baremetal-manager API).
// It is used as a backend pool member for load balancer virtual servers.
type Baremetal struct {
	Name             string      `json:"name"`     // friendly name, e.g. "amd-test2"
	Hostname         string      `json:"hostname"` // e.g. "n1-copper-bm09"
	State            string      `json:"state"`    // "Ready" | "Failed" | ...
	IPAddr           []string    `json:"ipAddr"`
	UUID             string      `json:"uuid"` // -> resource_id
	Flavor           string      `json:"flavor,omitempty"`
	PowerState       string      `json:"powerState,omitempty"`
	AvailabilityZone string      `json:"availabilityZone,omitempty"`
	NetworkInfo      NetworkInfo `json:"networkInfo"` // carries the backend port id
	// PortID is derived from NetworkInfo.PortID (networkInfo.portId in the
	// /server/{name} detail response) and maps to a virtual server node's
	// backend_port_id. It is populated by ResolveBaremetalNode, not by the API.
	PortID int `json:"-"`
}

// PrimaryIP returns the first fixed IP of the baremetal server, or "".
func (b *Baremetal) PrimaryIP() string {
	if len(b.IPAddr) > 0 {
		return b.IPAddr[0]
	}
	return ""
}
