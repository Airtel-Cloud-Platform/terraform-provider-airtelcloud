package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVPCPeeringResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVPCPeeringResourceConfig("test-peering"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vpc_peering.test", "name", "test-peering"),
					resource.TestCheckResourceAttr("airtelcloud_vpc_peering.test", "az", "south-1a"),
					resource.TestCheckResourceAttr("airtelcloud_vpc_peering.test", "region", "south-1"),
					resource.TestCheckResourceAttrSet("airtelcloud_vpc_peering.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vpc_peering.test", "vpc_source_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vpc_peering.test", "vpc_target_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vpc_peering.test", "state"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_vpc_peering.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"vpc_source_name",
					"vpc_target_name",
				},
			},
		},
	})
}

func testAccVPCPeeringResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_vpc" "source" {
  name       = "test-vpc-source"
  cidr_block = "10.0.0.0/16"
}

resource "airtelcloud_vpc" "target" {
  name       = "test-vpc-target"
  cidr_block = "10.1.0.0/16"
}

resource "airtelcloud_vpc_peering" "test" {
  name          = %[1]q
  vpc_source_id = airtelcloud_vpc.source.id
  vpc_target_id = airtelcloud_vpc.target.id
  az            = "south-1a"
  region        = "south-1"
}
`, name)
}
