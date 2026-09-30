package provider

import (
	"testing"

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
