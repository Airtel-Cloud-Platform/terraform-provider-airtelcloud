package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
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
