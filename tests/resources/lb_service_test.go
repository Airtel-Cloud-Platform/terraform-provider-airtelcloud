package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLBServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLBServiceResourceConfig("test-lb"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_lb_service.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_lb_service.test", "name", "test-lb"),
					resource.TestCheckResourceAttr("airtelcloud_lb_service.test", "flavor_id", "1"),
					resource.TestCheckResourceAttr("airtelcloud_lb_service.test", "vpc_id", "vpc-test-123"),
					resource.TestCheckResourceAttr("airtelcloud_lb_service.test", "vpc_name", "test-vpc"),
					resource.TestCheckResourceAttr("airtelcloud_lb_service.test", "ha", "false"),
					resource.TestCheckResourceAttrSet("airtelcloud_lb_service.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_lb_service.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ha",
					"timeouts",
				},
			},
		},
	})
}

func testAccLBServiceResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_lb_service" "test" {
  name        = %[1]q
  description = "Test LB service"
  flavor_id   = 1
  network_id  = "subnet-test-456"
  vpc_id      = "vpc-test-123"
  vpc_name    = "test-vpc"
  ha          = false

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name)
}

func TestAccLBVipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLBVipResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_lb_vip.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_lb_vip.test", "lb_service_id"),
					resource.TestCheckResourceAttrSet("airtelcloud_lb_vip.test", "status"),
				),
			},
		},
	})
}

func testAccLBVipResourceConfig() string {
	return `
resource "airtelcloud_lb_service" "test" {
  name        = "test-lb-for-vip"
  flavor_id   = 1
  network_id  = "subnet-test-456"
  vpc_id      = "vpc-test-123"
  vpc_name    = "test-vpc"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

resource "airtelcloud_lb_vip" "test" {
  lb_service_id = airtelcloud_lb_service.test.id
}
`
}

func TestAccLBCertificateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLBCertificateResourceConfig("test-cert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_lb_certificate.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_lb_certificate.test", "name", "test-cert"),
					resource.TestCheckResourceAttrSet("airtelcloud_lb_certificate.test", "lb_service_id"),
				),
			},
		},
	})
}

func testAccLBCertificateResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_lb_service" "test" {
  name        = "test-lb-for-cert"
  flavor_id   = 1
  network_id  = "subnet-test-456"
  vpc_id      = "vpc-test-123"
  vpc_name    = "test-vpc"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

resource "airtelcloud_lb_certificate" "test" {
  lb_service_id   = airtelcloud_lb_service.test.id
  name            = %[1]q
  ssl_cert        = "-----BEGIN CERTIFICATE-----\nMIIBxTCCAW...\n-----END CERTIFICATE-----"
  ssl_private_key = "-----BEGIN PRIVATE KEY-----\nMIIEvwIBAD...\n-----END PRIVATE KEY-----"
}
`, name)
}

func TestAccLBVirtualServerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLBVirtualServerResourceConfig("test-vs"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("airtelcloud_lb_virtual_server.test", "id"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "name", "test-vs"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "port", "80"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "routing_algorithm", "ROUND_ROBIN"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "interval", "30"),
					resource.TestCheckResourceAttr("airtelcloud_lb_virtual_server.test", "x_forwarded_for", "true"),
					resource.TestCheckResourceAttrSet("airtelcloud_lb_virtual_server.test", "status"),
				),
			},
		},
	})
}

func testAccLBVirtualServerResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_lb_service" "test" {
  name        = "test-lb-for-vs"
  flavor_id   = 1
  network_id  = "subnet-test-456"
  vpc_id      = "vpc-test-123"
  vpc_name    = "test-vpc"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

resource "airtelcloud_lb_vip" "test" {
  lb_service_id = airtelcloud_lb_service.test.id
}

resource "airtelcloud_lb_virtual_server" "test" {
  lb_service_id     = airtelcloud_lb_service.test.id
  name              = %[1]q
  vip_port_id       = tonumber(airtelcloud_lb_vip.test.id)
  protocol          = "HTTP"
  port              = 80
  routing_algorithm = "ROUND_ROBIN"
  vpc_id            = "vpc-test-123"
  interval          = 30
  x_forwarded_for   = true

  nodes = [
    {
      compute_id = 101
      compute_ip = "192.168.1.10"
      port       = 8080
    },
    {
      compute_id = 102
      compute_ip = "192.168.1.11"
      port       = 8080
    },
  ]

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name)
}
