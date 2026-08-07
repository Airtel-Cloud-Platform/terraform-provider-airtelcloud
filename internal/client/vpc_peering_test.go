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

func TestGetVPCPeering(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		id       string
		wantErr  bool
		wantName string
	}{
		{
			name:     "successful retrieval",
			id:       "test-peering-id",
			wantErr:  false,
			wantName: "test-peering",
		},
		{
			name:    "not found",
			id:      "nonexistent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id == "nonexistent-id" {
				mockServer.SetErrorResponse("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peering/nonexistent-id", 404, "Not found")
			}

			peering, err := client.GetVPCPeering(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVPCPeering() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if peering == nil {
					t.Error("GetVPCPeering() returned nil")
					return
				}
				if peering.Name != tt.wantName {
					t.Errorf("GetVPCPeering() Name = %v, want %v", peering.Name, tt.wantName)
				}
				if peering.ID != "test-peering-id" {
					t.Errorf("GetVPCPeering() ID = %v, want test-peering-id", peering.ID)
				}
				if peering.VPCSourceID != "vpc-source-1" {
					t.Errorf("GetVPCPeering() VPCSourceID = %v, want vpc-source-1", peering.VPCSourceID)
				}
				if peering.VPCTargetID != "vpc-target-1" {
					t.Errorf("GetVPCPeering() VPCTargetID = %v, want vpc-target-1", peering.VPCTargetID)
				}
				if peering.Region != "south-1" {
					t.Errorf("GetVPCPeering() Region = %v, want south-1", peering.Region)
				}
			}
		})
	}
}

func TestListVPCPeerings(t *testing.T) {
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
			name: "successful list with current API payload",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"count":1,"vpcPeerings":[{"name":"test-peering","description":"Test peering connection","vpcSourceId":"vpc-source-1","vpcTargetId":"vpc-target-1","peerVpcRegion":"south-1","state":"Active","vpcPeeringId":"test-peering-id","createdAt":"1785818438","az":["south-1a"],"region":"south-1"}]}`))
				})
			},
			wantCount: 1,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models.VPCPeeringListResponse{Count: 0, Items: []models.VPCPeering{}})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", 500, "Internal server error")
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

			response, err := client.ListVPCPeerings(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListVPCPeerings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if response.Count != tt.wantCount {
					t.Errorf("ListVPCPeerings() Count = %d, want %d", response.Count, tt.wantCount)
				}
				if len(response.Items) != tt.wantCount {
					t.Errorf("ListVPCPeerings() len(Items) = %d, want %d", len(response.Items), tt.wantCount)
				}
				if tt.wantCount == 1 {
					if response.Items[0].ID != "test-peering-id" {
						t.Errorf("ListVPCPeerings() first ID = %v, want test-peering-id", response.Items[0].ID)
					}
					if response.Items[0].Name != "test-peering" {
						t.Errorf("ListVPCPeerings() first Name = %v, want test-peering", response.Items[0].Name)
					}
					if response.Items[0].AZ != "south-1a" {
						t.Errorf("ListVPCPeerings() first AZ = %v, want south-1a", response.Items[0].AZ)
					}
				}
			}
		})
	}
}

func TestCreateVPCPeering(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.CreateVPCPeeringRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreateVPCPeeringRequest{
				Name:          "test-peering",
				VPCSourceID:   "vpc-source-1",
				VPCTargetID:   "vpc-target-1",
				AZ:            "south-1a",
				Region:        "south-1",
				PeerVpcRegion: "south-1",
			},
			wantErr: false,
		},
		{
			name: "server error on POST",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peering", 500, "Internal server error")
			},
			request: &models.CreateVPCPeeringRequest{
				Name:          "test-peering",
				VPCSourceID:   "vpc-source-1",
				VPCTargetID:   "vpc-target-1",
				AZ:            "south-1a",
				Region:        "south-1",
				PeerVpcRegion: "south-1",
			},
			wantErr: true,
		},
		{
			name: "successful creation with current API payload",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"count":1,"vpcPeerings":[{"name":"test-peering","description":"Test peering connection","vpcSourceId":"vpc-source-1","vpcTargetId":"vpc-target-1","peerVpcRegion":"south-1","state":"Active","vpcPeeringId":"test-peering-id","createdAt":"1785818438","az":["south-1a"],"region":"south-1"}]}`))
				})
			},
			request: &models.CreateVPCPeeringRequest{
				Name:          "test-peering",
				VPCSourceID:   "vpc-source-1",
				VPCTargetID:   "vpc-target-1",
				AZ:            "south-1a",
				Region:        "south-1",
				PeerVpcRegion: "south-1",
			},
			wantErr: false,
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

			peering, err := client.CreateVPCPeering(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVPCPeering() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if peering == nil {
					t.Error("CreateVPCPeering() returned nil")
					return
				}
				if peering.ID != "test-peering-id" {
					t.Errorf("CreateVPCPeering() ID = %v, want test-peering-id", peering.ID)
				}
				if peering.Name != "test-peering" {
					t.Errorf("CreateVPCPeering() Name = %v, want test-peering", peering.Name)
				}
			}
		})
	}
}

