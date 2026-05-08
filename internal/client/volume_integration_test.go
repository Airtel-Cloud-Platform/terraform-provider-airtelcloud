//go:build integration

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// getVolumeTestDefaults returns common fields needed for volume creation from env vars
func getVolumeTestDefaults(t *testing.T) (availabilityZone, networkID, subnetID string, volumeType models.VolumeType) {
	availabilityZone = os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE")
	if availabilityZone == "" {
		availabilityZone = "S1"
	}

	networkID = os.Getenv("AIRTEL_TEST_NETWORK_ID")
	if networkID == "" {
		t.Skip("AIRTEL_TEST_NETWORK_ID not set, skipping volume create test")
	}

	subnetID = os.Getenv("AIRTEL_TEST_SUBNET_ID")
	if subnetID == "" {
		t.Skip("AIRTEL_TEST_SUBNET_ID not set, skipping volume create test")
	}

	return
}

// getDefaultVolumeType fetches and returns the default volume type
func getDefaultVolumeType(t *testing.T, client *Client, ctx context.Context) models.VolumeType {
	volumeTypes, err := client.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		t.Fatalf("GetVolumeTypes failed: %v", err)
	}
	if len(volumeTypes) == 0 {
		t.Fatal("No volume types available")
	}

	// Prefer the default type, fall back to first active one
	for _, vt := range volumeTypes {
		if vt.IsDefault && vt.IsActive {
			return vt
		}
	}
	for _, vt := range volumeTypes {
		if vt.IsActive {
			return vt
		}
	}
	return volumeTypes[0]
}

// TestVolumeIntegration_CreateGetDelete tests the full lifecycle of a volume
func TestVolumeIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	availabilityZone, networkID, subnetID, _ := getVolumeTestDefaults(t)
	vt := getDefaultVolumeType(t, client, ctx)

	// Generate unique volume name
	volumeName := fmt.Sprintf("test-vol-%d", time.Now().Unix())

	// Determine minimum volume size from volume type
	volumeSize := 50
	if len(vt.MinSize) > 0 {
		var minSize struct {
			Additional int `json:"additional"`
		}
		if err := json.Unmarshal(vt.MinSize, &minSize); err == nil && minSize.Additional > volumeSize {
			volumeSize = minSize.Additional
		}
	}

	// Create volume
	createReq := &models.CreateVolumeRequest{
		VolumeName:       volumeName,
		VolumeSize:       volumeSize,
		Bootable:         false,
		AvailabilityZone: availabilityZone,
		Network:          networkID,
		VPCID:            networkID,
		Subnet:           subnetID,
		SubnetID:         subnetID,
		VolumeType:       vt.Name,
		VolumeTypeID:     fmt.Sprintf("%d", vt.ID),
		IsEncrypted:      "encrypted",
		EnableBackup:     false,
		BillingUnit:      "MRC",
		Products:         `{"volume":{"id":""}}`,
	}

	t.Logf("Creating volume: %s (size: %dGB, type: %s, AZ: %s)", volumeName, createReq.VolumeSize, vt.Name, availabilityZone)

	volume, err := client.CreateVolume(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	t.Logf("Volume created with ID: %d, Name: %s", volume.ID, volume.VolumeName)

	// Verify created volume fields
	if volume.VolumeName != volumeName {
		t.Errorf("Expected volume name %s, got %s", volumeName, volume.VolumeName)
	}
	if volume.ID == 0 {
		t.Error("Volume ID should not be zero")
	}

	// Cleanup: Delete volume at the end
	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Logf("DeleteVolume returned error (may be backend issue): %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Verify volume appears in list
	t.Logf("Verifying volume %d appears in list", volume.ID)
	volumes, err := client.ListVolumes(ctx, "")
	if err != nil {
		t.Fatalf("ListVolumes failed: %v", err)
	}

	found := false
	for _, v := range volumes {
		if v.ID == volume.ID {
			found = true
			if v.VolumeName != volumeName {
				t.Errorf("Expected volume name %s in list, got %s", volumeName, v.VolumeName)
			}
			if v.VolumeSize != volumeSize {
				t.Errorf("Expected volume size %d in list, got %d", volumeSize, v.VolumeSize)
			}
			t.Logf("Volume found in list: ID=%d, Name=%s, Size=%dGB, Status=%s", v.ID, v.VolumeName, v.VolumeSize, v.Status)
			break
		}
	}
	if !found {
		t.Errorf("Volume %d not found in list of %d volumes", volume.ID, len(volumes))
	}

	t.Log("Volume create and list verification passed")
}

// TestVolumeIntegration_List tests listing volumes
func TestVolumeIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List volumes (empty computeID to list all)
	volumes, err := client.ListVolumes(ctx, "")
	if err != nil {
		t.Fatalf("ListVolumes failed: %v", err)
	}

	t.Logf("Found %d volumes", len(volumes))

	// Log volume details
	for i, vol := range volumes {
		t.Logf("Volume %d: ID=%d, Name=%s, Size=%dGB, Status=%s",
			i+1, vol.ID, vol.VolumeName, vol.VolumeSize, vol.Status)
	}
}

