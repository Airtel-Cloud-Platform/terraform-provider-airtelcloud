package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	MSSQLStatusActive    = "Active"
	MSSQLStatusRequested = "Requested"
	MSSQLStatusInitiated = "Initiated"
	MSSQLStatusResizing  = "Resizing"
	MSSQLStatusDeleting  = "Deleting"
	MSSQLStatusFailed    = "Failed"

	MSSQLFixedUser               = "admin"
	MSSQLFixedTPS                = "500"
	MSSQLFixedPerTransactionSize = "0"
	MSSQLDefaultStorageType      = "High Performance"
	MSSQLStorageMinGiB           = 200
	MSSQLStorageMaxGiB           = 2048
	MSSQLStorageStepGiB          = 10
	MSSQLRAMDivisor              = 1024
)

// CreateMSSQLRequest is the POST body for an MSSQL instance.
type CreateMSSQLRequest struct {
	Name                 string                    `json:"name"`
	User                 string                    `json:"user"`
	Description          string                    `json:"description,omitempty"`
	AZName               string                    `json:"az_name"`
	AZIDs                []string                  `json:"az_ids"`
	CollationFeatures    MSSQLCollationFeatures    `json:"collation_features"`
	DBConfig             MSSQLDBConfig             `json:"db_config"`
	ComputeStorageConfig MSSQLComputeStorageConfig `json:"compute_storage_config"`
	Backup               MSSQLBackup               `json:"backup"`
	AdvancedSettings     MSSQLAdvancedSettings     `json:"advancedsettings"`
	VolumeSize           string                    `json:"volume_size"`
	VolumeTypeProviderID string                    `json:"volume_type_provider_id"`
	VolumeTypeID         int                       `json:"volume_type_id"`
	DBName               string                    `json:"db_name"`
	MSSQLUsername        string                    `json:"mssql_username"`
	Password             string                    `json:"password"`
	ConfirmPassword      string                    `json:"confirm_password"`
	AddLabels            []string                  `json:"add_labels"`
	EnableTDE            bool                      `json:"enable_tde"`
	SecurityGroup        *MSSQLSecurityGroup       `json:"security_group,omitempty"`
}

// MSSQLCollationFeatures is the collation and selected feature values.
type MSSQLCollationFeatures struct {
	Collation string   `json:"collation"`
	Features  []string `json:"features"`
}

// MSSQLDBConfig is the edition and version taken from the editions catalog.
type MSSQLDBConfig struct {
	Version string `json:"version"`
	Edition string `json:"edition"`
}

// MSSQLComputeStorageConfig is the compute and storage section of the create body.
type MSSQLComputeStorageConfig struct {
	Flavor             MSSQLFlavorRef    `json:"flavor"`
	FlavorIDs          []MSSQLFlavorAZID `json:"flavor_ids"`
	TPS                string            `json:"tps"`
	PerTransactionSize string            `json:"per_transaction_size"`
	DatabaseSize       string            `json:"database_size"`
	StorageType        string            `json:"storage_type"`
}

// MSSQLFlavorRef is the flavor summary sent on create. RAM is gigabytes.
type MSSQLFlavorRef struct {
	Name  string `json:"name"`
	VCPUs string `json:"vcpus"`
	RAM   int    `json:"ram"`
}

// MSSQLFlavorAZID binds a catalog flavor ID to an availability zone code.
type MSSQLFlavorAZID struct {
	AZID string `json:"az_id"`
	ID   int    `json:"id"`
}

// MSSQLBackup is the backup object. Netbackup fields are omitted when backup is disabled.
type MSSQLBackup struct {
	Enabled       bool                 `json:"enabled"`
	NetbackupName []MSSQLNetbackupName `json:"netbackup_name,omitempty"`
	NetbackupID   string               `json:"netbackup_id,omitempty"`
}

// MSSQLNetbackupName is one catalog plan name and description.
type MSSQLNetbackupName struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MSSQLAdvancedSettings carries only the enabled flag.
type MSSQLAdvancedSettings struct {
	Enabled bool `json:"enabled"`
}

// MSSQLSecurityGroup is the allowed-IP object on create.
type MSSQLSecurityGroup struct {
	AllowedIPs []string `json:"allowed_ips"`
}

