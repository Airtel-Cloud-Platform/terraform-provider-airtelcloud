package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComputeSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccComputeSnapshotResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_compute_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_compute_snapshot.test", "status"),
					resource.TestCheckResourceAttrSet("airtelcloud_compute_snapshot.test", "created"),
					resource.TestCheckResourceAttr("airtelcloud_compute_snapshot.test", "compute_id", "test-compute-id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_compute_snapshot.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"compute_id",
					"timeouts",
				},
			},
		},
	})
}

func testAccComputeSnapshotResourceConfig() string {
	return `
resource "airtelcloud_compute_snapshot" "test" {
  compute_id = "test-compute-id"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`
}
