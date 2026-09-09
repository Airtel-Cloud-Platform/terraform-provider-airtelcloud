package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	testPostgresID      = "e9953b5d-aa93-4c45-b9fe-6b977b2bd7be"
	testPostgresBase    = "/api/v1/dbaas/domain/test-org/project/test-project/postgres"
	testPostgresCluster = testPostgresBase + "/" + testPostgresID
	testConnString      = "postgresql://admin123:****@cluster.example:5451/terra"
)

func newPostgresTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	c, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func sampleCluster(status, conn string) models.PostgresCluster {
	cluster := models.PostgresCluster{
		UUID:             testPostgresID,
		Name:             "cluster-label",
		LevelName:        "cluster-label",
		Version:          "17",
		Topology:         models.PostgresTopologyPrimaryStandby,
		NumReplicas:      1,
		DatabaseName:     "terra",
		PostgresUsername: "admin123",
		Status:           status,
		Message:          "Cluster creation initiated",
	}
	if conn != "" {
		cluster.ConnectionString = &conn
	}
	return cluster
}

func TestCreatePostgresCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", testPostgresBase, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "" {
			t.Errorf("create should not require az header, got %q", r.Header.Get("ce-availability-zone"))
		}
		writeJSON(w, http.StatusOK, sampleCluster(models.PostgresStatusCreating, ""))
	})

	got, err := newPostgresTestClient(t, ms).CreatePostgresCluster(context.Background(), &models.CreatePostgresClusterRequest{
		Name:    "cluster-label",
		Version: "17",
	})
	if err != nil {
		t.Fatalf("CreatePostgresCluster() error = %v", err)
	}
	if got.UUID != testPostgresID {
		t.Fatalf("uuid = %q", got.UUID)
	}
	if got.Status != models.PostgresStatusCreating {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestCreatePostgresClusterError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("POST", testPostgresBase, 500, "Internal server error")

	_, err := newPostgresTestClient(t, ms).CreatePostgresCluster(context.Background(), &models.CreatePostgresClusterRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPostgresCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleCluster(models.PostgresStatusActive, testConnString))
	})

	got, err := newPostgresTestClient(t, ms).GetPostgresCluster(context.Background(), testPostgresID)
	if err != nil {
		t.Fatalf("GetPostgresCluster() error = %v", err)
	}
	if got.ConnectionString == nil || *got.ConnectionString != testConnString {
		t.Fatalf("connection_string = %v", got.ConnectionString)
	}
}

func TestGetPostgresClusterNotFound(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("GET", testPostgresCluster, 404, "Not found")

	_, err := newPostgresTestClient(t, ms).GetPostgresCluster(context.Background(), testPostgresID)
	if !IsNotFoundError(err) {
		t.Fatalf("IsNotFoundError() = false, err = %v", err)
	}
}

func TestDeletePostgresCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("DELETE", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := newPostgresTestClient(t, ms).DeletePostgresCluster(context.Background(), testPostgresID); err != nil {
		t.Fatalf("DeletePostgresCluster() error = %v", err)
	}
}

func TestDeletePostgresClusterNotFound(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("DELETE", testPostgresCluster, 404, "Not found")

	if err := newPostgresTestClient(t, ms).DeletePostgresCluster(context.Background(), testPostgresID); err != nil {
		t.Fatalf("404 should be success, got %v", err)
	}
}

func TestResolvePostgresAvailabilityZone(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", "/api/auth-mgmt/v1/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("regionCode") != "south" {
			t.Errorf("regionCode = %q, want south", r.URL.Query().Get("regionCode"))
		}
		writeJSON(w, http.StatusOK, models.PostgresZoneListResponse{
			Count: 2,
			Items: []models.PostgresZone{
				{Name: "South-AZ1", AZCode: "S1", RegionCode: "south", IsActive: true},
				{Name: "South-AZ2", AZCode: "S2", RegionCode: "south", IsActive: true},
			},
		})
	})

	c := newPostgresTestClient(t, ms)
	zone, err := c.ResolvePostgresAvailabilityZone(context.Background(), "S1")
	if err != nil {
		t.Fatalf("ResolvePostgresAvailabilityZone() error = %v", err)
	}
	if zone.AZCode != "S1" || zone.Name != "South-AZ1" {
		t.Fatalf("zone = %+v", zone)
	}

	if _, err := c.ResolvePostgresAvailabilityZone(context.Background(), "N1"); err == nil {
		t.Fatal("expected not found for N1")
	}
}