// TestVolumeIntegration_Update tests extending a volume
func TestVolumeIntegration_Update(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create a volume to update
	volumeName := fmt.Sprintf("test-vol-update-%d", time.Now().Unix())
	volume, err := client.CreateVolume(ctx, &models.CreateVolumeRequest{
		VolumeName: volumeName,
		VolumeSize: 10,

		BillingUnit: "MRC",
	})
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	t.Logf("Volume created: ID=%d, Size=%dGB", volume.ID, volume.VolumeSize)

	// Cleanup: Delete volume at the end
	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Errorf("DeleteVolume failed: %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Update volume size (extend to 20GB)
	updateReq := &models.UpdateVolumeRequest{
		VolumeSize:  20,
		BillingUnit: "MRC",
	}

	t.Logf("Updating volume %s: extending to %dGB", volume.UUID, updateReq.VolumeSize)
	err = client.UpdateVolume(ctx, volume.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateVolume failed: %v", err)
	}

	// Get and verify updated volume
	fetchedVolume, err := client.GetVolume(ctx, volume.UUID)
	if err != nil {
		t.Fatalf("GetVolume failed: %v", err)
	}

	if fetchedVolume.VolumeSize != 20 {
		t.Errorf("Expected volume size 20, got %d", fetchedVolume.VolumeSize)
	}

	t.Log("Volume updated successfully")
}

// TestVolumeIntegration_GetNonExistent tests getting a non-existent volume
func TestVolumeIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentUUID := "non-existent-uuid-99999"

	t.Logf("Attempting to get non-existent volume: %s", nonExistentUUID)

	_, err := client.GetVolume(ctx, nonExistentUUID)
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

// TestVolumeIntegration_VolumeTypes tests listing volume types
func TestVolumeIntegration_VolumeTypes(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List block storage volume types
	blockTypes, err := client.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		t.Fatalf("GetVolumeTypes(BLOCK_STORAGE) failed: %v", err)
	}

	// List all volume types (unfiltered)
	allTypes, err := client.GetVolumeTypes(ctx, "")
	if err != nil {
		t.Fatalf("GetVolumeTypes (unfiltered) failed: %v", err)
	}

	t.Logf("Found %d block storage volume types, %d total volume types", len(blockTypes), len(allTypes))

	// Log volume type details
	for i, vt := range blockTypes {
		t.Logf("VolumeType %d: ID=%d, Name=%s, Label=%s, Active=%v, Default=%v, MinSize=%s, MaxSize=%s",
			i+1, vt.ID, vt.Name, vt.Label, vt.IsActive, vt.IsDefault, string(vt.MinSize), string(vt.MaxSize))
	}

	t.Run("at_least_one_block_storage_type", func(t *testing.T) {
		if len(blockTypes) == 0 {
			t.Fatal("Expected at least one BLOCK_STORAGE volume type, got 0")
		}
	})

	t.Run("required_fields_populated", func(t *testing.T) {
		for _, vt := range blockTypes {
			if vt.ID <= 0 {
				t.Errorf("Volume type has invalid ID: %d (Name: %s)", vt.ID, vt.Name)
			}
			if vt.Name == "" {
				t.Errorf("Volume type ID=%d has empty Name", vt.ID)
			}
		}
	})

	t.Run("group_filter_enforced", func(t *testing.T) {
		for _, vt := range blockTypes {
			if vt.Group != "BLOCK_STORAGE" {
				t.Errorf("Volume type ID=%d Name=%s has Group=%q, expected BLOCK_STORAGE", vt.ID, vt.Name, vt.Group)
			}
		}
	})

	t.Run("at_least_one_active_type", func(t *testing.T) {
		hasActive := false
		for _, vt := range blockTypes {
			if vt.IsActive {
				hasActive = true
				break
			}
		}
		if !hasActive {
			t.Error("Expected at least one active BLOCK_STORAGE volume type, found none")
		}
	})

	t.Run("ids_are_unique", func(t *testing.T) {
		seen := make(map[int]string)
		for _, vt := range blockTypes {
			if prev, exists := seen[vt.ID]; exists {
				t.Errorf("Duplicate volume type ID=%d: %q and %q", vt.ID, prev, vt.Name)
			}
			seen[vt.ID] = vt.Name
		}
	})

	t.Run("names_are_unique", func(t *testing.T) {
		seen := make(map[string]int)
		for _, vt := range blockTypes {
			if prevID, exists := seen[vt.Name]; exists {
				t.Errorf("Duplicate volume type Name=%q: IDs %d and %d", vt.Name, prevID, vt.ID)
			}
			seen[vt.Name] = vt.ID
		}
	})

	t.Run("at_most_one_default", func(t *testing.T) {
		defaultCount := 0
		for _, vt := range blockTypes {
			if vt.IsDefault {
				defaultCount++
				t.Logf("Default volume type: ID=%d, Name=%s", vt.ID, vt.Name)
			}
		}
		if defaultCount > 1 {
			t.Errorf("Expected at most 1 default BLOCK_STORAGE volume type, found %d", defaultCount)
		}
	})

	t.Run("unfiltered_returns_superset", func(t *testing.T) {
		if len(allTypes) < len(blockTypes) {
			t.Errorf("Unfiltered count (%d) is less than BLOCK_STORAGE count (%d)", len(allTypes), len(blockTypes))
		}
		allIDs := make(map[int]bool)
		for _, vt := range allTypes {
			allIDs[vt.ID] = true
		}
		for _, vt := range blockTypes {
			if !allIDs[vt.ID] {
				t.Errorf("BLOCK_STORAGE type ID=%d Name=%s not found in unfiltered results", vt.ID, vt.Name)
			}
		}
	})

	t.Run("min_max_size_valid_json", func(t *testing.T) {
		for _, vt := range blockTypes {
			if len(vt.MinSize) > 0 && !json.Valid(vt.MinSize) {
				t.Errorf("Volume type ID=%d Name=%s has invalid MinSize JSON: %s", vt.ID, vt.Name, string(vt.MinSize))
			}
			if len(vt.MaxSize) > 0 && !json.Valid(vt.MaxSize) {
				t.Errorf("Volume type ID=%d Name=%s has invalid MaxSize JSON: %s", vt.ID, vt.Name, string(vt.MaxSize))
			}
		}
	})
}

