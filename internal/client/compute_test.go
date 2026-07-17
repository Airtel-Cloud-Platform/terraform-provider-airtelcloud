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

func TestCreateCompute(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		request *models.CreateComputeRequest
		wantErr bool
	}{
		{
			name: "successful compute creation",
			request: &models.CreateComputeRequest{
				InstanceName: "test-instance",
				ImageID:      "ubuntu-20.04",
				FlavorID:     "t2.micro",
				NetworkID:    "network-1",
				AZName:       "south-1a",
				VolumeSize:   20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compute, err := client.CreateCompute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCompute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if compute == nil {
					t.Error("CreateCompute() returned nil compute")
					return
				}

				if compute.ID == "" {
					t.Error("CreateCompute() returned compute with empty ID")
				}

				if compute.InstanceName != tt.request.InstanceName {
					t.Errorf("CreateCompute() instance name = %v, want %v",
						compute.InstanceName, tt.request.InstanceName)
				}
			}
		})
	}
}

func TestGetCompute(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		id       string
		wantErr  bool
		wantName string
	}{
		{
			name:     "successful compute retrieval",
			id:       "test-id",
			wantErr:  false,
			wantName: "test-instance",
		},
		{
			name:    "compute not found",
			id:      "nonexistent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up error response for nonexistent compute
			if tt.id == "nonexistent-id" {
				mockServer.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/nonexistent-id/", 404, "Not found")
			}

			compute, err := client.GetCompute(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCompute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if compute == nil {
					t.Error("GetCompute() returned nil compute")
					return
				}

				if compute.InstanceName != tt.wantName {
					t.Errorf("GetCompute() instance name = %v, want %v",
						compute.InstanceName, tt.wantName)
				}
			}
		})
	}
}

func TestUpdateCompute(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		id      string
		request *models.UpdateComputeRequest
		wantErr bool
	}{
		{
			name: "successful compute update",
			id:   "test-id",
			request: &models.UpdateComputeRequest{
				InstanceName:    "updated-instance",
				Description:     "Updated description",
				SecurityGroupID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compute, err := client.UpdateCompute(context.Background(), tt.id, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCompute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if compute == nil {
					t.Error("UpdateCompute() returned nil compute")
					return
				}

				if compute.ID != tt.id {
					t.Errorf("UpdateCompute() ID = %v, want %v", compute.ID, tt.id)
				}
			}
		})
	}
}

func TestDeleteCompute(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "successful compute deletion",
			id:      "test-id",
			wantErr: false,
		},
		{
			name:    "delete nonexistent compute",
			id:      "nonexistent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up error response for nonexistent compute
			if tt.id == "nonexistent-id" {
				mockServer.SetErrorResponse("DELETE", "/api/v2.1/computes/domain/test-org/project/test-project/computes/nonexistent-id/", 404, "Not found")
			}

			err := client.DeleteCompute(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCompute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListComputes(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 2,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.Compute{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", 500, "Internal server error")
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

			computes, err := client.ListComputes(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListComputes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(computes) != tt.wantCount {
					t.Errorf("ListComputes() count = %d, want %d", len(computes), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if computes[0].ID != "compute-1" {
						t.Errorf("ListComputes() first ID = %v, want compute-1", computes[0].ID)
					}
					if computes[1].InstanceName != "instance-2" {
						t.Errorf("ListComputes() second Name = %v, want instance-2", computes[1].InstanceName)
					}
				}
			}
		})
	}
}

func TestResolveComputeID(t *testing.T) {
	tests := []struct {
		name    string
		lookup  string
		setup   func(ms *testutil.MockServer)
		wantID  string
		wantErr bool
	}{
		{
			name:   "resolves first compute by name",
			lookup: "instance-1",
			wantID: "compute-1",
		},
		{
			name:   "resolves second compute by name",
			lookup: "instance-2",
			wantID: "compute-2",
		},
		{
			name:    "name not found",
			lookup:  "does-not-exist",
			wantErr: true,
		},
		{
			name:   "found but empty ID",
			lookup: "instance-1",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.Compute{{ID: "", InstanceName: "instance-1"}})
				})
			},
			wantErr: true,
		},
		{
			name:   "server error",
			lookup: "instance-1",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", 500, "Internal server error")
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

			id, err := client.ResolveComputeID(context.Background(), tt.lookup)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveComputeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id != tt.wantID {
				t.Errorf("ResolveComputeID() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestListFlavors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 2,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/flavors/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.Flavor{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/flavors/", 500, "Internal server error")
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

			flavors, err := client.ListFlavors(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListFlavors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(flavors) != tt.wantCount {
					t.Errorf("ListFlavors() count = %d, want %d", len(flavors), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if flavors[0].Name != "t2.micro" {
						t.Errorf("ListFlavors() first Name = %v, want t2.micro", flavors[0].Name)
					}
					if flavors[1].ID != 2 {
						t.Errorf("ListFlavors() second ID = %v, want 2", flavors[1].ID)
					}
				}
			}
		})
	}
}

func TestListImages(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 2,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/images/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.Image{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/images/", 500, "Internal server error")
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

			images, err := client.ListImages(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(images) != tt.wantCount {
					t.Errorf("ListImages() count = %d, want %d", len(images), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if images[0].Name != "ubuntu-20.04" {
						t.Errorf("ListImages() first Name = %v, want ubuntu-20.04", images[0].Name)
					}
					if images[1].OSType != "linux" {
						t.Errorf("ListImages() second OSType = %v, want linux", images[1].OSType)
					}
				}
			}
		})
	}
}

func TestListKeypairs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 2,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/ext/api/v1/keypairs/domain/test-org/keypairs", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"message":"Keypair fetched successfully.","data":{"items":[],"count":0,"offset":0,"limit":30}}`))
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/ext/api/v1/keypairs/domain/test-org/keypairs", 500, "Internal server error")
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

			keypairs, err := client.ListKeypairs(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListKeypairs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(keypairs) != tt.wantCount {
					t.Errorf("ListKeypairs() count = %d, want %d", len(keypairs), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if keypairs[0].Name != "keypair-1" {
						t.Errorf("ListKeypairs() first Name = %v, want keypair-1", keypairs[0].Name)
					}
					// The API nests items under "data" and keys the ID as "uuid".
					// Asserting the ID here guards against both the envelope and
					// the JSON tag regressing.
					if keypairs[0].ID != "aae08127-a18d-4aa3-b332-a77066ab7c90" {
						t.Errorf("ListKeypairs() first ID = %v, want aae08127-a18d-4aa3-b332-a77066ab7c90", keypairs[0].ID)
					}
					if keypairs[1].ID != "b1f2c3d4-5678-4abc-9def-0123456789ab" {
						t.Errorf("ListKeypairs() second ID = %v, want b1f2c3d4-5678-4abc-9def-0123456789ab", keypairs[1].ID)
					}
				}
			}
		})
	}
}
