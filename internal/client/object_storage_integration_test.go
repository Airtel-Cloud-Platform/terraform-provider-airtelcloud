//go:build integration

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// objectStorageTestConfig contains object storage-specific test configuration
type objectStorageTestConfig struct {
	APIEndpoint  string
	APIKey       string
	APISecret    string
	Region       string
	Organization string
	ProjectName  string
}

func getObjectStorageTestConfig(t *testing.T) *objectStorageTestConfig {
	config := &objectStorageTestConfig{
		APIEndpoint:  os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:       os.Getenv("AIRTEL_API_KEY"),
		APISecret:    os.Getenv("AIRTEL_API_SECRET"),
		Region:       os.Getenv("AIRTEL_REGION"),
		Organization: os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:  os.Getenv("AIRTEL_PROJECT_NAME"),
	}

	// Validate required fields
	if config.APIEndpoint == "" {
		t.Skip("AIRTEL_API_ENDPOINT not set, skipping integration test")
	}
	if config.APIKey == "" {
		t.Skip("AIRTEL_API_KEY not set, skipping integration test")
	}
	if config.APISecret == "" {
		t.Skip("AIRTEL_API_SECRET not set, skipping integration test")
	}
	if config.Organization == "" {
		t.Skip("AIRTEL_ORGANIZATION not set, skipping integration test")
	}
	if config.ProjectName == "" {
		t.Skip("AIRTEL_PROJECT_NAME not set, skipping integration test")
	}

	// Set defaults
	if config.Region == "" {
		config.Region = "south"
	}

	return config
}

func createObjectStorageTestClient(t *testing.T, config *objectStorageTestConfig) *Client {
	client, err := NewClient(
		config.APIEndpoint,
		config.APIKey,
		config.APISecret,
		config.Region,
		config.Organization,
		config.ProjectName,
		"",
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// TestObjectStorageIntegration_CreateGetDelete tests the full lifecycle of an object storage bucket
func TestObjectStorageIntegration_CreateGetDelete(t *testing.T) {
	config := getObjectStorageTestConfig(t)
	client := createObjectStorageTestClient(t, config)
	ctx := context.Background()

	// Generate unique bucket name (must be lowercase, no underscores)
	bucketName := fmt.Sprintf("test-bucket-%d", time.Now().Unix())

	// Create bucket request
	createReq := &models.CreateObjectStorageBucketRequest{
		Bucket: bucketName,
		Config: &models.BucketCreateConfig{
			Versioning: false,
			ObjLocking: false,
			Replication: &models.BucketReplicationConfig{
				ReplicationType: "Local",
				AZ:              "S1",
				Tag:             "south_S1",
			},
		},
		Tags: map[string]string{
			"env":     "test",
			"purpose": "integration-test",
		},
	}

	t.Logf("Creating object storage bucket: %s", bucketName)

	// Create bucket
	bucket, err := client.CreateObjectStorageBucket(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateObjectStorageBucket failed: %v", err)
	}

	t.Logf("Bucket created with name: %s", bucket.Name)

	// Verify created bucket fields
	if bucket.Name != bucketName {
		t.Errorf("Expected bucket name %s, got %s", bucketName, bucket.Name)
	}

	// Cleanup: Delete bucket at the end
	defer func() {
		t.Logf("Deleting bucket: %s", bucket.Name)
		err := client.DeleteObjectStorageBucket(ctx, bucket.Name)
		if err != nil {
			t.Errorf("DeleteObjectStorageBucket failed: %v", err)
		} else {
			t.Log("Bucket deleted successfully")
		}
	}()

	// Get bucket
	t.Logf("Getting bucket: %s", bucket.Name)
	fetchedBucket, err := client.GetObjectStorageBucket(ctx, bucket.Name)
	if err != nil {
		t.Fatalf("GetObjectStorageBucket failed: %v", err)
	}

	// Verify fetched bucket
	if fetchedBucket.Name != bucketName {
		t.Errorf("Expected bucket name %s, got %s", bucketName, fetchedBucket.Name)
	}

	t.Log("GetObjectStorageBucket returned correct data")
}

// TestObjectStorageIntegration_List tests listing object storage buckets
func TestObjectStorageIntegration_List(t *testing.T) {
	config := getObjectStorageTestConfig(t)
	client := createObjectStorageTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List buckets
	response, err := client.ListObjectStorageBuckets(ctx)
	if err != nil {
		t.Fatalf("ListObjectStorageBuckets failed: %v", err)
	}

	t.Logf("Found %d object storage buckets", response.Count)

	// Log bucket details
	for i, bucket := range response.Items {
		az := ""
		if bucket.ReplicationConfig != nil {
			az = bucket.ReplicationConfig.AZ
		}
		t.Logf("Bucket %d: Name=%s, AZ=%s, Versioning=%v",
			i+1, bucket.Name, az, bucket.Versioning)
	}
}

// TestObjectStorageIntegration_Update tests updating an object storage bucket
func TestObjectStorageIntegration_Update(t *testing.T) {
	config := getObjectStorageTestConfig(t)
	client := createObjectStorageTestClient(t, config)
	ctx := context.Background()

	// Generate unique bucket name
	bucketName := fmt.Sprintf("test-bucket-update-%d", time.Now().Unix())

	// Create bucket request
	createReq := &models.CreateObjectStorageBucketRequest{
		Bucket: bucketName,
		Config: &models.BucketCreateConfig{
			Versioning: false,
			ObjLocking: false,
			Replication: &models.BucketReplicationConfig{
				ReplicationType: "Local",
				AZ:              "S1",
				Tag:             "south_S1",
			},
		},
		Tags: map[string]string{
			"env": "test",
		},
	}

	t.Logf("Creating bucket for update test: %s", bucketName)

	// Create bucket
	bucket, err := client.CreateObjectStorageBucket(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateObjectStorageBucket failed: %v", err)
	}

	t.Logf("Bucket created with name: %s", bucket.Name)

	// Cleanup: Delete bucket at the end
	defer func() {
		t.Logf("Deleting bucket: %s", bucket.Name)
		err := client.DeleteObjectStorageBucket(ctx, bucket.Name)
		if err != nil {
			t.Errorf("DeleteObjectStorageBucket failed: %v", err)
		} else {
			t.Log("Bucket deleted successfully")
		}
	}()

	// Update bucket
	enableVersioning := true
	updateReq := &models.UpdateObjectStorageBucketRequest{
		Versioning: &enableVersioning,
		Tags: map[string]string{
			"env":     "test",
			"updated": "true",
		},
	}

	t.Logf("Updating bucket: %s", bucket.Name)
	updatedBucket, err := client.UpdateObjectStorageBucket(ctx, bucket.Name, updateReq)
	if err != nil {
		t.Fatalf("UpdateObjectStorageBucket failed: %v", err)
	}

	// Verify update
	if updatedBucket.Versioning == nil || !bool(*updatedBucket.Versioning) {
		t.Error("Expected versioning to be enabled")
	}

	t.Log("Bucket updated successfully")
}

// TestObjectStorageIntegration_GetNonExistent tests getting a non-existent bucket
func TestObjectStorageIntegration_GetNonExistent(t *testing.T) {
	config := getObjectStorageTestConfig(t)
	client := createObjectStorageTestClient(t, config)
	ctx := context.Background()

	nonExistentName := "non-existent-bucket-12345"

	t.Logf("Attempting to get non-existent bucket: %s", nonExistentName)

	_, err := client.GetObjectStorageBucket(ctx, nonExistentName)
	if err == nil {
		t.Error("Expected error for non-existent bucket, got nil")
		return
	}

	// Check if it's a 404 error
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent bucket")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}
