package provider

import (
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestProtectionPlanAPIName(t *testing.T) {
	tests := []struct {
		name, period, want string
	}{
		{"test", "weekly", "test-weekly"},
		{"test", "daily", "test-daily"},
		{"test-weekly", "weekly", "test-weekly"},
		{"test-daily", "daily", "test-daily"},
		{"TEST-WEEKLY", "weekly", "TEST-WEEKLY"},
		{"  teer  ", " weekly ", "teer-weekly"},
	}
	for _, tt := range tests {
		got := protectionPlanAPIName(tt.name, tt.period)
		if got != tt.want {
			t.Fatalf("protectionPlanAPIName(%q, %q) = %q, want %q", tt.name, tt.period, got, tt.want)
		}
	}
}

func TestProtectionPlanRecurrenceSeconds(t *testing.T) {
	if got := protectionPlanRecurrenceSeconds(1, "daily"); got != 86400 {
		t.Fatalf("daily 1 = %d, want 86400", got)
	}
	if got := protectionPlanRecurrenceSeconds(1, "weekly"); got != 604800 {
		t.Fatalf("weekly 1 = %d, want 604800", got)
	}
	if got := protectionPlanRecurrenceSeconds(21, "daily"); got != 1814400 {
		t.Fatalf("daily 21 = %d, want 1814400", got)
	}
}

func TestProtectionPlanRangeError(t *testing.T) {
	if !protectionPlanOutOfRange(31, 1, 30) || !protectionPlanOutOfRange(0, 1, 30) || protectionPlanOutOfRange(30, 1, 30) {
		t.Fatal("recurrence bounds")
	}
	if !protectionPlanOutOfRange(366, 1, 365) || protectionPlanOutOfRange(365, 1, 365) {
		t.Fatal("retention bounds")
	}
	msg := protectionPlanRangeError("retention", 400, 1, 365, "days")
	if !strings.Contains(msg, "400") || !strings.Contains(msg, "365") || !strings.Contains(msg, "not called") {
		t.Fatalf("error message = %q", msg)
	}
}

func TestProtectionPlanVMLocation(t *testing.T) {
	az, subnet := protectionPlanVMLocation(&models.Compute{AZName: "S1", NetworkID: "net-1"})
	if az != "S1" || subnet != "net-1" {
		t.Fatalf("got az=%q subnet=%q", az, subnet)
	}
	az, subnet = protectionPlanVMLocation(&models.Compute{AvailabilityZone: "S2", SubnetID: "sub-2"})
	if az != "S2" || subnet != "sub-2" {
		t.Fatalf("fallback got az=%q subnet=%q", az, subnet)
	}
}
