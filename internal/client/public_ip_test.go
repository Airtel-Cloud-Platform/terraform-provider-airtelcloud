package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const testPublicIPBasePath = "/ext/api/v1/domain/test-org/project/test-project/public-ip"

func newTestClientForPublicIP(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	c, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c
}

func TestListPublicIPs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(models.PublicIPListResponse{
						Items: []models.PublicIP{{
							UUID:       "test-public-ip-uuid",
							ObjectName: "test-public-ip",
							PublicIP:   "103.239.168.100",
							Status:     "Created",
						}},
						Count: 1,
					})
				})
			},
			wantCount: 1,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(models.PublicIPListResponse{Items: []models.PublicIP{}, Count: 0})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", testPublicIPBasePath, 500, "Internal server error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := testutil.NewMockServer()
			defer ms.Close()
			if tt.setup != nil {
				tt.setup(ms)
			}

			client := newTestClientForPublicIP(t, ms)
			resp, err := client.ListPublicIPs(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListPublicIPs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(resp.Items) != tt.wantCount {
				t.Fatalf("ListPublicIPs() count = %d, want %d", len(resp.Items), tt.wantCount)
			}
		})
	}
}

func TestDeletePublicIP(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("DELETE", testPublicIPBasePath+"/test-public-ip-uuid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClientForPublicIP(t, ms)
	if err := client.DeletePublicIP(context.Background(), "test-public-ip-uuid"); err != nil {
		t.Fatalf("DeletePublicIP() error = %v", err)
	}
}

func TestResolvePublicIPID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(ms *testutil.MockServer)
		lookup   string
		wantUUID string
		wantErr  bool
	}{
		{
			name: "resolves by object_name",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(models.PublicIPListResponse{
						Items: []models.PublicIP{{UUID: "test-public-ip-uuid", ObjectName: "test-public-ip"}},
						Count: 1,
					})
				})
			},
			lookup:   "test-public-ip",
			wantUUID: "test-public-ip-uuid",
		},
		{
			name: "name not found",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(models.PublicIPListResponse{Items: []models.PublicIP{}, Count: 0})
				})
			},
			lookup:  "missing-public-ip",
			wantErr: true,
		},
		{
			name: "empty uuid in response",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(models.PublicIPListResponse{
						Items: []models.PublicIP{{UUID: "", ObjectName: "test-public-ip"}},
						Count: 1,
					})
				})
			},
			lookup:  "test-public-ip",
			wantErr: true,
		},
		{
			name: "list error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", testPublicIPBasePath, 500, "Internal server error")
			},
			lookup:  "test-public-ip",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := testutil.NewMockServer()
			defer ms.Close()
			if tt.setup != nil {
				tt.setup(ms)
			}

			client := newTestClientForPublicIP(t, ms)
			uuid, err := client.ResolvePublicIPID(context.Background(), tt.lookup)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolvePublicIPID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && uuid != tt.wantUUID {
				t.Fatalf("ResolvePublicIPID() uuid = %q, want %q", uuid, tt.wantUUID)
			}
		})
	}
}

