package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const baremetalTestBasePath = "/api/baremetal-manager/v1/domain/test-org/project/test-project"

// newBaremetalTestClient wires a client to the mock server used across these tests.
func newBaremetalTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	client, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func TestResolveBaremetalNode(t *testing.T) {
	tests := []struct {
		name       string
		detailJSON string
		wantPortID int
	}{
		{
			name:       "resolves backend_port_id from wrapped detail response",
			detailJSON: `{"serverDetails":{"name":"bm-1","state":"Ready","uuid":"bm-uuid-1"},"networkInfo":{"portId":1234}}`,
			wantPortID: 1234,
		},
		{
			name:       "resolves backend_port_id from networkInfo.portId",
			detailJSON: `{"name":"bm-1","state":"Ready","uuid":"bm-uuid-1","networkInfo":{"portId":1234}}`,
			wantPortID: 1234,
		},
		{
			name:       "resolves backend_port_id when portId is a JSON string",
			detailJSON: `{"name":"bm-1","state":"Ready","uuid":"bm-uuid-1","networkInfo":{"portId":"17963"}}`,
			wantPortID: 17963,
		},
		{
			name:       "missing networkInfo yields zero port id",
			detailJSON: `{"name":"bm-1","state":"Ready","uuid":"bm-uuid-1"}`,
			wantPortID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := testutil.NewMockServer()
			defer ms.Close()

			ms.AddHandler("GET", baremetalTestBasePath+"/servers", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"count": 1,
					"items": []map[string]interface{}{
						{"name": "bm-1", "state": "Ready", "uuid": "bm-uuid-1", "availabilityZone": "S1"},
					},
				})
			})
			ms.AddHandler("GET", baremetalTestBasePath+"/server/bm-1", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.detailJSON))
			})

			client := newBaremetalTestClient(t, ms)

			bm, err := client.ResolveBaremetalNode(context.Background(), "", "bm-1")
			if err != nil {
				t.Fatalf("ResolveBaremetalNode() error = %v", err)
			}
			if bm.PortID != tt.wantPortID {
				t.Errorf("ResolveBaremetalNode() PortID = %d, want %d", bm.PortID, tt.wantPortID)
			}
		})
	}
}

func TestResolveBaremetalNodeNotReady(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", baremetalTestBasePath+"/servers", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count": 1,
			"items": []map[string]interface{}{
				{"name": "bm-1", "state": "Failed", "uuid": "bm-uuid-1"},
			},
		})
	})

	client := newBaremetalTestClient(t, ms)

	if _, err := client.ResolveBaremetalNode(context.Background(), "", "bm-1"); err == nil {
		t.Fatal("ResolveBaremetalNode() expected error for non-Ready server, got nil")
	}
}

func TestAllocateBaremetal(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("POST", baremetalTestBasePath+"/server", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, `"name":"bm-new"`) {
			t.Fatalf("request body missing name field: %s", body)
		}
		if !strings.Contains(body, `"flavor":"ccd.large"`) {
			t.Fatalf("request body missing flavor field: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	client := newBaremetalTestClient(t, ms)
	err := client.AllocateBaremetal(context.Background(), &models.AllocateBaremetalRequest{
		Name:    "bm-new",
		Flavor:  "ccd.large",
		OSImage: "Ubuntu22_04_Aug2026",
	})
	if err != nil {
		t.Fatalf("AllocateBaremetal() error = %v", err)
	}
}

func TestUpdateBaremetal(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("PUT", baremetalTestBasePath+"/server/bm-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, `"policyEnabled":true`) {
			t.Fatalf("request body missing policyEnabled field: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	client := newBaremetalTestClient(t, ms)
	enabled := true
	err := client.UpdateBaremetal(context.Background(), "bm-1", &models.UpdateBaremetalRequest{PolicyEnabled: &enabled})
	if err != nil {
		t.Fatalf("UpdateBaremetal() error = %v", err)
	}
}

func TestReleaseBaremetal(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("DELETE", baremetalTestBasePath+"/server/bm-1", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("systemId"); got != "sys-1" {
			t.Fatalf("systemId query = %q, want %q", got, "sys-1")
		}
		if got := r.URL.Query().Get("deleteDisks"); got != "true" {
			t.Fatalf("deleteDisks query = %q, want %q", got, "true")
		}
		if got := r.URL.Query().Get("secureErase"); got != "false" {
			t.Fatalf("secureErase query = %q, want %q", got, "false")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	client := newBaremetalTestClient(t, ms)
	deleteDisks := true
	secureErase := false
	err := client.ReleaseBaremetal(context.Background(), "bm-1", &models.ReleaseBaremetalOptions{
		SystemID:    "sys-1",
		DeleteDisks: &deleteDisks,
		SecureErase: &secureErase,
	})
	if err != nil {
		t.Fatalf("ReleaseBaremetal() error = %v", err)
	}
}
