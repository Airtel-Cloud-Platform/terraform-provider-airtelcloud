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
	testK8sName       = "test-kms"
	testK8sHostGroup  = "ccd.xLarge"
	testK8sBase       = "/api/airtel/v1/domain/test-org/project/test-project/cluster"
	testK8sPath       = testK8sBase + "/" + testK8sName
	testK8sHostGroups = "/api/airtel/v1/domain/test-org/project/test-project/hostgroup/" + testK8sHostGroup + "/clusters"
)

func newKubernetesTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	c, err := NewClient(strings.TrimSuffix(ms.URL, "/"), "test-api-key", "test-api-secret", "south", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func sampleK8sCluster(state string) models.KubernetesCluster {
	return models.KubernetesCluster{
		Name:       testK8sName,
		K8sVersion: "v1.33.7",
		State:      state,
		K8sInfo: &models.KubernetesKubeInfo{
			K8sName:     "CKP",
			K8sVersion:  "v1.33.7",
			CNIName:     "calico",
			CNIVersion:  "v3.30.6",
			MasterNodes: 3,
			WorkerNodes: 1,
		},
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
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", testK8sHostGroups, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.KubernetesClusterList{
			Count: 1,
			Items: []models.KubernetesCluster{sampleK8sCluster(models.KubernetesStateReady)},
		})
	})

	got, err := newKubernetesTestClient(t, ms).GetKubernetesCluster(context.Background(), testK8sName, []string{testK8sHostGroup})
	if err != nil {
		t.Fatalf("GetKubernetesCluster() error = %v", err)
	}
	if got.ClusterName() != testK8sName || got.State != models.KubernetesStateReady {
		t.Fatalf("got %+v", got)
	}
}

func TestDeleteKubernetesClusterNotFoundIsSuccess(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("DELETE", testK8sPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	if err := newKubernetesTestClient(t, ms).DeleteKubernetesCluster(context.Background(), testK8sName); err != nil {
		t.Fatalf("DeleteKubernetesCluster() error = %v", err)
	}
}

func TestWaitForKubernetesReady(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", testK8sHostGroups, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.KubernetesClusterList{
			Count: 1,
			Items: []models.KubernetesCluster{sampleK8sCluster(models.KubernetesStateReady)},
		})
	})

	got, err := newKubernetesTestClient(t, ms).WaitForKubernetesReady(context.Background(), testK8sName, []string{testK8sHostGroup}, time.Minute)
	if err != nil {
		t.Fatalf("WaitForKubernetesReady() error = %v", err)
	}
	if got.State != models.KubernetesStateReady {
		t.Fatalf("state = %q", got.State)
	}
}

func TestWaitForKubernetesDeleted(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", testK8sHostGroups, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.KubernetesClusterList{Count: 0, Items: []models.KubernetesCluster{}})
	})

	if err := newKubernetesTestClient(t, ms).WaitForKubernetesDeleted(context.Background(), testK8sName, []string{testK8sHostGroup}, time.Minute); err != nil {
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