func TestListIPAMServices(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)
	services, err := client.ListIPAMServices(context.Background(), "S1")
	if err != nil {
		t.Fatalf("ListIPAMServices() error = %v", err)
	}
	if len(services) != 4 {
		t.Fatalf("ListIPAMServices() count = %d, want 4", len(services))
	}
	if services[0].Name != "HTTP" {
		t.Fatalf("ListIPAMServices()[0].Name = %q, want %q", services[0].Name, "HTTP")
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
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("POST", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"message": "Policy Initiated successfully!!",
						"data":    map[string]any{"uuid": "test-public-ip-uuid-1"},
					})
				})
			},
			request: &models.CreatePublicIPPolicyRuleRequest{
				DisplayName: "test-rule",
				Source:      "any",
				ServiceList: []string{"uuid-http", "uuid-https"},
				Action:      "accept",
				TargetVIP:   "10.1.99.172",
				PublicIP:    "103.239.168.100",
				UUID:        "test-public-ip-uuid",
			},
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", 500, "Internal server error")
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
			ms := testutil.NewMockServer()
			defer ms.Close()
			if tt.setup != nil {
				tt.setup(ms)
			}

			client := newTestClientForPublicIP(t, ms)
			_, err := client.CreatePublicIPPolicyRule(context.Background(), tt.request, "S1")
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreatePublicIPPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreatePublicIPPolicyRule_SourceOfTruthPayload(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var payload map[string]any
	ms.AddHandler("POST", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Policy Initiated successfully!!",
			"data":    map[string]any{"uuid": "test-public-ip-uuid-1"},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	createdID, err := client.CreatePublicIPPolicyRule(context.Background(), &models.CreatePublicIPPolicyRuleRequest{
		DisplayName: "test-rule",
		Source:      "any",
		ServiceList: []string{"uuid-http"},
		Action:      "accept",
		TargetVIP:   "10.1.99.172",
		PublicIP:    "103.239.168.100",
		UUID:        "test-public-ip-uuid",
	}, "S1")
	if err != nil {
		t.Fatalf("CreatePublicIPPolicyRule() error = %v", err)
	}
	if createdID != "test-public-ip-uuid-1" {
		t.Fatalf("CreatePublicIPPolicyRule() id = %q, want %q", createdID, "test-public-ip-uuid-1")
	}

	if got, ok := payload["resource_type"].(string); !ok || got != "ipam" {
		t.Fatalf("payload resource_type = %v, want ipam", payload["resource_type"])
	}
	if got, ok := payload["rule_name"].(string); !ok || got != "test-rule" {
		t.Fatalf("payload rule_name = %v, want test-rule", payload["rule_name"])
	}
	sourceRaw, ok := payload["source"].([]any)
	if !ok || len(sourceRaw) != 1 {
		t.Fatalf("payload source = %v, want one source object", payload["source"])
	}
	sourceObj, ok := sourceRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("payload source[0] = %T, want object", sourceRaw[0])
	}
	if got, ok := sourceObj["create_new"].(bool); !ok || !got {
		t.Fatalf("payload source[0].create_new = %v, want true", sourceObj["create_new"])
	}
	if got, ok := sourceObj["source_type"].(string); !ok || got != "all" {
		t.Fatalf("payload source[0].source_type = %v, want all", sourceObj["source_type"])
	}
	if got, ok := payload["action"].(string); !ok || got != "accept" {
		t.Fatalf("payload action = %v, want accept", payload["action"])
	}
}

func TestCreatePublicIPPolicyRule_SourceOfTruthPayload_WithDetailedConfig(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var payload map[string]any
	ms.AddHandler("POST", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Policy Initiated successfully!!",
			"data":    map[string]any{"uuid": "test-public-ip-uuid-1"},
		})
	})

	createNewTrue := true
	createNewFalse := false
	isDefaultFalse := false

	client := newTestClientForPublicIP(t, ms)
	_, err := client.CreatePublicIPPolicyRule(context.Background(), &models.CreatePublicIPPolicyRuleRequest{
		DisplayName: "hello-rule",
		SourceConfig: []models.PublicIPPolicyRuleSourceInput{{
			CreateNew:  &createNewTrue,
			IPCIDR:     "1.2.1.0/24",
			SourceType: "ip_cidr",
		}},
		ServiceConfig: []models.PublicIPPolicyRuleServiceInput{{
			CreateNew: &createNewFalse,
			Name:      "RDP",
			IsDefault: &isDefaultFalse,
		}},
		Action:       "accept",
		ResourceType: "ipam",
		RevisionNote: "creating Policy",
		TargetVIP:    "10.1.99.172",
		PublicIP:     "103.239.168.100",
		UUID:         "test-public-ip-uuid",
	}, "S1")
	if err != nil {
		t.Fatalf("CreatePublicIPPolicyRule() error = %v", err)
	}

	if got, ok := payload["resource_type"].(string); !ok || got != "ipam" {
		t.Fatalf("payload resource_type = %v, want ipam", payload["resource_type"])
	}
	if got, ok := payload["revision_note"].(string); !ok || got != "creating Policy" {
		t.Fatalf("payload revision_note = %v, want creating Policy", payload["revision_note"])
	}

	sourceRaw, ok := payload["source"].([]any)
	if !ok || len(sourceRaw) != 1 {
		t.Fatalf("payload source = %v, want one source object", payload["source"])
	}
	sourceObj, ok := sourceRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("payload source[0] = %T, want object", sourceRaw[0])
	}
	if got, ok := sourceObj["create_new"].(bool); !ok || !got {
		t.Fatalf("payload source[0].create_new = %v, want true", sourceObj["create_new"])
	}
	if got, ok := sourceObj["source_type"].(string); !ok || got != "ip_cidr" {
		t.Fatalf("payload source[0].source_type = %v, want ip_cidr", sourceObj["source_type"])
	}
	if got, ok := sourceObj["ip_cidr"].(string); !ok || got != "1.2.1.0/24" {
		t.Fatalf("payload source[0].ip_cidr = %v, want 1.2.1.0/24", sourceObj["ip_cidr"])
	}

	servicesRaw, ok := payload["services"].([]any)
	if !ok || len(servicesRaw) != 1 {
		t.Fatalf("payload services = %v, want one service object", payload["services"])
	}
	serviceObj, ok := servicesRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("payload services[0] = %T, want object", servicesRaw[0])
	}
	if got, ok := serviceObj["create_new"].(bool); !ok || got {
		t.Fatalf("payload services[0].create_new = %v, want false", serviceObj["create_new"])
	}
	if got, ok := serviceObj["name"].(string); !ok || got != "RDP" {
		t.Fatalf("payload services[0].name = %v, want RDP", serviceObj["name"])
	}
	if got, ok := serviceObj["is_default"].(bool); !ok || got {
		t.Fatalf("payload services[0].is_default = %v, want false", serviceObj["is_default"])
	}
}

