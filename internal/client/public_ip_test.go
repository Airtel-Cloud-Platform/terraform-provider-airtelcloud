package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

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
			name: "successful list wrapped in data",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"message": "Public IPs fetched successfully.",
						"data": map[string]any{
							"items": []map[string]any{{
								"uuid":   "test-public-ip-uuid",
								"name":   "test-public-ip",
								"ip":     "103.239.168.100",
								"status": "reserved",
							}},
							"count": 1,
						},
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

func TestDeletePublicIPWithWait(t *testing.T) {
	t.Run("succeeds when backend reports deleted status", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		deletePath := testPublicIPBasePath + "/test-public-ip-uuid"
		getPath := testPublicIPBasePath + "/test-public-ip-uuid"

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Public IP deletion in progress",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "deleting",
				},
			})
		})

		ms.AddHandler("GET", getPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "Deleted",
				},
			})
		})

		client := newTestClientForPublicIP(t, ms)
		if err := client.DeletePublicIPWithWait(context.Background(), "test-public-ip-uuid", 2*time.Second); err != nil {
			t.Fatalf("DeletePublicIPWithWait() error = %v, want nil when backend reports a deleted status", err)
		}
	})

	t.Run("returns error when backend reports failed status", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		deletePath := testPublicIPBasePath + "/test-public-ip-uuid"
		getPath := testPublicIPBasePath + "/test-public-ip-uuid"

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Public IP deletion in progress",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "deleting",
				},
			})
		})

		ms.AddHandler("GET", getPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "Failed",
				},
			})
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPWithWait(context.Background(), "test-public-ip-uuid", 2*time.Second)
		if err == nil {
			t.Fatal("DeletePublicIPWithWait() expected failure on failed backend status")
		}
		if !strings.Contains(err.Error(), "delete failed") {
			t.Fatalf("DeletePublicIPWithWait() error = %v, want delete failed message", err)
		}
	})

	t.Run("returns error when backend reports timeout status", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		deletePath := testPublicIPBasePath + "/test-public-ip-uuid"
		getPath := testPublicIPBasePath + "/test-public-ip-uuid"

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Public IP deletion in progress",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "deleting",
				},
			})
		})

		ms.AddHandler("GET", getPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid",
					"public_ip": "103.239.168.100",
					"status":    "timed_out",
				},
			})
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPWithWait(context.Background(), "test-public-ip-uuid", 2*time.Second)
		if err == nil {
			t.Fatal("DeletePublicIPWithWait() expected failure on timed_out backend status")
		}
		if !strings.Contains(err.Error(), "delete failed") {
			t.Fatalf("DeletePublicIPWithWait() error = %v, want delete failed message for timeout state", err)
		}
	})
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
			name: "resolves wrapped list by name field",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", testPublicIPBasePath, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"message": "Public IPs fetched successfully.",
						"data": map[string]any{
							"items": []map[string]any{{
								"uuid": "wrapped-uuid",
								"name": "tft-pip-1",
							}},
							"count": 1,
						},
					})
				})
			},
			lookup:   "tft-pip-1",
			wantUUID: "wrapped-uuid",
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

