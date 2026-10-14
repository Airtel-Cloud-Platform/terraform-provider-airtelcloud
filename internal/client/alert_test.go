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
	testNotifID   = "55"
	testNotifBase = "/api/v1/notifications/"
	testNotifItem = testNotifBase + testNotifID
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

func TestCreateAlertNotification(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()
	ms.AddHandler("POST", testNotifBase, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got models.CreateAlertNotificationRequest
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got.Type != "email" || got.Target != "ops@example.com" || got.AlertID != 101 {
			t.Fatalf("create body = %+v", got)
		}
		writeJSON(w, http.StatusOK, models.AlertNotification{
			ID:        55,
			Type:      got.Type,
			Target:    got.Target,
			Enabled:   got.Enabled,
			AlertID:   got.AlertID,
			CreatedAt: "2026-09-30T12:00:00Z",
		})
	})

	got, err := newPostgresTestClient(t, ms).CreateAlertNotification(context.Background(), &models.CreateAlertNotificationRequest{
		Type:    "email",
		Target:  "ops@example.com",
		Enabled: true,
		AlertID: 101,
	})
	if err != nil {
		t.Fatalf("CreateAlertNotification() error = %v", err)
	}
	if got.ID != 55 || got.AlertID != 101 {
		t.Fatalf("notification = %+v", got)
	}
}

func TestGetUpdateDeleteAlertNotification(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.AddHandler("GET", testNotifItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, models.AlertNotification{
			ID: 55, Type: "email", Target: "ops@example.com", Enabled: true, AlertID: 101, CreatedAt: "2026-09-30T12:00:00Z",
		})
	})
	got, err := newPostgresTestClient(t, ms).GetAlertNotification(context.Background(), testNotifID)
	if err != nil {
		t.Fatalf("GetAlertNotification() error = %v", err)
	}
	if got.ID != 55 {
		t.Fatalf("get = %+v", got)
	}

	ms.AddHandler("PUT", testNotifItem, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req models.UpdateAlertNotificationRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		enabled := false
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		writeJSON(w, http.StatusOK, models.AlertNotification{
			ID: 55, Type: "email", Target: "ops@example.com", Enabled: enabled, AlertID: 101, CreatedAt: "2026-09-30T12:00:00Z",
		})
	})
	enabled := false
	updated, err := newPostgresTestClient(t, ms).UpdateAlertNotification(context.Background(), testNotifID, &models.UpdateAlertNotificationRequest{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UpdateAlertNotification() error = %v", err)
	}
	if updated.Enabled {
		t.Fatal("expected enabled=false")
	}

	ms.AddHandler("DELETE", testNotifItem, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
	})
	if err := newPostgresTestClient(t, ms).DeleteAlertNotification(context.Background(), testNotifID); err != nil {
		t.Fatalf("DeleteAlertNotification() error = %v", err)
	}
}
