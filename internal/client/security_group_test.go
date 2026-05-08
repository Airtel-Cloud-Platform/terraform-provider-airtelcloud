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

func TestCreateSecurityGroup(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		request *models.CreateSecurityGroupRequest
		wantErr bool
	}{
		{
			name: "successful security group creation",
			request: &models.CreateSecurityGroupRequest{
				SecurityGroupName: "test-sg",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg, err := client.CreateSecurityGroup(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if sg == nil {
					t.Error("CreateSecurityGroup() returned nil")
					return
				}
				if sg.ID == 0 {
					t.Error("CreateSecurityGroup() returned zero ID")
				}
				if sg.SecurityGroupName != "test-sg" {
					t.Errorf("CreateSecurityGroup() name = %v, want test-sg", sg.SecurityGroupName)
				}
				if sg.Status != "ACTIVE" {
					t.Errorf("CreateSecurityGroup() status = %v, want ACTIVE", sg.Status)
				}
			}
		})
	}
}

func TestCreateSecurityGroupWithAvailabilityZone(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	var receivedAZ string
	mockServer.AddHandler("POST", "/api/v1/networks/securitygroup/", func(w http.ResponseWriter, r *http.Request) {
		receivedAZ = r.Header.Get("ce-availability-zone")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		sg := models.SecurityGroupDetail{
			ID:                1,
			UUID:              "sg-uuid-az",
			SecurityGroupName: "test-sg-az",
			Status:            "ACTIVE",
			AZName:            "south-1b",
			AZRegion:          "south-1",
		}
		json.NewEncoder(w).Encode(sg)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")
	azClient := c.WithAvailabilityZone("south-1b")

	sg, err := azClient.CreateSecurityGroup(context.Background(), &models.CreateSecurityGroupRequest{
		SecurityGroupName: "test-sg-az",
	})
	if err != nil {
		t.Fatalf("CreateSecurityGroup() unexpected error: %v", err)
	}

	if receivedAZ != "south-1b" {
		t.Errorf("expected ce-availability-zone header = %q, got %q", "south-1b", receivedAZ)
	}
	if sg.AZName != "south-1b" {
		t.Errorf("CreateSecurityGroup() AZName = %v, want south-1b", sg.AZName)
	}
}

func TestGetSecurityGroup(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		id       int
		wantErr  bool
		wantName string
	}{
		{
			name:     "successful security group retrieval",
			id:       1,
			wantErr:  false,
			wantName: "test-sg",
		},
		{
			name:    "security group not found",
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id == 999 {
				mockServer.SetErrorResponse("GET", "/api/v1/networks/securitygroup/999/", 404, "Not found")
			}

			sg, err := client.GetSecurityGroup(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if sg == nil {
					t.Error("GetSecurityGroup() returned nil")
					return
				}
				if sg.SecurityGroupName != tt.wantName {
					t.Errorf("GetSecurityGroup() name = %v, want %v", sg.SecurityGroupName, tt.wantName)
				}
				if sg.ID != tt.id {
					t.Errorf("GetSecurityGroup() ID = %v, want %v", sg.ID, tt.id)
				}
				if len(sg.Rules) == 0 {
					t.Error("GetSecurityGroup() returned no rules")
				}
			}
		})
	}
}

