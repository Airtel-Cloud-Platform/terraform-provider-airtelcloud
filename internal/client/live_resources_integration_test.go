//go:build live

package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

type liveTestConfig struct {
	APIEndpoint  string
	APIKey       string
	APISecret    string
	Region       string
	Organization string
	ProjectName  string
	AZ           string
}

func getLiveTestConfig(t *testing.T) *liveTestConfig {
	t.Helper()
	cfg := &liveTestConfig{
		APIEndpoint:  os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:       os.Getenv("AIRTEL_API_KEY"),
		APISecret:    os.Getenv("AIRTEL_API_SECRET"),
		Region:       os.Getenv("AIRTEL_REGION"),
		Organization: os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:  os.Getenv("AIRTEL_PROJECT_NAME"),
		AZ:           os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE"),
	}
	if cfg.APIEndpoint == "" {
		t.Skip("AIRTEL_API_ENDPOINT not set")
	}
	if cfg.APIKey == "" {
		t.Skip("AIRTEL_API_KEY not set")
	}
	if cfg.APISecret == "" {
		t.Skip("AIRTEL_API_SECRET not set")
	}
	if cfg.Organization == "" {
		t.Skip("AIRTEL_ORGANIZATION not set")
	}
	if cfg.ProjectName == "" {
		t.Skip("AIRTEL_PROJECT_NAME not set")
	}
	if cfg.Region == "" {
		cfg.Region = "south"
	}
	if cfg.AZ == "" {
		cfg.AZ = "S1"
	}
	return cfg
}

func newLiveClient(t *testing.T, cfg *liveTestConfig) *Client {
	t.Helper()
	c, err := NewClient(cfg.APIEndpoint, cfg.APIKey, cfg.APISecret, cfg.Region, cfg.Organization, cfg.ProjectName, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func liveEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func liveBoolPtr(v bool) *bool          { return &v }
func liveFloat64Ptr(v float64) *float64 { return &v }

// TestLiveMonitoringDashboardCRUD creates, reads, updates, and deletes a dashboard.
// Requires monitoring-admin role. Falls back to tenant read when admin returns 403.
func TestLiveMonitoringDashboardCRUD(t *testing.T) {
	cfg := getLiveTestConfig(t)
	c := newLiveClient(t, cfg)
	ctx := context.Background()

	name := fmt.Sprintf("tft-dash-%d", time.Now().Unix()%100000)
	active := true
	desc := "terraform live test dashboard"
	req := &models.CreateMonitoringDashboardRequest{
		Name:         name,
		ResourceType: "compute",
		Active:       &active,
		Description:  &desc,
		Panels: []models.MonitoringDashboardPanel{
			{
				Title:        "CPU",
				ChartType:    "line",
				X:            0,
				Y:            0,
				W:            12,
				H:            4,
				DisplayOrder: 0,
				Queries: []models.MonitoringDashboardQuery{
					{
						QueryTemplate: `up{job="node"}`,
						QueryType:     "prometheus",
						RefID:         "A",
					},
				},
			},
		},
	}

	t.Logf("Creating monitoring dashboard %s", name)
	created, err := c.CreateMonitoringDashboard(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("monitoring-admin create forbidden for this principal: %v", err)
		}
		t.Fatalf("CreateMonitoringDashboard: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created dashboard id empty")
	}
	t.Logf("created id=%s", created.ID)

	defer func() {
		t.Logf("Deleting monitoring dashboard %s", created.ID)
		if err := c.DeleteMonitoringDashboard(ctx, created.ID); err != nil {
			t.Errorf("DeleteMonitoringDashboard: %v", err)
		}
	}()

	got, err := c.GetMonitoringDashboard(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetMonitoringDashboard: %v", err)
	}
	if got.Name != name {
		t.Fatalf("name = %q, want %q", got.Name, name)
	}

	updDesc := "updated by live test"
	updated, err := c.UpdateMonitoringDashboard(ctx, created.ID, &models.UpdateMonitoringDashboardRequest{
		Name:         got.Name,
		ResourceType: got.ResourceType,
		Active:       &got.Active,
		Description:  &updDesc,
		Panels:       got.Panels,
	})
	if err != nil {
		t.Fatalf("UpdateMonitoringDashboard: %v", err)
	}
	t.Logf("updated description=%v", updated.Description)
}

// TestLiveMonitoringProviderList lists providers via admin API, or skips on 403.
func TestLiveMonitoringProviderList(t *testing.T) {
	cfg := getLiveTestConfig(t)
	c := newLiveClient(t, cfg)
	ctx := context.Background()

	list, err := c.ListMonitoringProviders(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("monitoring-admin list forbidden for this principal: %v", err)
		}
		t.Fatalf("ListMonitoringProviders: %v", err)
	}
	t.Logf("providers count=%d", len(list))
	if len(list) > 0 {
		t.Logf("first provider id=%s name=%s type=%s", list[0].ID, list[0].Name, list[0].ProviderType)
	}

	dashboards, err := c.ListMonitoringDashboards(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skipf("monitoring-admin dashboards list forbidden: %v", err)
		}
		t.Fatalf("ListMonitoringDashboards: %v", err)
	}
	t.Logf("dashboards count=%d", len(dashboards))
}

