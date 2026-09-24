package provider

import (
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
)

func TestPublicIPAttachmentResourceTypes(t *testing.T) {
	for _, in := range []string{"vm", "lb", "baremetal"} {
		if _, err := client.NormalizePublicIPResourceType(in); err != nil {
			t.Fatalf("resource_type %q should be accepted: %v", in, err)
		}
	}

	if _, err := client.NormalizePublicIPResourceType("bucket"); err == nil {
		t.Fatal("resource_type \"bucket\" should be rejected")
	}
}

func TestValidatePublicIPAttachmentTargetVIP(t *testing.T) {
	t.Run("lb requires target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeLB, false, false, false); err == nil {
			t.Fatal("expected error when lb has no target_vip")
		}
	})

	t.Run("lb allows unknown target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeLB, true, true, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("lb accepts configured target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeLB, true, false, true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("vm rejects target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeVM, true, false, true); err == nil {
			t.Fatal("expected error when vm sets target_vip")
		}
	})

	t.Run("baremetal rejects target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeBaremetal, true, false, true); err == nil {
			t.Fatal("expected error when baremetal sets target_vip")
		}
	})

	t.Run("vm allows omitted target_vip", func(t *testing.T) {
		if err := validatePublicIPAttachmentTargetVIP(client.PublicIPResourceTypeVM, false, false, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPublicIPAttachmentRequestedVIP(t *testing.T) {
	got, err := publicIPAttachmentRequestedVIP("lb", "10.1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.1.2.3" {
		t.Fatalf("got %q, want 10.1.2.3", got)
	}

	if _, err := publicIPAttachmentRequestedVIP("lb", ""); err == nil {
		t.Fatal("expected error when lb has empty target_vip")
	}

	got, err = publicIPAttachmentRequestedVIP("vm", "10.1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("vm requested VIP = %q, want empty", got)
	}

	got, err = publicIPAttachmentRequestedVIP("baremetal", "10.1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("baremetal requested VIP = %q, want empty", got)
	}
}
