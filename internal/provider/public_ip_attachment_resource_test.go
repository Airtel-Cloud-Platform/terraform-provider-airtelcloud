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