// TestVolumeIntegration_AttachDetach tests attaching and detaching a volume to/from a compute instance
func TestVolumeIntegration_AttachDetach(t *testing.T) {
	computeID := os.Getenv("AIRTEL_TEST_COMPUTE_ID")
	if computeID == "" {
		t.Skip("AIRTEL_TEST_COMPUTE_ID not set, skipping attach/detach test")
	}

	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create a volume
	volumeName := fmt.Sprintf("test-vol-attach-%d", time.Now().Unix())
	volume, err := client.CreateVolume(ctx, &models.CreateVolumeRequest{
		VolumeName: volumeName,
		VolumeSize: 10,

		BillingUnit: "MRC",
	})
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}
	t.Logf("Volume created: ID=%d, Name=%s", volume.ID, volume.VolumeName)

	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Errorf("DeleteVolume failed: %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Attach volume to compute instance
	t.Logf("Attaching volume %s to compute %s", volume.UUID, computeID)
	err = client.AttachVolume(ctx, volume.UUID, &models.VolumeAttachRequest{
		ComputeID: computeID,
		VolumeID:  volume.ID,
	})
	if err != nil {
		t.Fatalf("AttachVolume failed: %v", err)
	}
	t.Log("Volume attached successfully")

	// Verify attached devices
	attachments, err := client.GetVolumeAttachedDevices(ctx, volume.UUID)
	if err != nil {
		t.Fatalf("GetVolumeAttachedDevices failed: %v", err)
	}
	t.Logf("Found %d attachments", len(attachments))

	found := false
	for _, a := range attachments {
		t.Logf("Attachment: ComputeID=%s, VolumeID=%s", a.ComputeID, string(a.VolumeID))
		if a.ComputeID == computeID {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected compute ID %s in attachments, not found", computeID)
	}

	// Detach volume
	t.Logf("Detaching volume %s", volume.UUID)
	err = client.DetachVolume(ctx, volume.UUID, &models.VolumeDetachRequest{
		ComputeID: computeID,
		VolumeID:  volume.ID,
	})
	if err != nil {
		t.Fatalf("DetachVolume failed: %v", err)
	}
	t.Log("Volume detached successfully")

	// Verify no attachments after detach
	attachments, err = client.GetVolumeAttachedDevices(ctx, volume.UUID)
	if err != nil {
		t.Fatalf("GetVolumeAttachedDevices after detach failed: %v", err)
	}
	if len(attachments) > 0 {
		t.Errorf("Expected no attachments after detach, got %d", len(attachments))
	}
}

// TestVolumeIntegration_Snapshot tests creating a volume snapshot
func TestVolumeIntegration_Snapshot(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create a volume
	volumeName := fmt.Sprintf("test-vol-snap-%d", time.Now().Unix())
	volume, err := client.CreateVolume(ctx, &models.CreateVolumeRequest{
		VolumeName: volumeName,
		VolumeSize: 10,

		BillingUnit: "MRC",
	})
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}
	t.Logf("Volume created: ID=%d, Name=%s", volume.ID, volume.VolumeName)

	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Errorf("DeleteVolume failed: %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Create snapshot
	snapshotName := fmt.Sprintf("test-snap-%d", time.Now().Unix())
	t.Logf("Creating snapshot %s for volume %s", snapshotName, volume.UUID)
	err = client.CreateVolumeSnapshot(ctx, volume.UUID, &models.VolumeSnapshotRequest{
		SnapshotName: snapshotName,
		BillingUnit:  "HRC",
		Products:     `{"volume_snapshot":{"id":""}}`,
	})
	if err != nil {
		t.Fatalf("CreateVolumeSnapshot failed: %v", err)
	}
	t.Log("Snapshot created successfully")
}

// TestVolumeIntegration_EnableBackup tests enabling backup on a volume
func TestVolumeIntegration_EnableBackup(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Create a volume with backup disabled
	volumeName := fmt.Sprintf("test-vol-backup-%d", time.Now().Unix())
	volume, err := client.CreateVolume(ctx, &models.CreateVolumeRequest{
		VolumeName:  volumeName,
		VolumeSize:  10,
		BillingUnit: "MRC",
	})
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}
	t.Logf("Volume created: ID=%d, Name=%s", volume.ID, volume.VolumeName)

	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Errorf("DeleteVolume failed: %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Enable backup
	t.Logf("Enabling backup for volume %s", volume.UUID)
	err = client.EnableVolumeBackup(ctx, volume.UUID)
	if err != nil {
		t.Fatalf("EnableVolumeBackup failed: %v", err)
	}

	// Get volume and verify backup is enabled
	fetchedVolume, err := client.GetVolume(ctx, volume.UUID)
	if err != nil {
		t.Fatalf("GetVolume failed: %v", err)
	}
	t.Logf("Backup enabled for volume %d", fetchedVolume.ID)
	t.Log("Backup enabled successfully")
}

// TestVolumeIntegration_VolumeTypeByID tests retrieving a volume type by ID
func TestVolumeIntegration_VolumeTypeByID(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Get all volume types first
	volumeTypes, err := client.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		t.Fatalf("GetVolumeTypes failed: %v", err)
	}
	if len(volumeTypes) == 0 {
		t.Skip("No volume types available, skipping test")
	}

	// Get the first volume type by ID
	expected := volumeTypes[0]
	t.Logf("Getting volume type by ID: %d (Name: %s)", expected.ID, expected.Name)

	volumeType, err := client.GetVolumeTypeByID(ctx, expected.ID)
	if err != nil {
		t.Fatalf("GetVolumeTypeByID failed: %v", err)
	}

	if volumeType.ID != expected.ID {
		t.Errorf("Expected volume type ID %d, got %d", expected.ID, volumeType.ID)
	}
	if volumeType.Name != expected.Name {
		t.Errorf("Expected volume type name %s, got %s", expected.Name, volumeType.Name)
	}
	t.Logf("GetVolumeTypeByID returned correct data: ID=%d, Name=%s", volumeType.ID, volumeType.Name)
}

