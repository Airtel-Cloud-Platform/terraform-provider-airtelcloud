package client

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	testMariaDBID      = "b941179e-c166-4b00-aece-3985896b0477"
	testMariaDBBase    = "/api/v1/dbaas/domain/test-org/project/test-project/mariadb"
	testMariaDBCluster = testMariaDBBase + "/" + testMariaDBID
	testMariaConnStr   = "mysql -u airtel -p'****' -h cluster.example -P 3350 -D admin"
)

func newMariaDBTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	c, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func sampleMariaDBCluster(status, conn string) models.MariaDBCluster {
	cluster := models.MariaDBCluster{
		UUID:            testMariaDBID,
		Name:            "cluster-label",
		LevelName:       "cluster-label",
		Version:         "10.6",
		Topology:        models.MariaDBTopologyMasterSlave,
		NumReplicas:     1,
		DatabaseName:    "admin",
		MariaDBUsername: "admin123",
		Port:            3350,
		Status:          status,
		Message:         "Cluster creation initiated",
	}
	if conn != "" {
		cluster.ConnectionString = &conn
	}
	return cluster
}

func TestCreateMariaDBCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", testMariaDBBase, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "" {
			t.Errorf("create should not require az header, got %q", r.Header.Get("ce-availability-zone"))
		}
		writeJSON(w, http.StatusOK, sampleMariaDBCluster(models.MariaDBStatusCreating, ""))
	})

	got, err := newMariaDBTestClient(t, ms).CreateMariaDBCluster(context.Background(), &models.CreateMariaDBClusterRequest{
		Name:    "cluster-label",
		Version: "10.6",
	})
	if err != nil {
		t.Fatalf("CreateMariaDBCluster() error = %v", err)
	}
	if got.UUID != testMariaDBID {
		t.Fatalf("uuid = %q", got.UUID)
	}
	if got.Status != models.MariaDBStatusCreating {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestCreateMariaDBClusterError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("POST", testMariaDBBase, 500, "Internal server error")

	_, err := newMariaDBTestClient(t, ms).CreateMariaDBCluster(context.Background(), &models.CreateMariaDBClusterRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMariaDBCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMariaDBCluster(models.MariaDBStatusActive, testMariaConnStr))
	})

	got, err := newMariaDBTestClient(t, ms).GetMariaDBCluster(context.Background(), testMariaDBID)
	if err != nil {
		t.Fatalf("GetMariaDBCluster() error = %v", err)
	}
	if got.ConnectionString == nil || *got.ConnectionString != testMariaConnStr {
		t.Fatalf("connection_string = %v", got.ConnectionString)
	}
	if got.Port != 3350 {
		t.Fatalf("port = %d", got.Port)
	}
}

func TestGetMariaDBClusterNotFound(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("GET", testMariaDBCluster, 404, "Not found")

	_, err := newMariaDBTestClient(t, ms).GetMariaDBCluster(context.Background(), testMariaDBID)
	if !IsNotFoundError(err) {
		t.Fatalf("IsNotFoundError() = false, err = %v", err)
	}
}

func TestDeleteMariaDBCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("DELETE", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := newMariaDBTestClient(t, ms).DeleteMariaDBCluster(context.Background(), testMariaDBID); err != nil {
		t.Fatalf("DeleteMariaDBCluster() error = %v", err)
	}
}

func TestDeleteMariaDBClusterNotFound(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("DELETE", testMariaDBCluster, 404, "Not found")

	if err := newMariaDBTestClient(t, ms).DeleteMariaDBCluster(context.Background(), testMariaDBID); err != nil {
		t.Fatalf("404 should be success, got %v", err)
	}
}

func TestResolveMariaDBAvailabilityZone(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", "/api/auth-mgmt/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("regionCode") != "south" {
			t.Errorf("regionCode = %q, want south", r.URL.Query().Get("regionCode"))
		}
		writeJSON(w, http.StatusOK, models.MariaDBZoneListResponse{
			Count: 2,
			Items: []models.MariaDBZone{
				{Name: "South-AZ1", AZCode: "S1", RegionCode: "south", IsActive: true},
				{Name: "South-AZ2", AZCode: "S2", RegionCode: "south", IsActive: true},
			},
		})
	})

	c := newMariaDBTestClient(t, ms)
	zone, err := c.ResolveMariaDBAvailabilityZone(context.Background(), "S1")
	if err != nil {
		t.Fatalf("ResolveMariaDBAvailabilityZone() error = %v", err)
	}
	if zone.AZCode != "S1" || zone.Name != "South-AZ1" {
		t.Fatalf("zone = %+v", zone)
	}

	if _, err := c.ResolveMariaDBAvailabilityZone(context.Background(), "N1"); err == nil {
		t.Fatal("expected not found for N1")
	}
}

