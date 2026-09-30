package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	testMSSQLID   = "617cbba2-f38a-470b-93a9-e9cdf36e13da"
	testMSSQLBase = "/api/v1/dbaas/domain/test-org/project/test-project/mssql"
	testMSSQLItem = testMSSQLBase + "/" + testMSSQLID
)

func TestResolveMSSQLEdition(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLBase+"/editions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []models.MSSQLEdition{{
			Edition:  "web",
			Versions: []string{"sql_server_2022"},
		}})
	})

	edition, version, err := newPostgresTestClient(t, ms).ResolveMSSQLEdition(context.Background(), "web")
	if err != nil {
		t.Fatalf("ResolveMSSQLEdition() error = %v", err)
	}
	if edition.Edition != "web" || version != "sql_server_2022" {
		t.Fatalf("edition=%q version=%q", edition.Edition, version)
	}
}

func TestResolveMSSQLCollation(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLBase+"/collations/web", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []models.MSSQLCollation{{
			Collation: "SQL_Latin1_General_CP1_CI_AS",
			Features: []models.MSSQLFeatureItem{{
				Value:    "SQLENGINE",
				Editions: []string{"web", "standard"},
			}, {
				Value:    "FULLTEXT",
				Editions: []string{"enterprise"},
			}},
		}})
	})

	c := newPostgresTestClient(t, ms)
	if err := c.ResolveMSSQLCollation(context.Background(), "web", "SQL_Latin1_General_CP1_CI_AS", []string{"SQLENGINE"}); err != nil {
		t.Fatalf("ResolveMSSQLCollation() error = %v", err)
	}
	if err := c.ResolveMSSQLCollation(context.Background(), "web", "SQL_Latin1_General_CP1_CI_AS", []string{"FULLTEXT"}); err == nil {
		t.Fatal("expected edition mismatch to fail")
	}
}

func TestResolveMSSQLFlavorRAM(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLBase+"/flavors/web", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q", r.Header.Get("ce-availability-zone"))
		}
		writeJSON(w, http.StatusOK, []models.MSSQLFlavor{{
			ID:    281,
			Name:  "db.mssqlweb.uhper.cci.xlarge",
			VCPUs: "4",
			RAM:   "8192",
		}})
	})

	flavor, ram, err := newPostgresTestClient(t, ms).ResolveMSSQLFlavor(context.Background(), "S1", "web", "db.mssqlweb.uhper.cci.xlarge")
	if err != nil {
		t.Fatalf("ResolveMSSQLFlavor() error = %v", err)
	}
	if flavor.ID != 281 || ram != 8 {
		t.Fatalf("id=%d ram=%d", flavor.ID, ram)
	}
}

func TestResolveMSSQLStorage(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLBase+"/volumetypes", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("group") != "BLOCK_STORAGE" {
			t.Errorf("group = %q", r.URL.Query().Get("group"))
		}
		writeJSON(w, http.StatusOK, []models.MSSQLVolumeType{{
			ID:           10,
			Name:         "s1_wkld_ntp02_5iops_backend",
			Label:        "High Performance",
			ProviderVTID: "4eb628fa-9139-43ed-af7c-e09e1e7d4fe5",
		}})
	})

	got, err := newPostgresTestClient(t, ms).ResolveMSSQLStorage(context.Background(), "S1", "High Performance")
	if err != nil {
		t.Fatalf("ResolveMSSQLStorage() error = %v", err)
	}
	if got.Name != "s1_wkld_ntp02_5iops_backend" || got.ID != 10 {
		t.Fatalf("storage = %+v", got)
	}
}

