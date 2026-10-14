package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestMonitoringDashboardResourceSchema(t *testing.T) {
	r := NewMonitoringDashboardResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	required := []string{"name", "resource_type"}
	for _, name := range required {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %q", name)
		}
		if !attr.IsRequired() {
			t.Fatalf("%q should be required", name)
		}
	}

	panels, ok := resp.Schema.Attributes["panels"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("panels must be ListNestedAttribute")
	}
	if _, ok := panels.NestedObject.Attributes["queries"].(schema.ListNestedAttribute); !ok {
		t.Fatal("panels.queries must be ListNestedAttribute")
	}
	chartType := panels.NestedObject.Attributes["chart_type"]
	if chartType == nil || !chartType.IsRequired() {
		t.Fatal("panels.chart_type must be required")
	}
}

func TestMonitoringProviderResourceSchema(t *testing.T) {
	r := NewMonitoringProviderResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	required := []string{
		"name", "resource_type", "dashboard_auth_url", "api_key_name",
		"api_token", "data_source_uid", "data_source_id",
	}
	for _, name := range required {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %q", name)
		}
		if !attr.IsRequired() {
			t.Fatalf("%q should be required", name)
		}
	}

	token := resp.Schema.Attributes["api_token"].(schema.StringAttribute)
	if !token.Sensitive {
		t.Fatal("api_token must be sensitive")
	}
}

func TestBuildMonitoringProviderUpdateOmitsUnchangedToken(t *testing.T) {
	plan := &MonitoringProviderResourceModel{
		Name:             types.StringValue("south-prometheus-compute"),
		ResourceType:     types.StringValue("compute"),
		DashboardAuthURL: types.StringValue("https://grafana.example.com"),
		APIKeyName:       types.StringValue("terraform-provider"),
		APIToken:         types.StringValue("same-token"),
		DataSourceUID:    types.StringValue("prom-uid-1"),
		DataSourceID:     types.StringValue("1"),
		Active:           types.BoolValue(true),
	}
	state := &MonitoringProviderResourceModel{
		APIToken: types.StringValue("same-token"),
	}

	req := buildMonitoringProviderUpdateRequest(plan, state)
	if req.APIToken != nil {
		t.Fatalf("expected api_token omitted when unchanged, got %v", *req.APIToken)
	}

	plan.APIToken = types.StringValue("rotated-token")
	req = buildMonitoringProviderUpdateRequest(plan, state)
	if req.APIToken == nil || *req.APIToken != "rotated-token" {
		t.Fatalf("expected rotated token, got %v", req.APIToken)
	}

	plan.APIToken = types.StringValue("")
	req = buildMonitoringProviderUpdateRequest(plan, state)
	if req.APIToken != nil {
		t.Fatalf("expected empty api_token omitted, got %v", *req.APIToken)
	}
}

func TestApplyMonitoringProviderPreservesToken(t *testing.T) {
	data := &MonitoringProviderResourceModel{
		APIToken: types.StringValue("keep-me"),
	}
	applyMonitoringProviderToState(data, &models.MonitoringProvider{
		ID:               "65f0b2c3d4e5f67890123456",
		Name:             "south-prometheus-compute",
		ResourceType:     "compute",
		ProviderType:     "prometheus",
		DashboardAuthURL: "https://grafana.example.com",
		APIKeyName:       "terraform-provider",
		DataSourceUID:    "prom-uid-1",
		DataSourceID:     "1",
		Active:           true,
	})
	// applyMonitoringProviderToState does not touch APIToken; callers must preserve it.
	if data.APIToken.ValueString() != "keep-me" {
		t.Fatalf("api_token overwritten: %q", data.APIToken.ValueString())
	}
	if data.ID.ValueString() != "65f0b2c3d4e5f67890123456" {
		t.Fatalf("id = %q", data.ID.ValueString())
	}
}

func TestMonitoringDashboardFlattenRoundTrip(t *testing.T) {
	desc := "CPU over time"
	created := "2026-09-30T12:00:00Z"
	order := int64(0)
	dashboard := &models.MonitoringDashboard{
		ID:           "65f0a1b2c3d4e5f678901234",
		Name:         "compute-default",
		ResourceType: "compute",
		Active:       true,
		Description:  &desc,
		CreatedAt:    &created,
		UpdatedAt:    &created,
		Panels: []models.MonitoringDashboardPanel{{
			Title:        "CPU Utilization",
			ChartType:    "line",
			Description:  &desc,
			X:            0,
			Y:            0,
			W:            12,
			H:            4,
			DisplayOrder: 0,
			Queries: []models.MonitoringDashboardQuery{{
				QueryTemplate: `avg(rate(cpu_usage{instance="{{instance}}"}[5m]))`,
				QueryType:     "prometheus",
				RefID:         "A",
				DisplayOrder:  &order,
				PlaceholderConfig: map[string]interface{}{
					"instance": map[string]interface{}{
						"source": "resource",
						"field":  "name",
					},
				},
			}},
		}},
	}

	var data MonitoringDashboardResourceModel
	diags := applyMonitoringDashboardToState(context.Background(), &data, dashboard)
	if diags.HasError() {
		t.Fatalf("apply diagnostics: %v", diags)
	}
	if data.ID.ValueString() != dashboard.ID {
		t.Fatalf("id = %q", data.ID.ValueString())
	}
	if data.Panels.IsNull() || data.Panels.IsUnknown() {
		t.Fatal("panels should be set")
	}

	req, diags := buildMonitoringDashboardCreateRequest(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("build diagnostics: %v", diags)
	}
	if req.Name != "compute-default" || len(req.Panels) != 1 || len(req.Panels[0].Queries) != 1 {
		t.Fatalf("create request = %+v", req)
	}
	if req.Panels[0].Queries[0].QueryType != "prometheus" {
		t.Fatalf("query_type = %q", req.Panels[0].Queries[0].QueryType)
	}
}