func TestWaitForPublicIPPolicyRuleReady(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	policyPath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
	var calls int
	ms.AddHandler("GET", policyPath, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls < 3 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid-1",
					"rule_name": "test-rule",
					"status":    "creating",
					"state":     "creating",
					"action":    "accept",
					"source":    []map[string]any{{"all": true}},
					"services":  []map[string]any{{"name": "HTTP"}},
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "",
			"data": map[string]any{
				"uuid":      "test-public-ip-uuid-1",
				"rule_name": "test-rule",
				"status":    "active",
				"state":     "active",
				"action":    "accept",
				"source":    []map[string]any{{"all": true}},
				"services":  []map[string]any{{"name": "HTTP"}},
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	rule, err := client.WaitForPublicIPPolicyRuleReady(context.Background(), "test-public-ip-uuid", "10.1.99.172", "103.239.168.100", "test-public-ip-uuid-1", 15*time.Second)
	if err != nil {
		t.Fatalf("WaitForPublicIPPolicyRuleReady() error = %v", err)
	}
	if rule == nil || rule.State != "active" {
		t.Fatalf("WaitForPublicIPPolicyRuleReady() state = %v, want active", rule)
	}
	if calls < 3 {
		t.Fatalf("WaitForPublicIPPolicyRuleReady() calls = %d, want at least 3", calls)
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
	if got, ok := sourceObj["create_new"].(bool); !ok || got {
		t.Fatalf("payload source[0].create_new = %v, want false", sourceObj["create_new"])
	}
	if got, ok := sourceObj["source_type"].(string); !ok || got != "all" {
		t.Fatalf("payload source[0].source_type = %v, want all", sourceObj["source_type"])
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
	if got, ok := serviceObj["name"].(string); !ok || got != "uuid-http" {
		t.Fatalf("payload services[0].name = %v, want uuid-http", serviceObj["name"])
	}
	if got, ok := payload["action"].(string); !ok || got != "accept" {
		t.Fatalf("payload action = %v, want accept", payload["action"])
	}
	if got, ok := payload["revision_note"].(string); !ok || got != "creating Policy" {
		t.Fatalf("payload revision_note = %v, want creating Policy", payload["revision_note"])
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

	createNewFalse := false
	isDefaultFalse := false

	client := newTestClientForPublicIP(t, ms)
	_, err := client.CreatePublicIPPolicyRule(context.Background(), &models.CreatePublicIPPolicyRuleRequest{
		DisplayName: "temporal-rule",
		SourceConfig: []models.PublicIPPolicyRuleSourceInput{
			{
				CreateNew:  &createNewFalse,
				IPCIDR:     "182.77.78.18/32",
				SourceType: "ip_cidr",
			},
			{
				SourceType: "geographic",
				Geographic: &models.PublicIPPolicyRuleGeographicInput{
					CountryCode: "IN",
					CountryName: "India",
				},
			},
		},
		ServiceConfig: []models.PublicIPPolicyRuleServiceInput{
			{CreateNew: &createNewFalse, Name: "tcp-443-443", IsDefault: &isDefaultFalse},
			{CreateNew: &createNewFalse, Name: "SSH", IsDefault: &isDefaultFalse},
			{CreateNew: &createNewFalse, Name: "HTTP", IsDefault: &isDefaultFalse},
			{CreateNew: &createNewFalse, Name: "HTTPS", IsDefault: &isDefaultFalse},
		},
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
	if !ok || len(sourceRaw) != 2 {
		t.Fatalf("payload source = %v, want two source objects", payload["source"])
	}
	cidrObj, ok := sourceRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("payload source[0] = %T, want object", sourceRaw[0])
	}
	if got, ok := cidrObj["create_new"].(bool); !ok || got {
		t.Fatalf("payload source[0].create_new = %v, want false", cidrObj["create_new"])
	}
	if got, ok := cidrObj["source_type"].(string); !ok || got != "ip_cidr" {
		t.Fatalf("payload source[0].source_type = %v, want ip_cidr", cidrObj["source_type"])
	}
	if got, ok := cidrObj["ip_cidr"].(string); !ok || got != "182.77.78.18/32" {
		t.Fatalf("payload source[0].ip_cidr = %v, want 182.77.78.18/32", cidrObj["ip_cidr"])
	}
	geoObj, ok := sourceRaw[1].(map[string]any)
	if !ok {
		t.Fatalf("payload source[1] = %T, want object", sourceRaw[1])
	}
	if _, hasCreateNew := geoObj["create_new"]; hasCreateNew {
		t.Fatalf("payload source[1] must not include create_new, got %v", geoObj)
	}
	if got, ok := geoObj["source_type"].(string); !ok || got != "geographic" {
		t.Fatalf("payload source[1].source_type = %v, want geographic", geoObj["source_type"])
	}
	geo, ok := geoObj["geographic"].(map[string]any)
	if !ok {
		t.Fatalf("payload source[1].geographic = %T, want object", geoObj["geographic"])
	}
	if got, ok := geo["country_code"].(string); !ok || got != "IN" {
		t.Fatalf("payload source[1].geographic.country_code = %v, want IN", geo["country_code"])
	}
	if got, ok := geo["country_name"].(string); !ok || got != "India" {
		t.Fatalf("payload source[1].geographic.country_name = %v, want India", geo["country_name"])
	}

	servicesRaw, ok := payload["services"].([]any)
	if !ok || len(servicesRaw) != 4 {
		t.Fatalf("payload services = %v, want four service objects", payload["services"])
	}
	serviceObj, ok := servicesRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("payload services[0] = %T, want object", servicesRaw[0])
	}
	if got, ok := serviceObj["create_new"].(bool); !ok || got {
		t.Fatalf("payload services[0].create_new = %v, want false", serviceObj["create_new"])
	}
	if got, ok := serviceObj["name"].(string); !ok || got != "tcp-443-443" {
		t.Fatalf("payload services[0].name = %v, want tcp-443-443", serviceObj["name"])
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

func TestDeletePublicIPPolicyRuleWithWait(t *testing.T) {
	t.Run("waits until rule is not found", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		rulePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
		deletePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy"

		var getCount int
		ms.AddHandler("GET", rulePath, func(w http.ResponseWriter, r *http.Request) {
			getCount++
			w.Header().Set("Content-Type", "application/json")
			if getCount < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"message": "",
					"data": map[string]any{
						"uuid":      "test-public-ip-uuid-1",
						"rule_name": "test-rule",
						"status":    "active",
						"action":    "accept",
						"source":    []map[string]any{{"all": true}},
						"services":  []map[string]any{{"name": "HTTP"}},
					},
				})
				return
			}

			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  404,
				"message": "Not found",
			})
		})

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPPolicyRuleWithWait(
			context.Background(),
			"test-public-ip-uuid",
			"10.1.99.172",
			"103.239.168.100",
			"test-public-ip-uuid-1",
			4*time.Second,
		)
		if err != nil {
			t.Fatalf("DeletePublicIPPolicyRuleWithWait() error = %v", err)
		}
	})

	t.Run("succeeds when backend reports deleted state", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		rulePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
		deletePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy"

		ms.AddHandler("GET", rulePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid-1",
					"rule_name": "test-rule",
					"status":    "Deleted",
					"state":     "deleted",
					"action":    "accept",
					"source":    []map[string]any{{"all": true}},
					"services":  []map[string]any{{"name": "HTTP"}},
				},
			})
		})

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPPolicyRuleWithWait(
			context.Background(),
			"test-public-ip-uuid",
			"10.1.99.172",
			"103.239.168.100",
			"test-public-ip-uuid-1",
			2*time.Second,
		)
		if err != nil {
			t.Fatalf("DeletePublicIPPolicyRuleWithWait() error = %v, want nil when backend reports a deleted state", err)
		}
	})

	t.Run("returns failure when backend reports failed state", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		rulePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
		deletePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy"

		ms.AddHandler("GET", rulePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":          "test-public-ip-uuid-1",
					"rule_name":     "test-rule",
					"status":        "Failed",
					"state":         "failed",
					"error_message": "rule creation failed",
					"action":        "accept",
					"source":        []map[string]any{{"all": true}},
					"services":      []map[string]any{{"name": "HTTP"}},
				},
			})
		})

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPPolicyRuleWithWait(
			context.Background(),
			"test-public-ip-uuid",
			"10.1.99.172",
			"103.239.168.100",
			"test-public-ip-uuid-1",
			2*time.Second,
		)
		if err == nil {
			t.Fatal("DeletePublicIPPolicyRuleWithWait() expected failure error for failed backend state")
		}
		if !strings.Contains(err.Error(), "failed") {
			t.Fatalf("DeletePublicIPPolicyRuleWithWait() error = %v, want failed-state message", err)
		}
	})

	t.Run("returns failure when backend reports timeout state", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		rulePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
		deletePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy"

		ms.AddHandler("GET", rulePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid-1",
					"rule_name": "test-rule",
					"status":    "timed_out",
					"state":     "timed_out",
					"action":    "accept",
					"source":    []map[string]any{{"all": true}},
					"services":  []map[string]any{{"name": "HTTP"}},
				},
			})
		})

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPPolicyRuleWithWait(
			context.Background(),
			"test-public-ip-uuid",
			"10.1.99.172",
			"103.239.168.100",
			"test-public-ip-uuid-1",
			2*time.Second,
		)
		if err == nil {
			t.Fatal("DeletePublicIPPolicyRuleWithWait() expected failure error for timeout backend state")
		}
		if !strings.Contains(err.Error(), "failed") {
			t.Fatalf("DeletePublicIPPolicyRuleWithWait() error = %v, want failed-state message for timeout state", err)
		}
	})

	t.Run("times out when rule remains present", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()

		rulePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy/test-public-ip-uuid-1"
		deletePath := "/ext/api/v1/domain/test-org/project/test-project/public-ip-id/test-public-ip-uuid/policy"

		ms.AddHandler("GET", rulePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "",
				"data": map[string]any{
					"uuid":      "test-public-ip-uuid-1",
					"rule_name": "test-rule",
					"status":    "active",
					"action":    "accept",
					"source":    []map[string]any{{"all": true}},
					"services":  []map[string]any{{"name": "HTTP"}},
				},
			})
		})

		ms.AddHandler("DELETE", deletePath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		client := newTestClientForPublicIP(t, ms)
		err := client.DeletePublicIPPolicyRuleWithWait(
			context.Background(),
			"test-public-ip-uuid",
			"10.1.99.172",
			"103.239.168.100",
			"test-public-ip-uuid-1",
			1*time.Second,
		)
		if err == nil {
			t.Fatal("DeletePublicIPPolicyRuleWithWait() expected timeout error")
		}
		if !strings.Contains(err.Error(), "still present") {
			t.Fatalf("DeletePublicIPPolicyRuleWithWait() error = %v, want timeout message", err)
		}
	})
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
	created, err := client.CreatePublicIP(context.Background(), &models.CreatePublicIPRequest{Name: "test", PortID: intPtr(19526)}, "N1")
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