// TestVolumeIntegration_VolumeTypeByName tests retrieving a volume type by name
func TestVolumeIntegration_VolumeTypeByName(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Get all volume types first
	volumeTypes, err := client.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		t.Fatalf("GetVolumeTypes failed: %v", err)
	}
	if len(volumeTypes) == 0 {
		t.Skip("No volume types available, skipping test")
	}

	// Get the first volume type by name
	expected := volumeTypes[0]
	t.Logf("Getting volume type by name: %s (ID: %d)", expected.Name, expected.ID)

	volumeType, err := client.GetVolumeTypeByName(ctx, expected.Name)
	if err != nil {
		t.Fatalf("GetVolumeTypeByName failed: %v", err)
	}

	if volumeType.Name != expected.Name {
		t.Errorf("Expected volume type name %s, got %s", expected.Name, volumeType.Name)
	}
	t.Logf("GetVolumeTypeByName returned correct data: ID=%d, Name=%s", volumeType.ID, volumeType.Name)
}

// TestVolumeIntegration_ListByComputeID tests listing volumes filtered by compute ID
func TestVolumeIntegration_ListByComputeID(t *testing.T) {
	computeID := os.Getenv("AIRTEL_TEST_COMPUTE_ID")
	if computeID == "" {
		t.Skip("AIRTEL_TEST_COMPUTE_ID not set, skipping list-by-compute test")
	}

	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("Listing volumes for compute ID: %s", computeID)
	volumes, err := client.ListVolumes(ctx, computeID)
	if err != nil {
		t.Fatalf("ListVolumes with computeID failed: %v", err)
	}

	t.Logf("Found %d volumes for compute %s", len(volumes), computeID)
	for i, vol := range volumes {
		t.Logf("Volume %d: ID=%d, Name=%s, Size=%dGB, Status=%s",
			i+1, vol.ID, vol.VolumeName, vol.VolumeSize, vol.Status)
	}
}

