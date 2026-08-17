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

type sourceOfTruthPublicIPPolicyService struct {
	CreateNew bool   `json:"create_new"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type publicIPPolicySourceEntry struct {
	CreateNew  bool   `json:"create_new"`
	IPCIDR     string `json:"ip_cidr,omitempty"`
	SourceType string `json:"source_type"`
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

	policySource := make([]publicIPPolicySourceEntry, 0, len(req.SourceConfig))
	for _, source := range req.SourceConfig {
		sourceType := strings.TrimSpace(strings.ToLower(source.SourceType))
		ipCIDR := strings.TrimSpace(source.IPCIDR)

		if sourceType == "" {
			if ipCIDR == "" || strings.EqualFold(ipCIDR, "any") || strings.EqualFold(ipCIDR, "all") {
				sourceType = "all"
			} else {
				sourceType = "ip_cidr"
			}
		}

		if sourceType == "ip_cidr" && ipCIDR == "" {
			continue
		}

		createNew := true
		if source.CreateNew != nil {
			createNew = *source.CreateNew
		}

		entry := publicIPPolicySourceEntry{
			CreateNew:  createNew,
			SourceType: sourceType,
		}
		if sourceType == "ip_cidr" {
			entry.IPCIDR = ipCIDR
		}

		policySource = append(policySource, entry)
	}

	if len(policySource) == 0 {
		source := strings.TrimSpace(strings.ToLower(req.Source))
		if source == "" || source == "any" || source == "all" {
			policySource = []publicIPPolicySourceEntry{{
				CreateNew:  true,
				SourceType: "all",
			}}
		} else {
			policySource = []publicIPPolicySourceEntry{{
				CreateNew:  true,
				SourceType: "ip_cidr",
				IPCIDR:     strings.TrimSpace(req.Source),
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
				time.Sleep(2 * time.Second)
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

		time.Sleep(2 * time.Second)
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
				time.Sleep(5 * time.Second)
				continue
			}
			if isAPIErrorStatus(err, 500) || isAPIErrorStatus(err, 503) {
				time.Sleep(5 * time.Second)
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

		time.Sleep(5 * time.Second)
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