func TestResolveMSSQLProtectionPlan(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLBase+"/netbackup/slos", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("region") != "south" || q.Get("az") != "S1" || q.Get("sort") != "name" {
			t.Errorf("query = %s", q.Encode())
		}
		writeJSON(w, http.StatusOK, models.MSSQLSLOListResponse{
			Data: []models.MSSQLSLO{{
				ID: "4fa392d3-3b54-4572-97d1-a98b8750f314",
				Attributes: models.MSSQLSLOAttributes{
					Name:        "CH26_AZ1_PaaS_MSSQLIP_DAILY_FULL_2100_PP",
					Description: "daily",
				},
			}},
		})
	})

	got, err := newPostgresTestClient(t, ms).ResolveMSSQLProtectionPlan(context.Background(), "S1", "CH26_AZ1_PaaS_MSSQLIP_DAILY_FULL_2100_PP")
	if err != nil {
		t.Fatalf("ResolveMSSQLProtectionPlan() error = %v", err)
	}
	if got.ID != "4fa392d3-3b54-4572-97d1-a98b8750f314" {
		t.Fatalf("id = %q", got.ID)
	}
}

func TestCreateMSSQLBody(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("POST", testMSSQLBase, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got models.CreateMSSQLRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.User != "admin" || got.ConfirmPassword != got.Password {
			t.Fatalf("user/password = %+v", got)
		}
		if got.ComputeStorageConfig.Flavor.RAM != 8 || got.ComputeStorageConfig.TPS != "500" {
			t.Fatalf("compute = %+v", got.ComputeStorageConfig)
		}
		if got.ComputeStorageConfig.DatabaseSize != "200" || got.VolumeSize != "200" {
			t.Fatalf("size database=%s volume=%s", got.ComputeStorageConfig.DatabaseSize, got.VolumeSize)
		}
		if got.Backup.Enabled || got.Backup.NetbackupID != "" {
			t.Fatalf("backup = %+v", got.Backup)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatal(err)
		}
		if _, exists := raw["vpc"]; exists {
			t.Fatal("vpc must not be sent")
		}
		writeJSON(w, http.StatusOK, models.MSSQLInstance{UUID: testMSSQLID, Name: got.Name})
	})

	got, err := newPostgresTestClient(t, ms).CreateMSSQL(context.Background(), &models.CreateMSSQLRequest{
		Name:            "hello",
		User:            models.MSSQLFixedUser,
		Password:        "secret",
		ConfirmPassword: "secret",
		ComputeStorageConfig: models.MSSQLComputeStorageConfig{
			Flavor:       models.MSSQLFlavorRef{RAM: 8},
			TPS:          models.MSSQLFixedTPS,
			DatabaseSize: "200",
		},
		VolumeSize: "200",
		Backup:     models.MSSQLBackup{Enabled: false},
		AddLabels:  []string{},
	})
	if err != nil {
		t.Fatalf("CreateMSSQL() error = %v", err)
	}
	if got.UUID != testMSSQLID {
		t.Fatalf("uuid = %q", got.UUID)
	}
}

func TestWaitForMSSQLReadyAndFailed(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMSSQLItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.MSSQLInstance{UUID: testMSSQLID, Status: models.MSSQLStatusActive})
	})
	if _, err := newPostgresTestClient(t, ms).WaitForMSSQLReady(context.Background(), testMSSQLID, time.Second); err != nil {
		t.Fatalf("WaitForMSSQLReady() error = %v", err)
	}

	ms.AddHandler("GET", testMSSQLItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.MSSQLInstance{UUID: testMSSQLID, Status: models.MSSQLStatusFailed, Message: "boom"})
	})
	if _, err := newPostgresTestClient(t, ms).WaitForMSSQLReady(context.Background(), testMSSQLID, time.Second); err == nil {
		t.Fatal("expected failed status to error")
	}

	ms.AddHandler("GET", testMSSQLItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.MSSQLInstance{UUID: testMSSQLID, Status: models.MSSQLStatusResizing})
	})
	_, err := newPostgresTestClient(t, ms).WaitForMSSQLReady(context.Background(), testMSSQLID, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("Resizing should keep polling, got %v", err)
	}
}

func TestWaitForMSSQLDeleted(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.SetErrorResponse("GET", testMSSQLItem, 404, "Not found")
	if err := newPostgresTestClient(t, ms).WaitForMSSQLDeleted(context.Background(), testMSSQLID, time.Second); err != nil {
		t.Fatalf("WaitForMSSQLDeleted() error = %v", err)
	}
}
