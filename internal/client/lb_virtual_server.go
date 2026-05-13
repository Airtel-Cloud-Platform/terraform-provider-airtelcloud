package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// CreateVirtualServer creates a virtual server on an LB service using query parameters
func (c *Client) CreateVirtualServer(ctx context.Context, lbServiceID string, params url.Values) (*models.LBVirtualServer, error) {
	path := fmt.Sprintf("%s/%s/virtual-servers", c.lbBasePath(), lbServiceID)

	var vs models.LBVirtualServer
	err := c.PostWithQueryParams(ctx, path, params, &vs)
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

// UpdateVirtualServer updates a virtual server using PATCH with query parameters
func (c *Client) UpdateVirtualServer(ctx context.Context, lbServiceID, vsID string, params url.Values) (*models.LBVirtualServer, error) {
	path := fmt.Sprintf("%s/%s/virtual-servers/%s", c.lbBasePath(), lbServiceID, vsID)

	var vs models.LBVirtualServer
	err := c.PatchWithQueryParams(ctx, path, params, &vs)
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

// BuildVirtualServerNodes serializes a slice of VirtualServerNode into url.Values
// Each node is JSON-serialized as a separate query parameter value (collectionFormat: multi)
func BuildVirtualServerNodes(nodes []models.VirtualServerNode) url.Values {
	params := url.Values{}
	for _, node := range nodes {
		nodeJSON, _ := json.Marshal(node)
		params.Add("nodes", string(nodeJSON))
	}
	return params
}

// BuildVirtualServerParams builds the full url.Values for creating a virtual server
func BuildVirtualServerParams(name, protocol, vpcID, routingAlgorithm, monitorProtocol, certificateID string,
	vipPortID, port, interval int,
	persistenceEnabled, xForwardedFor, redirectHTTPS bool,
	persistenceType string,
	nodes []models.VirtualServerNode,
) url.Values {
	params := url.Values{}

	if name != "" {
		params.Set("name", name)
	}
	params.Set("vip_port_id", strconv.Itoa(vipPortID))
	params.Set("protocol", protocol)
	params.Set("port", strconv.Itoa(port))
	params.Set("routing_algorithm", routingAlgorithm)
	params.Set("vpc_id", vpcID)
	params.Set("interval", strconv.Itoa(interval))

	if persistenceEnabled {
		params.Set("persistence_enabled", "true")
		if persistenceType != "" {
			params.Set("persistence_type", persistenceType)
		}
	}
	if xForwardedFor {
		params.Set("x_forwarded_for", "true")
	}
	if redirectHTTPS {
		params.Set("redirect_https", "true")
	}
	if monitorProtocol != "" {
		params.Set("monitor_protocol", monitorProtocol)
	}
	if certificateID != "" {
		params.Set("certificate_id", certificateID)
	}

	// Add nodes as repeated JSON query params
	for _, node := range nodes {
		nodeJSON, _ := json.Marshal(node)
		params.Add("nodes", string(nodeJSON))
	}

	return params
}
