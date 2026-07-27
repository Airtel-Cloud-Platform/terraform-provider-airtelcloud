package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// CreateVirtualServer creates a virtual server on an LB service. The payload is
// sent as a URL-encoded form body, with the pool-member collection carried in a
// single "nodes" field holding a JSON array (mirroring the security-group-rules
// bulk endpoint's "security_group_data" contract).
func (c *Client) CreateVirtualServer(ctx context.Context, lbServiceID string, formData map[string]interface{}) (*models.LBVirtualServer, error) {
	path := fmt.Sprintf("%s/%s/virtual-servers", c.lbBasePath(), lbServiceID)

	var vs models.LBVirtualServer
	err := c.PostURLEncodedForm(ctx, path, formData, &vs)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

// GetVirtualServer retrieves a virtual server by ID
func (c *Client) GetVirtualServer(ctx context.Context, lbServiceID, vsID string) (*models.LBVirtualServer, error) {
	var vs models.LBVirtualServer
	err := c.Get(ctx, fmt.Sprintf("%s/%s/virtual-servers/%s", c.lbBasePath(), lbServiceID, vsID), &vs)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

// ListVirtualServers lists all virtual servers for an LB service
func (c *Client) ListVirtualServers(ctx context.Context, lbServiceID string) ([]models.LBVirtualServer, error) {
	var servers []models.LBVirtualServer
	err := c.Get(ctx, fmt.Sprintf("%s/%s/virtual-servers", c.lbBasePath(), lbServiceID), &servers)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// UpdateVirtualServer updates a virtual server using PATCH with a URL-encoded form body.
func (c *Client) UpdateVirtualServer(ctx context.Context, lbServiceID, vsID string, formData map[string]interface{}) (*models.LBVirtualServer, error) {
	path := fmt.Sprintf("%s/%s/virtual-servers/%s", c.lbBasePath(), lbServiceID, vsID)

	var vs models.LBVirtualServer
	err := c.PatchURLEncodedForm(ctx, path, formData, &vs)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

// DeleteVirtualServer deletes a virtual server
func (c *Client) DeleteVirtualServer(ctx context.Context, lbServiceID, vsID string) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%s/virtual-servers/%s", c.lbBasePath(), lbServiceID, vsID))
}

// WaitForVirtualServerReady polls until the virtual server is ready
func (c *Client) WaitForVirtualServerReady(ctx context.Context, lbServiceID, vsID string, timeout time.Duration) (*models.LBVirtualServer, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		vs, err := c.GetVirtualServer(ctx, lbServiceID, vsID)
		if err != nil {
			return nil, err
		}

		switch vs.Status {
		case "Active", "active", "ACTIVE":
			return vs, nil
		case "Error", "error", "ERROR":
			return nil, fmt.Errorf("virtual server entered error state")
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("virtual server did not become ready within %v", timeout)
}

// WaitForVirtualServerDeleted polls until the virtual server is deleted (404)
func (c *Client) WaitForVirtualServerDeleted(ctx context.Context, lbServiceID, vsID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := c.GetVirtualServer(ctx, lbServiceID, vsID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}

		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("virtual server deletion timed out after %v", timeout)
}

// ResolveComputeNode resolves a VM pool member by id (UUID) or name and returns
// the full compute (including its Ports) needed to build the node payload.
func (c *Client) ResolveComputeNode(ctx context.Context, computeID, computeName string) (*models.Compute, error) {
	id := computeID
	if id == "" {
		resolved, err := c.ResolveComputeID(ctx, computeName)
		if err != nil {
			return nil, err
		}
		id = resolved
	}
	return c.GetCompute(ctx, id)
}

// BackendPortIDForIP scans a compute's ports for the one whose fixed_ips contains
// ip and returns its port id (used as a node's backend_port_id).
func BackendPortIDForIP(ports []models.Port, ip string) (int, error) {
	for _, p := range ports {
		for _, fixedIP := range p.FixedIPs {
			if fixedIP == ip {
				return p.ID, nil
			}
		}
	}
	return 0, fmt.Errorf("no port found with fixed_ip matching %s", ip)
}

// VirtualServerCreateParams holds the top-level create parameters for a virtual
// server, mirroring the API's form fields.
type VirtualServerCreateParams struct {
	Name               string
	Protocol           string
	VPCID              string
	RoutingAlgorithm   string
	MonitorProtocol    string
	CertificateID      string
	PoolName           string
	MonitorName        string
	PersistenceType    string
	VipPortID          int
	Port               int
	Interval           int
	MonitorPort        int
	Timeout            int
	PersistenceEnabled bool
	XForwardedFor      bool
	RedirectHTTPS      bool
	Nodes              []models.VirtualServerNode
}

// BuildVirtualServerFormData builds the URL-encoded form body for creating a
// virtual server. The pool-member collection is carried as repeated "nodes"
// fields in the body, each holding a single node's JSON object (confirmed
// against the live API: body is ...&nodes=<json-obj>&nodes=<json-obj>&...).
func BuildVirtualServerFormData(p VirtualServerCreateParams) map[string]interface{} {
	formData := map[string]interface{}{}

	if p.Name != "" {
		formData["name"] = p.Name
	}
	formData["vip_port_id"] = strconv.Itoa(p.VipPortID)
	formData["protocol"] = p.Protocol
	formData["port"] = strconv.Itoa(p.Port)
	formData["routing_algorithm"] = p.RoutingAlgorithm
	formData["vpc_id"] = p.VPCID
	formData["interval"] = strconv.Itoa(p.Interval)

	if p.PoolName != "" {
		formData["pool_name"] = p.PoolName
	}
	if p.MonitorName != "" {
		formData["monitor_name"] = p.MonitorName
	}
	if p.MonitorProtocol != "" {
		formData["monitor_protocol"] = p.MonitorProtocol
	}
	if p.MonitorPort > 0 {
		formData["monitor_port"] = strconv.Itoa(p.MonitorPort)
	}
	if p.Timeout > 0 {
		formData["timeout"] = strconv.Itoa(p.Timeout)
	}

	if p.PersistenceEnabled {
		formData["persistence_enabled"] = "true"
		if p.PersistenceType != "" {
			formData["persistence_type"] = p.PersistenceType
		}
	}
	if p.XForwardedFor {
		formData["x_forwarded_for"] = "true"
	}
	if p.RedirectHTTPS {
		formData["redirect_https"] = "true"
	}
	if p.CertificateID != "" {
		formData["certificate_id"] = p.CertificateID
	}

	// Carry each node as its own repeated "nodes" form field. A []string value
	// is expanded by doURLEncodedFormRequest into repeated fields, producing
	// ...&nodes=<json-obj>&nodes=<json-obj>&... as the backend expects.
	if len(p.Nodes) > 0 {
		nodeJSONs := make([]string, 0, len(p.Nodes))
		for _, node := range p.Nodes {
			nodeJSON, _ := json.Marshal(node)
			nodeJSONs = append(nodeJSONs, string(nodeJSON))
		}
		formData["nodes"] = nodeJSONs
	}

	return formData
}
