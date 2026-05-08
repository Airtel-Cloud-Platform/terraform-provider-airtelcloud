package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestCreateVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		request *models.CreateVolumeRequest
		wantErr bool
	}{
		{
			name: "successful volume creation",
			request: &models.CreateVolumeRequest{
				VolumeName:  "test-volume",
				VolumeSize:  10,
				VolumeType:  "gp2",
				BillingUnit: "monthly",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, err := client.CreateVolume(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if volume == nil {
					t.Error("CreateVolume() returned nil volume")
					return
				}

				if volume.UUID == "" {
					t.Error("CreateVolume() returned volume with empty UUID")
				}

				if volume.VolumeName != tt.request.VolumeName {
					t.Errorf("CreateVolume() volume name = %v, want %v",
						volume.VolumeName, tt.request.VolumeName)
				}

				if volume.VolumeSize != tt.request.VolumeSize {
					t.Errorf("CreateVolume() volume size = %v, want %v",
						volume.VolumeSize, tt.request.VolumeSize)
				}
			}
		})
	}
}

func TestGetVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name     string
		uuid     string
		wantErr  bool
		wantName string
	}{
		{
			name:     "successful volume retrieval",
			uuid:     "test-uuid-1",
			wantErr:  false,
			wantName: "test-volume",
		},
		{
			name:    "volume not found",
			uuid:    "nonexistent-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up error response for nonexistent volume
			if tt.uuid == "nonexistent-uuid" {
				mockServer.SetErrorResponse("GET", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/nonexistent-uuid/", 404, "Not found")
			}

			volume, err := client.GetVolume(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if volume == nil {
					t.Error("GetVolume() returned nil volume")
					return
				}

				if volume.VolumeName != tt.wantName {
					t.Errorf("GetVolume() volume name = %v, want %v",
						volume.VolumeName, tt.wantName)
				}

				if volume.UUID != tt.uuid {
					t.Errorf("GetVolume() volume UUID = %v, want %v",
						volume.UUID, tt.uuid)
				}
			}
		})
	}
}

func TestGetVolumeWithoutUUID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Register a custom handler that returns a valid volume WITHOUT uuid (mimics real API)
	mockServer.AddHandler("GET", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/no-uuid-vol/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Volume{
			ID:               1,
			ProviderVolumeID: "provider-vol-id",
			VolumeName:       "test-volume",
			VolumeSize:       10,
			Status:           "available",
			AvailabilityZone: "south-1a",
			VolumeTypeID:     json.RawMessage(`"gp2"`),
		})
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	volume, err := client.GetVolume(context.Background(), "no-uuid-vol")
	if err != nil {
		t.Fatalf("GetVolume() should succeed when UUID is missing from response, got error: %v", err)
	}
	if volume == nil {
		t.Fatal("GetVolume() returned nil volume")
	}
	if volume.ID != 1 {
		t.Errorf("GetVolume() volume ID = %d, want 1", volume.ID)
	}
	if volume.VolumeName != "test-volume" {
		t.Errorf("GetVolume() volume name = %q, want %q", volume.VolumeName, "test-volume")
	}
	if volume.UUID != "" {
		t.Errorf("GetVolume() volume UUID = %q, want empty", volume.UUID)
	}
}

func TestUpdateVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		request *models.UpdateVolumeRequest
		wantErr bool
	}{
		{
			name: "successful volume update",
			uuid: "test-uuid-1",
			request: &models.UpdateVolumeRequest{
				VolumeName:   "test-volume",
				VolumeSize:   20,
				Bootable:     false,
				EnableBackup: false,
				VolumeType:   "gp2",
				BillingUnit:  "MRC",
				VolumeRate:   0,
				ComputeID:    "test-id",
				IsEncrypted:  "encrypted",
			},
			wantErr: false,
		},
		{
			name: "update nonexistent volume",
			uuid: "nonexistent-uuid",
			request: &models.UpdateVolumeRequest{
				VolumeName:  "test-volume",
				VolumeSize:  20,
				VolumeType:  "gp2",
				BillingUnit: "MRC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up error response for nonexistent volume
			if tt.uuid == "nonexistent-uuid" {
				mockServer.SetErrorResponse("PUT", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/nonexistent-uuid/", 404, "Not found")
			}

			err := client.UpdateVolume(context.Background(), tt.uuid, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateVolumeFormData(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Add a custom handler that captures and validates form data
	var capturedForm map[string]string
	mockServer.AddHandler("PUT", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}
		w.WriteHeader(http.StatusOK)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	updateReq := &models.UpdateVolumeRequest{
		VolumeName:   "my-volume",
		VolumeSize:   50,
		Bootable:     true,
		EnableBackup: false,
		VolumeType:   "SSD",
		BillingUnit:  "MRC",
		VolumeRate:   0,
		ComputeID:    "compute-123",
		IsEncrypted:  "encrypted",
	}

	err := client.UpdateVolume(context.Background(), "test-uuid-1", updateReq)
	if err != nil {
		t.Fatalf("UpdateVolume() unexpected error: %v", err)
	}

	expectedFields := map[string]string{
		"volume_name":   "my-volume",
		"volume_size":   "50",
		"bootable":      "true",
		"enable_backup": "false",
		"volume_type":   "SSD",
		"billing_unit":  "MRC",
		"volume_rate":   "0",
		"compute_id":    "compute-123",
		"is_encrypted":  "encrypted",
	}

	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("UpdateVolume() missing form field %q", field)
		} else if got != expected {
			t.Errorf("UpdateVolume() form field %q = %q, want %q", field, got, expected)
		}
	}
}

func TestDeleteVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Create client configured to use the mock server
	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{
			name:    "successful volume deletion",
			uuid:    "test-uuid-1",
			wantErr: false,
		},
		{
			name:    "delete nonexistent volume",
			uuid:    "nonexistent-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up error response for nonexistent volume
			if tt.uuid == "nonexistent-uuid" {
				mockServer.SetErrorResponse("DELETE", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/nonexistent-uuid/", 404, "Not found")
			}

			err := client.DeleteVolume(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttachVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		request *models.VolumeAttachRequest
		wantErr bool
	}{
		{
			name: "successful volume attachment",
			uuid: "test-uuid-1",
			request: &models.VolumeAttachRequest{
				ComputeID: "test-compute-id",
				VolumeID:  1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.AttachVolume(context.Background(), tt.uuid, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("AttachVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttachVolumeFormData(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	var capturedForm map[string]string
	mockServer.AddHandler("PUT", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/volume_attach/test-uuid-1/", func(w http.ResponseWriter, r *http.Request) {
		// Should not be called — attach uses POST
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mockServer.AddHandler("POST", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/volume_attach/test-uuid-1/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}
		w.WriteHeader(http.StatusOK)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.AttachVolume(context.Background(), "test-uuid-1", &models.VolumeAttachRequest{
		ComputeID: "compute-abc",
		VolumeID:  12615,
	})
	if err != nil {
		t.Fatalf("AttachVolume() unexpected error: %v", err)
	}

	expectedFields := map[string]string{
		"compute_id": "compute-abc",
		"volume_id":  "12615",
	}
	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("AttachVolume() missing form field %q", field)
		} else if got != expected {
			t.Errorf("AttachVolume() form field %q = %q, want %q", field, got, expected)
		}
	}
}

func TestDetachVolume(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		request *models.VolumeDetachRequest
		wantErr bool
	}{
		{
			name: "successful volume detachment",
			uuid: "test-uuid-1",
			request: &models.VolumeDetachRequest{
				ComputeID: "test-id",
				VolumeID:  1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DetachVolume(context.Background(), tt.uuid, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetachVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetachVolumeFormData(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	var capturedForm map[string]string
	mockServer.AddHandler("POST", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/volume_detach/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}
		w.WriteHeader(http.StatusOK)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DetachVolume(context.Background(), "test-uuid-1", &models.VolumeDetachRequest{
		ComputeID: "compute-xyz",
		VolumeID:  12584,
	})
	if err != nil {
		t.Fatalf("DetachVolume() unexpected error: %v", err)
	}

	expectedFields := map[string]string{
		"compute_id": "compute-xyz",
		"volume_id":  "12584",
	}
	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("DetachVolume() missing form field %q", field)
		} else if got != expected {
			t.Errorf("DetachVolume() form field %q = %q, want %q", field, got, expected)
		}
	}
}

func TestCreateVolumeSnapshot(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		uuid    string
		request *models.VolumeSnapshotRequest
		wantErr bool
	}{
		{
			name: "successful snapshot creation",
			uuid: "test-uuid-1",
			request: &models.VolumeSnapshotRequest{
				SnapshotName: "my-snapshot",
				BillingUnit:  "HRC",
				Products:     `{"volume_snapshot":{"id":""}}`,
			},
			wantErr: false,
		},
		{
			name: "snapshot nonexistent volume",
			uuid: "nonexistent-uuid",
			request: &models.VolumeSnapshotRequest{
				SnapshotName: "my-snapshot",
				BillingUnit:  "HRC",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.uuid == "nonexistent-uuid" {
				mockServer.SetErrorResponse("POST", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/nonexistent-uuid/snapshots/", 404, "Not found")
			}

			err := client.CreateVolumeSnapshot(context.Background(), tt.uuid, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateVolumeSnapshotFormData(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	var capturedForm map[string]string
	mockServer.AddHandler("POST", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/test-uuid-1/snapshots/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}
		w.WriteHeader(http.StatusCreated)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.CreateVolumeSnapshot(context.Background(), "test-uuid-1", &models.VolumeSnapshotRequest{
		SnapshotName: "snap-test",
		BillingUnit:  "HRC",
		Products:     `{"volume_snapshot":{"id":""}}`,
	})
	if err != nil {
		t.Fatalf("CreateVolumeSnapshot() unexpected error: %v", err)
	}

	expectedFields := map[string]string{
		"snapshot_name": "snap-test",
		"billing_unit":  "HRC",
		"products":      `{"volume_snapshot":{"id":""}}`,
	}
	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("CreateVolumeSnapshot() missing form field %q", field)
		} else if got != expected {
			t.Errorf("CreateVolumeSnapshot() form field %q = %q, want %q", field, got, expected)
		}
	}
}

func TestListVolumes(t *testing.T) {
	tests := []struct {
		name      string
		computeID string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list all",
			computeID: "",
			wantCount: 2,
		},
		{
			name:      "list with compute_id filter",
			computeID: "test-compute-id",
			wantCount: 1,
		},
		{
			name:      "empty list",
			computeID: "",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.Volume{})
				})
			},
			wantCount: 0,
		},
		{
			name:      "server error",
			computeID: "",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/volumes/domain/test-org/project/test-project/volumes/", 500, "Internal server error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := testutil.NewMockServer()
			defer mockServer.Close()

			if tt.setup != nil {
				tt.setup(mockServer)
			}

			baseURL := strings.TrimSuffix(mockServer.URL, "/")
			client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

			volumes, err := client.ListVolumes(context.Background(), tt.computeID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(volumes) != tt.wantCount {
					t.Errorf("ListVolumes() count = %d, want %d", len(volumes), tt.wantCount)
				}
				if tt.wantCount == 2 {
					if volumes[0].VolumeName != "vol-1" {
						t.Errorf("ListVolumes() first Name = %v, want vol-1", volumes[0].VolumeName)
					}
					if volumes[1].ID != 2 {
						t.Errorf("ListVolumes() second ID = %v, want 2", volumes[1].ID)
					}
				}
			}
		})
	}
}

func TestGetVolumeTypes(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name      string
		group     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "all volume types",
			group:     "",
			wantCount: 4,
		},
		{
			name:      "block storage types only",
			group:     "BLOCK_STORAGE",
			wantCount: 3,
		},
		{
			name:      "file storage types only",
			group:     "FILE_STORAGE",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volumeTypes, err := client.GetVolumeTypes(context.Background(), tt.group)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetVolumeTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(volumeTypes) != tt.wantCount {
					t.Errorf("GetVolumeTypes() count = %d, want %d", len(volumeTypes), tt.wantCount)
				}
			}
		})
	}
}
