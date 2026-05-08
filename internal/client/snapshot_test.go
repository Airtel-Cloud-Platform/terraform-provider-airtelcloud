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

func TestCreateComputeSnapshot(t *testing.T) {
	tests := []struct {
		name      string
		computeID string
		setup     func(ms *testutil.MockServer)
		wantErr   bool
		wantUUID  string
	}{
		{
			name:      "successful snapshot creation",
			computeID: "test-id",
			wantErr:   false,
			wantUUID:  "snap-uuid-1234",
		},
		{
			name:      "server error",
			computeID: "bad-id",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/bad-id/snapshot/", 500, "Internal server error")
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

			snapshot, err := client.CreateComputeSnapshot(context.Background(), tt.computeID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateComputeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if snapshot == nil {
					t.Error("CreateComputeSnapshot() returned nil")
					return
				}
				if snapshot.UUID != tt.wantUUID {
					t.Errorf("CreateComputeSnapshot() UUID = %v, want %v", snapshot.UUID, tt.wantUUID)
				}
				if snapshot.Status != "active" {
					t.Errorf("CreateComputeSnapshot() Status = %v, want active", snapshot.Status)
				}
			}
		})
	}
}

func TestGetComputeSnapshot(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		uuid     string
		wantErr  bool
		wantName string
	}{
		{
			name:     "successful retrieval",
			uuid:     "snap-uuid-1234",
			wantErr:  false,
			wantName: "test-snapshot",
		},
		{
			name:    "not found",
			uuid:    "nonexistent-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := client.GetComputeSnapshot(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetComputeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if snapshot == nil {
					t.Error("GetComputeSnapshot() returned nil")
					return
				}
				if snapshot.SnapshotName != tt.wantName {
					t.Errorf("GetComputeSnapshot() SnapshotName = %v, want %v", snapshot.SnapshotName, tt.wantName)
				}
				if snapshot.UUID != "snap-uuid-1234" {
					t.Errorf("GetComputeSnapshot() UUID = %v, want snap-uuid-1234", snapshot.UUID)
				}
			}
		})
	}
}

func TestListComputeSnapshots(t *testing.T) {
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
				ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.ComputeSnapshot{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/snapshot/", 500, "Internal server error")
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

			snapshots, err := client.ListComputeSnapshots(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListComputeSnapshots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(snapshots) != tt.wantCount {
					t.Errorf("ListComputeSnapshots() count = %d, want %d", len(snapshots), tt.wantCount)
				}
			}
		})
	}
}

func TestDeleteComputeSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{
			name:    "successful deletion",
			uuid:    "snap-uuid-1234",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := testutil.NewMockServer()
			defer mockServer.Close()

			baseURL := strings.TrimSuffix(mockServer.URL, "/")
			client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

			err := client.DeleteComputeSnapshot(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComputeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
