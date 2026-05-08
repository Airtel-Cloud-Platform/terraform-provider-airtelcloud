package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ListFlavors retrieves all compute flavors
func (c *Client) ListFlavors(ctx context.Context) ([]models.Flavor, error) {
	var flavors []models.Flavor
	err := c.Get(ctx, fmt.Sprintf("%s/flavors/", c.computeBasePath()), &flavors)
	if err != nil {
		return nil, err
	}
	return flavors, nil
}

// ListImages retrieves all compute images
func (c *Client) ListImages(ctx context.Context) ([]models.Image, error) {
	var images []models.Image
	err := c.Get(ctx, fmt.Sprintf("%s/images/", c.computeBasePath()), &images)
	if err != nil {
		return nil, err
	}
	return images, nil
}

// ListKeypairs retrieves all SSH keypairs
func (c *Client) ListKeypairs(ctx context.Context) ([]models.Keypair, error) {
	var keypairs []models.Keypair
	err := c.Get(ctx, fmt.Sprintf("%s/keypairs/", c.computeBasePath()), &keypairs)
	if err != nil {
		return nil, err
	}
	return keypairs, nil
}

// ListSecurityGroups retrieves all security groups
func (c *Client) ListSecurityGroups(ctx context.Context) ([]models.SecurityGroup, error) {
	var groups []models.SecurityGroup
	err := c.Get(ctx, fmt.Sprintf("%s/security-groups/", c.computeBasePath()), &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// ResolveFlavorID resolves a flavor name to its ID
func (c *Client) ResolveFlavorID(ctx context.Context, name string) (string, error) {
	flavors, err := c.ListFlavors(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list flavors: %w", err)
	}
	for _, f := range flavors {
		if f.Name == name {
			return strconv.Itoa(f.ID), nil
		}
	}
	return "", fmt.Errorf("flavor with name %q not found", name)
}

// ResolveImageID resolves an image name to its ID
func (c *Client) ResolveImageID(ctx context.Context, name string) (string, error) {
	images, err := c.ListImages(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}
	for _, img := range images {
		if img.Name == name {
			return strconv.Itoa(img.ID), nil
		}
	}
	return "", fmt.Errorf("image with name %q not found", name)
}

// ResolveKeypairID resolves a keypair name to its ID
func (c *Client) ResolveKeypairID(ctx context.Context, name string) (string, error) {
	keypairs, err := c.ListKeypairs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list keypairs: %w", err)
	}
	for _, kp := range keypairs {
		if kp.Name == name {
			return strconv.Itoa(kp.ID), nil
		}
	}
	return "", fmt.Errorf("keypair with name %q not found", name)
}

// ResolveSecurityGroupID resolves a security group name to its ID
func (c *Client) ResolveSecurityGroupID(ctx context.Context, name string) (int, error) {
	groups, err := c.ListSecurityGroups(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list security groups: %w", err)
	}
	for _, sg := range groups {
		if sg.Name == name {
			return sg.ID, nil
		}
	}
	return 0, fmt.Errorf("security group with name %q not found", name)
}

// ResolveVPCID resolves a VPC name to its ID
func (c *Client) ResolveVPCID(ctx context.Context, name string) (string, error) {
	resp, err := c.ListVPCs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list VPCs: %w", err)
	}
	tflog.Debug(ctx, "ResolveVPCID: listing VPCs", map[string]interface{}{
		"vpc_count":     len(resp.Items),
		"searched_name": name,
	})
	for _, vpc := range resp.Items {
		if vpc.Name == name {
			if vpc.ID == "" {
				return "", fmt.Errorf("VPC %q found but has empty ID (API response field mismatch)", name)
			}
			tflog.Debug(ctx, "ResolveVPCID: resolved VPC", map[string]interface{}{
				"name": name,
				"id":   vpc.ID,
			})
			return vpc.ID, nil
		}
	}
	return "", fmt.Errorf("VPC with name %q not found", name)
}

// ResolveSubnetID resolves a subnet name to its ID within a given VPC (network)
func (c *Client) ResolveSubnetID(ctx context.Context, vpcID, name string) (string, error) {
	resp, err := c.ListSubnets(ctx, vpcID)
	if err != nil {
		return "", fmt.Errorf("failed to list subnets: %w", err)
	}
	for _, s := range resp.Items {
		if s.Name == name {
			return s.SubnetID, nil
		}
	}
	return "", fmt.Errorf("subnet with name %q not found in VPC %s", name, vpcID)
}