// TestLiveMonitoringTenantRead hits non-admin monitoring endpoints available to project users.
func TestLiveMonitoringTenantRead(t *testing.T) {
	cfg := getLiveTestConfig(t)
	c := newLiveClient(t, cfg)
	ctx := context.Background()

	var providers wrapStatusData
	if err := c.Get(ctx, "/api/v1/monitoring/providers", &providers); err != nil {
		t.Fatalf("GET /api/v1/monitoring/providers: %v", err)
	}
	t.Logf("tenant providers status=%s count=%d", providers.Status, len(providers.Data))

	path := fmt.Sprintf("/api/monitoring/v2/domain/%s/project/%s/dashboards/resource/compute", cfg.Organization, cfg.ProjectName)
	var dash models.MonitoringDashboard
	if err := c.Get(ctx, path, &dash); err != nil {
		t.Fatalf("GET tenant dashboard by resource_type: %v", err)
	}
	t.Logf("tenant dashboard id=%s name=%s panels=%d", dash.ID, dash.Name, len(dash.Panels))
}

type wrapStatusData struct {
	Status string                      `json:"status"`
	Data   []models.MonitoringProvider `json:"data"`
}

// TestLiveAlertCRUD creates, reads, updates, and deletes an alert via v2.1 API.
func TestLiveAlertCRUD(t *testing.T) {
	cfg := getLiveTestConfig(t)
	c := newLiveClient(t, cfg)
	ctx := context.Background()

	name := fmt.Sprintf("tft-alert-%d", time.Now().Unix()%100000)
	eval := "5"
	points := "3"
	desc := "terraform live test alert"
	resourceID := liveEnvOr("AIRTEL_TEST_ALERT_RESOURCE_ID", "1a4c1369-2e92-4463-8e5c-a7a1139d92f8")
	req := &models.CreateAlertRequest{
		Name:               name,
		Description:        &desc,
		Organization:       cfg.Organization,
		Project:            cfg.ProjectName,
		ResourceTypeName:   liveEnvOr("AIRTEL_TEST_ALERT_RESOURCE_TYPE", "Virtual Machine"),
		MetricType:         liveEnvOr("AIRTEL_TEST_ALERT_METRIC_TYPE", "CPUAverage"),
		MetricName:         liveEnvOr("AIRTEL_TEST_ALERT_METRIC_NAME", "Average"),
		Threshold:          80,
		ComparisonOperator: liveEnvOr("AIRTEL_TEST_ALERT_OPERATOR", "Greater than, equal to (>=)"),
		EvaluationPeriods:  &eval,
		DatapointsToAlarm:  &points,
		Enabled:            true,
		Severity:           2,
		AlertEmails:        []string{liveEnvOr("AIRTEL_TEST_ALERT_EMAIL", "noreply@example.com")},
		ResourcesInfo: []models.ResourceInfo{
			{
				ResourceID:         resourceID,
				ResourceName:       liveEnvOr("AIRTEL_TEST_ALERT_RESOURCE_NAME", "dbcorptest"),
				ResourceIP:         liveEnvOr("AIRTEL_TEST_ALERT_RESOURCE_IP", "10.29.16.181"),
				ProviderInstanceID: liveEnvOr("AIRTEL_TEST_ALERT_PROVIDER_INSTANCE_ID", resourceID),
			},
		},
	}

	t.Logf("Creating alert %s", name)
	created, err := c.CreateAlert(ctx, req)
	if err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}
	if created.UUIDID == "" {
		t.Fatal("CreateAlert returned empty UUIDID")
	}
	id := created.UUIDID
	t.Logf("created integer_id=%d uuid_id=%s state=%s", created.ID, id, created.State)

	defer func() {
		t.Logf("Deleting alert %s", id)
		if err := c.DeleteAlert(ctx, id); err != nil {
			t.Errorf("DeleteAlert: %v", err)
		}
	}()

	got, err := c.GetAlert(ctx, id)
	if err != nil {
		t.Fatalf("GetAlert: %v", err)
	}
	t.Logf("get name=%s metric=%s/%s", got.Name, got.MetricType, got.MetricName)

	updDesc := "updated by live test"
	org := cfg.Organization
	proj := cfg.ProjectName
	op := req.ComparisonOperator
	updated, err := c.UpdateAlert(ctx, id, &models.UpdateAlertRequest{
		Name:               name,
		Description:        &updDesc,
		Organization:       &org,
		Project:            &proj,
		ResourceTypeName:   req.ResourceTypeName,
		MetricType:         req.MetricType,
		MetricName:         req.MetricName,
		Threshold:          liveFloat64Ptr(75),
		ComparisonOperator: &op,
		EvaluationPeriods:  "5",
		DatapointsToAlarm:  "3",
		Enabled:            liveBoolPtr(true),
		Severity:           2,
		AlertEmails:        req.AlertEmails,
	})
	if err != nil {
		t.Fatalf("UpdateAlert: %v", err)
	}
	t.Logf("updated threshold=%v uuid_kept=%s", updated.Threshold, updated.UUIDID)
}
