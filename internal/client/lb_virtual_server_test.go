package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
				{ComputeID: 1, ComputeIP: "10.0.0.1", Port: 80},
			}

			params := BuildVirtualServerParams(
				"test-vs", "HTTP", "vpc-1", "ROUND_ROBIN", "", "",
				1, 80, 30,
				false, true, false,
				"",
				nodes,
			)

			vs, err := client.CreateVirtualServer(context.Background(), "lb-svc-1", params)

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

	params := url.Values{}
	params.Set("routing_algorithm", "LEAST_CONNECTIONS")
	params.Set("x_forwarded_for", "true")

	vs, err := client.UpdateVirtualServer(context.Background(), "lb-svc-1", "vs-1", params)
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

func TestBuildVirtualServerParams(t *testing.T) {
	nodes := []models.VirtualServerNode{
		{ComputeID: 1, ComputeIP: "10.0.0.1", Port: 80, Weight: 50},
		{ComputeID: 2, ComputeIP: "10.0.0.2", Port: 80, Weight: 50},
	}

	params := BuildVirtualServerParams(
		"test-vs", "HTTP", "vpc-1", "ROUND_ROBIN", "HTTP", "",
		1, 80, 30,
		true, true, false,
		"source_ip",
		nodes,
	)

	if params.Get("name") != "test-vs" {
		t.Errorf("BuildVirtualServerParams() name = %v, want test-vs", params.Get("name"))
	}
	if params.Get("protocol") != "HTTP" {
		t.Errorf("BuildVirtualServerParams() protocol = %v, want HTTP", params.Get("protocol"))
	}
	if params.Get("routing_algorithm") != "ROUND_ROBIN" {
		t.Errorf("BuildVirtualServerParams() routing_algorithm = %v, want ROUND_ROBIN", params.Get("routing_algorithm"))
	}
	if params.Get("persistence_enabled") != "true" {
		t.Errorf("BuildVirtualServerParams() persistence_enabled = %v, want true", params.Get("persistence_enabled"))
	}
	if params.Get("persistence_type") != "source_ip" {
		t.Errorf("BuildVirtualServerParams() persistence_type = %v, want source_ip", params.Get("persistence_type"))
	}

	nodeValues := params["nodes"]
	if len(nodeValues) != 2 {
		t.Errorf("BuildVirtualServerParams() nodes count = %d, want 2", len(nodeValues))
	}
}
