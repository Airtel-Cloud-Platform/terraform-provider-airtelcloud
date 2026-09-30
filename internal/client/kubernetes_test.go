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
	testK8sName   = "test-kms"
	testK8sBase   = "/api/airtel/v1/domain/test-org/project/test-project/cluster"
	testK8sDelete = "/api/compass/v1/domain/test-org/project/test-project/cluster/" + testK8sName
	testK8sStatus = testK8sDelete + "/status"
)

func newKubernetesTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	c, err := NewClient(strings.TrimSuffix(ms.URL, "/"), "test-api-key", "test-api-secret", "south", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func k8sStatusHandler(state, lifecycleState, failedStateError string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.KubernetesClusterStatus{
			Cluster: models.KubernetesCluster{
				Key:              models.KubernetesClusterKey{Cluster: testK8sName},
				LifeCycleState:   lifecycleState,
				FailedStateError: failedStateError,
			},
			State: state,
		})
	}
}

func TestCreateKubernetesCluster(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", testK8sBase, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("ce-availability-zone") != "S1" {
			t.Errorf("ce-availability-zone = %q", r.Header.Get("ce-availability-zone"))
		}
		body, _ := io.ReadAll(r.Body)
		var req models.CreateKubernetesClusterRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("body: %v", err)
		}
		if req.Cluster != testK8sName {
			t.Errorf("cluster = %q", req.Cluster)
		}
		if req.CKPProvider.Provider != models.KubernetesProviderBYOH {
			t.Errorf("provider = %q", req.CKPProvider.Provider)
		}
		w.WriteHeader(http.StatusOK)
	})

	err := newKubernetesTestClient(t, ms).WithAvailabilityZone("S1").CreateKubernetesCluster(context.Background(), &models.CreateKubernetesClusterRequest{
		Cluster: testK8sName,
		K8sInfo: models.KubernetesKubeInfo{K8sVersion: "v1.33.7", MasterNodes: 3, WorkerNodes: 1},
		CKPProvider: models.KubernetesCKPProvider{
			Provider: models.KubernetesProviderBYOH,
			BYOH: &models.KubernetesBYOHProvider{
				ControlPlaneProvider: models.KubernetesControlPlaneKamaji,
				NodePools: map[string]models.KubernetesNodePool{
					"md0": {HostGroup: "ccd.xLarge", Count: 1, Subnet: "subnet-1", AZ: "S1", OSDistribution: "Ubuntu"},
				},
			},
		},
		RequiredSchedulingTags: map[string]string{"availabilityZone": "S1"},
	})
	if err != nil {
		t.Fatalf("CreateKubernetesCluster() error = %v", err)
	}
}

func TestGetKubernetesCluster(t *testing.T) {
	t.Run("connected", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		ms.AddHandler("GET", testK8sStatus, k8sStatusHandler(models.KubernetesStateConnected, models.KubernetesStateReady, ""))

		got, err := newKubernetesTestClient(t, ms).GetKubernetesCluster(context.Background(), testK8sName)
		if err != nil {
			t.Fatalf("GetKubernetesCluster() error = %v", err)
		}
		if got.ClusterName() != testK8sName || got.State != models.KubernetesStateConnected || got.LifeCycleState != models.KubernetesStateReady {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("status endpoint not found", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		ms.AddHandler("GET", testK8sStatus, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":5,"message":"Cluster not found"}`))
		})

		_, err := newKubernetesTestClient(t, ms).GetKubernetesCluster(context.Background(), testK8sName)
		if !IsNotFoundError(err) {
			t.Fatalf("error = %v, want not found", err)
		}
	})
}

func TestDeleteKubernetesCluster(t *testing.T) {
	t.Run("compass delete", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		ms.AddHandler("DELETE", testK8sDelete, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("forceDelete"); got != "false" {
				t.Errorf("forceDelete = %q, want false", got)
			}
			w.WriteHeader(http.StatusOK)
		})

		if err := newKubernetesTestClient(t, ms).DeleteKubernetesCluster(context.Background(), testK8sName); err != nil {
			t.Fatalf("DeleteKubernetesCluster() error = %v", err)
		}
	})

	t.Run("not found is success", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		ms.AddHandler("DELETE", testK8sDelete, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		})

		if err := newKubernetesTestClient(t, ms).DeleteKubernetesCluster(context.Background(), testK8sName); err != nil {
			t.Fatalf("DeleteKubernetesCluster() error = %v", err)
		}
	})
}

func TestWaitForKubernetesReady(t *testing.T) {
	t.Run("compass connected", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testK8sStatus, k8sStatusHandler(models.KubernetesStateConnected, models.KubernetesStateReady, ""))

		got, err := newKubernetesTestClient(t, ms).WaitForKubernetesReady(context.Background(), testK8sName, time.Minute)
		if err != nil {
			t.Fatalf("WaitForKubernetesReady() error = %v", err)
		}
		if got.State != models.KubernetesStateConnected {
			t.Fatalf("state = %q", got.State)
		}
	})

	t.Run("not registered times out", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testK8sStatus, k8sStatusHandler(models.KubernetesStateNotRegistered, models.KubernetesStateCreating, ""))

		_, err := newKubernetesTestClient(t, ms).WaitForKubernetesReady(context.Background(), testK8sName, time.Millisecond)
		if err == nil {
			t.Fatal("WaitForKubernetesReady() error = nil, want timeout")
		}
		if !strings.Contains(err.Error(), models.KubernetesStateNotRegistered) {
			t.Fatalf("error = %v, want state Not Registered", err)
		}
	})

	t.Run("failed lifecycle returns error", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testK8sStatus, k8sStatusHandler("Not Connected", "Failed", "control plane bootstrap failed"))

		_, err := newKubernetesTestClient(t, ms).WaitForKubernetesReady(context.Background(), testK8sName, time.Minute)
		if err == nil || !strings.Contains(err.Error(), "control plane bootstrap failed") {
			t.Fatalf("error = %v, want failedStateError", err)
		}
	})
}

func TestWaitForKubernetesDeleted(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", testK8sStatus, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":5,"message":"Cluster not found"}`))
	})

	if err := newKubernetesTestClient(t, ms).WaitForKubernetesDeleted(context.Background(), testK8sName, time.Minute); err != nil {
		t.Fatalf("WaitForKubernetesDeleted() error = %v", err)
	}
}

func TestResolveSubnetIDByName(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/network/vpc-1/subnets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.SubnetListResponse{
			Count: 1,
			Items: []models.Subnet{{SubnetID: "subnet-uuid-1", Name: "app-subnet"}},
		})
	})
	ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/network/vpc-2/subnets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.SubnetListResponse{Count: 0, Items: []models.Subnet{}})
	})

	c := newKubernetesTestClient(t, ms)
	got, err := c.ResolveSubnetIDByName(context.Background(), "app-subnet", "vpc-default")
	if err != nil {
		t.Fatalf("scoped resolve: %v", err)
	}
	if got != "subnet-uuid-1" {
		t.Fatalf("scoped id = %q", got)
	}

	got, err = c.ResolveSubnetIDByName(context.Background(), "app-subnet", "")
	if err != nil {
		t.Fatalf("unscoped resolve: %v", err)
	}
	if got != "subnet-uuid-1" {
		t.Fatalf("unscoped id = %q", got)
	}
}
