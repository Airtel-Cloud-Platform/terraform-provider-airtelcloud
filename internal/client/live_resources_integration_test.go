//go:build live

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

type liveTestConfig struct {
	APIEndpoint   string
	APIKey        string
	APISecret     string
	Region        string
	Organization  string
	ProjectName   string
	AZ            string
	MariaPassword string
	CreateMaria   bool
}

func getLiveTestConfig(t *testing.T) *liveTestConfig {
	t.Helper()
	cfg := &liveTestConfig{
		APIEndpoint:   os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:        os.Getenv("AIRTEL_API_KEY"),
		APISecret:     os.Getenv("AIRTEL_API_SECRET"),
		Region:        os.Getenv("AIRTEL_REGION"),
		Organization:  os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:   os.Getenv("AIRTEL_PROJECT_NAME"),
		AZ:            os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE"),
		MariaPassword: os.Getenv("AIRTEL_TEST_MARIADB_PASSWORD"),
		CreateMaria:   os.Getenv("AIRTEL_LIVE_MARIADB_CREATE") == "1",
	}
	if cfg.APIEndpoint == "" {
		t.Skip("AIRTEL_API_ENDPOINT not set")
	}
	if cfg.APIKey == "" {
		t.Skip("AIRTEL_API_KEY not set")
	}
	if cfg.APISecret == "" {
		t.Skip("AIRTEL_API_SECRET not set")
	}
	if cfg.Organization == "" {
		t.Skip("AIRTEL_ORGANIZATION not set")
	}
	if cfg.ProjectName == "" {
		t.Skip("AIRTEL_PROJECT_NAME not set")
	}
	if cfg.Region == "" {
		cfg.Region = "south"
	}
	if cfg.AZ == "" {
		cfg.AZ = "S1"
	}
	return cfg
}

