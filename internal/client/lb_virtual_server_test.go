package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestCreateVirtualServer(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		wantErr bool
	}{
		{
			name:    "successful creation",
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers", 500, "Internal server error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := testutil.NewMockServer()
			defer mockServer.Close()

			if tt.setup != nil {
				tt.setup(mockServer)
			}

			baseURL := strings.TrimSuffix(mockServer.URL, "/")
			client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

			nodes := []models.VirtualServerNode{
				{
					ResourceID: "compute-uuid-1", InstanceName: "vm-1", ResourceIP: "10.0.0.1",
					BackendPortID: 101, SourceType: "vm", ResourceType: "compute", Port: 80,
				},
			}

			formData := BuildVirtualServerFormData(VirtualServerCreateParams{
				Name: "test-vs", Protocol: "HTTP", VPCID: "vpc-1", RoutingAlgorithm: "ROUND_ROBIN",
				VipPortID: 1, Port: 80, Interval: 30,
				XForwardedFor: true,
				Nodes:         nodes,
			})

			vs, err := client.CreateVirtualServer(context.Background(), "lb-svc-1", formData)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVirtualServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if vs == nil {
					t.Error("CreateVirtualServer() returned nil")
					return
				}
				if vs.ID != "vs-1" {
					t.Errorf("CreateVirtualServer() ID = %v, want vs-1", vs.ID)
				}
				if vs.Protocol != "HTTP" {
					t.Errorf("CreateVirtualServer() Protocol = %v, want HTTP", vs.Protocol)
				}
			}
		})
	}
}

func TestGetVirtualServer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		vsID    string
		wantErr bool
	}{
		{
			name:    "successful retrieval",
			vsID:    "vs-1",
			wantErr: false,
		},
		{
			name:    "not found",
			vsID:    "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs, err := client.GetVirtualServer(context.Background(), "lb-svc-1", tt.vsID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVirtualServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if vs == nil {
					t.Error("GetVirtualServer() returned nil")
					return
				}
				if vs.Name != "test-vs" {
					t.Errorf("GetVirtualServer() Name = %v, want test-vs", vs.Name)
				}
				if vs.RoutingAlgorithm != "ROUND_ROBIN" {
					t.Errorf("GetVirtualServer() RoutingAlgorithm = %v, want ROUND_ROBIN", vs.RoutingAlgorithm)
				}
			}
		})
	}
}

func TestListVirtualServers(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 1,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.LBVirtualServer{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb-svc-1/virtual-servers", 500, "Internal server error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := testutil.NewMockServer()
			defer mockServer.Close()

			if tt.setup != nil {
				tt.setup(mockServer)
			}

			baseURL := strings.TrimSuffix(mockServer.URL, "/")
			client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

			servers, err := client.ListVirtualServers(context.Background(), "lb-svc-1")

			if (err != nil) != tt.wantErr {
				t.Errorf("ListVirtualServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(servers) != tt.wantCount {
					t.Errorf("ListVirtualServers() count = %d, want %d", len(servers), tt.wantCount)
				}
			}
		})
	}
}

func TestUpdateVirtualServer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	formData := map[string]interface{}{
		"routing_algorithm": "LEAST_CONNECTIONS",
		"x_forwarded_for":   "true",
	}

	vs, err := client.UpdateVirtualServer(context.Background(), "lb-svc-1", "vs-1", formData)
	if err != nil {
		t.Fatalf("UpdateVirtualServer() error = %v", err)
	}

	if vs.RoutingAlgorithm != "LEAST_CONNECTIONS" {
		t.Errorf("UpdateVirtualServer() RoutingAlgorithm = %v, want LEAST_CONNECTIONS", vs.RoutingAlgorithm)
	}
}

func TestDeleteVirtualServer(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeleteVirtualServer(context.Background(), "lb-svc-1", "vs-1")
	if err != nil {
		t.Errorf("DeleteVirtualServer() error = %v", err)
	}
}

func TestBuildVirtualServerFormData(t *testing.T) {
	nodes := []models.VirtualServerNode{
		{
			ResourceID: "compute-uuid-1", InstanceName: "vm-1", ResourceIP: "10.0.0.1",
			BackendPortID: 101, SourceType: "vm", ResourceType: "compute", Port: 80, Weight: 50,
		},
		{
			ResourceID: "compute-uuid-2", InstanceName: "vm-2", ResourceIP: "10.0.0.2",
			BackendPortID: 102, SourceType: "vm", ResourceType: "compute", Port: 80, Weight: 50,
		},
	}

	formData := BuildVirtualServerFormData(VirtualServerCreateParams{
		Name: "test-vs", Protocol: "HTTP", VPCID: "vpc-1", RoutingAlgorithm: "ROUND_ROBIN",
		MonitorProtocol: "HTTP", VipPortID: 1, Port: 80, Interval: 30,
		PersistenceEnabled: true, XForwardedFor: true, PersistenceType: "source_ip",
		Nodes: nodes,
	})

	if formData["name"] != "test-vs" {
		t.Errorf("BuildVirtualServerFormData() name = %v, want test-vs", formData["name"])
	}
	if formData["protocol"] != "HTTP" {
		t.Errorf("BuildVirtualServerFormData() protocol = %v, want HTTP", formData["protocol"])
	}
	if formData["routing_algorithm"] != "ROUND_ROBIN" {
		t.Errorf("BuildVirtualServerFormData() routing_algorithm = %v, want ROUND_ROBIN", formData["routing_algorithm"])
	}
	if formData["persistence_enabled"] != "true" {
		t.Errorf("BuildVirtualServerFormData() persistence_enabled = %v, want true", formData["persistence_enabled"])
	}
	if formData["persistence_type"] != "source_ip" {
		t.Errorf("BuildVirtualServerFormData() persistence_type = %v, want source_ip", formData["persistence_type"])
	}

	// nodes must be repeated fields — a []string of per-node JSON objects.
	nodeJSONs, ok := formData["nodes"].([]string)
	if !ok {
		t.Fatalf("BuildVirtualServerFormData() nodes field is %T, want []string", formData["nodes"])
	}
	if len(nodeJSONs) != 2 {
		t.Fatalf("BuildVirtualServerFormData() nodes count = %d, want 2", len(nodeJSONs))
	}
	for i, raw := range nodeJSONs {
		var n models.VirtualServerNode
		if err := json.Unmarshal([]byte(raw), &n); err != nil {
			t.Fatalf("BuildVirtualServerFormData() node %d is not a JSON object: %v", i, err)
		}
		if n.ResourceID == "" || n.ResourceType == "" || n.ResourceIP == "" || n.BackendPortID == 0 {
			t.Errorf("BuildVirtualServerFormData() node %d missing required fields: %+v", i, n)
		}
	}
}