func TestResolveMariaDBFlavor(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBBase+"/flavors", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q, want S1", r.Header.Get("ce-availability-zone"))
		}
		writeJSON(w, http.StatusOK, []models.MariaDBFlavor{
			{ID: 290, Name: "db.maria.uhper.ccs.xlarge", VCPUs: "4", RAM: 16, IsActive: true},
			{ID: 291, Name: "db.maria.uhper.ccs.2xlarge", VCPUs: "8", RAM: 32, IsActive: true},
		})
	})

	c := newMariaDBTestClient(t, ms)
	flavor, err := c.ResolveMariaDBFlavor(context.Background(), "S1", "db.maria.uhper.ccs.xlarge")
	if err != nil {
		t.Fatalf("ResolveMariaDBFlavor() error = %v", err)
	}
	if flavor.ID != 290 || flavor.RAM != 16 || flavor.VCPUs != "4" {
		t.Fatalf("flavor = %+v", flavor)
	}

	if _, err := c.ResolveMariaDBFlavor(context.Background(), "S1", "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestResolveMariaDBTopology(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBBase+"/topologies", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []models.MariaDBTopology{
			{Topology: models.MariaDBTopologyMaster, MinNodes: 1, MaxNodes: 1},
			{Topology: models.MariaDBTopologyMasterSlave, MinNodes: 2, MaxNodes: 11},
		})
	})

	c := newMariaDBTestClient(t, ms)
	got, err := c.ResolveMariaDBTopology(context.Background(), false)
	if err != nil || got != models.MariaDBTopologyMaster {
		t.Fatalf("master: got %q err %v", got, err)
	}
	got, err = c.ResolveMariaDBTopology(context.Background(), true)
	if err != nil || got != models.MariaDBTopologyMasterSlave {
		t.Fatalf("ha: got %q err %v", got, err)
	}
}

func TestResolveMariaDBStorage(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", "/api/v1/dbaas/domain/test-org/project/test-project/mssql/volumetypes", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q, want S1", r.Header.Get("ce-availability-zone"))
		}
		if r.URL.Query().Get("group") != "BLOCK_STORAGE" {
			t.Errorf("group = %q", r.URL.Query().Get("group"))
		}
		writeJSON(w, http.StatusOK, []models.MariaDBVolumeType{
			{ID: 10, Name: "s1_wkld_ntp02_5iops_backend", Label: "High Performance", Group: "BLOCK_STORAGE", IsActive: true},
		})
	})

	c := newMariaDBTestClient(t, ms)
	vt, err := c.ResolveMariaDBStorage(context.Background(), "S1", "High Performance")
	if err != nil {
		t.Fatalf("ResolveMariaDBStorage() error = %v", err)
	}
	if vt.ID != 10 || vt.Name != "s1_wkld_ntp02_5iops_backend" {
		t.Fatalf("volume type = %+v", vt)
	}
}

func TestResolveMariaDBProtectionPlan(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBBase+"/protection-plans", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.MariaDBProtectionPlanListResponse{
			ProtectionPlans: []models.MariaDBProtectionPlan{
				{Value: "weekly-full-daily-incr", Label: "Weekly full and daily Incremental"},
			},
		})
	})

	c := newMariaDBTestClient(t, ms)
	plan, err := c.ResolveMariaDBProtectionPlan(context.Background(), "weekly-full-daily-incr")
	if err != nil {
		t.Fatalf("ResolveMariaDBProtectionPlan() error = %v", err)
	}
	if plan.Value != "weekly-full-daily-incr" {
		t.Fatalf("plan = %+v", plan)
	}
	if _, err := c.ResolveMariaDBProtectionPlan(context.Background(), "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestResolveMariaDBVersion(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBBase+"/versions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []models.MariaDBVersion{
			{Version: "10.6", Supported: true},
			{Version: "10.11", Supported: true},
			{Version: "11.4", Supported: true},
			{Version: "11.8", Supported: false},
		})
	})

	c := newMariaDBTestClient(t, ms)
	v, err := c.ResolveMariaDBVersion(context.Background(), "10.6")
	if err != nil {
		t.Fatalf("ResolveMariaDBVersion() error = %v", err)
	}
	if v.Version != "10.6" || !v.Supported {
		t.Fatalf("version = %+v", v)
	}
	if _, err := c.ResolveMariaDBVersion(context.Background(), "11.8"); err == nil {
		t.Fatal("expected unsupported version to fail")
	}
	if _, err := c.ResolveMariaDBVersion(context.Background(), "9.0"); err == nil {
		t.Fatal("expected missing version to fail")
	}
}

func TestWaitForMariaDBReady(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMariaDBCluster(models.MariaDBStatusActive, testMariaConnStr))
	})

	got, err := newMariaDBTestClient(t, ms).WaitForMariaDBReady(context.Background(), testMariaDBID, 15*time.Minute)
	if err != nil {
		t.Fatalf("WaitForMariaDBReady() error = %v", err)
	}
	if got.Status != models.MariaDBStatusActive {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestWaitForMariaDBReadyFailed(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		cluster := sampleMariaDBCluster(models.MariaDBStatusFailed, "")
		cluster.Message = "provisioning failed"
		writeJSON(w, http.StatusOK, cluster)
	})

	_, err := newMariaDBTestClient(t, ms).WaitForMariaDBReady(context.Background(), testMariaDBID, 15*time.Minute)
	if err == nil || !strings.Contains(err.Error(), "Failed") {
		t.Fatalf("expected failed-state error, got %v", err)
	}
}

func TestWaitForMariaDBReadyTimeout(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMariaDBCluster(models.MariaDBStatusCreating, ""))
	})

	_, err := newMariaDBTestClient(t, ms).WaitForMariaDBReady(context.Background(), testMariaDBID, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestWaitForMariaDBDeleted(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("GET", testMariaDBCluster, 404, "Not found")

	if err := newMariaDBTestClient(t, ms).WaitForMariaDBDeleted(context.Background(), testMariaDBID, 15*time.Minute); err != nil {
		t.Fatalf("WaitForMariaDBDeleted() error = %v", err)
	}
}

func TestWaitForMariaDBDeletedStatus(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMariaDBCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMariaDBCluster(models.MariaDBStatusDeleted, ""))
	})

	if err := newMariaDBTestClient(t, ms).WaitForMariaDBDeleted(context.Background(), testMariaDBID, 15*time.Minute); err != nil {
		t.Fatalf("WaitForMariaDBDeleted() error = %v", err)
	}
}
