package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Delays between public IP policy rule polls. Tests shorten them.
var (
	publicIPPolicyRuleReadyPollInterval  = 5 * time.Second
	publicIPPolicyRuleDeletePollInterval = 2 * time.Second
)

type sourceOfTruthPublicIPPolicyService struct {
	CreateNew bool   `json:"create_new"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type publicIPPolicyGeographic struct {
	CountryCode string `json:"country_code,omitempty"`
	CountryName string `json:"country_name,omitempty"`
}

type publicIPPolicySourceEntry struct {
	CreateNew  *bool                     `json:"create_new,omitempty"`
	IPCIDR     string                    `json:"ip_cidr,omitempty"`
	SourceType string                    `json:"source_type"`
	Geographic *publicIPPolicyGeographic `json:"geographic,omitempty"`
}

type sourceOfTruthCreatePublicIPPolicyRequest struct {
	ResourceType string                               `json:"resource_type"`
	RuleName     string                               `json:"rule_name"`
	Source       []publicIPPolicySourceEntry          `json:"source"`
	Services     []sourceOfTruthPublicIPPolicyService `json:"services"`
	Action       string                               `json:"action"`
	RevisionNote string                               `json:"revision_note"`
}

type sourceOfTruthCreatePublicIPPolicyResponse struct {
	Message string `json:"message"`
	Data    struct {
		UUID string `json:"uuid"`
	} `json:"data"`
}

type sourceOfTruthPublicIPPolicyRuleDetailResponse struct {
	Message string `json:"message"`
	Data    struct {
		UUID     string `json:"uuid"`
		RuleName string `json:"rule_name"`
		State    string `json:"state"`
		Status   string `json:"status"`
		Action   string `json:"action"`
		Source   []struct {
			All        bool   `json:"all,omitempty"`
			IPCIDR     string `json:"ip_cidr,omitempty"`
			Geographic any    `json:"geographic,omitempty"`
		} `json:"source"`
		Services []struct {
			Name string `json:"name"`
		} `json:"services"`
	} `json:"data"`
}

type sourceOfTruthDeletePublicIPPolicyRequest struct {
	PolicyIDs []string `json:"policy_ids"`
}

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

// AttachPublicIP binds a reserved public IP to a port.
func (c *Client) AttachPublicIP(ctx context.Context, uuid string, portID int, availabilityZone string) error {
	scopedClient := c.WithAvailabilityZone(availabilityZone)
	path := fmt.Sprintf("%s/%s/attach", scopedClient.ipamBasePath(), uuid)
	req := &models.AttachPublicIPRequest{PortID: portID}

	tflog.Debug(ctx, "AttachPublicIP request", map[string]interface{}{
		"availability_zone": availabilityZone,
		"public_ip_uuid":    uuid,
		"port_id":           portID,
	})

	return scopedClient.Post(ctx, path, req, nil)
}

// DetachPublicIP unbinds a public IP from a port and returns it to reserved.
func (c *Client) DetachPublicIP(ctx context.Context, uuid string, portID int, availabilityZone string) error {
	scopedClient := c.WithAvailabilityZone(availabilityZone)
	path := fmt.Sprintf("%s/%s/detach", scopedClient.ipamBasePath(), uuid)
	req := &models.AttachPublicIPRequest{PortID: portID}

	tflog.Debug(ctx, "DetachPublicIP request", map[string]interface{}{
		"availability_zone": availabilityZone,
		"public_ip_uuid":    uuid,
		"port_id":           portID,
	})

	return scopedClient.Post(ctx, path, req, nil)
}

// Resource types accepted when attaching a public IP.
const (
	PublicIPResourceTypeVM        = "vm"
	PublicIPResourceTypeLB        = "lb"
	PublicIPResourceTypeBaremetal = "baremetal"
)

// NormalizePublicIPResourceType maps user-facing aliases to canonical types.
func NormalizePublicIPResourceType(resourceType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case "vm", "compute", "instance", "virtual_machine":
		return PublicIPResourceTypeVM, nil
	case "lb", "load_balancer", "loadbalancer":
		return PublicIPResourceTypeLB, nil
	case "baremetal", "bare_metal", "bm":
		return PublicIPResourceTypeBaremetal, nil
	default:
		return "", fmt.Errorf("unsupported resource_type %q: must be one of vm, lb, baremetal", resourceType)
	}
}

// FindPortForResource resolves the attach port ID for a named VM, load balancer,
// or baremetal server. It always uses the resource's primary private IP/VIP
// (first NIC, first LB VIP, or first baremetal ipAddr). It returns the port ID
// and that private IP.
func (c *Client) FindPortForResource(ctx context.Context, resourceType, resourceName, targetVIP, availabilityZone string) (int, string, error) {
	canonicalType, err := NormalizePublicIPResourceType(resourceType)
	if err != nil {
		return 0, "", err
	}

	name := strings.TrimSpace(resourceName)
	if name == "" {
		return 0, "", fmt.Errorf("resource_name is required")
	}

	vip := strings.TrimSpace(targetVIP)
	if vip != "" && net.ParseIP(vip) == nil {
		return 0, "", fmt.Errorf("invalid target_vip %q", targetVIP)
	}

	tflog.Debug(ctx, "FindPortForResource: resolving attach port", map[string]interface{}{
		"resource_type":     canonicalType,
		"resource_name":     name,
		"target_vip":        vip,
		"availability_zone": availabilityZone,
	})

	scopedClient := c.WithAvailabilityZone(availabilityZone)

	switch canonicalType {
	case PublicIPResourceTypeVM:
		return scopedClient.findVMPortForResource(ctx, name, vip, availabilityZone)
	case PublicIPResourceTypeLB:
		return scopedClient.findLBPortForResource(ctx, name, vip, availabilityZone)
	default:
		return scopedClient.findBaremetalPortForResource(ctx, name, vip, availabilityZone)
	}
}

func (c *Client) findVMPortForResource(ctx context.Context, name, vip, availabilityZone string) (int, string, error) {
	computes, err := c.ListComputes(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to list computes to find port for VM %q in availability zone %s: %w", name, availabilityZone, err)
	}

	for _, compute := range computes {
		if compute.InstanceName != name {
			continue
		}

		ports := compute.Ports
		if len(ports) == 0 && compute.ID != "" {
			fullCompute, getErr := c.GetCompute(ctx, compute.ID)
			if getErr != nil {
				return 0, "", fmt.Errorf("failed to fetch compute %q details: %w", name, getErr)
			}
			ports = fullCompute.Ports
		}
		if len(ports) == 0 {
			return 0, "", fmt.Errorf("VM %q has no ports in availability zone %s", name, availabilityZone)
		}

		if portID, matchedVIP, ok := matchPortByVIP(ports, vip); ok {
			return portID, matchedVIP, nil
		}
		return 0, "", fmt.Errorf("VM %q has no port with private IP %s", name, vip)
	}

	return 0, "", fmt.Errorf("VM with name %q not found in availability zone %s", name, availabilityZone)
}

func (c *Client) findLBPortForResource(ctx context.Context, name, vip, availabilityZone string) (int, string, error) {
	var mappings []models.NetworkVIPPort
	if err := c.Get(ctx, c.networkPortsVipsBasePath(), &mappings); err == nil {
		for _, item := range mappings {
			if !strings.EqualFold(strings.TrimSpace(item.LBName), name) || item.PortID == 0 {
				continue
			}
			allowedIP := strings.TrimSpace(item.AllowedIPAddress)
			if vip == "" || sameIP(allowedIP, vip) {
				return item.PortID, allowedIP, nil
			}
		}
	} else {
		tflog.Warn(ctx, "FindPortForResource: networks VIP lookup failed; falling back to LB services", map[string]interface{}{
			"lb_name": name,
			"error":   err.Error(),
		})
	}

	services, err := c.ListLBServices(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to list load balancers to find port for %q: %w", name, err)
	}

	normalizedAZ := strings.TrimSpace(availabilityZone)
	for _, svc := range services {
		if !strings.EqualFold(strings.TrimSpace(svc.Name), name) {
			continue
		}
		if normalizedAZ != "" && svc.AZName != "" && !strings.EqualFold(strings.TrimSpace(svc.AZName), normalizedAZ) {
			continue
		}

		lbScopedClient := c
		if strings.TrimSpace(svc.NetworkID) != "" {
			lbScopedClient = c.WithSubnetID(strings.TrimSpace(svc.NetworkID))
		}

		lbVips, vipErr := lbScopedClient.ListLBVips(ctx, svc.ID)
		if vipErr != nil {
			return 0, "", fmt.Errorf("failed to list VIPs for load balancer %q: %w", name, vipErr)
		}
		if len(lbVips) == 0 {
			return 0, "", fmt.Errorf("load balancer %q has no VIP ports", name)
		}

		ports := make([]models.Port, 0, len(lbVips))
		for _, lbVIP := range lbVips {
			ports = append(ports, models.Port{ID: lbVIP.ID, FixedIPs: lbVIP.FixedIPs})
		}
		if portID, matchedVIP, ok := matchPortByVIP(ports, vip); ok {
			return portID, matchedVIP, nil
		}
		return 0, "", fmt.Errorf("load balancer %q has no VIP port with private IP %s", name, vip)
	}

	return 0, "", fmt.Errorf("load balancer with name %q not found in availability zone %s", name, availabilityZone)
}

func (c *Client) findBaremetalPortForResource(ctx context.Context, name, vip, availabilityZone string) (int, string, error) {
	server, err := c.GetBaremetal(ctx, name, availabilityZone)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch baremetal server %q: %w", name, err)
	}

	portID := int(server.NetworkInfo.PortID)
	if portID == 0 {
		return 0, "", fmt.Errorf("baremetal server %q has no port id", name)
	}

	if vip == "" {
		return portID, server.PrimaryIP(), nil
	}
	for _, ip := range server.IPAddr {
		if sameIP(ip, vip) {
			return portID, strings.TrimSpace(ip), nil
		}
	}
	return 0, "", fmt.Errorf("baremetal server %q has no private IP %s", name, vip)
}

// matchPortByVIP returns the port carrying vip, or the first port when vip is empty.
func matchPortByVIP(ports []models.Port, vip string) (int, string, bool) {
	if len(ports) == 0 {
		return 0, "", false
	}

	if vip == "" {
		first := ports[0]
		fixedIP := ""
		if len(first.FixedIPs) > 0 {
			fixedIP = strings.TrimSpace(first.FixedIPs[0])
		}
		return first.ID, fixedIP, true
	}

	for _, port := range ports {
		for _, fixedIP := range port.FixedIPs {
			if sameIP(fixedIP, vip) {
				return port.ID, strings.TrimSpace(fixedIP), true
			}
		}
	}
	return 0, "", false
}

func sameIP(a, b string) bool {
	parsedA := net.ParseIP(strings.TrimSpace(a))
	parsedB := net.ParseIP(strings.TrimSpace(b))
	if parsedA == nil || parsedB == nil {
		return false
	}
	return parsedA.Equal(parsedB)
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
	var wrapped struct {
		Message string `json:"message"`
		Data    struct {
			Items []models.PublicIP `json:"items"`
			Count int               `json:"count"`
		} `json:"data"`
		Items []models.PublicIP `json:"items"`
		Count int               `json:"count"`
	}
	err := c.Get(ctx, fmt.Sprintf("%s?offset=0&limit=1000", c.ipamBasePath()), &wrapped)
	if err != nil {
		return nil, err
	}

	items := wrapped.Data.Items
	count := wrapped.Data.Count
	if len(items) == 0 {
		items = wrapped.Items
		count = wrapped.Count
	}

	return &models.PublicIPListResponse{Items: items, Count: count}, nil
}

// DeletePublicIP deallocates a public IP by UUID
func (c *Client) DeletePublicIP(ctx context.Context, uuid string) error {
	path := fmt.Sprintf("%s/%s", c.ipamBasePath(), uuid)
	tflog.Debug(ctx, "DeletePublicIP: issuing delete request", map[string]interface{}{
		"public_ip_uuid": uuid,
		"delete_path":    path,
	})

	err := c.Delete(ctx, path)
	if err != nil {
		tflog.Error(ctx, "DeletePublicIP: delete request failed", map[string]interface{}{
			"public_ip_uuid": uuid,
			"delete_path":    path,
			"error":          err.Error(),
		})
		return err
	}

	tflog.Debug(ctx, "DeletePublicIP: delete request completed", map[string]interface{}{
		"public_ip_uuid": uuid,
		"delete_path":    path,
	})
	return nil
}

// DeletePublicIPWithWait deletes a public IP and waits until the backend reports
// a final deleted state or the resource is no longer found.
func (c *Client) DeletePublicIPWithWait(ctx context.Context, uuid string, timeout time.Duration) error {
	err := c.DeletePublicIP(ctx, uuid)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ip, readErr := c.GetPublicIP(ctx, uuid)
		if readErr != nil {
			if IsNotFoundError(readErr) {
				return nil
			}
			if isAPIErrorStatus(readErr, 500) || isAPIErrorStatus(readErr, 503) {
				time.Sleep(2 * time.Second)
				continue
			}
			return readErr
		}

		if ip != nil {
			if isPublicIPDeleted(ip.Status) {
				tflog.Debug(ctx, "DeletePublicIPWithWait: backend reported deleted status", map[string]interface{}{
					"public_ip_uuid": uuid,
					"status":         ip.Status,
				})
				return nil
			}

			if isPublicIPFailed(ip.Status) {
				return fmt.Errorf("public IP %s delete failed: status=%s", uuid, ip.Status)
			}

			tflog.Debug(ctx, "DeletePublicIPWithWait: public IP still present after delete request", map[string]interface{}{
				"public_ip_uuid": uuid,
				"status":         ip.Status,
				"deadline":       deadline.Format(time.RFC3339Nano),
				"remaining":      time.Until(deadline).String(),
			})
		}

		time.Sleep(2 * time.Second)
	}

	tflog.Debug(ctx, "DeletePublicIPWithWait: delete wait timeout reached; public IP still present", map[string]interface{}{
		"public_ip_uuid": uuid,
		"timeout":        timeout,
	})
	return fmt.Errorf("public IP %s is still present after delete wait timeout %v", uuid, timeout)
}

func isPublicIPDeleted(status string) bool {
	statusNormalized := strings.ToLower(strings.TrimSpace(status))
	if statusNormalized == "" {
		return false
	}

	switch statusNormalized {
	case "deleted", "delete", "soft-deleted", "removed", "not_found", "notfound":
		return true
	default:
		return false
	}
}

func isPublicIPFailed(status string) bool {
	statusNormalized := strings.ToLower(strings.TrimSpace(status))
	if statusNormalized == "" {
		return false
	}

	switch statusNormalized {
	case "failed", "failure", "error", "timed_out", "timeout", "timedout", "cancelled", "canceled", "rollback", "deletion_failed", "deletionfailed":
		return true
	default:
		return false
	}
}

// GetPublicIPByName resolves a public IP object_name to the full object.
func (c *Client) GetPublicIPByName(ctx context.Context, name string) (*models.PublicIP, error) {
	resp, err := c.ListPublicIPs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list public IPs: %w", err)
	}
	tflog.Debug(ctx, "GetPublicIPByName: listing public IPs", map[string]interface{}{
		"public_ip_count": len(resp.Items),
		"searched_name":   name,
	})
	want := strings.TrimSpace(name)
	for _, ip := range resp.Items {
		if !strings.EqualFold(models.PublicIPDisplayName(ip), want) {
			continue
		}
		if ip.UUID == "" {
			return nil, fmt.Errorf("public IP %q found but has empty UUID (API response field mismatch)", name)
		}
		full, getErr := c.GetPublicIP(ctx, ip.UUID)
		if getErr != nil || full == nil || strings.TrimSpace(full.UUID) == "" {
			copied := ip
			return &copied, nil
		}
		return full, nil
	}
	return nil, fmt.Errorf("public IP with name %q not found", name)
}

// ResolvePublicIPID resolves a public IP name (object_name) to its UUID
func (c *Client) ResolvePublicIPID(ctx context.Context, name string) (string, error) {
	ip, err := c.GetPublicIPByName(ctx, name)
	if err != nil {
		return "", err
	}
	return ip.UUID, nil
}

// --- Public IP Policy Rule ---

// ipamAdminBasePath returns the base path for IPAM admin (policy rule) endpoints
func (c *Client) ipamAdminBasePath() string {
	return "/api/v1/admin/ipam_vip"
}

func (c *Client) publicIPPolicyBasePath(publicIPUUID string) string {
	return fmt.Sprintf("/ext/api/v1/domain/%s/project/%s/public-ip-id/%s/policy", c.Organization, c.ProjectName, publicIPUUID)
}

// ListIPAMServices retrieves available services/ports for policy rules
func (c *Client) ListIPAMServices(ctx context.Context, availabilityZone string) ([]models.IPAMService, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	var services []models.IPAMService
	err := scopedClient.Get(ctx, "/api/v1/admin/ipam_vip/ipam_port?port_type=all", &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// CreatePublicIPPolicyRule creates a NAT policy rule for a public IP and returns the created policy UUID.
func (c *Client) CreatePublicIPPolicyRule(ctx context.Context, req *models.CreatePublicIPPolicyRuleRequest, availabilityZone string) (string, error) {
	scopedClient := c.WithAvailabilityZone(availabilityZone)

	falseVal := false
	policySource := make([]publicIPPolicySourceEntry, 0, len(req.SourceConfig)+1)
	for _, source := range req.SourceConfig {
		sourceType := strings.TrimSpace(strings.ToLower(source.SourceType))
		ipCIDR := strings.TrimSpace(source.IPCIDR)

		var geographic *publicIPPolicyGeographic
		if source.Geographic != nil {
			code := strings.TrimSpace(source.Geographic.CountryCode)
			name := strings.TrimSpace(source.Geographic.CountryName)
			if code != "" || name != "" {
				geographic = &publicIPPolicyGeographic{CountryCode: code, CountryName: name}
			}
		}

		if sourceType == "" {
			switch {
			case geographic != nil:
				sourceType = "geographic"
			case ipCIDR == "" || strings.EqualFold(ipCIDR, "any") || strings.EqualFold(ipCIDR, "all"):
				sourceType = "all"
			default:
				sourceType = "ip_cidr"
			}
		}
		if sourceType == "any" {
			sourceType = "all"
		}

		if sourceType == "ip_cidr" && ipCIDR == "" {
			continue
		}
		if sourceType == "geographic" && geographic == nil {
			continue
		}

		entry := publicIPPolicySourceEntry{SourceType: sourceType}
		switch sourceType {
		case "geographic":
			entry.Geographic = geographic
		case "ip_cidr":
			entry.IPCIDR = ipCIDR
			entry.CreateNew = &falseVal
			if source.CreateNew != nil {
				entry.CreateNew = source.CreateNew
			}
		case "all":
			entry.CreateNew = &falseVal
			if source.CreateNew != nil {
				entry.CreateNew = source.CreateNew
			}
		default:
			if ipCIDR != "" {
				entry.IPCIDR = ipCIDR
			}
			if geographic != nil {
				entry.Geographic = geographic
			}
			if source.CreateNew != nil {
				entry.CreateNew = source.CreateNew
			}
		}
		policySource = append(policySource, entry)
	}

	if len(policySource) == 0 {
		source := strings.TrimSpace(req.Source)
		lower := strings.ToLower(source)
		if source == "" || lower == "any" || lower == "all" {
			policySource = []publicIPPolicySourceEntry{{
				CreateNew:  &falseVal,
				SourceType: "all",
			}}
		} else {
			policySource = []publicIPPolicySourceEntry{{
				CreateNew:  &falseVal,
				SourceType: "ip_cidr",
				IPCIDR:     source,
			}}
		}
	}

	policyServices := make([]sourceOfTruthPublicIPPolicyService, 0, len(req.ServiceConfig)+len(req.ServiceList))
	for _, service := range req.ServiceConfig {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			continue
		}

		createNew := false
		if service.CreateNew != nil {
			createNew = *service.CreateNew
		}

		isDefault := false
		if service.IsDefault != nil {
			isDefault = *service.IsDefault
		}

		policyServices = append(policyServices, sourceOfTruthPublicIPPolicyService{
			CreateNew: createNew,
			Name:      name,
			IsDefault: isDefault,
		})
	}

	if len(policyServices) == 0 {
		for _, service := range req.ServiceList {
			name := strings.TrimSpace(service)
			if name == "" {
				continue
			}
			policyServices = append(policyServices, sourceOfTruthPublicIPPolicyService{
				CreateNew: false,
				Name:      name,
				IsDefault: false,
			})
		}
	}
	if len(policyServices) == 0 {
		policyServices = []sourceOfTruthPublicIPPolicyService{{CreateNew: false, Name: "ALL", IsDefault: false}}
	}

	resourceType := strings.TrimSpace(req.ResourceType)
	if resourceType == "" {
		resourceType = "ipam"
	}

	revisionNote := strings.TrimSpace(req.RevisionNote)
	if revisionNote == "" {
		revisionNote = "creating Policy"
	}

	payload := sourceOfTruthCreatePublicIPPolicyRequest{
		ResourceType: resourceType,
		RuleName:     req.DisplayName,
		Source:       policySource,
		Services:     policyServices,
		Action:       req.Action,
		RevisionNote: revisionNote,
	}

	policyPath := scopedClient.publicIPPolicyBasePath(req.UUID)
	tflog.Debug(ctx, "CreatePublicIPPolicyRule: creating policy rule via source-of-truth route", map[string]interface{}{
		"public_ip_id":       req.UUID,
		"availability_zone":  availabilityZone,
		"rule_name":          req.DisplayName,
		"source":             payload.Source,
		"service_name_count": len(payload.Services),
		"action":             req.Action,
		"policy_path":        policyPath,
	})

	/*
		Legacy create path retained for reference:
		- Endpoint: /api/v1/admin/ipam_vip/nat_rule
		- Payload shape: {display_name, source, service_list, action, target_vip, public_ip, uuid}
	*/

	var result sourceOfTruthCreatePublicIPPolicyResponse
	if err := scopedClient.Post(ctx, policyPath, &payload, &result); err != nil {
		return "", err
	}

	policyUUID := strings.TrimSpace(result.Data.UUID)
	if policyUUID == "" {
		return "", fmt.Errorf("policy creation succeeded but response did not contain policy UUID")
	}

	return policyUUID, nil
}

// ListPublicIPPolicyRules lists all policy rules for a public IP
func (c *Client) ListPublicIPPolicyRules(ctx context.Context, publicIPUUID, targetVIP, publicIP string) (*models.PublicIPPolicyRuleListResponse, error) {
	tflog.Debug(ctx, "ListPublicIPPolicyRules: requesting policy list via admin ipam route", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"target_vip":   targetVIP,
		"public_ip":    publicIP,
	})

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
	tflog.Debug(ctx, "ListPublicIPPolicyRules: admin ipam response received", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"count":        len(response.Items),
	})

	return &response, nil
}

// GetPublicIPPolicyRule retrieves a specific policy rule by UUID.
func (c *Client) GetPublicIPPolicyRule(ctx context.Context, publicIPUUID, targetVIP, publicIP, ruleUUID string) (*models.PublicIPPolicyRule, error) {
	tflog.Debug(ctx, "GetPublicIPPolicyRule: requesting policy via source-of-truth route", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"policy_uuid":  ruleUUID,
		"target_vip":   targetVIP,
		"public_ip":    publicIP,
		"policy_path":  fmt.Sprintf("%s/%s", c.publicIPPolicyBasePath(publicIPUUID), ruleUUID),
	})

	path := fmt.Sprintf("%s/%s", c.publicIPPolicyBasePath(publicIPUUID), ruleUUID)

	var response sourceOfTruthPublicIPPolicyRuleDetailResponse
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(response.Data.UUID) == "" {
		return nil, &APIError{StatusCode: 404, Message: "policy rule not found"}
	}

	rule := &models.PublicIPPolicyRule{
		UUID:        response.Data.UUID,
		DisplayName: response.Data.RuleName,
		SourceIP:    extractPolicyRuleSource(response.Data.Source),
		Services:    extractPolicyRuleServices(response.Data.Services),
		Action:      response.Data.Action,
		State:       extractPolicyRuleState(response.Data.State, response.Data.Status),
	}

	tflog.Debug(ctx, "GetPublicIPPolicyRule: fetched policy via source-of-truth route", map[string]interface{}{
		"public_ip_id":  publicIPUUID,
		"policy_uuid":   rule.UUID,
		"state":         rule.State,
		"service_count": len(rule.Services),
	})

	return rule, nil
}

// DeletePublicIPPolicyRule deletes a NAT policy rule
func (c *Client) DeletePublicIPPolicyRule(ctx context.Context, publicIPUUID, ruleUUID string) error {
	path := c.publicIPPolicyBasePath(publicIPUUID)
	payload := sourceOfTruthDeletePublicIPPolicyRequest{PolicyIDs: []string{ruleUUID}}

	tflog.Debug(ctx, "DeletePublicIPPolicyRule: deleting rule via source-of-truth route", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"policy_uuid":  ruleUUID,
		"policy_path":  path,
		"request_body": payload,
	})

	tflog.Debug(ctx, "DeletePublicIPPolicyRule: issuing delete request with policy payload", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"policy_uuid":  ruleUUID,
		"policy_path":  path,
		"payload":      payload,
	})

	err := c.DeleteWithBody(ctx, path, &payload, nil)
	if err != nil {
		tflog.Error(ctx, "DeletePublicIPPolicyRule: delete request failed", map[string]interface{}{
			"public_ip_id": publicIPUUID,
			"policy_uuid":  ruleUUID,
			"policy_path":  path,
			"error":        err.Error(),
		})
		return err
	}

	tflog.Debug(ctx, "DeletePublicIPPolicyRule: delete request completed", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"policy_uuid":  ruleUUID,
		"policy_path":  path,
	})
	return nil
}

// DeletePublicIPPolicyRuleWithWait deletes a NAT policy rule and waits until
// a follow-up read confirms the rule is absent in backend.
func (c *Client) DeletePublicIPPolicyRuleWithWait(ctx context.Context, publicIPUUID, targetVIP, publicIP, ruleUUID string, timeout time.Duration) error {
	err := c.DeletePublicIPPolicyRule(ctx, publicIPUUID, ruleUUID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rule, readErr := c.GetPublicIPPolicyRule(ctx, publicIPUUID, targetVIP, publicIP, ruleUUID)
		if readErr != nil {
			if IsNotFoundError(readErr) {
				return nil
			}

			if isAPIErrorStatus(readErr, 500) || isAPIErrorStatus(readErr, 503) {
				if !waitBeforeNextPoll(ctx, publicIPPolicyRuleDeletePollInterval, deadline) {
					break
				}
				continue
			}

			return readErr
		}

		if rule != nil {
			if isPublicIPPolicyRuleDeleted(rule.State) {
				tflog.Debug(ctx, "DeletePublicIPPolicyRuleWithWait: backend reported deleted state", map[string]interface{}{
					"public_ip_id": publicIPUUID,
					"policy_uuid":  ruleUUID,
					"state":        rule.State,
				})
				return nil
			}

			if isPublicIPPolicyRuleFailed(rule.State) {
				return fmt.Errorf("public IP policy rule %s delete failed: state=%s", ruleUUID, rule.State)
			}

			tflog.Debug(ctx, "DeletePublicIPPolicyRuleWithWait: policy still present after delete request", map[string]interface{}{
				"public_ip_id": publicIPUUID,
				"policy_uuid":  ruleUUID,
				"state":        rule.State,
				"deadline":     deadline.Format(time.RFC3339Nano),
				"remaining":    time.Until(deadline).String(),
			})
		}

		if !waitBeforeNextPoll(ctx, publicIPPolicyRuleDeletePollInterval, deadline) {
			break
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	tflog.Debug(ctx, "DeletePublicIPPolicyRuleWithWait: delete wait timeout reached; policy still present", map[string]interface{}{
		"public_ip_id": publicIPUUID,
		"policy_uuid":  ruleUUID,
		"timeout":      timeout,
	})
	return fmt.Errorf("public IP policy rule %s is still present after delete wait timeout %v", ruleUUID, timeout)
}

func isPublicIPPolicyRuleDeleted(state string) bool {
	stateNormalized := strings.ToLower(strings.TrimSpace(state))
	if stateNormalized == "" {
		return false
	}

	switch stateNormalized {
	case "deleted", "delete", "soft-deleted", "removed", "not_found", "notfound":
		return true
	default:
		return false
	}
}

func isPublicIPPolicyRuleFailed(state string) bool {
	stateNormalized := strings.ToLower(strings.TrimSpace(state))
	if stateNormalized == "" {
		return false
	}

	switch stateNormalized {
	case "failed", "failure", "error", "timed_out", "timeout", "timedout", "cancelled", "canceled", "rollback", "deletion_failed", "deletionfailed":
		return true
	default:
		return false
	}
}

func isPublicIPPolicyRuleReady(state string) bool {
	stateNormalized := strings.ToLower(strings.TrimSpace(state))
	if stateNormalized == "" {
		return false
	}

	switch stateNormalized {
	case "active", "enabled", "ready", "created", "accepted", "applied", "succeeded", "completed":
		return true
	default:
		return false
	}
}

// WaitForPublicIPPolicyRuleReady polls until the policy rule reaches a ready state.
func (c *Client) WaitForPublicIPPolicyRuleReady(ctx context.Context, publicIPUUID, targetVIP, publicIP, ruleUUID string, timeout time.Duration) (*models.PublicIPPolicyRule, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rule, err := c.GetPublicIPPolicyRule(ctx, publicIPUUID, targetVIP, publicIP, ruleUUID)
		if err != nil {
			if IsNotFoundError(err) {
				tflog.Debug(ctx, "WaitForPublicIPPolicyRuleReady: rule not found yet; continuing wait", map[string]interface{}{
					"public_ip_id": publicIPUUID,
					"policy_uuid":  ruleUUID,
				})
				if !waitBeforeNextPoll(ctx, publicIPPolicyRuleReadyPollInterval, deadline) {
					break
				}
				continue
			}
			if isAPIErrorStatus(err, 500) || isAPIErrorStatus(err, 503) {
				if !waitBeforeNextPoll(ctx, publicIPPolicyRuleReadyPollInterval, deadline) {
					break
				}
				continue
			}
			return nil, err
		}

		if rule != nil {
			if isPublicIPPolicyRuleReady(rule.State) {
				return rule, nil
			}
			if isPublicIPPolicyRuleFailed(rule.State) {
				return nil, fmt.Errorf("public IP policy rule %s failed while becoming ready: state=%s", ruleUUID, rule.State)
			}
			tflog.Debug(ctx, "WaitForPublicIPPolicyRuleReady: policy still pending", map[string]interface{}{
				"public_ip_id": publicIPUUID,
				"policy_uuid":  ruleUUID,
				"state":        rule.State,
				"deadline":     deadline.Format(time.RFC3339Nano),
			})
		}

		if !waitBeforeNextPoll(ctx, publicIPPolicyRuleReadyPollInterval, deadline) {
			break
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("public IP policy rule %s did not become ready within %v", ruleUUID, timeout)
}

// Legacy retry settings retained for backwards-compatible unit tests.
const publicIPPolicyRuleCreateRetryMaxAttempts = 6

// Legacy retry backoff retained for backwards-compatible unit tests.
var publicIPPolicyRuleCreateRetryBackoff = 5 * time.Second

// Legacy helper retained for backwards-compatible unit tests.
func isPublicIPPolicyRuleRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode != 400 {
		return false
	}

	message := strings.ToLower(apiErr.Message)
	return strings.Contains(message, "public ip allocation") && strings.Contains(message, "in progress")
}

func isAPIErrorStatus(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.StatusCode == statusCode
}

func extractPolicyRuleSource(source []struct {
	All        bool   `json:"all,omitempty"`
	IPCIDR     string `json:"ip_cidr,omitempty"`
	Geographic any    `json:"geographic,omitempty"`
}) string {
	for _, entry := range source {
		if entry.All {
			return "any"
		}
		if entry.IPCIDR != "" {
			return entry.IPCIDR
		}
	}
	return ""
}

func extractPolicyRuleServices(services []struct {
	Name string `json:"name"`
}) []string {
	names := make([]string, 0, len(services))
	for _, service := range services {
		if strings.TrimSpace(service.Name) == "" {
			continue
		}
		names = append(names, service.Name)
	}
	return names
}

func extractPolicyRuleState(state, status string) string {
	if strings.TrimSpace(status) != "" {
		return status
	}
	return state
}

// WaitForPublicIPReady polls until the public IP is reserved (or created on older APIs).
func (c *Client) WaitForPublicIPReady(ctx context.Context, uuid string, timeout time.Duration) (*models.PublicIP, error) {
	return c.WaitForPublicIPStatus(ctx, uuid, timeout, isPublicIPReservedStatus, "reserved")
}

// WaitForPublicIPAttached polls until the public IP is attached to a port.
func (c *Client) WaitForPublicIPAttached(ctx context.Context, uuid string, timeout time.Duration) (*models.PublicIP, error) {
	return c.WaitForPublicIPStatus(ctx, uuid, timeout, isPublicIPAttachedStatus, "attached")
}

func (c *Client) WaitForPublicIPStatus(ctx context.Context, uuid string, timeout time.Duration, ready func(string) bool, want string) (*models.PublicIP, error) {
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

		if ready(ip.Status) {
			return ip, nil
		}
		if isPublicIPFailed(ip.Status) {
			return nil, fmt.Errorf("public IP entered error state %q", ip.Status)
		}

		time.Sleep(15 * time.Second)
	}

	return nil, fmt.Errorf("public IP did not become %s within %v", want, timeout)
}

func isPublicIPReservedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "reserved", "created":
		return true
	default:
		return false
	}
}

func isPublicIPAttachedStatus(status string) bool {
	return strings.ToLower(strings.TrimSpace(status)) == "attached"
}

// ErrPublicIPNotAttached is returned when a policy is created against a
// public IP that is not bound to a VM, load balancer, or baremetal server.
var ErrPublicIPNotAttached = errors.New("public IP is not attached to a VM, load balancer, or baremetal")

// IsPublicIPAttached reports whether a public IP status allows policy rules.
func IsPublicIPAttached(status string) bool {
	return isPublicIPAttachedStatus(status)
}

// PublicIPBoundToResource reports whether the public IP is attached to a
// compute, load balancer, or baremetal port (status attached plus port or VIP).
func PublicIPBoundToResource(ip *models.PublicIP) bool {
	if ip == nil {
		return false
	}
	if !IsPublicIPAttached(ip.Status) {
		return false
	}
	return ip.PortID != 0 || strings.TrimSpace(ip.TargetVIP) != ""
}

// RequirePublicIPAttached loads a public IP and errors unless it is attached
// to a VM, load balancer, or baremetal server.
func (c *Client) RequirePublicIPAttached(ctx context.Context, uuid string) (*models.PublicIP, error) {
	ip, err := c.GetPublicIP(ctx, uuid)
	if err != nil {
		return nil, err
	}
	if ip == nil {
		return nil, fmt.Errorf("public IP %q not found", uuid)
	}
	if !PublicIPBoundToResource(ip) {
		return ip, fmt.Errorf("%w: %q is in %q state (port_id=%d target_vip=%q); attach first then add policy",
			ErrPublicIPNotAttached, uuid, ip.Status, ip.PortID, ip.TargetVIP)
	}
	return ip, nil
}
