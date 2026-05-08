package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetSubnet retrieves a subnet by ID
func (c *Client) GetSubnet(ctx context.Context, networkID, subnetID string) (*models.Subnet, error) {
	var subnet models.Subnet
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s/subnet/%s", c.Organization, c.ProjectName, networkID, subnetID)
	err := c.Get(ctx, path, &subnet)
	if err != nil {
		return nil, err
	}
	return &subnet, nil
}

// ListSubnets retrieves all subnets in a network
func (c *Client) ListSubnets(ctx context.Context, networkID string) (*models.SubnetListResponse, error) {
	var response models.SubnetListResponse
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s/subnets?limit=1000", c.Organization, c.ProjectName, networkID)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DefaultSubnetTimeout is the default timeout for subnet operations
const DefaultSubnetTimeout = 10 * time.Minute

// CreateSubnet creates a new subnet and waits for it to reach 'Allocated' state
func (c *Client) CreateSubnet(ctx context.Context, networkID string, req *models.CreateSubnetRequest) (*models.Subnet, error) {
	return c.CreateSubnetWithTimeout(ctx, networkID, req, DefaultSubnetTimeout)
}

// CreateSubnetWithTimeout creates a new subnet and waits for it to reach 'Allocated' state with configurable timeout
func (c *Client) CreateSubnetWithTimeout(ctx context.Context, networkID string, req *models.CreateSubnetRequest, timeout time.Duration) (*models.Subnet, error) {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s/subnet", c.Organization, c.ProjectName, networkID)

	err := c.Post(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	// The API returns empty response on success, so we need to find the subnet by name
	var subnet *models.Subnet
	for time.Now().Before(deadline) {
		subnets, err := c.ListSubnets(ctx, networkID)
		if err != nil {
			return nil, fmt.Errorf("failed to list subnets after creation: %w", err)
		}

		for _, s := range subnets.Items {
			if s.Name == req.Name {
				subnet = &s
				break
			}
		}

		if subnet != nil {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if subnet == nil {
		return nil, fmt.Errorf("subnet created but not found in list after %v", timeout)
	}

	// Wait for subnet to reach 'Allocated' state or fail
	for time.Now().Before(deadline) {
		currentSubnet, err := c.GetSubnet(ctx, networkID, subnet.SubnetID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subnet state: %w", err)
		}

		switch currentSubnet.State {
		case "Allocated":
			return currentSubnet, nil
		case "Failed":
			errMsg := currentSubnet.ErrorMessage
			if errMsg == "" {
				errMsg = "subnet creation failed"
			}
			return nil, fmt.Errorf("subnet creation failed: %s", errMsg)
		}

		// Still in progress (e.g., "Allocating", "Pending", etc.)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}

	return nil, fmt.Errorf("subnet creation timed out after %v waiting for 'Allocated' state, current state: %s", timeout, subnet.State)
}

// UpdateSubnet updates an existing subnet
func (c *Client) UpdateSubnet(ctx context.Context, networkID, subnetID string, req *models.UpdateSubnetRequest) (*models.Subnet, error) {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s/subnet/%s", c.Organization, c.ProjectName, networkID, subnetID)

	err := c.Put(ctx, path, req, nil)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated subnet
	return c.GetSubnet(ctx, networkID, subnetID)
}

// IsNotFoundError checks if the error indicates a 404 not found response
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for *APIError type
	if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
		return true
	}
	// Fallback: check error message for HTTP 404 (when JSON parsing fails)
	errMsg := err.Error()
	return strings.Contains(errMsg, "HTTP 404") || strings.Contains(errMsg, "404 Not Found")
}

// DeleteSubnet deletes a subnet
func (c *Client) DeleteSubnet(ctx context.Context, networkID, subnetID string) error {
	return c.DeleteSubnetWithTimeout(ctx, networkID, subnetID, DefaultSubnetTimeout)
}

// DeleteSubnetWithTimeout deletes a subnet with configurable timeout
func (c *Client) DeleteSubnetWithTimeout(ctx context.Context, networkID, subnetID string, timeout time.Duration) error {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s/subnet/%s", c.Organization, c.ProjectName, networkID, subnetID)
	err := c.Delete(ctx, path)
	if err != nil {
		// If subnet is already not found, consider it deleted
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	// Wait for subnet to be deleted (poll until 404)
	for time.Now().Before(deadline) {
		_, err := c.GetSubnet(ctx, networkID, subnetID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}

	return fmt.Errorf("subnet deletion timed out after %v", timeout)
}
