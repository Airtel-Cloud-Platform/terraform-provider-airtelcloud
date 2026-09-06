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

const objectStorageTestBasePath = "/api/storage-plugin/v1/domain/test-org/project/test-project/object-storage"

func TestCreateObjectStorageBucketRequestBody(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", objectStorageTestBasePath+"/bucket", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, `"bucket":"temporal-server-01"`) {
			t.Fatalf("missing bucket name: %s", body)
		}
		if !strings.Contains(body, `"versioning":true`) {
			t.Fatalf("missing versioning: %s", body)
		}
		if !strings.Contains(body, `"objLocking":true`) {
			t.Fatalf("missing objLocking: %s", body)
		}
		if !strings.Contains(body, `"objLockValidityDays":30`) {
			t.Fatalf("missing objLockValidityDays: %s", body)
		}
		if !strings.Contains(body, `"replicationType":"Local"`) {
			t.Fatalf("missing replicationType: %s", body)
		}
		if !strings.Contains(body, `"az":"S1"`) {
			t.Fatalf("missing az: %s", body)
		}
		if !strings.Contains(body, `"tag":"south_S1"`) {
			t.Fatalf("missing tag: %s", body)
		}
		if !strings.Contains(body, `"test":"test"`) {
			t.Fatalf("missing tags: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	ms.AddHandler("GET", objectStorageTestBasePath+"/bucket/temporal-server-01", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":  "temporal-server-01",
			"state": "Active",
		})
	})

	c := newBaremetalTestClient(t, ms)
	_, err := c.CreateObjectStorageBucket(context.Background(), &models.CreateObjectStorageBucketRequest{
		Bucket: "temporal-server-01",
		Tags:   map[string]string{"test": "test"},
		Config: &models.BucketCreateConfig{
			Versioning:          true,
			ObjLocking:          true,
			ObjLockValidityDays: 30,
			Replication: &models.BucketReplicationConfig{
				ReplicationType: "Local",
				AZ:              "S1",
				Tag:             "south_S1",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateObjectStorageBucket() error = %v", err)
	}
}

func TestDeleteObjectStorageBucketWaitsUntilGone(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	deleted := false
	ms.AddHandler("PUT", objectStorageTestBasePath+"/bucket/temporal-server-01", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	ms.AddHandler("DELETE", objectStorageTestBasePath+"/bucket/temporal-server-01", func(w http.ResponseWriter, r *http.Request) {
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	})
	ms.AddHandler("GET", objectStorageTestBasePath+"/bucket/temporal-server-01", func(w http.ResponseWriter, r *http.Request) {
		if deleted {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"name": "temporal-server-01", "state": "Active"})
	})

	c := newBaremetalTestClient(t, ms)
	if err := c.DeleteObjectStorageBucket(context.Background(), "temporal-server-01", "S1", true); err != nil {
		t.Fatalf("DeleteObjectStorageBucket() error = %v", err)
	}
}
