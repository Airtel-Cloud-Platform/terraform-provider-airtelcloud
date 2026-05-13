package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// ipamBasePath returns the base path for IPAM (public IP) endpoints
func (c *Client) ipamBasePath() string {
	return fmt.Sprintf("/v1/ipam/domain/%s/project/%s", c.Organization, c.ProjectName)
}

// CreatePublicIP allocates a new public IP in the specified availability zone
func (c *Client) CreatePublicIP(ctx context.Context, req *models.CreatePublicIPRequest, availabilityZone string) (*models.PublicIP, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	var publicIP models.PublicIP
	err := scopedClient.Post(ctx, scopedClient.ipamBasePath(), req, &publicIP)
	if err != nil {
		return nil, err
	}
	return &publicIP, nil
}

// FindPortIDByVIP lists all compute instances and returns the port ID
// whose fixed_ips contain the given VIP address.
func (c *Client) FindPortIDByVIP(ctx context.Context, vip string) (int, error) {
	computes, err := c.ListComputes(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list computes to find port for VIP %s: %w", vip, err)
	}

	for _, compute := range computes {
		for _, port := range compute.Ports {
			for _, fixedIP := range port.FixedIPs {
				if fixedIP == vip {
					return port.ID, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("no port found with fixed_ip matching VIP %s", vip)
}

// MapPublicIP associates a public IP with an internal VIP by creating a vip_object mapping
func (c *Client) MapPublicIP(ctx context.Context, req *models.MapPublicIPRequest, availabilityZone string) error {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	var result map[string]interface{}
	err := scopedClient.Post(ctx, fmt.Sprintf("%s/vip_object", scopedClient.ipamAdminBasePath()), req, &result)
	if err != nil {
		return fmt.Errorf("failed to map public IP: %w", err)
	}
	return nil
}

// GetPublicIP retrieves a public IP by UUID
func (c *Client) GetPublicIP(ctx context.Context, uuid string) (*models.PublicIP, error) {
	var publicIP models.PublicIP
	err := c.Get(ctx, fmt.Sprintf("%s/%s", c.ipamBasePath(), uuid), &publicIP)
	if err != nil {
		return nil, err
	}
	return &publicIP, nil
}

// ListPublicIPs retrieves all public IPs
func (c *Client) ListPublicIPs(ctx context.Context) (*models.PublicIPListResponse, error) {
	var response models.PublicIPListResponse
	err := c.Get(ctx, fmt.Sprintf("%s?offset=0&limit=1000", c.ipamBasePath()), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeletePublicIP deallocates a public IP by UUID
func (c *Client) DeletePublicIP(ctx context.Context, uuid string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%s", c.ipamBasePath(), uuid))
}

// --- Public IP Policy Rule ---

// ipamAdminBasePath returns the base path for IPAM admin (policy rule) endpoints
func (c *Client) ipamAdminBasePath() string {
	return "/api/v1/admin/ipam_vip"
}

// ListIPAMServices retrieves available services/ports for policy rules
func (c *Client) ListIPAMServices(ctx context.Context, availabilityZone string) ([]models.IPAMService, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	var services []models.IPAMService
	err := scopedClient.Get(ctx, fmt.Sprintf("%s/ipam_port", scopedClient.ipamAdminBasePath()), &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// CreatePublicIPPolicyRule creates a NAT policy rule for a public IP
func (c *Client) CreatePublicIPPolicyRule(ctx context.Context, req *models.CreatePublicIPPolicyRuleRequest, availabilityZone string) error {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	var result map[string]interface{}
	err := scopedClient.Post(ctx, fmt.Sprintf("%s/nat_rule", scopedClient.ipamAdminBasePath()), req, &result)
	if err != nil {
		return err
	}
	return nil
}

// ListPublicIPPolicyRules lists all policy rules for a public IP
func (c *Client) ListPublicIPPolicyRules(ctx context.Context, publicIPUUID, targetVIP, publicIP string) (*models.PublicIPPolicyRuleListResponse, error) {
	var response models.PublicIPPolicyRuleListResponse
	path := fmt.Sprintf("%s/%s/rules?offset=0&limit=1000&target_vip=%s&public_ip=%s",
		c.ipamAdminBasePath(), publicIPUUID, targetVIP, publicIP)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetPublicIPPolicyRule retrieves a specific policy rule by listing and filtering by rule UUID
func (c *Client) GetPublicIPPolicyRule(ctx context.Context, publicIPUUID, targetVIP, publicIP, ruleUUID string) (*models.PublicIPPolicyRule, error) {
	response, err := c.ListPublicIPPolicyRules(ctx, publicIPUUID, targetVIP, publicIP)
	if err != nil {
		return nil, err
	}

	for _, rule := range response.Items {
		if rule.UUID == ruleUUID {
			return &rule, nil
		}
	}

	return nil, &APIError{StatusCode: 404, Message: "policy rule not found"}
}

// DeletePublicIPPolicyRule deletes a NAT policy rule
func (c *Client) DeletePublicIPPolicyRule(ctx context.Context, ruleUUID string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/nat_rule/%s", c.ipamAdminBasePath(), ruleUUID))
}

// WaitForPublicIPReady polls until the public IP reaches "Created" status
func (c *Client) WaitForPublicIPReady(ctx context.Context, uuid string, timeout time.Duration) (*models.PublicIP, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		ip, err := c.GetPublicIP(ctx, uuid)
		if err != nil {
			return nil, err
		}

		switch ip.Status {
		case "Created", "created", "CREATED":
			return ip, nil
		case "Error", "error", "ERROR":
			return nil, fmt.Errorf("public IP entered error state")
		}

		time.Sleep(15 * time.Second)
	}

	return nil, fmt.Errorf("public IP did not become ready within %v", timeout)
}
