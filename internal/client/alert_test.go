package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	testAlertID   = "b4ca963b-821c-46d2-a59a-2299f04780a1"
	testAlertBase = "/api/v2.1/alert/domain/test-org/project/test-project/alerts"
	testAlertItem = testAlertBase + "/" + testAlertID
)

func TestCreateAlert(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("POST", testAlertBase, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got models.CreateAlertRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.Name != "high-cpu-vm1" || got.Organization != "test-org" || got.Project != "test-project" {
			t.Fatalf("create body = %+v", got)
		}
		if got.Threshold != 80 || got.ComparisonOperator != ">" || len(got.ResourcesInfo) != 1 {
			t.Fatalf("create body = %+v", got)
		}
		if got.ResourcesInfo[0].ResourceID != "vm-uuid-1" {
			t.Fatalf("resources_info = %+v", got.ResourcesInfo)
		}
		writeJSON(w, http.StatusOK, models.Alert{
			ID:                 101,
			Name:               got.Name,
			ResourceTypeName:   got.ResourceTypeName,
			MetricType:         got.MetricType,
			MetricName:         got.MetricName,
			Threshold:          got.Threshold,
			ComparisonOperator: got.ComparisonOperator,
			Enabled:            got.Enabled,
			Severity:           got.Severity,
			State:              "OK",
			CreatedAt:          "2026-09-30T12:00:00Z",
		})
	})
	ms.AddHandler("GET", testAlertBase, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.AlertListResponse{
			AlertData: []models.AlertListItem{{
				UUIDID: testAlertID,
				Name:   "high-cpu-vm1",
			}},
			Total: 1,
		})
	})

	got, err := newPostgresTestClient(t, ms).CreateAlert(context.Background(), &models.CreateAlertRequest{
		Name:               "high-cpu-vm1",
		Organization:       "test-org",
		Project:            "test-project",
		ResourceTypeName:   "compute",
		MetricType:         "cpu",
		MetricName:         "cpu_utilization",
		Threshold:          80,
		ComparisonOperator: ">",
		Enabled:            true,
		Severity:           2,
		ResourcesInfo: []models.ResourceInfo{{
			ResourceID:         "vm-uuid-1",
			ResourceName:       "app-vm-1",
			ResourceIP:         "10.0.0.10",
			ProviderInstanceID: "i-123",
		}},
		AlertEmails: []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}
	if got.ID != 101 || got.State != "OK" {
		t.Fatalf("alert = %+v", got)
	}
	if got.UUIDID != testAlertID {
		t.Fatalf("uuid_id = %q, want %q", got.UUIDID, testAlertID)
	}
}

func TestGetAlertWithoutID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testAlertItem, func(w http.ResponseWriter, r *http.Request) {
		// AlertListShow shape: no id
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"name":                "high-cpu-vm1",
			"metric_type":         "cpu",
			"resource_type_name":  "compute",
			"metric_name":         "cpu_utilization",
			"threshold":           80,
			"comparison_operator": ">",
			"severity":            2,
			"enabled":             true,
			"description":         "CPU above 80%",
			"resource_info":       []map[string]interface{}{},
			"alert_emails":        []string{"ops@example.com"},
		})
	})

	got, err := newPostgresTestClient(t, ms).GetAlert(context.Background(), testAlertID)
	if err != nil {
		t.Fatalf("GetAlert() error = %v", err)
	}
	if got.ID != 0 {
		t.Fatalf("expected id absent/zero, got %d", got.ID)
	}
	if got.Name != "high-cpu-vm1" || got.Threshold != 80 || !got.Enabled {
		t.Fatalf("get = %+v", got)
	}
}

func TestGetAlertWithID(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("GET", testAlertItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.AlertGetResponse{
			ID:                 101,
			Name:               "high-cpu-vm1",
			ResourceTypeName:   "compute",
			MetricType:         "cpu",
			MetricName:         "cpu_utilization",
			Threshold:          80,
			ComparisonOperator: ">",
			Enabled:            true,
			Severity:           2,
			State:              "OK",
			CreatedAt:          "2026-09-30T12:00:00Z",
		})
	})

	got, err := newPostgresTestClient(t, ms).GetAlert(context.Background(), testAlertID)
	if err != nil {
		t.Fatalf("GetAlert() error = %v", err)
	}
	if got.ID != 101 || got.State != "OK" {
		t.Fatalf("get = %+v", got)
	}
}

