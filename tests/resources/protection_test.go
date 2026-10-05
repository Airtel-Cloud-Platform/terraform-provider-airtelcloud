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
			{
				Config: testAccProtectionPlanResourceConfig("test-plan"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_protection_plan.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "name", "test-plan"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "vm_name", "test-vm"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "recurrence", "1"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "recurrence_period", "daily"),
					resource.TestCheckResourceAttr("airtelcloud_protection_plan.test", "retention", "1"),
				),
			},
			{
				ResourceName:            "airtelcloud_protection_plan.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name", "description", "recurrence", "recurrence_period", "retention", "vm_name", "selector_value", "subnet_id"},
			},
		},
	})
}

func testAccProtectionPlanResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_protection_plan" "test" {
  name              = %[1]q
  description       = "Test plan"
  vm_name           = "test-vm"
  recurrence        = 1
  recurrence_period = "daily"
  retention         = 1
}
`, name)
}

func TestAccProtectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_protection.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "vm_name", "test-vm"),
					resource.TestCheckResourceAttrSet("airtelcloud_protection.test", "compute_id"),
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "protection_plan", "daily-plan"),
					resource.TestCheckResourceAttrSet("airtelcloud_protection.test", "status"),
				),
			},
			{
				ResourceName:      "airtelcloud_protection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"vm_name",
					"enable_scheduler",
					"start_date",
					"end_date",
					"start_time",
					"weekday",
				},
			},
			{
				Config: testAccProtectionResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_protection.test", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccProtectionResourceConfig() string {
	return `
resource "airtelcloud_protection" "test" {
  description      = "Test protection"
  vm_name          = "test-vm"
  protection_plan  = "daily-plan"
  start_date       = "10/06/2026"
  start_time       = "02:00 AM"
}
`
}

func testAccProtectionResourceConfigUpdated() string {
	return `
resource "airtelcloud_protection" "test" {
  description      = "Updated description"
  vm_name          = "test-vm"
  protection_plan  = "daily-plan"
  start_date       = "10/06/2026"
  start_time       = "02:00 AM"
}
`
}
