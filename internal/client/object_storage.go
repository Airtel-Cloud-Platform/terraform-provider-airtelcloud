package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// GetObjectStorageBucket retrieves an object storage bucket by name
func (c *Client) GetObjectStorageBucket(ctx context.Context, name string) (*models.ObjectStorageBucket, error) {
	var bucket models.ObjectStorageBucket
	path := fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/bucket/%s", c.Organization, c.ProjectName, name)
	err := c.Get(ctx, path, &bucket)
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// ListObjectStorageBuckets retrieves all object storage buckets in the project
func (c *Client) ListObjectStorageBuckets(ctx context.Context) (*models.ObjectStorageBucketListResponse, error) {
	var response models.ObjectStorageBucketListResponse
	path := fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/buckets", c.Organization, c.ProjectName)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateObjectStorageBucket creates a new object storage bucket
func (c *Client) CreateObjectStorageBucket(ctx context.Context, req *models.CreateObjectStorageBucketRequest) (*models.ObjectStorageBucket, error) {
	var response struct {
		Bucket      *models.ObjectStorageBucket `json:"bucket"`
		OperationID string                      `json:"operation_id,omitempty"`
	}

	err := c.Post(ctx, fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/bucket", c.Organization, c.ProjectName), req, &response)
	if err != nil {
		return nil, err
	}

	// If there's an operation ID, wait for completion
	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// If the response included bucket details, return them
	if response.Bucket != nil {
		return response.Bucket, nil
	}

	// Otherwise fetch the bucket by name (the create API may return an empty body)
	return c.GetObjectStorageBucket(ctx, req.Bucket)
}

// UpdateObjectStorageBucket updates an existing object storage bucket
func (c *Client) UpdateObjectStorageBucket(ctx context.Context, name string, req *models.UpdateObjectStorageBucketRequest) (*models.ObjectStorageBucket, error) {
	var response struct {
		Bucket      *models.ObjectStorageBucket `json:"bucket"`
		OperationID string                      `json:"operation_id,omitempty"`
	}

	err := c.Put(ctx, fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/bucket/%s", c.Organization, c.ProjectName, name), req, &response)
	if err != nil {
		return nil, err
	}

	// If there's an operation ID, wait for completion
	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to update bucket: %w", err)
		}
	}

	// If the response included bucket details, return them
	if response.Bucket != nil {
		return response.Bucket, nil
	}

	// Otherwise fetch the bucket by name (the update API may return an empty body)
	return c.GetObjectStorageBucket(ctx, name)
}

// CreateAccessKey creates a new object storage access key
func (c *Client) CreateAccessKey(ctx context.Context, req *models.CreateAccessKeyRequest) (*models.CreateAccessKeyResponse, error) {
	var response models.CreateAccessKeyResponse
	path := fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/accesskey", c.Organization, c.ProjectName)
	err := c.Post(ctx, path, req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ListAccessKeys retrieves all object storage access keys in the project
func (c *Client) ListAccessKeys(ctx context.Context) (*models.AccessKeyListResponse, error) {
	var response models.AccessKeyListResponse
	path := fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/accesskeys", c.Organization, c.ProjectName)
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteAccessKey deletes an object storage access key by its ID
func (c *Client) DeleteAccessKey(ctx context.Context, accessKeyID string) error {
	path := fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/accesskey/%s", c.Organization, c.ProjectName, accessKeyID)
	return c.Delete(ctx, path)
}

// DeleteObjectStorageBucket deletes an object storage bucket
func (c *Client) DeleteObjectStorageBucket(ctx context.Context, name string) error {
	err := c.Delete(ctx, fmt.Sprintf("/storage-plugin/v1/domain/%s/project/%s/object-storage/bucket/%s", c.Organization, c.ProjectName, name))
	if err != nil {
		return err
	}

	// Wait for bucket to be deleted (poll until 404)
	for i := 0; i < 60; i++ {
		_, err := c.GetObjectStorageBucket(ctx, name)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("bucket deletion timed out")
}
