package client

import (
	"context"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// baremetalBasePath returns the base path for baremetal server endpoints.
// Scoped by domain (organization) and project, same as lbBasePath().
func (c *Client) baremetalBasePath() string {
	return fmt.Sprintf("/api/baremetal-manager/v1/domain/%s/project/%s",
		c.Organization, c.ProjectName)
}

// BaremetalStateReady is the only server state usable as a pool member.
const BaremetalStateReady = "Ready"

// ListBaremetals retrieves all baremetal servers for the domain/project.
func (c *Client) ListBaremetals(ctx context.Context) ([]models.Baremetal, error) {
	var response models.BaremetalListResponse
	err := c.Get(ctx, fmt.Sprintf("%s/servers", c.baremetalBasePath()), &response)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

// GetBaremetal retrieves a single baremetal server by name. The detail endpoint
// is scoped by the server's availability zone (ce-availability-zone header) and
// carries the backend port id used for load balancer pool membership.
func (c *Client) GetBaremetal(ctx context.Context, name, az string) (*models.Baremetal, error) {
	scopedClient := c.WithAvailabilityZone(az)

	var bm models.Baremetal
	err := scopedClient.Get(ctx, fmt.Sprintf("%s/server/%s", c.baremetalBasePath(), name), &bm)
	if err != nil {
		return nil, err
	}
	return &bm, nil
}

// ResolveBaremetalNode resolves a baremetal pool member by UUID or name and
// returns the full server (including its backend port id from the detail
// endpoint). Only servers in the "Ready" state are usable.
func (c *Client) ResolveBaremetalNode(ctx context.Context, id, name string) (*models.Baremetal, error) {
	servers, err := c.ListBaremetals(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list baremetal servers: %w", err)
	}
	tflog.Debug(ctx, "ResolveBaremetalNode: listing baremetal servers", map[string]interface{}{
		"server_count":  len(servers),
		"searched_id":   id,
		"searched_name": name,
	})

	var match *models.Baremetal
	for i := range servers {
		s := &servers[i]
		if (id != "" && s.UUID == id) || (name != "" && s.Name == name) {
			match = s
			break
		}
	}
	if match == nil {
		if id != "" {
			return nil, fmt.Errorf("baremetal server with id %q not found", id)
		}
		return nil, fmt.Errorf("baremetal server with name %q not found", name)
	}
	if match.State != BaremetalStateReady {
		return nil, fmt.Errorf("baremetal server %q is in state %q, only %q servers can be used as pool members",
			match.Name, match.State, BaremetalStateReady)
	}

	// Fetch the detail (by name, scoped to the server's AZ) to obtain the port id.
	detail, err := c.GetBaremetal(ctx, match.Name, match.AvailabilityZone)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch baremetal server %q detail: %w", match.Name, err)
	}
	// The list response carries fields the detail may omit; keep list values as
	// the source of truth and layer the detail's port id on top.
	detail.Name = match.Name
	detail.UUID = match.UUID
	if len(detail.IPAddr) == 0 {
		detail.IPAddr = match.IPAddr
	}
	if detail.AvailabilityZone == "" {
		detail.AvailabilityZone = match.AvailabilityZone
	}

	// PortID becomes the node's backend_port_id, sourced from networkInfo.portId
	// in the /server/{name} detail response. A zero value means the detail did
	// not carry that field, which makes the backend reject the member with
	// "Missing required fields in pool member node". Surface it loudly rather
	// than silently sending backend_port_id: 0.
	detail.PortID = int(detail.NetworkInfo.PortID)
	if detail.PortID == 0 {
		tflog.Warn(ctx, "baremetal server detail returned no networkInfo.portId; backend_port_id will be 0 and the LB pool member will be rejected.", map[string]interface{}{
			"server_name": match.Name,
			"server_uuid": match.UUID,
		})
	}
	return detail, nil
}
