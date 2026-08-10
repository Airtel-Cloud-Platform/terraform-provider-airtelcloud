package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientRedirect_RewritesInternalCCPExtensionHost(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ext/api/v1/domain/test-org/project/test-project/public-ip/test-public-ip-uuid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://ccp-extension:8000/ext/api/v1/domain/test-org/project/test-project/public-ip/test-public-ip-uuid/ready")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/ext/api/v1/domain/test-org/project/test-project/public-ip/test-public-ip-uuid/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uuid":   "test-public-ip-uuid",
			"status": "Created",
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-api-key", "test-api-secret", "north", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var out map[string]any
	err = c.Get(context.Background(), "/ext/api/v1/domain/test-org/project/test-project/public-ip/test-public-ip-uuid", &out)
	if err != nil {
		t.Fatalf("Get() unexpected error = %v", err)
	}

	if got := fmt.Sprintf("%v", out["status"]); got != "Created" {
		t.Fatalf("redirected response status = %q, want %q", got, "Created")
	}
}