// MSSQLInstance is the create and get response. Only fields used by Terraform are mapped.
type MSSQLInstance struct {
	UUID                 string                    `json:"uuid"`
	Name                 string                    `json:"name"`
	DBName               string                    `json:"db_name"`
	Description          string                    `json:"description"`
	AZName               string                    `json:"az_name"`
	MSSQLUsername        string                    `json:"mssql_username"`
	ConnectionString     *string                   `json:"connection_string"`
	Status               string                    `json:"status"`
	Message              string                    `json:"message"`
	CreatedAt            string                    `json:"created_at"`
	Port                 int                       `json:"port"`
	VolumeSize           FlexInt                   `json:"volume_size"`
	EnableTDE            bool                      `json:"enable_tde"`
	AddLabels            []string                  `json:"add_labels"`
	AllowedIPs           []string                  `json:"allowed_ips"`
	CollationFeatures    MSSQLCollationFeatures    `json:"collation_features"`
	DBConfig             MSSQLDBConfig             `json:"db_config"`
	ComputeStorageConfig MSSQLComputeStorageConfig `json:"compute_storage_config"`
	AdvancedSettings     MSSQLAdvancedSettings     `json:"advancedsettings"`
	Backup               MSSQLBackup               `json:"backup"`
}

// MSSQLEdition is one editions-catalog item.
type MSSQLEdition struct {
	Edition  string   `json:"edition"`
	Versions []string `json:"versions"`
}

// MSSQLCollation is one collations-catalog item for an edition.
type MSSQLCollation struct {
	Collation string             `json:"collation"`
	Features  []MSSQLFeatureItem `json:"features"`
}

// MSSQLFeatureItem is a selectable feature on a collation.
type MSSQLFeatureItem struct {
	Label    string   `json:"label"`
	Value    string   `json:"value"`
	Editions []string `json:"editions"`
}

// MSSQLFlavor is a flavors-catalog item. RAM is megabytes, as a string.
type MSSQLFlavor struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	VCPUs  string `json:"vcpus"`
	RAM    string `json:"ram"`
	Public bool   `json:"public"`
}

// RAMGiB converts catalog RAM from megabytes to gigabytes.
func (f MSSQLFlavor) RAMGiB() (int, error) {
	mb, err := strconv.Atoi(strings.TrimSpace(f.RAM))
	if err != nil {
		return 0, fmt.Errorf("flavor %q ram %q is not a number of megabytes", f.Name, f.RAM)
	}
	if mb <= 0 || mb%MSSQLRAMDivisor != 0 {
		return 0, fmt.Errorf("flavor %q ram %d MB is not a whole number of GB", f.Name, mb)
	}
	return mb / MSSQLRAMDivisor, nil
}

// MSSQLVolumeType is a volume-types catalog item.
type MSSQLVolumeType struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	ProviderVTID string `json:"provider_vt_id"`
	Group        string `json:"group"`
	IsActive     bool   `json:"is_active"`
}

// MSSQLSLOListResponse wraps GET .../mssql/netbackup/slos.
type MSSQLSLOListResponse struct {
	Data []MSSQLSLO `json:"data"`
}

// MSSQLSLO is one protection plan in the netbackup catalog.
type MSSQLSLO struct {
	ID         string             `json:"id"`
	Attributes MSSQLSLOAttributes `json:"attributes"`
}

// MSSQLSLOAttributes holds the plan name the user selects.
type MSSQLSLOAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FlexInt unmarshals a JSON number or a numeric string.
type FlexInt int

// UnmarshalJSON accepts either 200 or "200".
func (f *FlexInt) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if strings.TrimSpace(s) == "" {
			return nil
		}
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return err
		}
		*f = FlexInt(n)
		return nil
	}
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = FlexInt(n)
	return nil
}

// MSSQLStorageSizeAllowed reports whether size is inside the database-size catalog.
func MSSQLStorageSizeAllowed(size int64) bool {
	if size < MSSQLStorageMinGiB || size > MSSQLStorageMaxGiB {
		return false
	}
	return (size-MSSQLStorageMinGiB)%MSSQLStorageStepGiB == 0
}
