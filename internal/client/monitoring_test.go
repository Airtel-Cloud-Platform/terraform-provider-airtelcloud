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

const (
	testMonitoringDashboardID = "65f0a1b2c3d4e5f678901234"
	testMonitoringProviderID  = "65f0b2c3d4e5f67890123456"
	testMonitoringDashboards  = "/api/monitoring-admin/v2/dashboards"
	testMonitoringProviders   = "/api/monitoring-admin/v2/providers"
)

func newMonitoringTestClient(t *testing.T, ms *testutil.MockServer) *Client {
	t.Helper()
	baseURL := strings.TrimSuffix(ms.URL, "/")
	c, err := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c
}

func sampleMonitoringDashboard() models.MonitoringDashboard {
	desc := "Default compute dashboard"
	created := "2026-09-30T12:00:00Z"
	return models.MonitoringDashboard{
		ID:           testMonitoringDashboardID,
		Name:         "compute-default",
		ResourceType: "compute",
		Active:       true,
		Description:  &desc,
		Panels: []models.MonitoringDashboardPanel{{
			Title:        "CPU Utilization",
			ChartType:    "line",
			X:            0,
			Y:            0,
			W:            12,
			H:            4,
			DisplayOrder: 0,
			Queries: []models.MonitoringDashboardQuery{{
				QueryTemplate: `avg(rate(cpu_usage{instance="{{instance}}"}[5m]))`,
				QueryType:     "prometheus",
				RefID:         "A",
			}},
		}},
		CreatedAt: &created,
		UpdatedAt: &created,
	}
}

func sampleMonitoringProvider() models.MonitoringProvider {
	az := "S1"
	desc := "Prometheus for compute in S1"
	created := "2026-09-30T12:00:00Z"
	return models.MonitoringProvider{
		ID:               testMonitoringProviderID,
		Name:             "south-prometheus-compute",
		ResourceType:     "compute",
		ProviderType:     "prometheus",
		DashboardAuthURL: "https://grafana.example.com",
		APIKeyName:       "terraform-provider",
		DataSourceUID:    "prom-uid-1",
		DataSourceID:     "1",
		AZName:           &az,
		Active:           true,
		Description:      &desc,
		CreatedAt:        &created,
		UpdatedAt:        &created,
	}
}

func TestCreateMonitoringDashboard(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var gotBody models.CreateMonitoringDashboardRequest
	ms.AddHandler("POST", testMonitoringDashboards, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		writeJSON(w, http.StatusOK, sampleMonitoringDashboard())
	})

	active := true
	resp, err := newMonitoringTestClient(t, ms).CreateMonitoringDashboard(context.Background(), &models.CreateMonitoringDashboardRequest{
		Name:         "compute-default",
		ResourceType: "compute",
		Active:       &active,
		Panels: []models.MonitoringDashboardPanel{{
			Title:     "CPU Utilization",
			ChartType: "line",
			W:         12,
			H:         4,
		}},
	})
	if err != nil {
		t.Fatalf("CreateMonitoringDashboard() error = %v", err)
	}
	if resp.ID != testMonitoringDashboardID {
		t.Fatalf("id = %q", resp.ID)
	}
	if gotBody.Name != "compute-default" || gotBody.ResourceType != "compute" {
		t.Fatalf("request body = %+v", gotBody)
	}
}

func TestGetMonitoringDashboard(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMonitoringDashboards+"/"+testMonitoringDashboardID, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMonitoringDashboard())
	})

	resp, err := newMonitoringTestClient(t, ms).GetMonitoringDashboard(context.Background(), testMonitoringDashboardID)
	if err != nil {
		t.Fatalf("GetMonitoringDashboard() error = %v", err)
	}
	if resp.Name != "compute-default" || len(resp.Panels) != 1 {
		t.Fatalf("response = %+v", resp)
	}

	ms.SetErrorResponse("GET", testMonitoringDashboards+"/missing", 404, "Not found")
	if _, err := newMonitoringTestClient(t, ms).GetMonitoringDashboard(context.Background(), "missing"); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestUpdateMonitoringDashboard(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var gotBody models.UpdateMonitoringDashboardRequest
	ms.AddHandler("PUT", testMonitoringDashboards+"/"+testMonitoringDashboardID, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		updated := sampleMonitoringDashboard()
		updated.Name = gotBody.Name
		writeJSON(w, http.StatusOK, updated)
	})

	active := true
	resp, err := newMonitoringTestClient(t, ms).UpdateMonitoringDashboard(context.Background(), testMonitoringDashboardID, &models.UpdateMonitoringDashboardRequest{
		Name:         "compute-updated",
		ResourceType: "compute",
		Active:       &active,
	})
	if err != nil {
		t.Fatalf("UpdateMonitoringDashboard() error = %v", err)
	}
	if resp.Name != "compute-updated" || gotBody.Name != "compute-updated" {
		t.Fatalf("name=%q body=%+v", resp.Name, gotBody)
	}
}