func TestDeleteSecurityGroup(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "successful security group deletion",
			id:      1,
			wantErr: false,
		},
		{
			name:    "delete nonexistent security group",
			id:      999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id == 999 {
				mockServer.SetErrorResponse("DELETE", "/api/v1/networks/securitygroup/999/", 404, "Not found")
			}

			err := client.DeleteSecurityGroup(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListSecurityGroupsDetailed(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list all",
			wantCount: 2,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v1/networks/securitygroup/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.SecurityGroupDetail{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v1/networks/securitygroup/", 500, "Internal server error")
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

			sgs, err := client.ListSecurityGroupsDetailed(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSecurityGroupsDetailed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(sgs) != tt.wantCount {
					t.Errorf("ListSecurityGroupsDetailed() count = %d, want %d", len(sgs), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if sgs[0].SecurityGroupName != "test-sg" {
						t.Errorf("ListSecurityGroupsDetailed() first name = %v, want test-sg", sgs[0].SecurityGroupName)
					}
					if sgs[1].ID != 2 {
						t.Errorf("ListSecurityGroupsDetailed() second ID = %v, want 2", sgs[1].ID)
					}
				}
			}
		})
	}
}

func TestCreateSecurityGroupRule(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		sgID    int
		request *models.CreateSecurityGroupRuleRequest
		wantErr bool
	}{
		{
			name: "successful rule creation",
			sgID: 1,
			request: &models.CreateSecurityGroupRuleRequest{
				Direction:      "ingress",
				Protocol:       "tcp",
				PortRangeMin:   "22",
				PortRangeMax:   "22",
				RemoteIPPrefix: "0.0.0.0/0",
				Ethertype:      "IPv4",
				Description:    "Allow SSH",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := client.CreateSecurityGroupRule(context.Background(), tt.sgID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSecurityGroupRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rule == nil {
					t.Error("CreateSecurityGroupRule() returned nil")
					return
				}
				if rule.ID == 0 {
					t.Error("CreateSecurityGroupRule() returned zero ID")
				}
				if rule.Direction != "ingress" {
					t.Errorf("CreateSecurityGroupRule() direction = %v, want ingress", rule.Direction)
				}
				if rule.Protocol != "tcp" {
					t.Errorf("CreateSecurityGroupRule() protocol = %v, want tcp", rule.Protocol)
				}
			}
		})
	}
}

func TestGetSecurityGroupRule(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		sgID    int
		ruleID  int
		wantErr bool
	}{
		{
			name:    "successful rule retrieval",
			sgID:    1,
			ruleID:  10,
			wantErr: false,
		},
		{
			name:    "rule not found",
			sgID:    1,
			ruleID:  999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ruleID == 999 {
				mockServer.SetErrorResponse("GET", "/api/v1/networks/securitygroup/1/securitygrouprule/999/", 404, "Not found")
			}

			rule, err := client.GetSecurityGroupRule(context.Background(), tt.sgID, tt.ruleID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecurityGroupRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if rule == nil {
					t.Error("GetSecurityGroupRule() returned nil")
					return
				}
				if rule.ID != tt.ruleID {
					t.Errorf("GetSecurityGroupRule() ID = %v, want %v", rule.ID, tt.ruleID)
				}
				if rule.Direction != "ingress" {
					t.Errorf("GetSecurityGroupRule() direction = %v, want ingress", rule.Direction)
				}
			}
		})
	}
}

func TestDeleteSecurityGroupRule(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		sgID    int
		ruleID  int
		wantErr bool
	}{
		{
			name:    "successful rule deletion",
			sgID:    1,
			ruleID:  10,
			wantErr: false,
		},
		{
			name:    "delete nonexistent rule",
			sgID:    1,
			ruleID:  999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ruleID == 999 {
				mockServer.SetErrorResponse("DELETE", "/api/v1/networks/securitygroup/1/securitygrouprule/999/", 404, "Not found")
			}

			err := client.DeleteSecurityGroupRule(context.Background(), tt.sgID, tt.ruleID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSecurityGroupRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListSecurityGroupRules(t *testing.T) {
	tests := []struct {
		name      string
		sgID      int
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list all rules",
			sgID:      1,
			wantCount: 2,
		},
		{
			name: "empty list",
			sgID: 1,
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v1/networks/securitygroup/1/securitygrouprule/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.SecurityGroupRuleDetail{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			sgID: 1,
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v1/networks/securitygroup/1/securitygrouprule/", 500, "Internal server error")
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

			rules, err := client.ListSecurityGroupRules(context.Background(), tt.sgID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSecurityGroupRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(rules) != tt.wantCount {
					t.Errorf("ListSecurityGroupRules() count = %d, want %d", len(rules), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if rules[0].Direction != "ingress" {
						t.Errorf("ListSecurityGroupRules() first direction = %v, want ingress", rules[0].Direction)
					}
					if rules[1].Direction != "egress" {
						t.Errorf("ListSecurityGroupRules() second direction = %v, want egress", rules[1].Direction)
					}
				}
			}
		})
	}
}
