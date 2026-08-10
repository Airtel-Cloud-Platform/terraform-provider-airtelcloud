package client

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ipamBasePath returns the base path for IPAM (public IP) endpoints
func (c *Client) ipamBasePath() string {
	return fmt.Sprintf("/ext/api/v1/domain/%s/project/%s/public-ip", c.Organization, c.ProjectName)
}

// CreatePublicIP allocates a new public IP in the specified availability zone
func (c *Client) CreatePublicIP(ctx context.Context, req *models.CreatePublicIPRequest, availabilityZone string) (*models.PublicIP, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	tflog.Debug(ctx, "CreatePublicIP request", map[string]interface{}{
		"availability_zone": availabilityZone,
		"request_body":      fmt.Sprintf("%+v", req),
	})

	// The IPAM create endpoint wraps the actual payload under "data".
	// Decode the wrapper so created.UUID is correctly populated.
	var createResp struct {
		Message string          `json:"message"`
		Data    models.PublicIP `json:"data"`
	}
	err := scopedClient.Post(ctx, scopedClient.ipamBasePath(), req, &createResp)
	if err != nil {
		return nil, err
	}
	return &createResp.Data, nil
}

// FindPortIDByVIP lists all compute instances and returns the port ID
// whose fixed_ips contain the given VIP address.
func (c *Client) FindPortIDByVIP(ctx context.Context, vip, availabilityZone string) (int, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	normalizedVIP := net.ParseIP(strings.TrimSpace(vip))
	if normalizedVIP == nil {
		return 0, fmt.Errorf("invalid VIP address %q", vip)
	}

	// Preferred LB lookup path: networks VIPs API exposes allowed_ip_address to port_id
	// mappings directly and is reliable for load-balancer VIPs.
	if lbPortID, lbFound, err := scopedClient.findNetworkVipPortIDByVIP(ctx, normalizedVIP); err != nil {
		tflog.Warn(ctx, "FindPortIDByVIP: networks VIP lookup failed; falling back", map[string]interface{}{
			"availability_zone": availabilityZone,
			"vip":               normalizedVIP.String(),
			"error":             err.Error(),
		})
	} else if lbFound {
		return lbPortID, nil
	}

	// LB VIPs are not always represented in compute port listings. Try the
	// LB VIP API first and return immediately when a matching VIP is found.
	if lbPortID, lbFound, err := scopedClient.findLBVipPortIDByVIP(ctx, normalizedVIP, availabilityZone); err != nil {
		tflog.Warn(ctx, "FindPortIDByVIP: LB VIP lookup failed; falling back to compute scan", map[string]interface{}{
			"availability_zone": availabilityZone,
			"vip":               normalizedVIP.String(),
			"error":             err.Error(),
		})
	} else if lbFound {
		return lbPortID, nil
	}

	computes, err := scopedClient.ListComputes(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list computes to find port for VIP %s in availability zone %s: %w", vip, availabilityZone, err)
	}

	tflog.Debug(ctx, "FindPortIDByVIP: listing computes", map[string]interface{}{
		"availability_zone": availabilityZone,
		"compute_count":     len(computes),
		"searched_vip":      normalizedVIP.String(),
	})

	matchInPorts := func(ports []models.Port) (int, bool) {
		for _, port := range ports {
			for _, fixedIP := range port.FixedIPs {
				parsedFixedIP := net.ParseIP(strings.TrimSpace(fixedIP))
				if parsedFixedIP == nil {
					continue
				}

				tflog.Debug(ctx, "FindPortIDByVIP: checking fixed IP", map[string]interface{}{
					"port_id":  port.ID,
					"fixed_ip": parsedFixedIP.String(),
					"vip":      normalizedVIP.String(),
				})

				if parsedFixedIP.Equal(normalizedVIP) {
					return port.ID, true
				}
			}
		}

		return 0, false
	}

	for _, compute := range computes {
		if portID, ok := matchInPorts(compute.Ports); ok {
			return portID, nil
		}

		if len(compute.Ports) == 0 && compute.ID != "" {
			fullCompute, err := scopedClient.GetCompute(ctx, compute.ID)
			if err != nil {
				tflog.Warn(ctx, "FindPortIDByVIP: failed to fetch full compute details", map[string]interface{}{
					"compute_id": compute.ID,
					"error":      err.Error(),
				})
				continue
			}

			tflog.Debug(ctx, "FindPortIDByVIP: fetched full compute details", map[string]interface{}{
				"compute_id":  compute.ID,
				"ports_count": len(fullCompute.Ports),
			})

			if portID, ok := matchInPorts(fullCompute.Ports); ok {
				return portID, nil
			}
		}
	}

	return 0, fmt.Errorf("no port found with fixed_ip matching VIP %s in availability zone %s", normalizedVIP.String(), availabilityZone)
}

func (c *Client) networkPortsVipsBasePath() string {
	return fmt.Sprintf("/api/v2.1/networks/domain/%s/project/%s/networks/ports/vips", c.Organization, c.ProjectName)
}

