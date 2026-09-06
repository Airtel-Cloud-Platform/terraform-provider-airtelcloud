package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const blockStorageTestBasePath = "/api/storage-plugin/v1/domain/test-org/project/test-project/block-storage"

func TestCreateBlockStorageVolume(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", blockStorageTestBasePath+"/volume", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, `"name":"ak-store"`) {
			t.Fatalf("request body missing name: %s", body)
		}
		if !strings.Contains(body, `"availabilityZone":"S1"`) {
			t.Fatalf("request body missing availabilityZone: %s", body)
		}
		if !strings.Contains(body, `"size":10`) {
			t.Fatalf("request body missing numeric size: %s", body)
		}
		if !strings.Contains(body, `"description":"Create new baremetel storage"`) {
			t.Fatalf("request body missing description: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c := newBaremetalTestClient(t, ms)
	err := c.CreateBlockStorageVolume(context.Background(), &models.CreateBlockStorageVolumeRequest{
		Name:             "ak-store",
		AvailabilityZone: "S1",
		Size:             10,
		Description:      "Create new baremetel storage",
	})
	if err != nil {
		t.Fatalf("CreateBlockStorageVolume() error = %v", err)
	}
}

func TestGetBlockStorageVolume(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", blockStorageTestBasePath+"/volume/ak-store", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("availabilityZone") != "S1" {
			t.Fatalf("availabilityZone query = %q, want S1", r.URL.Query().Get("availabilityZone"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":             "ak-store",
			"description":      "Create new baremetel storage",
			"availabilityZone": "S1",
			"size":             "10",
			"state":            "Active",
			"uuid":             "vol-uuid-1",
		})
	})

	c := newBaremetalTestClient(t, ms)
	vol, err := c.GetBlockStorageVolume(context.Background(), "ak-store", "S1")
	if err != nil {
		t.Fatalf("GetBlockStorageVolume() error = %v", err)
	}
	if vol.Name != "ak-store" || int64(vol.Size) != 10 || vol.State != models.BlockStorageStateActive {
		t.Fatalf("unexpected volume: %+v", vol)
	}
}

func TestDeleteBlockStorageVolume(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	deleted := false
	ms.AddHandler("DELETE", blockStorageTestBasePath+"/volume/ak-store", func(w http.ResponseWriter, r *http.Request) {
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	})
	ms.AddHandler("GET", blockStorageTestBasePath+"/volume/ak-store", func(w http.ResponseWriter, r *http.Request) {
		if deleted {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"ak-store","state":"Active","size":"10"}`))
	})

	c := newBaremetalTestClient(t, ms)
	if err := c.DeleteBlockStorageVolume(context.Background(), "ak-store", "S1", ""); err != nil {
		t.Fatalf("DeleteBlockStorageVolume() error = %v", err)
	}
}
