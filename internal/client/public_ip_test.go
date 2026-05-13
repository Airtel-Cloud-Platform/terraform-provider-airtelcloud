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

func TestCreatePublicIP(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.CreatePublicIPRequest
		az      string
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreatePublicIPRequest{
				ObjectName: "test-public-ip",
				VIP:        "10.1.99.172",
			},
			az:      "S1",
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v1/ipam/domain/test-org/project/test-project", 500, "Internal server error")
			},
			request: &models.CreatePublicIPRequest{
				ObjectName: "test-public-ip",
				VIP:        "10.1.99.172",
			},
			az:      "S1",
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

			ip, err := client.CreatePublicIP(context.Background(), tt.request, tt.az)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePublicIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if ip == nil {
					t.Error("CreatePublicIP() returned nil")
					return
				}
				if ip.UUID != "test-public-ip-uuid" {
					t.Errorf("CreatePublicIP() UUID = %v, want test-public-ip-uuid", ip.UUID)
				}
				if ip.PublicIP != "103.239.168.100" {
					t.Errorf("CreatePublicIP() PublicIP = %v, want 103.239.168.100", ip.PublicIP)
				}
			}
		})
	}
}

func TestGetPublicIP(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{
			name:    "successful retrieval",
			uuid:    "test-public-ip-uuid",
			wantErr: false,
		},
		{
			name:    "not found",
			uuid:    "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := client.GetPublicIP(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if ip == nil {
					t.Error("GetPublicIP() returned nil")
					return
				}
				if ip.ObjectName != "test-public-ip" {
					t.Errorf("GetPublicIP() ObjectName = %v, want test-public-ip", ip.ObjectName)
				}
				if ip.IP != "103.239.168.100" {
					t.Errorf("GetPublicIP() IP = %v, want 103.239.168.100", ip.IP)
				}
				if ip.Status != "Created" {
					t.Errorf("GetPublicIP() Status = %v, want Created", ip.Status)
				}
				if ip.AZName != "S1" {
					t.Errorf("GetPublicIP() AZName = %v, want S1", ip.AZName)
				}
			}
		})
	}
}

func TestListPublicIPs(t *testing.T) {
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
				ms.AddHandler("GET", "/api/v1/ipam/domain/test-org/project/test-project", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models.PublicIPListResponse{Items: []models.PublicIP{}, Count: 0})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v1/ipam/domain/test-org/project/test-project", 500, "Internal server error")
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

			resp, err := client.ListPublicIPs(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPublicIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(resp.Items) != tt.wantCount {
					t.Errorf("ListPublicIPs() count = %d, want %d", len(resp.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestDeletePublicIP(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeletePublicIP(context.Background(), "test-public-ip-uuid")
	if err != nil {
		t.Errorf("DeletePublicIP() error = %v", err)
	}
}

// --- Policy Rule Tests ---

func TestListIPAMServices(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	services, err := client.ListIPAMServices(context.Background(), "S1")
	if err != nil {
		t.Fatalf("ListIPAMServices() error = %v", err)
	}

	if len(services) != 4 {
		t.Errorf("ListIPAMServices() count = %d, want 4", len(services))
	}

	if services[0].Name != "HTTP" {
		t.Errorf("ListIPAMServices()[0].Name = %v, want HTTP", services[0].Name)
	}
}

func TestCreatePublicIPPolicyRule(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.CreatePublicIPPolicyRuleRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreatePublicIPPolicyRuleRequest{
				DisplayName: "test-rule",
				Source:      "any",
				ServiceList: []string{"uuid-http", "uuid-https"},
				Action:      "accept",
				TargetVIP:   "10.1.99.172",
				PublicIP:    "103.239.168.100",
				UUID:        "test-public-ip-uuid",
			},
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v1/admin/ipam_vip/nat_rule", 500, "Internal server error")
			},
			request: &models.CreatePublicIPPolicyRuleRequest{
				DisplayName: "test-rule",
				Source:      "any",
				ServiceList: []string{"uuid-http"},
				Action:      "accept",
				TargetVIP:   "10.1.99.172",
				PublicIP:    "103.239.168.100",
				UUID:        "test-public-ip-uuid",
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

			err := client.CreatePublicIPPolicyRule(context.Background(), tt.request, "S1")

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePublicIPPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListPublicIPPolicyRules(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	resp, err := client.ListPublicIPPolicyRules(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100")
	if err != nil {
		t.Fatalf("ListPublicIPPolicyRules() error = %v", err)
	}

	if len(resp.Items) != 1 {
		t.Errorf("ListPublicIPPolicyRules() count = %d, want 1", len(resp.Items))
	}

	if resp.Items[0].DisplayName != "test-rule" {
		t.Errorf("ListPublicIPPolicyRules()[0].DisplayName = %v, want test-rule", resp.Items[0].DisplayName)
	}

	if resp.Items[0].Action != "accept" {
		t.Errorf("ListPublicIPPolicyRules()[0].Action = %v, want accept", resp.Items[0].Action)
	}
}

func TestGetPublicIPPolicyRule(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		ruleUUID string
		wantErr  bool
	}{
		{
			name:     "successful retrieval",
			ruleUUID: "test-public-ip-uuid-1",
			wantErr:  false,
		},
		{
			name:     "not found",
			ruleUUID: "nonexistent-rule",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := client.GetPublicIPPolicyRule(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100", tt.ruleUUID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicIPPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rule.DisplayName != "test-rule" {
					t.Errorf("GetPublicIPPolicyRule() DisplayName = %v, want test-rule", rule.DisplayName)
				}
				if len(rule.Services) != 2 {
					t.Errorf("GetPublicIPPolicyRule() Services count = %d, want 2", len(rule.Services))
				}
			}
		})
	}
}

func TestDeletePublicIPPolicyRule(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeletePublicIPPolicyRule(context.Background(), "test-public-ip-uuid-1")
	if err != nil {
		t.Errorf("DeletePublicIPPolicyRule() error = %v", err)
	}
}

// --- FindPortIDByVIP and MapPublicIP Tests ---

func TestFindPortIDByVIP(t *testing.T) {
	tests := []struct {
		name       string
		vip        string
		wantPortID int
		wantErr    bool
	}{
		{
			name:       "found matching port",
			vip:        "10.1.99.172",
			wantPortID: 101,
			wantErr:    false,
		},
		{
			name:    "no matching port",
			vip:     "192.168.1.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := testutil.NewMockServer()
			defer mockServer.Close()

			baseURL := strings.TrimSuffix(mockServer.URL, "/")
			client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

			portID, err := client.FindPortIDByVIP(context.Background(), tt.vip)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindPortIDByVIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && portID != tt.wantPortID {
				t.Errorf("FindPortIDByVIP() portID = %d, want %d", portID, tt.wantPortID)
			}
		})
	}
}

func TestMapPublicIP(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.MapPublicIPRequest
		az      string
		wantErr bool
	}{
		{
			name: "successful mapping",
			request: &models.MapPublicIPRequest{
				TargetVIP: "10.1.99.172",
				PublicIP:  "103.239.168.100",
				UUID:      "test-public-ip-uuid",
				PortID:    101,
			},
			az:      "S1",
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v1/admin/ipam_vip/vip_object", 500, "Internal server error")
			},
			request: &models.MapPublicIPRequest{
				TargetVIP: "10.1.99.172",
				PublicIP:  "103.239.168.100",
				UUID:      "test-public-ip-uuid",
				PortID:    101,
			},
			az:      "S1",
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

			err := client.MapPublicIP(context.Background(), tt.request, tt.az)

			if (err != nil) != tt.wantErr {
				t.Errorf("MapPublicIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
