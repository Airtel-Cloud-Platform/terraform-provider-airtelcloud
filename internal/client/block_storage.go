package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) blockStorageBasePath() string {
	return fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/block-storage", c.Organization, c.ProjectName)
}

func (c *Client) blockStorageVolumePath(name, availabilityZone string) string {
	path := fmt.Sprintf("%s/volume/%s", c.blockStorageBasePath(), url.PathEscape(name))
	q := url.Values{}
	if availabilityZone != "" {
		q.Set("availabilityZone", availabilityZone)
		q.Set("az", availabilityZone)
	}
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return path
}

func (c *Client) blockStorageClient(availabilityZone string) *Client {
	if availabilityZone == "" {
		return c
	}
	return c.WithAvailabilityZone(availabilityZone)
}

// GetBlockStorageVolume retrieves a block storage volume by name.
func (c *Client) GetBlockStorageVolume(ctx context.Context, name, availabilityZone string) (*models.BlockStorageVolume, error) {
	var volume models.BlockStorageVolume
	err := c.blockStorageClient(availabilityZone).Get(ctx, c.blockStorageVolumePath(name, availabilityZone), &volume)
	if err != nil {
		return nil, err
	}
	if volume.IsDeleted {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("block storage volume %s is deleted", name)}
	}
	return &volume, nil
}

// ListBlockStorageVolumes retrieves all block storage volumes in a project.
func (c *Client) ListBlockStorageVolumes(ctx context.Context) (*models.BlockStorageVolumeListResponse, error) {
	var response models.BlockStorageVolumeListResponse
	err := c.Get(ctx, c.blockStorageBasePath()+"/volumes", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateBlockStorageVolume creates a baremetal block storage volume.
func (c *Client) CreateBlockStorageVolume(ctx context.Context, req *models.CreateBlockStorageVolumeRequest) error {
	return c.blockStorageClient(req.AvailabilityZone).Post(ctx, c.blockStorageBasePath()+"/volume", req, nil)
}

// UpdateBlockStorageVolume updates an existing block storage volume.
func (c *Client) UpdateBlockStorageVolume(ctx context.Context, name string, req *models.UpdateBlockStorageVolumeRequest) error {
	az := req.AvailabilityZone
	return c.blockStorageClient(az).Put(ctx, c.blockStorageVolumePath(name, az), req, nil)
}

// DeleteBlockStorageVolume deletes a volume by name (and volId if still present) and waits until it is gone.
func (c *Client) DeleteBlockStorageVolume(ctx context.Context, name, availabilityZone, volID string) error {
	if err := c.deleteBlockStorageVolumeOnce(ctx, name, availabilityZone); err != nil && !IsNotFoundError(err) {
		return err
	}

	gone, err := c.blockStorageVolumeGone(ctx, name, availabilityZone)
	if err != nil {
		return err
	}
	if gone {
		return nil
	}

	if volID != "" && volID != name {
		if err := c.deleteBlockStorageVolumeOnce(ctx, volID, availabilityZone); err != nil && !IsNotFoundError(err) {
			return err
		}
	}

	return c.WaitForBlockStorageVolumeDeleted(ctx, name, availabilityZone, 10*time.Minute)
}

func (c *Client) deleteBlockStorageVolumeOnce(ctx context.Context, id, availabilityZone string) error {
	return c.blockStorageClient(availabilityZone).Delete(ctx, c.blockStorageVolumePath(id, availabilityZone))
}

func (c *Client) blockStorageVolumeGone(ctx context.Context, name, availabilityZone string) (bool, error) {
	_, err := c.GetBlockStorageVolume(ctx, name, availabilityZone)
	if err != nil {
		if IsNotFoundError(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// WaitForBlockStorageVolumeDeleted waits until GET returns 404 or isDeleted.
func (c *Client) WaitForBlockStorageVolumeDeleted(ctx context.Context, name, availabilityZone string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gone, err := c.blockStorageVolumeGone(ctx, name, availabilityZone)
		if err != nil {
			return err
		}
		if gone {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	return fmt.Errorf("timeout waiting for block storage volume %s to be deleted", name)
}

// WaitForBlockStorageVolumeReady waits for a volume to become Active.
func (c *Client) WaitForBlockStorageVolumeReady(ctx context.Context, name, availabilityZone string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		volume, err := c.GetBlockStorageVolume(ctx, name, availabilityZone)
		if err != nil {
			return err
		}

		switch volume.State {
		case models.BlockStorageStateActive:
			return nil
		case models.BlockStorageStateCreateFailed, models.BlockStorageStateUpdateFailed, models.BlockStorageStateDeleteFailed:
			return fmt.Errorf("block storage volume %s failed with state: %s, error: %s", name, volume.State, volume.FailedStateError)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	return fmt.Errorf("timeout waiting for block storage volume %s to become ready", name)
}