func (c *Client) findNetworkVipPortIDByVIP(ctx context.Context, vip net.IP) (int, bool, error) {
	var mappings []models.NetworkVIPPort
	if err := c.Get(ctx, c.networkPortsVipsBasePath(), &mappings); err != nil {
		return 0, false, fmt.Errorf("failed to list network VIP ports for VIP %s: %w", vip.String(), err)
	}

	tflog.Debug(ctx, "FindPortIDByVIP: listing network VIP ports", map[string]interface{}{
		"mapping_count": len(mappings),
		"searched_vip":  vip.String(),
	})

	for _, item := range mappings {
		parsedAllowedIP := net.ParseIP(strings.TrimSpace(item.AllowedIPAddress))
		if parsedAllowedIP == nil {
			continue
		}

		tflog.Debug(ctx, "FindPortIDByVIP: checking network VIP mapping", map[string]interface{}{
			"lb_name":              item.LBName,
			"vs_name":              item.VSName,
			"port_id":              item.PortID,
			"allowed_ip_address":   parsedAllowedIP.String(),
			"searched_vip_address": vip.String(),
		})

		if parsedAllowedIP.Equal(vip) {
			return item.PortID, true, nil
		}
	}

	return 0, false, nil
}

func (c *Client) findLBVipPortIDByVIP(ctx context.Context, vip net.IP, availabilityZone string) (int, bool, error) {
	services, err := c.ListLBServices(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("failed to list LB services for VIP %s: %w", vip.String(), err)
	}

	tflog.Debug(ctx, "FindPortIDByVIP: listing LB services", map[string]interface{}{
		"availability_zone": availabilityZone,
		"service_count":     len(services),
		"searched_vip":      vip.String(),
	})

	normalizedAZ := strings.TrimSpace(availabilityZone)
	for _, svc := range services {
		if normalizedAZ != "" && svc.AZName != "" && !strings.EqualFold(strings.TrimSpace(svc.AZName), normalizedAZ) {
			continue
		}

		lbScopedClient := c
		if strings.TrimSpace(svc.NetworkID) != "" {
			lbScopedClient = c.WithSubnetID(strings.TrimSpace(svc.NetworkID))
		}

		lbVips, err := lbScopedClient.ListLBVips(ctx, svc.ID)
		if err != nil {
			tflog.Warn(ctx, "FindPortIDByVIP: failed to list LB VIPs for service", map[string]interface{}{
				"lb_service_id": svc.ID,
				"network_id":    svc.NetworkID,
				"error":         err.Error(),
			})
			continue
		}

		for _, lbVIP := range lbVips {
			for _, fixedIP := range lbVIP.FixedIPs {
				parsedFixedIP := net.ParseIP(strings.TrimSpace(fixedIP))
				if parsedFixedIP == nil {
					continue
				}

				tflog.Debug(ctx, "FindPortIDByVIP: checking LB VIP fixed IP", map[string]interface{}{
					"lb_service_id": svc.ID,
					"vip_port_id":   lbVIP.ID,
					"fixed_ip":      parsedFixedIP.String(),
					"vip":           vip.String(),
				})

				if parsedFixedIP.Equal(vip) {
					return lbVIP.ID, true, nil
				}
			}
		}
	}

	return 0, false, nil
}

// GetPublicIP retrieves a public IP by UUID
func (c *Client) GetPublicIP(ctx context.Context, uuid string) (*models.PublicIP, error) {
	var response struct {
		Message string          `json:"message"`
		Data    models.PublicIP `json:"data"`
	}
	err := c.Get(ctx, fmt.Sprintf("%s/%s", c.ipamBasePath(), uuid), &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
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

// ResolvePublicIPID resolves a public IP name (object_name) to its UUID
func (c *Client) ResolvePublicIPID(ctx context.Context, name string) (string, error) {
	resp, err := c.ListPublicIPs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list public IPs: %w", err)
	}
	tflog.Debug(ctx, "ResolvePublicIPID: listing public IPs", map[string]interface{}{
		"public_ip_count": len(resp.Items),
		"searched_name":   name,
	})
	for _, ip := range resp.Items {
		if ip.ObjectName == name {
			if ip.UUID == "" {
				return "", fmt.Errorf("public IP %q found but has empty UUID (API response field mismatch)", name)
			}
			tflog.Debug(ctx, "ResolvePublicIPID: resolved public IP", map[string]interface{}{
				"name": name,
				"uuid": ip.UUID,
			})
			return ip.UUID, nil
		}
	}
	return "", fmt.Errorf("public IP with name %q not found", name)
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

	q := url.Values{}
	q.Set("offset", "0")
	q.Set("limit", "1000")
	q.Set("target_vip", targetVIP)
	q.Set("public_ip", publicIP)

	path := fmt.Sprintf("%s/%s/rules?%s", c.ipamAdminBasePath(), publicIPUUID, q.Encode())

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
