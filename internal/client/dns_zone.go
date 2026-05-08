package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetDNSZone retrieves a DNS zone by ID
func (c *Client) GetDNSZone(ctx context.Context, zoneID string) (*models.DNSZone, error) {
	var zone models.DNSZone
	err := c.Get(ctx, fmt.Sprintf("/v1/zones/%s", zoneID), &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// ListDNSZones retrieves all DNS zones
func (c *Client) ListDNSZones(ctx context.Context) (*models.DNSZoneListResponse, error) {
	var response models.DNSZoneListResponse
	err := c.Get(ctx, "/v1/zones?limit=1000", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateDNSZone creates a new DNS zone
func (c *Client) CreateDNSZone(ctx context.Context, req *models.CreateDNSZoneRequest) (*models.DNSZone, error) {
	var zone models.DNSZone

	err := c.Post(ctx, "/v1/zones", req, &zone)
	if err != nil {
		return nil, err
	}

	// If the response doesn't contain the zone data, fetch it
	if zone.UUID == "" {
		// The API might return just a success message, so we need to find the zone by name
		zones, err := c.ListDNSZones(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list zones after creation: %w", err)
		}
		for _, z := range zones.Items {
			if z.ZoneName == req.ZoneName {
				return &z, nil
			}
		}
		return nil, fmt.Errorf("zone created but not found in list")
	}

	return &zone, nil
}

// UpdateDNSZone updates an existing DNS zone
func (c *Client) UpdateDNSZone(ctx context.Context, zoneID string, req *models.UpdateDNSZoneRequest) (*models.DNSZone, error) {
	var zone models.DNSZone

	err := c.Patch(ctx, fmt.Sprintf("/v1/zones/%s", zoneID), req, &zone)
	if err != nil {
		return nil, err
	}

	// If the response doesn't contain the zone data, fetch it
	if zone.UUID == "" {
		return c.GetDNSZone(ctx, zoneID)
	}

	return &zone, nil
}

// DeleteDNSZone deletes a DNS zone
func (c *Client) DeleteDNSZone(ctx context.Context, zoneID string) error {
	err := c.Delete(ctx, fmt.Sprintf("/v1/zones/%s", zoneID))
	if err != nil {
		return err
	}

	// Wait for zone to be deleted (poll until 404)
	for i := 0; i < 60; i++ {
		_, err := c.GetDNSZone(ctx, zoneID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("DNS zone deletion timed out")
}
