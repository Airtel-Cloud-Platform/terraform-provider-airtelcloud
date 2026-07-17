package tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr("airtelcloud_compute_snapshot.test", "snapshot_name", "tf-acc-snap"),
				),
			},
			// ImportState testing using the "<compute_id>:<snapshot_uuid>" form.
			{
				ResourceName:      "airtelcloud_compute_snapshot.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["airtelcloud_compute_snapshot.test"]
					return rs.Primary.Attributes["compute_id"] + ":" + rs.Primary.ID, nil
				},
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
			},
		},
	})
}

// TestAccComputeSnapshotResourceByName exercises the compute_name path, where the
// provider resolves the name to compute_id before creating the snapshot.
func TestAccComputeSnapshotResourceByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeSnapshotResourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_compute_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_compute_snapshot.test", "compute_id"),
					resource.TestCheckResourceAttr("airtelcloud_compute_snapshot.test", "compute_name", "test-instance"),
					resource.TestCheckResourceAttr("airtelcloud_compute_snapshot.test", "snapshot_name", "tf-acc-snap-byname"),
				),
			},
		},
	})
}

// TestAccComputeSnapshotResourceValidation asserts the exactly-one-of rule between
// compute_id and compute_name enforced by ValidateConfig.
func TestAccComputeSnapshotResourceValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "airtelcloud_compute_snapshot" "test" {
  compute_id    = "test-compute-id"
  compute_name  = "test-instance"
  snapshot_name = "tf-acc-snap"
}
`,
				ExpectError: regexp.MustCompile("Only one of compute_id or compute_name"),
			},
			{
				Config: `
resource "airtelcloud_compute_snapshot" "test" {
  snapshot_name = "tf-acc-snap"
}
`,
				ExpectError: regexp.MustCompile("One of compute_id or compute_name must be specified"),
			},
		},
	})
}

func testAccComputeSnapshotResourceConfig() string {
	return `
resource "airtelcloud_compute_snapshot" "test" {
  compute_id    = "test-compute-id"
  snapshot_name = "tf-acc-snap"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`
}

func testAccComputeSnapshotResourceConfigByName() string {
	return `
resource "airtelcloud_compute_snapshot" "test" {
  compute_name  = "test-instance"
  snapshot_name = "tf-acc-snap-byname"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`
}