func TestDeleteVPCPeering(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(ms *testutil.MockServer)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			id:      "test-peering-id",
			wantErr: false,
		},
		{
			name: "already deleted (404 on DELETE is not an error)",
			id:   "test-peering-id",
			setup: func(ms *testutil.MockServer) {
				// DELETE returns 404 — client treats this as success
				ms.SetErrorResponse("DELETE", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peering/test-peering-id", 404, "Not found")
			},
			wantErr: false,
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

			err := client.DeleteVPCPeering(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteVPCPeering() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVPCPeeringStateClassification(t *testing.T) {
	t.Run("ready states", func(t *testing.T) {
		states := []string{"ACTIVE", "active", " Established ", "ready", "Succeeded"}
		for _, state := range states {
			if !isVPCPeeringReadyState(state) {
				t.Fatalf("expected state %q to be ready", state)
			}
		}
	})

	t.Run("failed states", func(t *testing.T) {
		states := []string{"FAILED", "error", " Rejected ", "terminated", "canceled"}
		for _, state := range states {
			if !isVPCPeeringFailedState(state) {
				t.Fatalf("expected state %q to be failed", state)
			}
		}
	})
}

func TestCreateVPCPeering_PendingStateHonorsContext(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	mockServer.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"vpcPeerings":[{"name":"test-peering","description":"Test peering connection","vpcSourceId":"vpc-source-1","vpcTargetId":"vpc-target-1","peerVpcRegion":"south-1","state":"Initializing","vpcPeeringId":"test-peering-id","createdAt":"1785818438","az":["south-1a"],"region":"south-1"}]}`))
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	createReq := &models.CreateVPCPeeringRequest{
		Name:          "test-peering",
		VPCSourceID:   "vpc-source-1",
		VPCTargetID:   "vpc-target-1",
		AZ:            "south-1a",
		Region:        "south-1",
		PeerVpcRegion: "south-1",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.CreateVPCPeeringWithTimeout(ctx, createReq, time.Minute)
	if err == nil {
		t.Fatal("expected error while peering remains in pending state")
	}
	if !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("expected context deadline exceeded, got: %v", err)
	}
}

func TestCreateVPCPeering_FailedStateReturnsError(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	mockServer.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/vpc-peerings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"vpcPeerings":[{"name":"test-peering","description":"Test peering connection","vpcSourceId":"vpc-source-1","vpcTargetId":"vpc-target-1","peerVpcRegion":"south-1","state":"Failed","vpcPeeringId":"test-peering-id","createdAt":"1785818438","az":["south-1a"],"region":"south-1"}]}`))
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	createReq := &models.CreateVPCPeeringRequest{
		Name:          "test-peering",
		VPCSourceID:   "vpc-source-1",
		VPCTargetID:   "vpc-target-1",
		AZ:            "south-1a",
		Region:        "south-1",
		PeerVpcRegion: "south-1",
	}

	_, err := client.CreateVPCPeeringWithTimeout(context.Background(), createReq, time.Minute)
	if err == nil {
		t.Fatal("expected failure-state error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "failure state") {
		t.Fatalf("expected failure-state error, got: %v", err)
	}
}
