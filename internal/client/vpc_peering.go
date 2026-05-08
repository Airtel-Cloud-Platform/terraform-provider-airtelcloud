package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetVPCPeering retrieves a VPC peering connection by ID
func (c *Client) GetVPCPeering(ctx context.Context, id string) (*models.VPCPeering, error) {
	var peering models.VPCPeering
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/vpc-peering/%s", c.Organization, c.ProjectName, id)
	err := c.Get(ctx, path, &peering)
	if err != nil {
		return nil, err
	}
	return &peering, nil
}

// ListVPCPeerings retrieves all VPC peering connections
func (c *Client) ListVPCPeerings(ctx context.Context) (*models.VPCPeeringListResponse, error) {
	var response models.VPCPeeringListResponse
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/vpc-peerings?region=%s", c.Organization, c.ProjectName, c.Region)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DefaultVPCPeeringTimeout is the default timeout for VPC peering create/delete operations.
const DefaultVPCPeeringTimeout = 5 * time.Minute

// CreateVPCPeering creates a new VPC peering connection using the default timeout.
func (c *Client) CreateVPCPeering(ctx context.Context, req *models.CreateVPCPeeringRequest) (*models.VPCPeering, error) {
	return c.CreateVPCPeeringWithTimeout(ctx, req, DefaultVPCPeeringTimeout)
}

// CreateVPCPeeringWithTimeout creates a new VPC peering connection with a configurable timeout.
// The API returns an empty/incomplete response, so we list and find by name after creation.
func (c *Client) CreateVPCPeeringWithTimeout(ctx context.Context, req *models.CreateVPCPeeringRequest, timeout time.Duration) (*models.VPCPeering, error) {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/vpc-peering", c.Organization, c.ProjectName)

	err := c.Post(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}

	// The API returns empty response on success, so find the peering by name
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		peerings, err := c.ListVPCPeerings(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list VPC peerings after creation: %w", err)
		}

		for _, p := range peerings.Items {
			if p.Name == req.Name {
				return &p, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	return nil, fmt.Errorf("VPC peering created but not found in list after %v", timeout)
}

// DeleteVPCPeering deletes a VPC peering connection using the default timeout.
func (c *Client) DeleteVPCPeering(ctx context.Context, id string) error {
	return c.DeleteVPCPeeringWithTimeout(ctx, id, DefaultVPCPeeringTimeout)
}

// DeleteVPCPeeringWithTimeout deletes a VPC peering connection and waits for deletion
// with a configurable timeout.
func (c *Client) DeleteVPCPeeringWithTimeout(ctx context.Context, id string, timeout time.Duration) error {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/vpc-peering/%s", c.Organization, c.ProjectName, id)
	err := c.Delete(ctx, path)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}

	// Wait for VPC peering to be deleted (poll until 404)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := c.GetVPCPeering(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	return fmt.Errorf("VPC peering deletion timed out after %v", timeout)
}