func TestCreatePublicIP_SendsNullPortID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := testPublicIPBasePath
	ms.AddHandler("POST", path, func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode create body: %v", err)
		}
		if payload["name"] != "temporal-pip" {
			t.Errorf("name = %v, want temporal-pip", payload["name"])
		}
		if payload["description"] != "new pip" {
			t.Errorf("description = %v, want new pip", payload["description"])
		}
		if payload["port_id"] != nil {
			t.Errorf("port_id = %v, want null", payload["port_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"uuid":      "pip-uuid-reserve",
				"public_ip": "45.112.58.189",
				"status":    "reserved",
			},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	created, err := client.CreatePublicIP(context.Background(), &models.CreatePublicIPRequest{
		Name:        "temporal-pip",
		Description: "new pip",
		PortID:      nil,
	}, "S1")
	if err != nil {
		t.Fatalf("CreatePublicIP() unexpected error = %v", err)
	}
	if created.UUID != "pip-uuid-reserve" {
		t.Fatalf("CreatePublicIP() UUID = %q, want pip-uuid-reserve", created.UUID)
	}
}

func TestAttachPublicIP_PostsPortID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := testPublicIPBasePath + "/a1ad4176-7955-44cc-ac8a-4d72ac77559a/attach"
	ms.AddHandler("POST", path, func(w http.ResponseWriter, r *http.Request) {
		var payload models.AttachPublicIPRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode attach body: %v", err)
		}
		if payload.PortID != 27148 {
			t.Errorf("port_id = %d, want 27148", payload.PortID)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	client := newTestClientForPublicIP(t, ms)
	err := client.AttachPublicIP(context.Background(), "a1ad4176-7955-44cc-ac8a-4d72ac77559a", 27148, "S1")
	if err != nil {
		t.Fatalf("AttachPublicIP() unexpected error = %v", err)
	}
}

func TestNormalizePublicIPResourceType(t *testing.T) {
	cases := map[string]string{
		"vm":        PublicIPResourceTypeVM,
		"VM":        PublicIPResourceTypeVM,
		"compute":   PublicIPResourceTypeVM,
		"lb":        PublicIPResourceTypeLB,
		"baremetal": PublicIPResourceTypeBaremetal,
	}
	for in, want := range cases {
		got, err := NormalizePublicIPResourceType(in)
		if err != nil {
			t.Fatalf("NormalizePublicIPResourceType(%q) error = %v", in, err)
		}
		if got != want {
			t.Fatalf("NormalizePublicIPResourceType(%q) = %q, want %q", in, got, want)
		}
	}

	if _, err := NormalizePublicIPResourceType("volume"); err == nil {
		t.Fatal("NormalizePublicIPResourceType() expected error for unsupported type")
	}
}

func TestFindPortForResource_VM(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	client := newTestClientForPublicIP(t, ms)

	t.Run("matches target vip", func(t *testing.T) {
		portID, vip, err := client.FindPortForResource(context.Background(), "vm", "instance-1", "10.1.99.172", "S1")
		if err != nil {
			t.Fatalf("FindPortForResource() unexpected error = %v", err)
		}
		if portID != 101 || vip != "10.1.99.172" {
			t.Fatalf("FindPortForResource() = (%d, %q), want (101, 10.1.99.172)", portID, vip)
		}
	})

	t.Run("defaults to primary port", func(t *testing.T) {
		portID, vip, err := client.FindPortForResource(context.Background(), "vm", "instance-2", "", "S1")
		if err != nil {
			t.Fatalf("FindPortForResource() unexpected error = %v", err)
		}
		if portID != 202 || vip != "10.1.99.200" {
			t.Fatalf("FindPortForResource() = (%d, %q), want (202, 10.1.99.200)", portID, vip)
		}
	})

	t.Run("rejects vip not on resource", func(t *testing.T) {
		if _, _, err := client.FindPortForResource(context.Background(), "vm", "instance-1", "10.9.9.9", "S1"); err == nil {
			t.Fatal("FindPortForResource() expected error for mismatched target_vip")
		}
	})
}

func TestFindPortForResource_LB(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	path := "/api/v2.1/networks/domain/test-org/project/test-project/networks/ports/vips"
	ms.AddHandler("GET", path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]models.NetworkVIPPort{
			{LBName: "gardener-vip-lb", PortID: 855, AllowedIPAddress: "10.101.21.119"},
			{LBName: "gardener-vip-lb", PortID: 17664, AllowedIPAddress: "10.101.21.35"},
		})
	})

	client := newTestClientForPublicIP(t, ms)
	portID, vip, err := client.FindPortForResource(context.Background(), "lb", "gardener-vip-lb", "10.101.21.35", "S1")
	if err != nil {
		t.Fatalf("FindPortForResource() unexpected error = %v", err)
	}
	if portID != 17664 || vip != "10.101.21.35" {
		t.Fatalf("FindPortForResource() = (%d, %q), want (17664, 10.101.21.35)", portID, vip)
	}
}

