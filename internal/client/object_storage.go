package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) objectStorageBasePath() string {
	return fmt.Sprintf("/api/storage-plugin/v1/domain/%s/project/%s/object-storage", c.Organization, c.ProjectName)
}

func (c *Client) objectStorageBucketPath(name string) string {
	return fmt.Sprintf("%s/bucket/%s", c.objectStorageBasePath(), url.PathEscape(name))
}

// GetObjectStorageBucket retrieves an object storage bucket by name.
func (c *Client) GetObjectStorageBucket(ctx context.Context, name string) (*models.ObjectStorageBucket, error) {
	var bucket models.ObjectStorageBucket
	err := c.Get(ctx, c.objectStorageBucketPath(name), &bucket)
	if err != nil {
		return nil, err
	}
	if bucket.IsDeleted {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("object storage bucket %s is deleted", name)}
	}
	if bucket.Name == "" {
		bucket.Name = name
	}
	return &bucket, nil
}

// ListObjectStorageBuckets retrieves all object storage buckets in the project.
func (c *Client) ListObjectStorageBuckets(ctx context.Context) (*models.ObjectStorageBucketListResponse, error) {
	var response models.ObjectStorageBucketListResponse
	err := c.Get(ctx, c.objectStorageBasePath()+"/buckets", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateObjectStorageBucket creates a new object storage bucket and waits until it is ready.
func (c *Client) CreateObjectStorageBucket(ctx context.Context, req *models.CreateObjectStorageBucketRequest) (*models.ObjectStorageBucket, error) {
	var response struct {
		Bucket      *models.ObjectStorageBucket `json:"bucket"`
		OperationID string                      `json:"operation_id,omitempty"`
	}

	az := ""
	if req.Config != nil && req.Config.Replication != nil {
		az = req.Config.Replication.AZ
	}
	err := c.objectStorageClient(az).Post(ctx, c.objectStorageBasePath()+"/bucket", req, &response)
	if err != nil {
		return nil, err
	}

	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	if err := c.WaitForObjectStorageBucketReady(ctx, req.Bucket, 10*time.Minute); err != nil {
		return nil, err
	}

	if response.Bucket != nil && response.Bucket.Name != "" && !response.Bucket.IsDeleted {
		return response.Bucket, nil
	}

	return c.GetObjectStorageBucket(ctx, req.Bucket)
}

func (c *Client) objectStorageClient(az string) *Client {
	if az == "" {
		return c
	}
	return c.WithAvailabilityZone(az)
}

// UpdateObjectStorageBucket updates an existing object storage bucket.
func (c *Client) UpdateObjectStorageBucket(ctx context.Context, name string, req *models.UpdateObjectStorageBucketRequest) (*models.ObjectStorageBucket, error) {
	var response struct {
		Bucket      *models.ObjectStorageBucket `json:"bucket"`
		OperationID string                      `json:"operation_id,omitempty"`
	}

	err := c.Put(ctx, c.objectStorageBucketPath(name), req, &response)
	if err != nil {
		return nil, err
	}

	if response.OperationID != "" {
		err = c.WaitForOperation(ctx, response.OperationID, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to update bucket: %w", err)
		}
	}

	if err := c.WaitForObjectStorageBucketReady(ctx, name, 10*time.Minute); err != nil {
		return nil, err
	}

	if response.Bucket != nil && response.Bucket.Name != "" {
		return response.Bucket, nil
	}

	return c.GetObjectStorageBucket(ctx, name)
}

// CreateAccessKey creates a new object storage access key.
func (c *Client) CreateAccessKey(ctx context.Context, req *models.CreateAccessKeyRequest) (*models.CreateAccessKeyResponse, error) {
	var response models.CreateAccessKeyResponse
	path := fmt.Sprintf("%s/accesskey", c.objectStorageBasePath())
	err := c.Post(ctx, path, req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ListAccessKeys retrieves all object storage access keys in the project.
func (c *Client) ListAccessKeys(ctx context.Context) (*models.AccessKeyListResponse, error) {
	var response models.AccessKeyListResponse
	path := fmt.Sprintf("%s/accesskeys", c.objectStorageBasePath())
	err := c.Get(ctx, path, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteAccessKey deletes an object storage access key by its ID.
func (c *Client) DeleteAccessKey(ctx context.Context, accessKeyID string) error {
	path := fmt.Sprintf("%s/accesskey/%s", c.objectStorageBasePath(), url.PathEscape(accessKeyID))
	return c.Delete(ctx, path)
}

// DeleteObjectStorageBucket deletes a bucket and waits until GET returns 404.
// availabilityZone is required by the storage plugin to authorize delete against the backing store.
func (c *Client) DeleteObjectStorageBucket(ctx context.Context, name, availabilityZone string, objectLocking bool) error {
	cli := c.objectStorageClient(availabilityZone)

	if objectLocking {
		locking := false
		_ = cli.Put(ctx, c.objectStorageBucketPath(name), &models.UpdateObjectStorageBucketRequest{ObjLocking: &locking}, nil)
	}

	q := url.Values{}
	if availabilityZone != "" {
		q.Set("az", availabilityZone)
		q.Set("availabilityZone", availabilityZone)
	}
	q.Set("force", "true")
	path := c.objectStorageBucketPath(name) + "?" + q.Encode()

	err := cli.Delete(ctx, path)
	if err != nil && !IsNotFoundError(err) {
		return fmt.Errorf("%w; object lock or missing AZ on delete can surface as a 403", err)
	}

	return cli.WaitForObjectStorageBucketDeleted(ctx, name, 10*time.Minute)
}

func (c *Client) objectStorageBucketGone(ctx context.Context, name string) (bool, error) {
	_, err := c.GetObjectStorageBucket(ctx, name)
	if err != nil {
		if IsNotFoundError(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// WaitForObjectStorageBucketDeleted waits until the bucket is gone.
func (c *Client) WaitForObjectStorageBucketDeleted(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gone, err := c.objectStorageBucketGone(ctx, name)
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
	return fmt.Errorf("timeout waiting for object storage bucket %s to be deleted", name)
}

// WaitForObjectStorageBucketReady waits until the bucket is Active.
func (c *Client) WaitForObjectStorageBucketReady(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		bucket, err := c.GetObjectStorageBucket(ctx, name)
		if err != nil {
			if IsNotFoundError(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(5 * time.Second):
				}
				continue
			}
			return err
		}

		switch bucket.State {
		case "Active", "Ready", "available", "":
			return nil
		case "CreateFailed", "Failed", "DeleteFailed":
			return fmt.Errorf("object storage bucket %s failed with state: %s", name, bucket.State)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for object storage bucket %s to become ready", name)
}
