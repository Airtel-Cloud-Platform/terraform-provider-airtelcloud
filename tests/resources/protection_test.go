package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProtectionPlanResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProtectionPlanResourceConfig("test-plan"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_protection_plan.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "name", "test-plan"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "subnet_id", "35df162d-5211-4d58-84ed-6a499626949c"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_protection_plan.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProtectionPlanResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_protection_plan" "test" {
  name           = %[1]q
  description    = "Test plan"
  retention      = 1
  retention_unit = "DAYS"
  recurrence     = 86400
  selector_key   = "AZ"
  selector_value = "S1"
  subnet_id      = "35df162d-5211-4d58-84ed-6a499626949c"
}
`, name)
}

func TestAccProtectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProtectionResourceConfig("test-protection"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_protection.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "name", "test-protection"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "compute_id", "test-compute-id"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "protection_plan", "daily-plan"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "enable_scheduler", "true"),
					resource.TestCheckResourceAttrSet("airtelcloud_protection.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_protection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"enable_scheduler",
					"start_date",
					"end_date",
					"start_time",
				},
			},
			// Update testing
			{
				Config: testAccProtectionResourceConfigUpdated("updated-protection"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "name", "updated-protection"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccProtectionResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_protection" "test" {
  name             = %[1]q
  description      = "Test protection"
  compute_id       = "test-compute-id"
  protection_plan  = "daily-plan"
  enable_scheduler = "true"
}
`, name)
}

func testAccProtectionResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_protection" "test" {
  name             = %[1]q
  description      = "Updated description"
  compute_id       = "test-compute-id"
  protection_plan  = "daily-plan"
  enable_scheduler = "true"
}
`, name)
}
