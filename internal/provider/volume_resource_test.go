package provider

import (
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBothKnownSet(t *testing.T) {
	t.Parallel()

	if bothKnownSet(types.StringUnknown(), types.StringUnknown()) {
		t.Fatal("unknown pair must not fail XOR")
	}
	if bothKnownSet(types.StringValue("a"), types.StringUnknown()) {
		t.Fatal("unknown sibling must not fail XOR")
	}
	if bothKnownSet(types.StringValue("a"), types.StringNull()) {
		t.Fatal("id + omitted name must be allowed")
	}
	if !bothKnownSet(types.StringValue("a"), types.StringValue("b")) {
		t.Fatal("two known values must fail XOR")
	}
}

func TestParseVolumeLabels(t *testing.T) {
	t.Parallel()

	got := parseVolumeLabels([]byte(`{"labels": ["tft"], "default_labels": []}`))
	if len(got) != 1 || got[0] != "tft" {
		t.Fatalf("parse=%v", got)
	}
}

func TestMatchVolumeType(t *testing.T) {
	t.Parallel()

	typesIn := []models.VolumeType{
		{ID: 1, Name: "s1_wkld_ntp02_4iops_backend", Label: "Standard Performance", IsActive: true},
		{ID: 2, Name: "high-perf", Label: "High Performance", IsActive: true},
		{ID: 3, Name: "old", Label: "Deprecated", IsActive: false},
	}

	got, _, ok := matchVolumeType(typesIn, "Standard Performance")
	if !ok || got.ID != 1 {
		t.Fatalf("label match failed: ok=%v id=%d", ok, got.ID)
	}
	got, _, ok = matchVolumeType(typesIn, "s1_wkld_ntp02_4iops_backend")
	if !ok || got.ID != 1 {
		t.Fatalf("name match failed: ok=%v id=%d", ok, got.ID)
	}
	if _, _, ok := matchVolumeType(typesIn, "Deprecated"); ok {
		t.Fatal("inactive type must not match")
	}
}
