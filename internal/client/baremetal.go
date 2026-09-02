package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

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

// AllocateBaremetal allocates a new baremetal server.
func (c *Client) AllocateBaremetal(ctx context.Context, req *models.AllocateBaremetalRequest) error {
	return c.Post(ctx, fmt.Sprintf("%s/server", c.baremetalBasePath()), req, nil)
}

// GetBaremetal retrieves a single baremetal server by name. The detail endpoint
// is scoped by the server's availability zone (ce-availability-zone header) and
// carries the backend port id used for load balancer pool membership.
func (c *Client) GetBaremetal(ctx context.Context, name, az string) (*models.Baremetal, error) {
	scopedClient := c.WithAvailabilityZone(az)

	// Swagger defines /server/{name} as a wrapped shape with serverDetails and
	// networkInfo, while some backend responses are flat. Decode both shapes.
	var resp struct {
		models.Baremetal
		ServerDetails models.Baremetal   `json:"serverDetails"`
		NetworkInfo   models.NetworkInfo `json:"networkInfo"`
	}
	err := scopedClient.Get(ctx, fmt.Sprintf("%s/server/%s", c.baremetalBasePath(), name), &resp)
	if err != nil {
		return nil, err
	}

	bm := resp.Baremetal
	if resp.ServerDetails.Name != "" || resp.ServerDetails.UUID != "" {
		bm = resp.ServerDetails
	}
	if bm.NetworkInfo.PortID == 0 && resp.NetworkInfo.PortID != 0 {
		bm.NetworkInfo = resp.NetworkInfo
	}
	return &bm, nil
}

// UpdateBaremetal updates mutable baremetal settings.
func (c *Client) UpdateBaremetal(ctx context.Context, name string, req *models.UpdateBaremetalRequest) error {
	return c.Put(ctx, fmt.Sprintf("%s/server/%s", c.baremetalBasePath(), name), req, nil)
}

// ReleaseBaremetal releases (deletes) a baremetal server by name.
func (c *Client) ReleaseBaremetal(ctx context.Context, name string, opts *models.ReleaseBaremetalOptions) error {
	path := fmt.Sprintf("%s/server/%s", c.baremetalBasePath(), name)
	if opts == nil {
		return c.Delete(ctx, path)
	}

	q := url.Values{}
	if opts.SystemID != "" {
		q.Set("systemId", opts.SystemID)
	}
	if opts.DeleteDisks != nil {
		q.Set("deleteDisks", strconv.FormatBool(*opts.DeleteDisks))
	}
	if opts.SecureErase != nil {
		q.Set("secureErase", strconv.FormatBool(*opts.SecureErase))
	}
	if encoded := q.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	return c.Delete(ctx, path)
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
