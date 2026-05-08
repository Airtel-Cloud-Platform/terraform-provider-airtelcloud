package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetVPC retrieves a VPC by ID
func (c *Client) GetVPC(ctx context.Context, id string) (*models.VPC, error) {
	var vpc models.VPC
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s", c.Organization, c.ProjectName, id)
	err := c.Get(ctx, path, &vpc)
	if err != nil {
		return nil, err
	}
	return &vpc, nil
}

// ListVPCs retrieves all VPCs
func (c *Client) ListVPCs(ctx context.Context) (*models.VPCListResponse, error) {
	var response models.VPCListResponse
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/networks", c.Organization, c.ProjectName)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateVPC creates a new VPC
func (c *Client) CreateVPC(ctx context.Context, req *models.CreateVPCRequest) (*models.VPC, error) {
	var response struct {
		VPC         *models.VPC `json:"vpc"`
		OperationID string      `json:"operation_id,omitempty"`
	}

	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network", c.Organization, c.ProjectName)
	err := c.Post(ctx, path, req, &response)
	if err != nil {
		return nil, err
	}

	// If there's an operation ID, wait for completion
	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to create VPC: %w", err)
		}

		// Retrieve the created VPC
		return c.GetVPC(ctx, response.VPC.ID)
	}

	return response.VPC, nil
}

// UpdateVPC updates an existing VPC
func (c *Client) UpdateVPC(ctx context.Context, id string, req *models.UpdateVPCRequest) (*models.VPC, error) {
	var response struct {
		VPC         *models.VPC `json:"vpc"`
		OperationID string      `json:"operation_id,omitempty"`
	}

	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s", c.Organization, c.ProjectName, id)
	err := c.Put(ctx, path, req, &response)
	if err != nil {
		return nil, err
	}

	// If there's an operation ID, wait for completion
	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to update VPC: %w", err)
		}

		// Retrieve the updated VPC
		return c.GetVPC(ctx, id)
	}

	return response.VPC, nil
}

// DeleteVPC deletes a VPC
func (c *Client) DeleteVPC(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/network-manager/v1/domain/%s/project/%s/network/%s", c.Organization, c.ProjectName, id)
	err := c.Delete(ctx, path)
	if err != nil {
		return err
	}

	// Wait for VPC to be deleted (poll until 404)
	for i := 0; i < 60; i++ {
		_, err := c.GetVPC(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("VPC deletion timed out")
}