func TestIsPublicIPAttached(t *testing.T) {
	if !IsPublicIPAttached("attached") || !IsPublicIPAttached("Attached") {
		t.Fatal("expected attached statuses to pass")
	}
	if IsPublicIPAttached("reserved") || IsPublicIPAttached("created") || IsPublicIPAttached("") {
		t.Fatal("expected non-attached statuses to fail")
	}
}

func TestRequirePublicIPAttached(t *testing.T) {
	t.Run("allows attached with port", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testPublicIPBasePath+"/pip-uuid-123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"uuid": "pip-uuid-123", "status": "attached", "port_id": 27148, "target_vip": "10.10.3.237"},
			})
		})
		client := newTestClientForPublicIP(t, ms)
		ip, err := client.RequirePublicIPAttached(context.Background(), "pip-uuid-123")
		if err != nil {
			t.Fatalf("RequirePublicIPAttached() unexpected error = %v", err)
		}
		if ip.UUID != "pip-uuid-123" {
			t.Fatalf("UUID = %q", ip.UUID)
		}
	})

	t.Run("rejects reserved", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testPublicIPBasePath+"/pip-uuid-123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"uuid": "pip-uuid-123", "status": "reserved", "object_name": "temporal-pip"},
			})
		})
		client := newTestClientForPublicIP(t, ms)
		ip, err := client.RequirePublicIPAttached(context.Background(), "pip-uuid-123")
		if err == nil {
			t.Fatal("RequirePublicIPAttached() expected error for reserved")
		}
		if ip == nil || ip.Status != "reserved" {
			t.Fatalf("expected reserved public IP in error path, got %+v", ip)
		}
		if !errors.Is(err, ErrPublicIPNotAttached) {
			t.Fatalf("error = %v, want ErrPublicIPNotAttached", err)
		}
	})

	t.Run("rejects attached without resource", func(t *testing.T) {
		ms := testutil.NewMockServer()
		defer ms.Close()
		ms.AddHandler("GET", testPublicIPBasePath+"/pip-uuid-123", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"uuid": "pip-uuid-123", "status": "attached"},
			})
		})
		client := newTestClientForPublicIP(t, ms)
		_, err := client.RequirePublicIPAttached(context.Background(), "pip-uuid-123")
		if err == nil {
			t.Fatal("RequirePublicIPAttached() expected error when no port or target_vip")
		}
		if !errors.Is(err, ErrPublicIPNotAttached) {
			t.Fatalf("error = %v, want ErrPublicIPNotAttached", err)
		}
	})
}

func intPtr(v int) *int { return &v }
