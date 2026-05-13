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

func TestListSubnets(t *testing.T) {
	tests := []struct {
		name      string
		networkID string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			networkID: "test-network-id",
			wantCount: 2,
		},
		{
			name:      "empty list",
			networkID: "test-network-id",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/network/test-network-id/subnets", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models.SubnetListResponse{Count: 0, Items: []models.Subnet{}})
				})
			},
			wantCount: 0,
		},
		{
			name:      "server error",
			networkID: "test-network-id",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/network-manager/v1/domain/test-org/project/test-project/network/test-network-id/subnets", 500, "Internal server error")
			},
			wantErr: true,
		},
		{
			name:      "non-existent network",
			networkID: "non-existent-network",
			wantErr:   true,
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

			response, err := client.ListSubnets(context.Background(), tt.networkID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if response.Count != tt.wantCount {
					t.Errorf("ListSubnets() Count = %d, want %d", response.Count, tt.wantCount)
				}
				if len(response.Items) != tt.wantCount {
					t.Errorf("ListSubnets() len(Items) = %d, want %d", len(response.Items), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if response.Items[0].SubnetID != "subnet-1" {
						t.Errorf("ListSubnets() first SubnetID = %v, want subnet-1", response.Items[0].SubnetID)
					}
					if response.Items[0].Name != "subnet-a" {
						t.Errorf("ListSubnets() first Name = %v, want subnet-a", response.Items[0].Name)
					}
					if response.Items[1].IPv4AddressSpace != "10.0.2.0/24" {
						t.Errorf("ListSubnets() second IPv4AddressSpace = %v, want 10.0.2.0/24", response.Items[1].IPv4AddressSpace)
					}
				}
			}
		})
	}
}
