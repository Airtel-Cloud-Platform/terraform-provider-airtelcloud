package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecurityGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSecurityGroupResourceConfig("test-sg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("airtelcloud_security_group.test", "security_group_name", "test-sg"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group.test", "uuid"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "airtelcloud_security_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSecurityGroupResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "airtelcloud_security_group" "test" {
  security_group_name = %[1]q
}
`, name)
}

func TestAccSecurityGroupRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create security group and rule
			{
				Config: testAccSecurityGroupRuleResourceConfig("ingress", "tcp", "22", "22", "0.0.0.0/0", "IPv4", "Allow SSH"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify security group
					resource.TestCheckResourceAttr("airtelcloud_security_group.test", "security_group_name", "test-sg-for-rules"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group.test", "id"),
					// Verify rule
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "direction", "ingress"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "port_range_min", "22"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "port_range_max", "22"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "remote_ip_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "ethertype", "IPv4"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.test", "description", "Allow SSH"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group_rule.test", "id"),
					resource.TestCheckResourceAttrSet("airtelcloud_security_group_rule.test", "status"),
				),
			},
			// ImportState testing for rule (format: sg_id/rule_id)
			{
				ResourceName:      "airtelcloud_security_group_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSecurityGroupRuleResourceConfig(direction, protocol, portMin, portMax, remoteIP, ethertype, description string) string {
	return fmt.Sprintf(`
resource "airtelcloud_security_group" "test" {
  security_group_name = "test-sg-for-rules"
}

resource "airtelcloud_security_group_rule" "test" {
  security_group_id = airtelcloud_security_group.test.id
  direction         = %[1]q
  protocol          = %[2]q
  port_range_min    = %[3]q
  port_range_max    = %[4]q
  remote_ip_prefix  = %[5]q
  ethertype         = %[6]q
  description       = %[7]q
}
`, direction, protocol, portMin, portMax, remoteIP, ethertype, description)
}

func TestAccSecurityGroupWithMultipleRules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupMultipleRulesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify security group
					resource.TestCheckResourceAttr("airtelcloud_security_group.test", "security_group_name", "web-servers"),
					// Verify SSH rule
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.ssh", "direction", "ingress"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.ssh", "protocol", "tcp"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.ssh", "port_range_min", "22"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.ssh", "port_range_max", "22"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.ssh", "remote_ip_prefix", "10.0.0.0/8"),
					// Verify HTTP rule
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.http", "direction", "ingress"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.http", "protocol", "tcp"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.http", "port_range_min", "80"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.http", "port_range_max", "80"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.http", "remote_ip_prefix", "0.0.0.0/0"),
					// Verify HTTPS egress rule
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.https_out", "direction", "egress"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.https_out", "protocol", "tcp"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.https_out", "port_range_min", "443"),
					resource.TestCheckResourceAttr("airtelcloud_security_group_rule.https_out", "port_range_max", "443"),
				),
			},
		},
	})
}

func testAccSecurityGroupMultipleRulesConfig() string {
	return `
resource "airtelcloud_security_group" "test" {
  security_group_name = "web-servers"
}

resource "airtelcloud_security_group_rule" "ssh" {
  security_group_id = airtelcloud_security_group.test.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "22"
  port_range_max    = "22"
  remote_ip_prefix  = "10.0.0.0/8"
  ethertype         = "IPv4"
  description       = "Allow SSH from internal network"
}

resource "airtelcloud_security_group_rule" "http" {
  security_group_id = airtelcloud_security_group.test.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = "80"
  port_range_max    = "80"
  remote_ip_prefix  = "0.0.0.0/0"
  ethertype         = "IPv4"
  description       = "Allow HTTP"
}

resource "airtelcloud_security_group_rule" "https_out" {
  security_group_id = airtelcloud_security_group.test.id
  direction         = "egress"
  protocol          = "tcp"
  port_range_min    = "443"
  port_range_max    = "443"
  remote_ip_prefix  = "0.0.0.0/0"
  ethertype         = "IPv4"
  description       = "Allow HTTPS outbound"
}
`
}
