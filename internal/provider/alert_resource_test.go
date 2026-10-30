package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAlertSchema(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	(&AlertResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	required := []string{
		"name", "emails", "resource_type", "severity",
		"metric_type", "function", "condition", "value",
	}
	for _, name := range required {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %q", name)
		}
		if !attr.IsRequired() {
			t.Fatalf("%s must be required", name)
		}
	}

	if attr, ok := resp.Schema.Attributes["description"]; !ok || !attr.IsOptional() {
		t.Fatal("description must be optional")
	}

	for _, name := range []string{"id", "state", "created_at", "updated_at"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok || !attr.IsComputed() {
			t.Fatalf("%s must be computed", name)
		}
	}

	for _, name := range []string{
		"resources", "resources_info", "evaluation_periods", "datapoints_to_alarm",
		"enabled", "numeric_id", "alert_emails", "metric_name", "comparison_operator",
		"threshold", "resource_type_name",
	} {
		if _, ok := resp.Schema.Attributes[name]; ok {
			t.Fatalf("%s should not be in schema", name)
		}
	}
}

func TestAlertValidateConfigValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   types.Float64
		wantErr bool
	}{
		{name: "valid mid", value: types.Float64Value(50)},
		{name: "valid zero", value: types.Float64Value(0)},
		{name: "valid hundred", value: types.Float64Value(100)},
		{name: "below", value: types.Float64Value(-1), wantErr: true},
		{name: "above", value: types.Float64Value(100.1), wantErr: true},
		{name: "unknown skips", value: types.Float64Unknown()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.IsNull() || tt.value.IsUnknown() {
				if tt.wantErr {
					t.Fatal("unknown/null should not error")
				}
				return
			}
			v := tt.value.ValueFloat64()
			gotErr := v < 0 || v > 100
			if gotErr != tt.wantErr {
				t.Fatalf("value %v wantErr=%v gotErr=%v", v, tt.wantErr, gotErr)
			}
		})
	}
}
