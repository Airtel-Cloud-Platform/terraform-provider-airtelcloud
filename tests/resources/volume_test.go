package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVolumeResourceConfig("test-volume", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_name", "test-volume"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_size", "10"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "status"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "provider_volume_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing (increase size)
			{
				Config: testAccVolumeResourceConfig("test-volume-updated", 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_name", "test-volume-updated"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_size", "20"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeResourceConfig(name string, size int) string {
	return fmt.Sprintf(`
resource "airtelcloud_volume" "test" {
  volume_name  = %[1]q
  volume_size  = %[2]d
}
`, name, size)
}

func TestAccVolumeResourceFull(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceFullConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_name", "test-full-volume"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_size", "50"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "availability_zone", "S1"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "is_encrypted", "true"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "bootable", "false"),
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "enable_backup", "false"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "status"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "provider_volume_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"compute_id",
					"subnet_id",
				},
			},
		},
	})
}

func testAccVolumeResourceFullConfig() string {
	return `
resource "airtelcloud_vpc" "test" {
  name       = "test-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "airtelcloud_subnet" "test" {
  name              = "test-subnet"
  vpc_id            = airtelcloud_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "S1"
}

resource "airtelcloud_volume" "test" {
  volume_name       = "test-full-volume"
  volume_size       = 50
  availability_zone = "S1"
  vpc_id            = airtelcloud_vpc.test.id
  subnet_id         = airtelcloud_subnet.test.id
  is_encrypted      = true
  bootable          = false
  enable_backup     = false
}
`
}

func TestAccVolumeResourceWithAttachment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create volume and VM, then test attachment
			{
				Config: testAccVolumeWithComputeConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_volume.test", "volume_name", "test-attached-volume"),
					resource.TestCheckResourceAttr("airtelcloud_vm.test", "instance_name", "test-compute-for-volume"),
					resource.TestCheckResourceAttrSet("airtelcloud_volume.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_vm.test", "id"),
				),
			},
		},
	})
}

func testAccVolumeWithComputeConfig() string {
	return `
# Create dependencies for compute instance
resource "airtelcloud_vpc" "test" {
  name       = "test-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "airtelcloud_subnet" "test" {
  name              = "test-subnet"
  vpc_id            = airtelcloud_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "south-1a"
}

# Create compute instance
resource "airtelcloud_vm" "test" {
  instance_name    = "test-compute-for-volume"
  compute_type     = "t2.micro"
  image_id         = "ubuntu-20.04"
  network_id       = airtelcloud_vpc.test.id
  subnet_id        = airtelcloud_subnet.test.id
  disk_size        = 20
}

# Create volume attached to compute
resource "airtelcloud_volume" "test" {
  volume_name       = "test-attached-volume"
  volume_size       = 10
  vpc_id            = airtelcloud_vpc.test.id
  subnet_id         = airtelcloud_subnet.test.id
  compute_id        = airtelcloud_vm.test.id
  is_encrypted      = true
}
`
}
