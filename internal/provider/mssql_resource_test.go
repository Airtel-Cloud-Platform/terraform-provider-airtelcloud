package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestMSSQLStorageSizeAllowed(t *testing.T) {
	if !models.MSSQLStorageSizeAllowed(200) || !models.MSSQLStorageSizeAllowed(210) || !models.MSSQLStorageSizeAllowed(2040) {
		t.Fatal("expected sizes on the 10 GiB grid to be allowed")
	}
	for _, size := range []int64{199, 201, 2048, 2050} {
		if models.MSSQLStorageSizeAllowed(size) {
			t.Fatalf("size %d should be rejected", size)
		}
	}
}

func TestMSSQLSchemaNetworkAccess(t *testing.T) {
	t.Parallel()

	var resp resource.SchemaResponse
	(&MSSQLResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if attr, ok := resp.Schema.Attributes["vpc"].(schema.StringAttribute); !ok || !attr.IsOptional() {
		t.Fatal("vpc must be optional")
	}
	if attr, ok := resp.Schema.Attributes["subnet"].(schema.StringAttribute); !ok || !attr.IsOptional() {
		t.Fatal("subnet must be optional")
	}
	if attr, ok := resp.Schema.Attributes["custom_cidr"].(schema.ListAttribute); !ok || !attr.IsOptional() {
		t.Fatal("custom_cidr must be optional")
	}
	if _, exists := resp.Schema.Attributes["security_group"]; exists {
		t.Fatal("security_group must not be a Terraform attribute")
	}
}