func TestResolvePostgresFlavor(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresBase+"/flavors", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q, want S1", r.Header.Get("ce-availability-zone"))
		}
		writeJSON(w, http.StatusOK, []models.PostgresFlavor{
			{ID: 249, Name: "db.postgres.uhper.ccs.xlarge", RAM: 16, IsActive: true},
			{ID: 279, Name: "db.postgres.uhper.ccs.2xlarge", RAM: 32, IsActive: true},
		})
	})

	c := newPostgresTestClient(t, ms)
	flavor, err := c.ResolvePostgresFlavor(context.Background(), "S1", "db.postgres.uhper.ccs.xlarge")
	if err != nil {
		t.Fatalf("ResolvePostgresFlavor() error = %v", err)
	}
	if flavor.ID != 249 || flavor.RAM != 16 {
		t.Fatalf("flavor = %+v", flavor)
	}

	if _, err := c.ResolvePostgresFlavor(context.Background(), "S1", "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestResolvePostgresTopology(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresBase+"/topologies", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []models.PostgresTopology{
			{Topology: models.PostgresTopologyStandalone, MinNodes: 1, MaxNodes: 1},
			{Topology: models.PostgresTopologyPrimaryStandby, MinNodes: 2, MaxNodes: 3},
		})
	})

	c := newPostgresTestClient(t, ms)
	got, err := c.ResolvePostgresTopology(context.Background(), false)
	if err != nil || got != models.PostgresTopologyStandalone {
		t.Fatalf("standalone: got %q err %v", got, err)
	}
	got, err = c.ResolvePostgresTopology(context.Background(), true)
	if err != nil || got != models.PostgresTopologyPrimaryStandby {
		t.Fatalf("ha: got %q err %v", got, err)
	}
}

func TestResolvePostgresStorage(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", "/api/v1/dbaas/domain/test-org/project/test-project/mssql/volumetypes", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q, want S1", r.Header.Get("ce-availability-zone"))
		}
		if r.URL.Query().Get("group") != "BLOCK_STORAGE" {
			t.Errorf("group = %q", r.URL.Query().Get("group"))
		}
		writeJSON(w, http.StatusOK, []models.PostgresVolumeType{
			{ID: 10, Name: "s1_wkld_ntp02_5iops_backend", Label: "High Performance", Group: "BLOCK_STORAGE", IsActive: true},
		})
	})

	c := newPostgresTestClient(t, ms)
	vt, err := c.ResolvePostgresStorage(context.Background(), "S1", "High Performance")
	if err != nil {
		t.Fatalf("ResolvePostgresStorage() error = %v", err)
	}
	if vt.ID != 10 || vt.Name != "s1_wkld_ntp02_5iops_backend" {
		t.Fatalf("volume type = %+v", vt)
	}
}

func TestResolvePostgresProtectionPlan(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresBase+"/protection-plans", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.PostgresProtectionPlanListResponse{
			ProtectionPlans: []models.PostgresProtectionPlan{
				{Value: "weekly-full-daily-incr", Label: "Weekly full and daily Incremental"},
			},
		})
	})

	c := newPostgresTestClient(t, ms)
	plan, err := c.ResolvePostgresProtectionPlan(context.Background(), "weekly-full-daily-incr")
	if err != nil {
		t.Fatalf("ResolvePostgresProtectionPlan() error = %v", err)
	}
	if plan.Value != "weekly-full-daily-incr" {
		t.Fatalf("plan = %+v", plan)
	}
	if _, err := c.ResolvePostgresProtectionPlan(context.Background(), "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestWaitForPostgresReady(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleCluster(models.PostgresStatusActive, testConnString))
	})

	got, err := newPostgresTestClient(t, ms).WaitForPostgresReady(context.Background(), testPostgresID, 15*time.Minute)
	if err != nil {
		t.Fatalf("WaitForPostgresReady() error = %v", err)
	}
	if got.Status != models.PostgresStatusActive {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestWaitForPostgresReadyFailed(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		cluster := sampleCluster(models.PostgresStatusFailed, "")
		cluster.Message = "provisioning failed"
		writeJSON(w, http.StatusOK, cluster)
	})

	_, err := newPostgresTestClient(t, ms).WaitForPostgresReady(context.Background(), testPostgresID, 15*time.Minute)
	if err == nil || !strings.Contains(err.Error(), "Failed") {
		t.Fatalf("expected failed-state error, got %v", err)
	}
}

func TestWaitForPostgresReadyTimeout(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleCluster(models.PostgresStatusCreating, ""))
	})

	_, err := newPostgresTestClient(t, ms).WaitForPostgresReady(context.Background(), testPostgresID, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestWaitForPostgresDeleted(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("GET", testPostgresCluster, 404, "Not found")

	if err := newPostgresTestClient(t, ms).WaitForPostgresDeleted(context.Background(), testPostgresID, 15*time.Minute); err != nil {
		t.Fatalf("WaitForPostgresDeleted() error = %v", err)
	}
}

func TestWaitForPostgresDeletedStatus(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testPostgresCluster, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleCluster(models.PostgresStatusDeleted, ""))
	})

	if err := newPostgresTestClient(t, ms).WaitForPostgresDeleted(context.Background(), testPostgresID, 15*time.Minute); err != nil {
		t.Fatalf("WaitForPostgresDeleted() error = %v", err)
	}
}
