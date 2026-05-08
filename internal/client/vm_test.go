package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client/testutil"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func TestResizeCompute(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		flavorID string
		setup    func(ms *testutil.MockServer)
		wantErr  bool
	}{
		{
			name:     "successful resize",
			id:       "test-id",
			flavorID: "2",
			wantErr:  false,
		},
		{
			name:     "resize nonexistent compute",
			id:       "nonexistent-id",
			flavorID: "2",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/nonexistent-id/resize/2", 404, "Not found")
			},
			wantErr: true,
		},
		{
			name:     "resize with invalid flavor",
			id:       "test-id",
			flavorID: "999",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/resize/999", 400, "Invalid flavor")
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

			err := client.ResizeCompute(context.Background(), tt.id, tt.flavorID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResizeCompute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveFlavorID(t *testing.T) {
	tests := []struct {
		name    string
		flavor  string
		setup   func(ms *testutil.MockServer)
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve existing flavor",
			flavor:  "t2.micro",
			wantID:  "1",
			wantErr: false,
		},
		{
			name:    "resolve second flavor",
			flavor:  "t2.small",
			wantID:  "2",
			wantErr: false,
		},
		{
			name:    "flavor not found",
			flavor:  "nonexistent-flavor",
			wantErr: true,
		},
		{
			name:   "server error on list",
			flavor: "t2.micro",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/flavors/", 500, "Internal server error")
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

			id, err := client.ResolveFlavorID(context.Background(), tt.flavor)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveFlavorID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && id != tt.wantID {
				t.Errorf("ResolveFlavorID() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestResolveImageID(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		setup   func(ms *testutil.MockServer)
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve existing image",
			image:   "ubuntu-20.04",
			wantID:  "1",
			wantErr: false,
		},
		{
			name:    "resolve second image",
			image:   "centos-8",
			wantID:  "2",
			wantErr: false,
		},
		{
			name:    "image not found",
			image:   "nonexistent-image",
			wantErr: true,
		},
		{
			name:  "server error on list",
			image: "ubuntu-20.04",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/images/", 500, "Internal server error")
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

			id, err := client.ResolveImageID(context.Background(), tt.image)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveImageID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && id != tt.wantID {
				t.Errorf("ResolveImageID() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestResolveKeypairID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	id, err := client.ResolveKeypairID(context.Background(), "keypair-1")
	if err != nil {
		t.Fatalf("ResolveKeypairID() error = %v", err)
	}
	if id != "1" {
		t.Errorf("ResolveKeypairID() = %v, want 1", id)
	}

	_, err = client.ResolveKeypairID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("ResolveKeypairID() expected error for nonexistent keypair")
	}
}

func TestResolveSecurityGroupID(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	id, err := client.ResolveSecurityGroupID(context.Background(), "default")
	if err != nil {
		t.Fatalf("ResolveSecurityGroupID() error = %v", err)
	}
	if id != 1 {
		t.Errorf("ResolveSecurityGroupID() = %v, want 1", id)
	}

	_, err = client.ResolveSecurityGroupID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("ResolveSecurityGroupID() expected error for nonexistent security group")
	}
}

func TestCreateComputeFormData(t *testing.T) {
	var capturedForm map[string]string

	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	mockServer.AddHandler("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := []models.Compute{{
			ID:           "test-id",
			InstanceName: "test-instance",
			Status:       "ACTIVE",
			FloatingIP:   "10.0.0.100",
		}}
		json.NewEncoder(w).Encode(response)
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	req := &models.CreateComputeRequest{
		InstanceName:   "test-instance",
		ImageID:        "ubuntu-20.04",
		FlavorID:       "1",
		VPCID:          "vpc-1",
		SubnetID:       "subnet-1",
		NetworkID:      "subnet-1",
		AZName:         "south-1a",
		OSType:         "linux",
		VolumeSize:     20,
		BootFromVolume: true,
	}

	compute, err := client.CreateCompute(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateCompute() error = %v", err)
	}
	if compute.ID != "test-id" {
		t.Errorf("CreateCompute() ID = %v, want test-id", compute.ID)
	}

	// Verify form data was sent correctly
	expectedFields := map[string]string{
		"instance_name":    "test-instance",
		"image_id":         "ubuntu-20.04",
		"flavor_id":        "1",
		"vpc_id":           "vpc-1",
		"os_type":          "linux",
		"volume_size":      "20",
		"boot_from_volume": "true",
	}

	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("CreateCompute() form missing field %q", field)
		} else if got != expected {
			t.Errorf("CreateCompute() form field %q = %q, want %q", field, got, expected)
		}
	}
}

func TestPerformComputeAction(t *testing.T) {
	tests := []struct {
		name    string
		action  models.ComputeAction
		setup   func(ms *testutil.MockServer)
		wantErr bool
	}{
		{
			name:   "start action",
			action: models.ComputeActionStart,
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/start", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			wantErr: false,
		},
		{
			name:   "stop action",
			action: models.ComputeActionStop,
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/stop", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			wantErr: false,
		},
		{
			name:   "reboot action",
			action: models.ComputeActionReboot,
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("POST", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/reboot", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			wantErr: false,
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

			err := client.PerformComputeAction(context.Background(), "test-id", tt.action)

			if (err != nil) != tt.wantErr {
				t.Errorf("PerformComputeAction(%s) error = %v, wantErr %v", tt.action, err, tt.wantErr)
			}
		})
	}
}

func TestGetComputeConsoleURL(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	mockServer.AddHandler("GET", "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/console_url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"console_url": "https://console.example.com/test-id"})
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	url, err := client.GetComputeConsoleURL(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("GetComputeConsoleURL() error = %v", err)
	}
	if url != "https://console.example.com/test-id" {
		t.Errorf("GetComputeConsoleURL() = %v, want https://console.example.com/test-id", url)
	}
}

func TestWaitForComputeReady(t *testing.T) {
	computePath := "/api/v2.1/computes/domain/test-org/project/test-project/computes/test-id/"

	t.Run("immediately active", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		mockServer.AddHandler("GET", computePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.Compute{
				ID:     "test-id",
				Status: "ACTIVE",
			})
		})

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		compute, err := c.WaitForComputeReady(context.Background(), "test-id", 30*time.Second)
		if err != nil {
			t.Fatalf("WaitForComputeReady() error = %v", err)
		}
		if compute.Status != "ACTIVE" {
			t.Errorf("WaitForComputeReady() status = %v, want ACTIVE", compute.Status)
		}
	})

	t.Run("transitions from BUILD to ACTIVE", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		var callCount int32
		mockServer.AddHandler("GET", computePath, func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			w.Header().Set("Content-Type", "application/json")
			status := "BUILD"
			if count >= 2 {
				status = "ACTIVE"
			}
			json.NewEncoder(w).Encode(models.Compute{
				ID:     "test-id",
				Status: status,
			})
		})

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		compute, err := c.WaitForComputeReady(context.Background(), "test-id", 30*time.Second)
		if err != nil {
			t.Fatalf("WaitForComputeReady() error = %v", err)
		}
		if compute.Status != "ACTIVE" {
			t.Errorf("WaitForComputeReady() status = %v, want ACTIVE", compute.Status)
		}
	})

	t.Run("error state", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		mockServer.AddHandler("GET", computePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.Compute{
				ID:     "test-id",
				Status: "ERROR",
			})
		})

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		_, err := c.WaitForComputeReady(context.Background(), "test-id", 30*time.Second)
		if err == nil {
			t.Fatal("WaitForComputeReady() expected error for ERROR state")
		}
		if !strings.Contains(err.Error(), "error state") {
			t.Errorf("WaitForComputeReady() error = %v, want error containing 'error state'", err)
		}
	})

	t.Run("timeout exceeded", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		mockServer.AddHandler("GET", computePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.Compute{
				ID:     "test-id",
				Status: "BUILD",
			})
		})

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		_, err := c.WaitForComputeReady(context.Background(), "test-id", 1*time.Second)
		if err == nil {
			t.Fatal("WaitForComputeReady() expected error for timeout")
		}
		if !strings.Contains(err.Error(), "did not become ready") {
			t.Errorf("WaitForComputeReady() error = %v, want error containing 'did not become ready'", err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		mockServer.AddHandler("GET", computePath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.Compute{
				ID:     "test-id",
				Status: "BUILD",
			})
		})

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		_, err := c.WaitForComputeReady(ctx, "test-id", 30*time.Second)
		if err == nil {
			t.Fatal("WaitForComputeReady() expected error for cancelled context")
		}
	})

	t.Run("API error during polling", func(t *testing.T) {
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		mockServer.SetErrorResponse("GET", computePath, 500, "Internal server error")

		baseURL := strings.TrimSuffix(mockServer.URL, "/")
		c, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

		_, err := c.WaitForComputeReady(context.Background(), "test-id", 30*time.Second)
		if err == nil {
			t.Fatal("WaitForComputeReady() expected error for API error")
		}
	})
}
