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

func TestListVPCs(t *testing.T) {
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
				ms.AddHandler("GET", "/api/network-manager/v1/domain/test-org/project/test-project/networks", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models.VPCListResponse{Count: 0, Items: []models.VPC{}})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/network-manager/v1/domain/test-org/project/test-project/networks", 500, "Internal server error")
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

			response, err := client.ListVPCs(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if response.Count != tt.wantCount {
					t.Errorf("ListVPCs() Count = %d, want %d", response.Count, tt.wantCount)
				}
				if len(response.Items) != tt.wantCount {
					t.Errorf("ListVPCs() len(Items) = %d, want %d", len(response.Items), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if response.Items[0].ID != "vpc-1" {
						t.Errorf("ListVPCs() first ID = %v, want vpc-1", response.Items[0].ID)
					}
					if response.Items[0].Name != "vpc-default" {
						t.Errorf("ListVPCs() first Name = %v, want vpc-default", response.Items[0].Name)
					}
					if response.Items[1].CIDRBlock != "172.16.0.0/16" {
						t.Errorf("ListVPCs() second CIDRBlock = %v, want 172.16.0.0/16", response.Items[1].CIDRBlock)
					}
				}
			}
		})
	}
}