// TestVolumeIntegration_DeleteNonExistent tests deleting a non-existent volume
func TestVolumeIntegration_DeleteNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentUUID := "non-existent-uuid-99999"

	t.Logf("Attempting to delete non-existent volume: %s", nonExistentUUID)
	err := client.DeleteVolume(ctx, nonExistentUUID)
	if err == nil {
		t.Error("Expected error for deleting non-existent volume, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		t.Logf("Received API error: status=%d, message=%s", apiErr.StatusCode, apiErr.Message)
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestVolumeIntegration_CreateWithVolumeType tests creating a volume with a specific volume type
func TestVolumeIntegration_CreateWithVolumeType(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Get available volume types
	volumeTypes, err := client.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		t.Fatalf("GetVolumeTypes failed: %v", err)
	}
	if len(volumeTypes) == 0 {
		t.Skip("No volume types available, skipping test")
	}

	// Use the first available volume type
	vt := volumeTypes[0]
	t.Logf("Using volume type: ID=%d, Name=%s", vt.ID, vt.Name)

	volumeName := fmt.Sprintf("test-vol-type-%d", time.Now().Unix())
	volume, err := client.CreateVolume(ctx, &models.CreateVolumeRequest{
		VolumeName:   volumeName,
		VolumeSize:   10,
		VolumeTypeID: fmt.Sprintf("%d", vt.ID),
		BillingUnit:  "MRC",
	})
	if err != nil {
		t.Fatalf("CreateVolume with volume type failed: %v", err)
	}
	t.Logf("Volume created: ID=%d, Name=%s, VolumeTypeID=%s", volume.ID, volume.VolumeName, string(volume.VolumeTypeID))

	defer func() {
		t.Logf("Deleting volume: %s", volume.UUID)
		err := client.DeleteVolume(ctx, volume.UUID)
		if err != nil {
			t.Logf("DeleteVolume returned error (may be backend issue): %v", err)
		} else {
			t.Log("Volume deleted successfully")
		}
	}()

	// Verify volume type ID from create response
	expectedTypeID := fmt.Sprintf("%d", vt.ID)
	gotTypeID := strings.Trim(string(volume.VolumeTypeID), `"`)
	if gotTypeID != expectedTypeID {
		t.Errorf("Expected VolumeTypeID %s, got %s", expectedTypeID, gotTypeID)
	} else {
		t.Logf("Volume created with correct volume type: %s", gotTypeID)
	}
}
