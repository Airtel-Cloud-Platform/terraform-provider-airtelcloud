package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetFileStorageVolume retrieves a file storage volume by name
func (c *Client) GetFileStorageVolume(ctx context.Context, name, availabilityZone string) (*models.FileStorageVolume, error) {
	var volume models.FileStorageVolume
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/volume/%s?availabilityZone=%s", c.Organization, c.ProjectName, name, availabilityZone)
	err := c.Get(ctx, path, &volume)
	if err != nil {
		return nil, err
	}
	return &volume, nil
}

// ListFileStorageVolumes retrieves all file storage volumes in a project
func (c *Client) ListFileStorageVolumes(ctx context.Context) (*models.FileStorageVolumeListResponse, error) {
	var response models.FileStorageVolumeListResponse
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/volumes", c.Organization, c.ProjectName)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateFileStorageVolume creates a new file storage volume
func (c *Client) CreateFileStorageVolume(ctx context.Context, req *models.CreateFileStorageVolumeRequest) error {
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/volume", c.Organization, c.ProjectName)
	err := c.Post(ctx, path, req, nil)
	if err != nil {
		return err
	}
	return nil
}

// UpdateFileStorageVolume updates an existing file storage volume
func (c *Client) UpdateFileStorageVolume(ctx context.Context, name string, req *models.UpdateFileStorageVolumeRequest) error {
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/volume/%s", c.Organization, c.ProjectName, name)
	err := c.Put(ctx, path, req, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteFileStorageVolume deletes a file storage volume by name
func (c *Client) DeleteFileStorageVolume(ctx context.Context, name, availabilityZone string) error {
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/volume/%s?availabilityZone=%s", c.Organization, c.ProjectName, name, availabilityZone)
	return c.Delete(ctx, path)
}

// WaitForFileStorageVolumeReady waits for a file storage volume to become ready
func (c *Client) WaitForFileStorageVolumeReady(ctx context.Context, name, availabilityZone string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		volume, err := c.GetFileStorageVolume(ctx, name, availabilityZone)
		if err != nil {
			return err
		}

		switch volume.State {
		case models.FileStorageStateActive:
			return nil
		case models.FileStorageStateCreateFailed, models.FileStorageStateUpdateFailed, models.FileStorageStateDeleteFailed:
			return fmt.Errorf("file storage volume %s failed with state: %s, error: %s", name, volume.State, volume.FailedStateError)
		}

		// Wait before next check
		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("timeout waiting for file storage volume %s to become ready", name)
}

// GetFileStorageExportPath retrieves an NFS export path by path ID
func (c *Client) GetFileStorageExportPath(ctx context.Context, pathID string) (*models.FileStorageExportPath, error) {
	var exportPath models.FileStorageExportPath
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/path/%s", c.Organization, c.ProjectName, pathID)
	err := c.Get(ctx, path, &exportPath)
	if err != nil {
		return nil, err
	}
	return &exportPath, nil
}

// ListFileStorageExportPaths retrieves all NFS export paths for a volume
func (c *Client) ListFileStorageExportPaths(ctx context.Context, volume string) (*models.FileStorageExportPathListResponse, error) {
	var response models.FileStorageExportPathListResponse
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/paths?volume=%s", c.Organization, c.ProjectName, volume)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateFileStorageExportPath creates a new NFS export path for a file storage volume
func (c *Client) CreateFileStorageExportPath(ctx context.Context, req *models.CreateFileStorageExportPathRequest) (*models.FileStorageExportPathKey, error) {
	var response models.FileStorageExportPathKey
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/path", c.Organization, c.ProjectName)
	err := c.Post(ctx, path, req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// UpdateFileStorageExportPath updates an existing NFS export path
func (c *Client) UpdateFileStorageExportPath(ctx context.Context, req *models.UpdateFileStorageExportPathRequest) error {
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/path", c.Organization, c.ProjectName)
	err := c.Put(ctx, path, req, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeleteFileStorageExportPath deletes an NFS export path by path ID
func (c *Client) DeleteFileStorageExportPath(ctx context.Context, pathID, availabilityZone string) error {
	path := fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/file-storage/path/%s?availabilityZone=%s", c.Organization, c.ProjectName, pathID, availabilityZone)
	return c.Delete(ctx, path)
}
