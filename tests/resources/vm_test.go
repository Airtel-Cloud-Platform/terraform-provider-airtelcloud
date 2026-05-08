package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVMResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing — use flavor_name / image_name for API validation
			{
				Config: testAccVMResourceConfig("test-vm"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-vm"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "os_type", "linux"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "boot_from_volume", "true"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "disk_size", "20"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "status"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "flavor_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "image_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "vpc_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "private_ip"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"flavor_name",
					"image_name",
					"vpc_name",
					"subnet_name",
					"os_type",
					"boot_from_volume",
					"disk_size",
					"tags",
				},
			},
			// Update and Read testing — change instance_name and description
			{
				Config: testAccVMResourceConfigUpdated("test-vm-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-vm-updated"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "description", "Updated VM"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVMResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_vm" "test" {
  instance_name     = %[1]q
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
  description       = "Test VM"
}
`, name)
}

func testAccVMResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_vm" "test" {
  instance_name     = %[1]q
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
  description       = "Updated VM"
}
`, name)
}

func TestAccVMResourceWithFlavorID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMResourceConfigWithFlavorID("test-vm-flavor-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-vm-flavor-id"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "flavor_id", "1"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "image_id", "1"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "id"),
				),
			},
		},
	})
}

func testAccVMResourceConfigWithFlavorID(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_vm" "test" {
  instance_name     = %[1]q
  os_type           = "linux"
  flavor_id         = "1"
  image_id          = "1"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
}
`, name)
}

func TestAccVMResourceWithSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMResourceConfigWithSecurityGroup("test-vm-sg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-vm-sg"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "security_group_id"),
				),
			},
		},
	})
}

func testAccVMResourceConfigWithSecurityGroup(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_security_group" "test" {
  security_group_name = "test-sg-for-vm"
  availability_zone   = "S2"
}

resource "airtelcloud_vm" "test" {
  instance_name     = %[1]q
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  security_group_id = airtelcloud_security_group.test.id
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
}
`, name)
}

func TestAccVMResourceWithTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMResourceConfigWithTags("test-vm-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-vm-tags"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "tags.Environment", "test"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "tags.Team", "platform"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "id"),
				),
			},
		},
	})
}

func testAccVMResourceConfigWithTags(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_vm" "test" {
  instance_name     = %[1]q
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"

  tags = {
    Environment = "test"
    Team        = "platform"
  }
}
`, name)
}

func TestAccVMResourceMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create two VMs and verify both
			{
				Config: testAccVMResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_vm.web1", "instance_name", "web-server-1"),
					resource.TestCheckResourceAttr("airtelcloud_vm.web2", "instance_name", "web-server-2"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.web1", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.web2", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.web1", "private_ip"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.web2", "private_ip"),
				),
			},
		},
	})
}

func testAccVMResourceConfigMultiple() string {
	return `
resource "airtelcloud_vm" "web1" {
  instance_name     = "web-server-1"
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
  description       = "Web server 1"
}

resource "airtelcloud_vm" "web2" {
  instance_name     = "web-server-2"
  os_type           = "linux"
  flavor_name       = "t2.micro"
  image_name        = "ubuntu-20.04"
  vpc_id            = "029ac9b8-d93e-4691-a7cb-2f651c607cfe"
  subnet_id         = "35df162d-5211-4d58-84ed-6a499626949c"
  boot_from_volume  = true
  disk_size         = 20
  availability_zone = "S2"
  description       = "Web server 2"
}
`
}
