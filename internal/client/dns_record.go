package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetDNSRecord retrieves a DNS record by zone ID and record ID
func (c *Client) GetDNSRecord(ctx context.Context, zoneID, recordID string) (*models.DNSRecord, error) {
	var record models.DNSRecord
	err := c.Get(ctx, fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID), &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListDNSRecords retrieves all DNS records for a zone
func (c *Client) ListDNSRecords(ctx context.Context, zoneID string) (*models.DNSRecordListResponse, error) {
	var response models.DNSRecordListResponse
	err := c.Get(ctx, fmt.Sprintf("/v1/zones/%s/records?limit=1000", zoneID), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateDNSRecord creates a new DNS record in a zone
func (c *Client) CreateDNSRecord(ctx context.Context, zoneID string, req *models.CreateDNSRecordRequest) (*models.DNSRecord, error) {
	var record models.DNSRecord

	err := c.Post(ctx, fmt.Sprintf("/v1/zones/%s/records", zoneID), req, &record)
	if err != nil {
		return nil, err
	}

	// If the response doesn't contain the record data, fetch it
	if record.UUID == "" {
		// The API might return just a success message, so we need to find the record
		records, err := c.ListDNSRecords(ctx, zoneID)
		if err != nil {
			return nil, fmt.Errorf("failed to list records after creation: %w", err)
		}
		// Find the newly created record by matching owner and record type
		for _, r := range records.Items {
			owner := ""
			if req.Owner != nil {
				owner = *req.Owner
			}
			data := ""
			if req.Data != nil {
				data = *req.Data
			}
			if r.Owner == owner && r.RecordType == req.RecordType && r.Data == data {
				return &r, nil
			}
		}
		// Return the first matching record by type if not found exactly
		for _, r := range records.Items {
			if r.RecordType == req.RecordType {
				return &r, nil
			}
		}
		return nil, fmt.Errorf("record created but not found in list")
	}

	return &record, nil
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(ctx context.Context, zoneID, recordID string, req *models.UpdateDNSRecordRequest) (*models.DNSRecord, error) {
	var record models.DNSRecord

	err := c.Patch(ctx, fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID), req, &record)
	if err != nil {
		return nil, err
	}

	// If the response doesn't contain the record data, fetch it
	if record.UUID == "" {
		return c.GetDNSRecord(ctx, zoneID, recordID)
	}

	return &record, nil
}

// DeleteDNSRecord deletes a DNS record
func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	err := c.Delete(ctx, fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID))
	if err != nil {
		return err
	}

	// Wait for record to be deleted (poll until 404)
	for i := 0; i < 60; i++ {
		_, err := c.GetDNSRecord(ctx, zoneID, recordID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("DNS record deletion timed out")
}
