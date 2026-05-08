package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// CreateComputeSnapshot creates a snapshot from a compute instance
func (c *Client) CreateComputeSnapshot(ctx context.Context, computeID string) (*models.ComputeSnapshot, error) {
	var snapshot models.ComputeSnapshot
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/snapshot/", c.computeBasePath(), computeID), nil, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// GetComputeSnapshot retrieves a compute snapshot by UUID
func (c *Client) GetComputeSnapshot(ctx context.Context, snapshotUUID string) (*models.ComputeSnapshot, error) {
	var snapshot models.ComputeSnapshot
	err := c.Get(ctx, fmt.Sprintf("%s/snapshot/%s/", c.computeBasePath(), snapshotUUID), &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ListComputeSnapshots retrieves all compute snapshots
func (c *Client) ListComputeSnapshots(ctx context.Context) ([]models.ComputeSnapshot, error) {
	var snapshots []models.ComputeSnapshot
	err := c.Get(ctx, fmt.Sprintf("%s/snapshot/", c.computeBasePath()), &snapshots)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

// DeleteComputeSnapshot deletes a compute snapshot by UUID and polls until 404
func (c *Client) DeleteComputeSnapshot(ctx context.Context, snapshotUUID string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/snapshot/%s/", c.computeBasePath(), snapshotUUID))
	if err != nil {
		return err
	}

	// Poll until 404
	for i := 0; i < 60; i++ {
		_, err := c.GetComputeSnapshot(ctx, snapshotUUID)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("compute snapshot deletion timed out")
}

// WaitForSnapshotReady polls until snapshot reaches a stable status
func (c *Client) WaitForSnapshotReady(ctx context.Context, snapshotUUID string, timeout time.Duration) (*models.ComputeSnapshot, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		snapshot, err := c.GetComputeSnapshot(ctx, snapshotUUID)
		if err != nil {
			return nil, err
		}

		switch snapshot.Status {
		case "active", "available":
			return snapshot, nil
		case "error", "error_deleting":
			return nil, fmt.Errorf("snapshot entered error state: %s", snapshot.Status)
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("snapshot did not become ready within %v", timeout)
}
