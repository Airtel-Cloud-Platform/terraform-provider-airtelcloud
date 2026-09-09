package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPostgresResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPostgresResourceConfig("test-pg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_postgres.test", "cluster_name", "test-pg"),
					resource.TestCheckResourceAttr("airtelcloud_postgres.test", "version", "17"),
					resource.TestCheckResourceAttr("airtelcloud_postgres.test", "database_name", "terra"),
					resource.TestCheckResourceAttr("airtelcloud_postgres.test", "high_availability", "true"),
					resource.TestCheckResourceAttr("airtelcloud_postgres.test", "num_replicas", "1"),
					resource.TestCheckResourceAttrSet("airtelcloud_postgres.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_postgres.test", "status"),
					resource.TestCheckResourceAttrSet("airtelcloud_postgres.test", "topology"),
				),
			},
			{
				ResourceName:      "airtelcloud_postgres.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func testAccPostgresResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_postgres" "test" {
  cluster_name      = %[1]q
  version           = "17"
  database_name     = "terra"
  postgres_username = "admin123"
  password          = "T3rraform!Pass"
  compute_size      = "db.postgres.uhper.ccs.xlarge"
  storage_size      = 200
  availability_zone = "S1"
  high_availability = true
  num_replicas      = 1
}
`, name)
}