func TestCreatePublicIPPolicyRule_ReturnsAllocationInProgressError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var calls int
	ms.AddHandler("POST", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Please wait, Public IP allocation is in progress",
			"code":    0,
		})
	})

	client := newTestClientForPublicIP(t, ms)
	_, err := client.CreatePublicIPPolicyRule(context.Background(), &models.CreatePublicIPPolicyRuleRequest{
		DisplayName: "test-rule",
		Source:      "any",
		ServiceList: []string{"uuid-http"},
		Action:      "accept",
		TargetVIP:   "10.1.99.172",
		PublicIP:    "103.239.168.100",
		UUID:        "test-public-ip-uuid",
	}, "S1")
	if err == nil {
		t.Fatal("CreatePublicIPPolicyRule() expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("CreatePublicIPPolicyRule() calls = %d, want 1", calls)
	}
	if !strings.Contains(err.Error(), "Public IP allocation is in progress") {
		t.Fatalf("CreatePublicIPPolicyRule() error = %v, want allocation-in-progress error", err)
	}
}

func TestIsPublicIPPolicyRuleRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "matching allocation message",
			err:  &APIError{StatusCode: 400, Message: "Please wait, Public IP allocation is in progress", Code: 0},
			want: true,
		},
		{
			name: "different 400",
			err:  &APIError{StatusCode: 400, Message: "invalid request", Code: 0},
			want: false,
		},
		{
			name: "server error",
			err:  &APIError{StatusCode: 500, Message: "backend error", Code: 0},
			want: false,
		},
		{
			name: "plain error",
			err:  fmt.Errorf("transport error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPublicIPPolicyRuleRetryableError(tt.err); got != tt.want {
				t.Fatalf("isPublicIPPolicyRuleRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListPublicIPPolicyRules(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)
	resp, err := client.ListPublicIPPolicyRules(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100")
	if err != nil {
		t.Fatalf("ListPublicIPPolicyRules() error = %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("ListPublicIPPolicyRules() count = %d, want 1", len(resp.Items))
	}
}

func TestGetPublicIPPolicyRule(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "",
			"data": map[string]any{
				"uuid":      "test-public-ip-uuid-1",
				"rule_name": "test-rule",
				"status":    "active",
				"action":    "accept",
				"source":    []map[string]any{{"all": true}},
				"services":  []map[string]any{{"name": "HTTP"}, {"name": "HTTPS"}},
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)

	rule, err := client.GetPublicIPPolicyRule(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100", "test-public-ip-uuid-1")
	if err != nil {
		t.Fatalf("GetPublicIPPolicyRule() error = %v", err)
	}
	if rule.DisplayName != "test-rule" {
		t.Fatalf("GetPublicIPPolicyRule() DisplayName = %q, want %q", rule.DisplayName, "test-rule")
	}

	_, err = client.GetPublicIPPolicyRule(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100", "nonexistent-rule")
	if err == nil {
		t.Fatal("GetPublicIPPolicyRule() expected error for missing rule, got nil")
	}
}

func TestDeletePublicIPPolicyRule(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("DELETE", "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			PolicyIDs []string `json:"policy_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode delete payload: %v", err)
		}
		if len(payload.PolicyIDs) != 1 || payload.PolicyIDs[0] != "test-public-ip-uuid-1" {
			t.Fatalf("unexpected policy_ids payload: %#v", payload.PolicyIDs)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClientForPublicIP(t, ms)
	if err := client.DeletePublicIPPolicyRule(context.Background(), "test-public-ip-uuid", "test-public-ip-uuid-1"); err != nil {
		t.Fatalf("DeletePublicIPPolicyRule() error = %v", err)
	}
}

func TestCreatePublicIP_DecodesWrappedResponseData(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := "/ext/api/v1/domain/test-org/project/test-project/public-ip"
	ms.AddHandler("POST", path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Public IP request accepted",
			"data": map[string]any{
				"uuid":       "pip-uuid-123",
				"public_ip":  "45.112.58.189",
				"target_vip": "10.10.14.67",
				"port_id":    19526,
				"status":     "creating",
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	created, err := client.CreatePublicIP(context.Background(), &models.CreatePublicIPRequest{Name: "test", PortID: 19526}, "N1")
	if err != nil {
		t.Fatalf("CreatePublicIP() unexpected error = %v", err)
	}
	if created == nil {
		t.Fatal("CreatePublicIP() returned nil")
	}
	if created.UUID != "pip-uuid-123" {
		t.Fatalf("CreatePublicIP() UUID = %q, want %q", created.UUID, "pip-uuid-123")
	}
}

func TestGetPublicIP_DecodesWrappedResponseData(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := "/ext/api/v1/domain/test-org/project/test-project/public-ip/pip-uuid-123"
	ms.AddHandler("GET", path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "",
			"data": map[string]any{
				"uuid":       "pip-uuid-123",
				"public_ip":  "45.112.58.189",
				"target_vip": "10.10.14.67",
				"port_id":    19526,
				"status":     "created",
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	publicIP, err := client.GetPublicIP(context.Background(), "pip-uuid-123")
	if err != nil {
		t.Fatalf("GetPublicIP() unexpected error = %v", err)
	}
	if publicIP == nil {
		t.Fatal("GetPublicIP() returned nil")
	}
	if publicIP.UUID != "pip-uuid-123" {
		t.Fatalf("GetPublicIP() UUID = %q, want %q", publicIP.UUID, "pip-uuid-123")
	}
	if publicIP.Status != "created" {
		t.Fatalf("GetPublicIP() Status = %q, want %q", publicIP.Status, "created")
	}
}

func TestFindPortIDByVIP_FindsMatchInListComputesPorts(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)

	portID, err := client.FindPortIDByVIP(context.Background(), "10.1.99.172", "N1")
	if err != nil {
		t.Fatalf("FindPortIDByVIP() unexpected error = %v", err)
	}
	if portID != 101 {
		t.Fatalf("FindPortIDByVIP() portID = %d, want 101", portID)
	}
}

func TestFindPortIDByVIP_InvalidVIP(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)

	_, err := client.FindPortIDByVIP(context.Background(), "not-an-ip", "N1")
	if err == nil {
		t.Fatal("FindPortIDByVIP() expected error for invalid VIP, got nil")
	}
}

func TestFindPortIDByVIP_NoMatch(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)

	_, err := client.FindPortIDByVIP(context.Background(), "10.255.255.255", "N1")
	if err == nil {
		t.Fatal("FindPortIDByVIP() expected no-match error, got nil")
	}
}

func TestFindPortIDByVIP_FallbackToGetComputeWhenListOmitsPorts(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	listPath := "/api/v2.1/computes/domain/test-org/project/test-project/computes/"
	getPath := "/api/v2.1/computes/domain/test-org/project/test-project/computes/c-1/"

	ms.AddHandler("GET", listPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]models.Compute{{
			ID:           "c-1",
			InstanceName: "vm-no-ports-in-list",
		}})
	})

	ms.AddHandler("GET", getPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.Compute{
			ID: "c-1",
			Ports: []models.Port{{
				ID:       909,
				FixedIPs: []string{"10.55.0.9"},
			}},
		})
	})

	client := newTestClientForPublicIP(t, ms)

	portID, err := client.FindPortIDByVIP(context.Background(), "10.55.0.9", "N1")
	if err != nil {
		t.Fatalf("FindPortIDByVIP() unexpected error = %v", err)
	}
	if portID != 909 {
		t.Fatalf("FindPortIDByVIP() portID = %d, want 909", portID)
	}
}

func TestFindPortIDByVIP_ResolvesLBVipPortID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)

	portID, err := client.FindPortIDByVIP(context.Background(), "10.0.0.100", "N1")
	if err != nil {
		t.Fatalf("FindPortIDByVIP() unexpected error = %v", err)
	}
	if portID != 1 {
		t.Fatalf("FindPortIDByVIP() portID = %d, want 1", portID)
	}
}

func TestFindPortIDByVIP_ResolvesLBVipPortIDFromNetworkVIPsAPI(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := "/api/v2.1/networks/domain/test-org/project/test-project/networks/ports/vips"
	ms.AddHandler("GET", path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]models.NetworkVIPPort{
			{
				LBName:           "gardener-vip-lb",
				VSName:           "parul",
				PortID:           855,
				AllowedIPAddress: "10.101.21.119",
			},
			{
				LBName:           "gardener-vip-lb",
				VSName:           "newvip",
				PortID:           17664,
				AllowedIPAddress: "10.101.21.35",
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)

	portID, err := client.FindPortIDByVIP(context.Background(), "10.101.21.35", "N1")
	if err != nil {
		t.Fatalf("FindPortIDByVIP() unexpected error = %v", err)
	}
	if portID != 17664 {
		t.Fatalf("FindPortIDByVIP() portID = %d, want 17664", portID)
	}
}
