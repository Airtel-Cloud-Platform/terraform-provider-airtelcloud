package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// volumeBasePath returns the v2.1 base path for volume endpoints
func (c *Client) volumeBasePath() string {
	return fmt.Sprintf("/api/v2.1/volumes/domain/%s/project/%s/volumes",
		c.Organization, c.ProjectName)
}

// GetVolume retrieves a volume by UUID
func (c *Client) GetVolume(ctx context.Context, uuid string) (*models.Volume, error) {
	var volume models.Volume
	err := c.Get(ctx, fmt.Sprintf("%s/%s/", c.volumeBasePath(), uuid), &volume)
	if err != nil {
		return nil, err
	}
	if volume.ID == 0 {
		return nil, &APIError{StatusCode: 404, Message: "volume not found"}
	}
	return &volume, nil
}

// ListVolumes retrieves all volumes with optional filtering
func (c *Client) ListVolumes(ctx context.Context, computeID string) ([]models.Volume, error) {
	path := fmt.Sprintf("%s/", c.volumeBasePath())
	if computeID != "" {
		path += "?compute_id=" + computeID
	}

	var volumes []models.Volume
	err := c.Get(ctx, path, &volumes)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

// CreateVolume creates a new volume using create-and-attach endpoint
func (c *Client) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.Volume, error) {
	// Convert struct to form data map
	formData := structToFormData(req)

	var volume models.Volume
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/create-and-attach/", c.volumeBasePath()), formData, &volume)
	if err != nil {
		return nil, err
	}

	return &volume, nil
}

// UpdateVolume extends an existing volume (size increase only)
func (c *Client) UpdateVolume(ctx context.Context, uuid string, req *models.UpdateVolumeRequest) error {
	// Convert struct to form data map
	formData := structToFormData(req)

	return c.PutURLEncodedForm(ctx, fmt.Sprintf("%s/%s/", c.volumeBasePath(), uuid), formData, nil)
}

// DeleteVolume deletes a volume
func (c *Client) DeleteVolume(ctx context.Context, uuid string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s/", c.volumeBasePath(), uuid))
	if err != nil {
		return err
	}

	// Wait for volume to be deleted (poll until 404 or soft-deleted)
	for i := 0; i < 60; i++ {
		volume, err := c.GetVolume(ctx, uuid)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if volume.Deleted != nil || volume.Status == "deleted" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("volume deletion timed out")
}

// AttachVolume attaches a volume to a compute instance
func (c *Client) AttachVolume(ctx context.Context, volumeUUID string, req *models.VolumeAttachRequest) error {
	// Convert struct to form data map
	formData := structToFormData(req)

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/volume_attach/%s/", c.volumeBasePath(), volumeUUID), formData, nil)
}

// DetachVolume detaches a volume from a compute instance
func (c *Client) DetachVolume(ctx context.Context, volumeUUID string, req *models.VolumeDetachRequest) error {
	formData := structToFormData(req)
	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/volume_detach/", c.volumeBasePath(), volumeUUID), formData, nil)
}

// WaitForVolumeDetached polls until the volume has no attachments.
func (c *Client) WaitForVolumeDetached(ctx context.Context, uuid string) error {
	for i := 0; i < 60; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		volume, err := c.GetVolume(ctx, uuid)
		if err != nil {
			return fmt.Errorf("error polling volume during detach: %w", err)
		}
		if len(volume.VolumeAttachments) == 0 {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for volume %s to detach", uuid)
}

// WaitForVolumeAttached polls until the volume is attached to the given compute instance.
func (c *Client) WaitForVolumeAttached(ctx context.Context, uuid, computeID string) error {
	for i := 0; i < 60; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		volume, err := c.GetVolume(ctx, uuid)
		if err != nil {
			return fmt.Errorf("error polling volume during attach: %w", err)
		}
		for _, a := range volume.VolumeAttachments {
			if a.ComputeID == computeID {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for volume %s to attach to %s", uuid, computeID)
}

// GetVolumeAttachedDevices gets the devices a volume is attached to
func (c *Client) GetVolumeAttachedDevices(ctx context.Context, volumeUUID string) ([]models.VolumeAttachment, error) {
	var attachments []models.VolumeAttachment
	err := c.Get(ctx, fmt.Sprintf("%s/%s/attached_devices/", c.volumeBasePath(), volumeUUID), &attachments)
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

// CreateVolumeSnapshot creates a snapshot of a volume
func (c *Client) CreateVolumeSnapshot(ctx context.Context, volumeUUID string, req *models.VolumeSnapshotRequest) error {
	// Convert struct to form data map
	formData := structToFormData(req)

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/snapshots/", c.volumeBasePath(), volumeUUID), formData, nil)
}

// EnableVolumeBackup enables backup for a volume
func (c *Client) EnableVolumeBackup(ctx context.Context, volumeUUID string) error {
	formData := map[string]interface{}{
		"billing_unit": "MRC",
	}

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/backup", c.volumeBasePath(), volumeUUID), formData, nil)
}

// GetVolumeTypes retrieves available volume types, optionally filtered by group
func (c *Client) GetVolumeTypes(ctx context.Context, group string) ([]models.VolumeType, error) {
	path := fmt.Sprintf("%s/volume_types", c.volumeBasePath())
	if group != "" {
		path += "?group=" + group
	}
	var volumeTypes []models.VolumeType
	err := c.Get(ctx, path, &volumeTypes)
	if err != nil {
		return nil, err
	}
	return volumeTypes, nil
}

// GetVolumeTypeByID retrieves a specific volume type by ID
func (c *Client) GetVolumeTypeByID(ctx context.Context, id int) (*models.VolumeType, error) {
	var volumeType models.VolumeType
	err := c.Get(ctx, fmt.Sprintf("%s/volume_types/%d/", c.volumeBasePath(), id), &volumeType)
	if err != nil {
		return nil, err
	}
	return &volumeType, nil
}

// GetVolumeTypeByName retrieves a specific volume type by name
func (c *Client) GetVolumeTypeByName(ctx context.Context, name string) (*models.VolumeType, error) {
	var volumeType models.VolumeType
	err := c.Get(ctx, fmt.Sprintf("%s/volume_types/%s", c.volumeBasePath(), name), &volumeType)
	if err != nil {
		return nil, err
	}
	return &volumeType, nil
}
