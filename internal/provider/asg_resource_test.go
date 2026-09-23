package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestValidateRulePairs(t *testing.T) {
	tests := []struct {
		name    string
		up      []models.ASGScalingRule
		down    []models.ASGScalingRule
		wantErr bool
	}{
		{
			name: "valid pair",
			up:   []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
			down: []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
		},
		{
			name:    "missing scale down",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 60}},
			wantErr: true,
		},
		{
			name:    "invalid target ordering",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			down:    []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			wantErr: true,
		},
		{
			name:    "aggregation mismatch",
			up:      []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "max", TargetValue: 60}},
			down:    []models.ASGScalingRule{{MetricType: "cpu", AggregationType: "avg", TargetValue: 30}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateRulePairs(&diags, tt.up, tt.down)
			if diags.HasError() != tt.wantErr {
				t.Fatalf("HasError() = %v, want %v; diagnostics: %v", diags.HasError(), tt.wantErr, diags)
			}
		})
	}
}

func TestSubnetAvailabilityZoneMismatch(t *testing.T) {
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "S2", "S1"); msg == "" {
		t.Fatal("expected mismatch error")
	}
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "S1", "S1"); msg != "" {
		t.Fatalf("unexpected mismatch: %s", msg)
	}
	if msg := subnetAvailabilityZoneMismatch("subnet1213", "", "S1"); msg != "" {
		t.Fatalf("empty subnet zone should skip: %s", msg)
	}
}

func TestApplyASGAvailabilityZone(t *testing.T) {
	t.Run("keeps configured zone and warns on mismatch", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringValue("S1")}

		diags := applyASGAvailabilityZone(data, "S2")

		if diags.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", diags)
		}
		if diags.WarningsCount() != 1 {
			t.Fatalf("warnings = %d, want 1", diags.WarningsCount())
		}
		if got := data.AvailabilityZone.ValueString(); got != "S1" {
			t.Fatalf("availability_zone = %q, want S1", got)
		}
	})

	t.Run("adopts API zone on import", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringNull()}

		diags := applyASGAvailabilityZone(data, "S2")

		if diags.WarningsCount() != 0 {
			t.Fatalf("warnings = %d, want 0", diags.WarningsCount())
		}
		if got := data.AvailabilityZone.ValueString(); got != "S2" {
			t.Fatalf("availability_zone = %q, want S2", got)
		}
	})

	t.Run("no diagnostics when zones match", func(t *testing.T) {
		data := &ASGResourceModel{AvailabilityZone: types.StringValue("S1")}

		if diags := applyASGAvailabilityZone(data, "S1"); len(diags) != 0 {
			t.Fatalf("diagnostics = %v, want none", diags)
		}
	})
}

func TestSettleASGComputed(t *testing.T) {
	data := &ASGResourceModel{
		KeypairID:          types.StringUnknown(),
		KeypairName:        types.StringUnknown(),
		OSType:             types.StringUnknown(),
		ImageID:            types.Int64Unknown(),
		SecurityGroupNames: types.ListUnknown(types.StringType),
		FlavorName:         types.StringValue("ccd.Large"),
	}

	settleASGComputed(data)

	if data.KeypairID.IsUnknown() || data.KeypairName.IsUnknown() {
		t.Fatalf("keypair values still unknown: id=%v name=%v", data.KeypairID, data.KeypairName)
	}
	if data.OSType.IsUnknown() || data.ImageID.IsUnknown() {
		t.Fatalf("image values still unknown: os_type=%v image_id=%v", data.OSType, data.ImageID)
	}
	if data.SecurityGroupNames.IsUnknown() {
		t.Fatalf("security_group_names still unknown")
	}
	if data.FlavorName.ValueString() != "ccd.Large" {
		t.Fatalf("flavor_name = %v, want ccd.Large", data.FlavorName)
	}
}

func TestImageOSFamily(t *testing.T) {
	if got := (models.Image{OSType: "linux", OS: "windows"}).OSFamily(); got != "linux" {
		t.Fatalf("OSType should win, got %q", got)
	}
	if got := (models.Image{OS: "windows"}).OSFamily(); got != "windows" {
		t.Fatalf("OS fallback = %q, want windows", got)
	}
}

func TestDefaultASGDiskSize(t *testing.T) {
	if got := defaultASGDiskSize("ubuntu"); got != 100 {
		t.Fatalf("linux default = %d, want 100", got)
	}
	if got := defaultASGDiskSize("windows"); got != 200 {
		t.Fatalf("windows default = %d, want 200", got)
	}
	if got := defaultASGDiskSize(""); got != 100 {
		t.Fatalf("empty OS default = %d, want 100", got)
	}
}
