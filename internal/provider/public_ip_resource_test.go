package provider

import (
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestGetPublicIPAllocatedTimePrefersAllocatedTimeThenCreatedAt(t *testing.T) {
	ip := &models.PublicIP{AllocatedTime: "2026-08-10T01:02:03Z", CreatedAt: "2026-08-10T01:01:01Z"}
	if got := getPublicIPAllocatedTime(ip); got != "2026-08-10T01:02:03Z" {
		t.Fatalf("getPublicIPAllocatedTime() = %q, want %q", got, "2026-08-10T01:02:03Z")
	}

	ip = &models.PublicIP{CreatedAt: "2026-08-10T01:01:01Z"}
	if got := getPublicIPAllocatedTime(ip); got != "2026-08-10T01:01:01Z" {
		t.Fatalf("getPublicIPAllocatedTime() = %q, want %q", got, "2026-08-10T01:01:01Z")
	}
}

func TestGetPublicIPAZNamePrefersAZNameThenAZ(t *testing.T) {
	ip := &models.PublicIP{AZName: "N1", AZ: "north-1"}
	if got := getPublicIPAZName(ip); got != "N1" {
		t.Fatalf("getPublicIPAZName() = %q, want %q", got, "N1")
	}

	ip = &models.PublicIP{AZ: "north-1"}
	if got := getPublicIPAZName(ip); got != "north-1" {
		t.Fatalf("getPublicIPAZName() = %q, want %q", got, "north-1")
	}
}
