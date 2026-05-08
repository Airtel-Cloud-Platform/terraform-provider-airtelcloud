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

// nfsTestConfig extends testConfig with NFS-specific fields
type nfsTestConfig struct {
	APIEndpoint      string
	APIKey           string
	APISecret        string
	Region           string
	Organization     string
	ProjectName      string
	AvailabilityZone string
}

func getNFSTestConfig(t *testing.T) *nfsTestConfig {
	config := &nfsTestConfig{
		APIEndpoint:      os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:           os.Getenv("AIRTEL_API_KEY"),
		APISecret:        os.Getenv("AIRTEL_API_SECRET"),
		Region:           os.Getenv("AIRTEL_REGION"),
		Organization:     os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:      os.Getenv("AIRTEL_PROJECT_NAME"),
		AvailabilityZone: os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE"),
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
	if config.AvailabilityZone == "" {
		t.Skip("AIRTEL_TEST_AVAILABILITY_ZONE not set, skipping integration test")
	}

	// Set defaults
	if config.Region == "" {
		config.Region = "south"
	}

	return config
}

func createNFSTestClient(t *testing.T, config *nfsTestConfig) *Client {
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

// TestFileStorageVolumeIntegration_CreateGetDelete tests the full lifecycle of a file storage volume
func TestFileStorageVolumeIntegration_CreateGetDelete(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	// Generate unique volume name
	volumeName := fmt.Sprintf("test-volume-%d", time.Now().Unix())

	// Create volume request
	createReq := &models.CreateFileStorageVolumeRequest{
		Name:             volumeName,
		Description:      "Integration test volume",
		Size:             "100",
		AvailabilityZone: config.AvailabilityZone,
	}

	t.Logf("Creating file storage volume: %s", volumeName)

	// Create volume
	err := client.CreateFileStorageVolume(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateFileStorageVolume failed: %v", err)
	}

	t.Logf("File storage volume create request submitted: %s", volumeName)

	// Cleanup: Delete volume at the end
	defer func() {
		t.Logf("Deleting file storage volume: %s", volumeName)
		err := client.DeleteFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
		if err != nil {
			t.Errorf("DeleteFileStorageVolume failed: %v", err)
		} else {
			t.Log("File storage volume deleted successfully")
		}
	}()

	// Wait for volume to become ready
	t.Logf("Waiting for volume to become ready: %s", volumeName)
	err = client.WaitForFileStorageVolumeReady(ctx, volumeName, config.AvailabilityZone, 10*time.Minute)
	if err != nil {
		t.Fatalf("WaitForFileStorageVolumeReady failed: %v", err)
	}

	// Get volume
	t.Logf("Getting file storage volume: %s", volumeName)
	volume, err := client.GetFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
	if err != nil {
		t.Fatalf("GetFileStorageVolume failed: %v", err)
	}

	// Verify fetched volume
	if volume.Name != volumeName {
		t.Errorf("Expected volume name %s, got %s", volumeName, volume.Name)
	}
	if volume.Size != createReq.Size {
		t.Errorf("Expected size %s, got %s", createReq.Size, volume.Size)
	}
	if volume.State != models.FileStorageStateActive {
		t.Errorf("Expected state Active, got %s", volume.State)
	}

	t.Log("GetFileStorageVolume returned correct data")
}

// TestFileStorageVolumeIntegration_List tests listing file storage volumes
func TestFileStorageVolumeIntegration_List(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization (domain): %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List volumes
	response, err := client.ListFileStorageVolumes(ctx)
	if err != nil {
		t.Fatalf("ListFileStorageVolumes failed: %v", err)
	}

	t.Logf("Found %d file storage volumes", response.Count)

	// Log volume details
	for i, volume := range response.Items {
		t.Logf("Volume %d: Name=%s, Size=%s, State=%s, AZ=%s",
			i+1, volume.Name, volume.Size, volume.State, volume.AvailabilityZone)
	}
}

// TestFileStorageVolumeIntegration_Update tests updating a file storage volume
func TestFileStorageVolumeIntegration_Update(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	// Generate unique volume name
	volumeName := fmt.Sprintf("test-volume-update-%d", time.Now().Unix())
	updatedDescription := "Updated integration test volume"
	updatedSize := "200"

	// Create volume request
	createReq := &models.CreateFileStorageVolumeRequest{
		Name:             volumeName,
		Description:      "Original description",
		Size:             "100",
		AvailabilityZone: config.AvailabilityZone,
	}

	t.Logf("Creating file storage volume for update test: %s", volumeName)

	// Create volume
	err := client.CreateFileStorageVolume(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateFileStorageVolume failed: %v", err)
	}

	// Cleanup: Delete volume at the end
	defer func() {
		t.Logf("Deleting file storage volume: %s", volumeName)
		err := client.DeleteFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
		if err != nil {
			t.Errorf("DeleteFileStorageVolume failed: %v", err)
		} else {
			t.Log("File storage volume deleted successfully")
		}
	}()

	// Wait for volume to become ready
	t.Logf("Waiting for volume to become ready: %s", volumeName)
	err = client.WaitForFileStorageVolumeReady(ctx, volumeName, config.AvailabilityZone, 10*time.Minute)
	if err != nil {
		t.Fatalf("WaitForFileStorageVolumeReady failed: %v", err)
	}

	// Update volume
	updateReq := &models.UpdateFileStorageVolumeRequest{
		Name:             volumeName,
		Description:      updatedDescription,
		Size:             updatedSize,
		AvailabilityZone: config.AvailabilityZone,
	}

	t.Logf("Updating file storage volume: %s", volumeName)
	err = client.UpdateFileStorageVolume(ctx, volumeName, updateReq)
	if err != nil {
		t.Fatalf("UpdateFileStorageVolume failed: %v", err)
	}

	// Wait for update to complete
	err = client.WaitForFileStorageVolumeReady(ctx, volumeName, config.AvailabilityZone, 10*time.Minute)
	if err != nil {
		t.Fatalf("WaitForFileStorageVolumeReady after update failed: %v", err)
	}

	// Verify update by fetching volume
	volume, err := client.GetFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
	if err != nil {
		t.Fatalf("GetFileStorageVolume after update failed: %v", err)
	}

	if volume.Description != updatedDescription {
		t.Errorf("Expected description %s, got %s", updatedDescription, volume.Description)
	}
	if volume.Size != updatedSize {
		t.Errorf("Expected size %s, got %s", updatedSize, volume.Size)
	}

	t.Log("File storage volume updated successfully")
}

// TestFileStorageVolumeIntegration_GetNonExistent tests getting a non-existent file storage volume
func TestFileStorageVolumeIntegration_GetNonExistent(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	nonExistentName := "non-existent-volume-12345"

	t.Logf("Attempting to get non-existent file storage volume: %s", nonExistentName)

	_, err := client.GetFileStorageVolume(ctx, nonExistentName, config.AvailabilityZone)
	if err == nil {
		t.Error("Expected error for non-existent volume, got nil")
		return
	}

	// Check if it's a 404 error
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent volume")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestFileStorageExportPathIntegration_CreateGetDelete tests the full lifecycle of an export path
func TestFileStorageExportPathIntegration_CreateGetDelete(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	// First, create a volume to attach the export path to
	volumeName := fmt.Sprintf("test-volume-export-%d", time.Now().Unix())

	volumeReq := &models.CreateFileStorageVolumeRequest{
		Name:             volumeName,
		Description:      "Integration test volume for export path",
		Size:             "100",
		AvailabilityZone: config.AvailabilityZone,
	}

	t.Logf("Creating file storage volume for export path test: %s", volumeName)

	err := client.CreateFileStorageVolume(ctx, volumeReq)
	if err != nil {
		t.Fatalf("CreateFileStorageVolume failed: %v", err)
	}

	// Cleanup: Delete volume at the end (this will also delete export paths)
	defer func() {
		t.Logf("Deleting file storage volume: %s", volumeName)
		err := client.DeleteFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
		if err != nil {
			t.Errorf("DeleteFileStorageVolume failed: %v", err)
		} else {
			t.Log("File storage volume deleted successfully")
		}
	}()

	// Wait for volume to become ready
	t.Logf("Waiting for volume to become ready: %s", volumeName)
	err = client.WaitForFileStorageVolumeReady(ctx, volumeName, config.AvailabilityZone, 10*time.Minute)
	if err != nil {
		t.Fatalf("WaitForFileStorageVolumeReady failed: %v", err)
	}

	// Create export path request
	createReq := &models.CreateFileStorageExportPathRequest{
		Volume:           volumeName,
		Description:      "Integration test export path",
		Protocol:         models.NFSProtocolV4,
		AvailabilityZone: config.AvailabilityZone,
		NFSInfo: &models.NFSExportInfo{
			DefaultAccessType: models.NFSAccessReadWrite,
			DefaultUserSquash: models.NFSSquashNone,
		},
	}

	t.Log("Creating file storage export path")

	// Create export path
	exportPathKey, err := client.CreateFileStorageExportPath(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateFileStorageExportPath failed: %v", err)
	}

	t.Logf("Export path created with PathID: %s", exportPathKey.PathID)

	if exportPathKey.PathID == "" {
		t.Error("PathID should not be empty")
	}

	// Cleanup: Delete export path before volume
	defer func() {
		t.Logf("Deleting export path: %s", exportPathKey.PathID)
		err := client.DeleteFileStorageExportPath(ctx, exportPathKey.PathID, config.AvailabilityZone)
		if err != nil {
			t.Errorf("DeleteFileStorageExportPath failed: %v", err)
		} else {
			t.Log("Export path deleted successfully")
		}
	}()

	// Get export path
	t.Logf("Getting export path: %s", exportPathKey.PathID)
	exportPath, err := client.GetFileStorageExportPath(ctx, exportPathKey.PathID)
	if err != nil {
		t.Fatalf("GetFileStorageExportPath failed: %v", err)
	}

	// Verify fetched export path
	if exportPath.PathID != exportPathKey.PathID {
		t.Errorf("Expected PathID %s, got %s", exportPathKey.PathID, exportPath.PathID)
	}
	if exportPath.Volume != volumeName {
		t.Errorf("Expected volume %s, got %s", volumeName, exportPath.Volume)
	}

	t.Log("GetFileStorageExportPath returned correct data")
}

// TestFileStorageExportPathIntegration_List tests listing export paths for a volume
func TestFileStorageExportPathIntegration_List(t *testing.T) {
	config := getNFSTestConfig(t)
	client := createNFSTestClient(t, config)
	ctx := context.Background()

	// First, create a volume
	volumeName := fmt.Sprintf("test-volume-list-%d", time.Now().Unix())

	volumeReq := &models.CreateFileStorageVolumeRequest{
		Name:             volumeName,
		Description:      "Integration test volume for listing export paths",
		Size:             "100",
		AvailabilityZone: config.AvailabilityZone,
	}

	t.Logf("Creating file storage volume for list test: %s", volumeName)

	err := client.CreateFileStorageVolume(ctx, volumeReq)
	if err != nil {
		t.Fatalf("CreateFileStorageVolume failed: %v", err)
	}

	// Cleanup: Delete volume at the end
	defer func() {
		t.Logf("Deleting file storage volume: %s", volumeName)
		err := client.DeleteFileStorageVolume(ctx, volumeName, config.AvailabilityZone)
		if err != nil {
			t.Errorf("DeleteFileStorageVolume failed: %v", err)
		} else {
			t.Log("File storage volume deleted successfully")
		}
	}()

	// Wait for volume to become ready
	err = client.WaitForFileStorageVolumeReady(ctx, volumeName, config.AvailabilityZone, 10*time.Minute)
	if err != nil {
		t.Fatalf("WaitForFileStorageVolumeReady failed: %v", err)
	}

	// List export paths (should be empty initially)
	t.Logf("Listing export paths for volume: %s", volumeName)
	response, err := client.ListFileStorageExportPaths(ctx, volumeName)
	if err != nil {
		t.Fatalf("ListFileStorageExportPaths failed: %v", err)
	}

	t.Logf("Found %d export paths", response.Count)

	// Log export path details
	for i, ep := range response.Items {
		t.Logf("Export Path %d: PathID=%s, Volume=%s, Protocol=%s",
			i+1, ep.PathID, ep.Volume, ep.Protocol)
	}
}