func TestDeleteMonitoringDashboard(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("DELETE", testMonitoringDashboards+"/"+testMonitoringDashboardID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err := newMonitoringTestClient(t, ms).DeleteMonitoringDashboard(context.Background(), testMonitoringDashboardID); err != nil {
		t.Fatalf("DeleteMonitoringDashboard() error = %v", err)
	}

	ms.SetErrorResponse("DELETE", testMonitoringDashboards+"/missing", 404, "Not found")
	if err := newMonitoringTestClient(t, ms).DeleteMonitoringDashboard(context.Background(), "missing"); err != nil {
		t.Fatalf("DeleteMonitoringDashboard() 404 should succeed, got %v", err)
	}
}

func TestCreateMonitoringProvider(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var gotBody models.CreateMonitoringProviderRequest
	ms.AddHandler("POST", testMonitoringProviders, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		writeJSON(w, http.StatusOK, sampleMonitoringProvider())
	})

	providerType := "prometheus"
	active := true
	az := "S1"
	resp, err := newMonitoringTestClient(t, ms).CreateMonitoringProvider(context.Background(), &models.CreateMonitoringProviderRequest{
		Name:             "south-prometheus-compute",
		ResourceType:     "compute",
		ProviderType:     &providerType,
		DashboardAuthURL: "https://grafana.example.com",
		APIKeyName:       "terraform-provider",
		APIToken:         "glsa_secret",
		DataSourceUID:    "prom-uid-1",
		DataSourceID:     "1",
		AZName:           &az,
		Active:           &active,
	})
	if err != nil {
		t.Fatalf("CreateMonitoringProvider() error = %v", err)
	}
	if resp.ID != testMonitoringProviderID {
		t.Fatalf("id = %q", resp.ID)
	}
	if gotBody.APIToken != "glsa_secret" {
		t.Fatalf("api_token not sent, body=%+v", gotBody)
	}
}

func TestGetMonitoringProvider(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testMonitoringProviders+"/"+testMonitoringProviderID, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, sampleMonitoringProvider())
	})

	resp, err := newMonitoringTestClient(t, ms).GetMonitoringProvider(context.Background(), testMonitoringProviderID)
	if err != nil {
		t.Fatalf("GetMonitoringProvider() error = %v", err)
	}
	if resp.Name != "south-prometheus-compute" || resp.DataSourceUID != "prom-uid-1" {
		t.Fatalf("response = %+v", resp)
	}

	ms.SetErrorResponse("GET", testMonitoringProviders+"/missing", 404, "Not found")
	if _, err := newMonitoringTestClient(t, ms).GetMonitoringProvider(context.Background(), "missing"); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestUpdateMonitoringProvider(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	var rawBody map[string]json.RawMessage
	ms.AddHandler("PUT", testMonitoringProviders+"/"+testMonitoringProviderID, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		writeJSON(w, http.StatusOK, sampleMonitoringProvider())
	})

	// Omit token so existing API token is kept.
	_, err := newMonitoringTestClient(t, ms).UpdateMonitoringProvider(context.Background(), testMonitoringProviderID, &models.UpdateMonitoringProviderRequest{
		Name:             "south-prometheus-compute",
		ResourceType:     "compute",
		DashboardAuthURL: "https://grafana.example.com",
		APIKeyName:       "terraform-provider",
		DataSourceUID:    "prom-uid-1",
		DataSourceID:     "1",
	})
	if err != nil {
		t.Fatalf("UpdateMonitoringProvider() error = %v", err)
	}
	if _, ok := rawBody["api_token"]; ok {
		t.Fatalf("api_token should be omitted when empty, body keys=%v", rawBody)
	}

	token := "new-token"
	ms.AddHandler("PUT", testMonitoringProviders+"/"+testMonitoringProviderID, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		writeJSON(w, http.StatusOK, sampleMonitoringProvider())
	})
	_, err = newMonitoringTestClient(t, ms).UpdateMonitoringProvider(context.Background(), testMonitoringProviderID, &models.UpdateMonitoringProviderRequest{
		Name:             "south-prometheus-compute",
		ResourceType:     "compute",
		DashboardAuthURL: "https://grafana.example.com",
		APIKeyName:       "terraform-provider",
		APIToken:         &token,
		DataSourceUID:    "prom-uid-1",
		DataSourceID:     "1",
	})
	if err != nil {
		t.Fatalf("UpdateMonitoringProvider() with token error = %v", err)
	}
	var sentToken string
	if err := json.Unmarshal(rawBody["api_token"], &sentToken); err != nil || sentToken != "new-token" {
		t.Fatalf("api_token = %s err=%v", rawBody["api_token"], err)
	}
}

func TestDeleteMonitoringProvider(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("DELETE", testMonitoringProviders+"/"+testMonitoringProviderID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err := newMonitoringTestClient(t, ms).DeleteMonitoringProvider(context.Background(), testMonitoringProviderID); err != nil {
		t.Fatalf("DeleteMonitoringProvider() error = %v", err)
	}

	ms.SetErrorResponse("DELETE", testMonitoringProviders+"/missing", 404, "Not found")
	if err := newMonitoringTestClient(t, ms).DeleteMonitoringProvider(context.Background(), "missing"); err != nil {
		t.Fatalf("DeleteMonitoringProvider() 404 should succeed, got %v", err)
	}
}