func newLiveClient(t *testing.T, cfg *liveTestConfig) *Client {
	t.Helper()
	c, err := NewClient(cfg.APIEndpoint, cfg.APIKey, cfg.APISecret, cfg.Region, cfg.Organization, cfg.ProjectName, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func liveEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// TestLiveMariaDBCatalog resolves MariaDB dependency catalogs against the live API.
func TestLiveMariaDBCatalog(t *testing.T) {
	cfg := getLiveTestConfig(t)
	c := newLiveClient(t, cfg)
	ctx := context.Background()

	zone, err := c.ResolveMariaDBAvailabilityZone(ctx, cfg.AZ)
	if err != nil {
		t.Fatalf("ResolveMariaDBAvailabilityZone(%s): %v", cfg.AZ, err)
	}
	t.Logf("AZ ok: %s (%s)", zone.AZCode, zone.Name)

	versions, err := c.ListMariaDBVersions(ctx)
	if err != nil {
		t.Fatalf("ListMariaDBVersions: %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("no MariaDB versions returned")
	}
	t.Logf("versions: %d (first=%s supported=%v)", len(versions), versions[0].Version, versions[0].Supported)

	resolved, err := c.ResolveMariaDBVersion(ctx, versions[0].Version)
	if err != nil {
		t.Fatalf("ResolveMariaDBVersion: %v", err)
	}
	t.Logf("resolved version: %s", resolved.Version)

	topologies, err := c.ListMariaDBTopologies(ctx)
	if err != nil {
		t.Fatalf("ListMariaDBTopologies: %v", err)
	}
	t.Logf("topologies: %d", len(topologies))
	topo, err := c.ResolveMariaDBTopology(ctx, false)
	if err != nil {
		t.Fatalf("ResolveMariaDBTopology(false): %v", err)
	}
	if topo != models.MariaDBTopologyMaster {
		t.Fatalf("topology = %q, want master", topo)
	}

	flavors, err := c.WithAvailabilityZone(cfg.AZ).ListMariaDBFlavors(ctx)
	if err != nil {
		t.Fatalf("ListMariaDBFlavors: %v", err)
	}
	if len(flavors) == 0 {
		t.Fatal("no MariaDB flavors returned")
	}
	t.Logf("flavors: %d (first=%s ram=%d vcpus=%s)", len(flavors), flavors[0].Name, flavors[0].RAM, flavors[0].VCPUs)

	storage, err := c.ResolveMariaDBStorage(ctx, cfg.AZ, "High Performance")
	if err != nil {
		t.Fatalf("ResolveMariaDBStorage: %v", err)
	}
	t.Logf("storage: name=%s id=%d label=%s", storage.Name, storage.ID, storage.Label)

	plans, err := c.ListMariaDBProtectionPlans(ctx)
	if err != nil {
		t.Fatalf("ListMariaDBProtectionPlans: %v", err)
	}
	t.Logf("protection plans: %d", len(plans.ProtectionPlans))
}

// TestLiveMariaDBCreateGetDelete creates a real cluster when AIRTEL_LIVE_MARIADB_CREATE=1.
func TestLiveMariaDBCreateGetDelete(t *testing.T) {
	cfg := getLiveTestConfig(t)
	if !cfg.CreateMaria {
		t.Skip("set AIRTEL_LIVE_MARIADB_CREATE=1 to run MariaDB create/delete (long-running)")
	}
	if cfg.MariaPassword == "" {
		t.Fatal("AIRTEL_TEST_MARIADB_PASSWORD required when AIRTEL_LIVE_MARIADB_CREATE=1")
	}

	c := newLiveClient(t, cfg)
	ctx := context.Background()

	flavorName := os.Getenv("AIRTEL_TEST_MARIADB_FLAVOR")
	var flavor *models.MariaDBFlavor
	var err error
	if flavorName != "" {
		flavor, err = c.ResolveMariaDBFlavor(ctx, cfg.AZ, flavorName)
		if err != nil {
			t.Fatalf("ResolveMariaDBFlavor(%s): %v", flavorName, err)
		}
	} else {
		flavors, lerr := c.WithAvailabilityZone(cfg.AZ).ListMariaDBFlavors(ctx)
		if lerr != nil || len(flavors) == 0 {
			t.Fatalf("ListMariaDBFlavors: %v", lerr)
		}
		flavor = &flavors[0]
		t.Logf("using first flavor %s", flavor.Name)
	}

	storage, err := c.ResolveMariaDBStorage(ctx, cfg.AZ, "High Performance")
	if err != nil {
		t.Fatalf("ResolveMariaDBStorage: %v", err)
	}
	version := liveEnvOr("AIRTEL_TEST_MARIADB_VERSION", "10.6")
	if _, err := c.ResolveMariaDBVersion(ctx, version); err != nil {
		t.Fatalf("ResolveMariaDBVersion(%s): %v", version, err)
	}

	name := fmt.Sprintf("tft-maria-%d", time.Now().Unix()%100000)
	req := &models.CreateMariaDBClusterRequest{
		Name:            name,
		Version:         version,
		Topology:        models.MariaDBTopologyMaster,
		NumReplicas:     0,
		DatabaseName:    "admindb",
		MariaDBUsername: "tfadmin",
		Password:        cfg.MariaPassword,
		IsSuperuser:     false,
		NetworkType:     "private",
		AZIDs:           []string{cfg.AZ},
		Backup:          &models.MariaDBBackupConfig{Enabled: false},
		SecurityGroup:   &models.MariaDBSecurityGroup{AllowedIPs: []string{"10.0.0.0/8"}},
		ComputeStorageConfig: models.MariaDBComputeStorageConfig{
			Flavor: models.MariaDBFlavorRef{
				Name:  flavor.Name,
				VCPUs: flavor.VCPUs,
				RAM:   flavor.RAM,
			},
			FlavorIDs: []models.MariaDBFlavorAZID{
				{AZID: cfg.AZ, ID: flavor.ID},
			},
			StorageSizeGB:    200,
			StorageType:      storage.Name,
			StorageName:      storage.Label,
			DataVolumeTypeID: storage.ID,
		},
	}

	t.Logf("Creating MariaDB cluster %s ...", name)
	created, err := c.CreateMariaDBCluster(ctx, req)
	if err != nil {
		t.Fatalf("CreateMariaDBCluster: %v", err)
	}
	t.Logf("created uuid=%s status=%s", created.UUID, created.Status)

	defer func() {
		t.Logf("Deleting MariaDB cluster %s", created.UUID)
		if err := c.DeleteMariaDBCluster(ctx, created.UUID); err != nil {
			t.Errorf("DeleteMariaDBCluster: %v", err)
			return
		}
		if err := c.WaitForMariaDBDeleted(ctx, created.UUID, 30*time.Minute); err != nil {
			t.Errorf("WaitForMariaDBDeleted: %v", err)
		}
	}()

	got, err := c.GetMariaDBCluster(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetMariaDBCluster: %v", err)
	}
	t.Logf("get status=%s name=%s", got.Status, got.Name)

	ready, err := c.WaitForMariaDBReady(ctx, created.UUID, 45*time.Minute)
	if err != nil {
		t.Fatalf("WaitForMariaDBReady: %v", err)
	}
	t.Logf("ready status=%s port=%d", ready.Status, ready.Port)
}