func TestUpdateAlert(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("PUT", testAlertItem, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got models.UpdateAlertRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.EvaluationPeriods != "5" || got.DatapointsToAlarm != "3" {
			t.Fatalf("update body = %+v", got)
		}
		writeJSON(w, http.StatusOK, models.Alert{
			ID:                 101,
			Name:               got.Name,
			ResourceTypeName:   got.ResourceTypeName,
			MetricType:         got.MetricType,
			MetricName:         got.MetricName,
			Severity:           got.Severity,
			Threshold:          90,
			ComparisonOperator: ">=",
			State:              "OK",
			CreatedAt:          "2026-09-30T12:00:00Z",
		})
	})

	threshold := 90.0
	op := ">="
	got, err := newPostgresTestClient(t, ms).UpdateAlert(context.Background(), testAlertID, &models.UpdateAlertRequest{
		Name:               "high-cpu-vm1",
		ResourceTypeName:   "compute",
		MetricType:         "cpu",
		MetricName:         "cpu_utilization",
		Severity:           2,
		Threshold:          &threshold,
		ComparisonOperator: &op,
		EvaluationPeriods:  "5",
		DatapointsToAlarm:  "3",
	})
	if err != nil {
		t.Fatalf("UpdateAlert() error = %v", err)
	}
	if got.ID != 101 || got.Threshold != 90 {
		t.Fatalf("update = %+v", got)
	}
}

func TestDeleteAlert(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("DELETE", testAlertItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
	})
	if err := newPostgresTestClient(t, ms).DeleteAlert(context.Background(), testAlertID); err != nil {
		t.Fatalf("DeleteAlert() error = %v", err)
	}

	ms.SetErrorResponse("DELETE", testAlertItem, 404, "Not found")
	if err := newPostgresTestClient(t, ms).DeleteAlert(context.Background(), testAlertID); err != nil {
		t.Fatalf("DeleteAlert() 404 should succeed, got %v", err)
	}
}

func TestResolveAlertCreateParams(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", models.AlertStaticResourceTypesPath, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []string{"Virtual Machine"})
	})
	ms.AddHandler("GET", models.AlertStaticFunctionsPath, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []string{"Average", "Minimum", "Maximum"})
	})
	ms.AddHandler("GET", models.AlertStaticConditionsPath, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []string{"Greater than, equal to (>=)", "Equal to (==)"})
	})
	ms.AddHandler("GET", models.AlertStaticMetricTypesPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("resource_type") != "Virtual Machine" {
			t.Fatalf("resource_type query = %q", r.URL.Query().Get("resource_type"))
		}
		writeJSON(w, http.StatusOK, []string{"CPUAverage", "MemoryUsed"})
	})

	c := newPostgresTestClient(t, ms)
	resType, metType, metName, cond, err := c.ResolveAlertCreateParams(
		context.Background(),
		"Virtual Machine",
		"CPUAverage",
		"Average",
		">=",
	)
	if err != nil {
		t.Fatalf("ResolveAlertCreateParams() error = %v", err)
	}
	if resType != "Virtual Machine" || metType != "CPUAverage" || metName != "Average" {
		t.Fatalf("got %q %q %q", resType, metType, metName)
	}
	if cond != "Greater than, equal to (>=)" {
		t.Fatalf("condition = %q", cond)
	}

	if _, _, _, _, err := c.ResolveAlertCreateParams(context.Background(), "Virtual Machine", "NoSuchMetric", "Average", ">="); err == nil {
		t.Fatal("expected metric_type error")
	}
}

func TestResolveAlertResourcesForType(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []map[string]interface{}{
			{
				"id": "vm-active", "provider_instance_id": "prov-1", "instance_name": "app-vm-1",
				"status": "Active", "floating_ip": "10.0.0.10",
			},
			{
				"id": "vm-fixed", "instance_name": "app-vm-2", "status": "Active",
				"floating_ip": nil, "fixed_ips": []string{"10.0.0.20"},
			},
			{
				// Active but no IP anywhere: must be skipped (API requires resource_ip).
				"id": "vm-noip", "instance_name": "no-ip-vm", "status": "Active",
				"floating_ip": nil, "fixed_ips": []string{},
			},
			{
				"id": "vm-stopped", "instance_name": "old-vm", "status": "Stopped", "floating_ip": "10.0.0.11",
			},
		})
	})

	c := newPostgresTestClient(t, ms)
	got, err := c.ResolveAlertResourcesForType(context.Background(), "Virtual Machine")
	if err != nil {
		t.Fatalf("ResolveAlertResourcesForType() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (Active VMs with an IP)", len(got))
	}
	if got[0].ResourceID != "vm-active" || got[0].ResourceName != "app-vm-1" || got[0].ResourceIP != "10.0.0.10" {
		t.Fatalf("got[0] = %+v", got[0])
	}
	if got[1].ResourceID != "vm-fixed" || got[1].ResourceIP != "10.0.0.20" || got[1].ProviderInstanceID != "vm-fixed" {
		t.Fatalf("got[1] = %+v", got[1])
	}

	if _, err := c.ResolveAlertResourcesForType(context.Background(), "Other"); err == nil {
		t.Fatal("expected error for unsupported resource type")
	}
}

func TestMatchAlertConditionSymbol(t *testing.T) {
	conds := []string{"Greater than, equal to (>=)", "Equal to (==)"}
	if got := matchAlertConditionSymbol(conds, ">="); got != conds[0] {
		t.Fatalf("got %q", got)
	}
	if got := matchAlertConditionSymbol(conds, "=="); got != conds[1] {
		t.Fatalf("got %q", got)
	}
	if got := matchAlertConditionSymbol(conds, ">"); got != "" {
		t.Fatalf("unexpected %q", got)
	}
}
