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

func TestCreateProtection(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.CreateProtectionRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreateProtectionRequest{
				Name:           "test-protection",
				ComputeID:      "compute-1",
				ProtectionPlan: "daily-plan",
			},
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protections/", 500, "Internal server error")
			},
			request: &models.CreateProtectionRequest{
				Name: "test-protection",
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

			protection, err := client.CreateProtection(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProtection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if protection == nil {
					t.Error("CreateProtection() returned nil")
					return
				}
				if protection.ID != 1 {
					t.Errorf("CreateProtection() ID = %v, want 1", protection.ID)
				}
				if protection.Name != "test-protection" {
					t.Errorf("CreateProtection() Name = %v, want test-protection", protection.Name)
				}
				if protection.Status != "ACTIVE" {
					t.Errorf("CreateProtection() Status = %v, want ACTIVE", protection.Status)
				}
			}
		})
	}
}

func TestGetProtection(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "successful retrieval",
			id:      1,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protection, err := client.GetProtection(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProtection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if protection == nil {
					t.Error("GetProtection() returned nil")
					return
				}
				if protection.Name != "test-protection" {
					t.Errorf("GetProtection() Name = %v, want test-protection", protection.Name)
				}
			}
		})
	}
}

func TestListProtections(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 1,
		},
		{
			name: "empty list",
			setup: func(ms *testutil.MockServer) {
				ms.AddHandler("GET", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protections/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.VeritasProtection{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protections/", 500, "Internal server error")
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

			protections, err := client.ListProtections(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListProtections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(protections) != tt.wantCount {
					t.Errorf("ListProtections() count = %d, want %d", len(protections), tt.wantCount)
				}
			}
		})
	}
}

func TestUpdateProtection(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	updateReq := &models.UpdateProtectionRequest{
		Name:        "updated-protection",
		Description: "Updated protection policy",
	}

	protection, err := client.UpdateProtection(context.Background(), 1, updateReq)
	if err != nil {
		t.Fatalf("UpdateProtection() error = %v", err)
	}

	if protection.Name != "updated-protection" {
		t.Errorf("UpdateProtection() Name = %v, want updated-protection", protection.Name)
	}
}

func TestDeleteProtection(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		setup   func(ms *testutil.MockServer)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			id:      1,
			wantErr: false,
		},
		{
			name: "already deleted (404 is not an error)",
			id:   1,
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("DELETE", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protections/1/", 404, "Not found")
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

			err := client.DeleteProtection(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateProtectionPlan(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	createReq := &models.CreateProtectionPlanRequest{
		Name:          "test-plan",
		Retention:     1,
		RetentionUnit: "DAYS",
		Recurrence:    86400,
		SelectorKey:   "AZ",
		SelectorValue: "S1",
	}

	plan, err := client.CreateProtectionPlan(context.Background(), createReq, "test-subnet-id")
	if err != nil {
		t.Fatalf("CreateProtectionPlan() error = %v", err)
	}

	if plan.ID != "plan-uuid-1234" {
		t.Errorf("CreateProtectionPlan() ID = %v, want plan-uuid-1234", plan.ID)
	}
}

func TestGetProtectionPlan(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	plan, err := client.GetProtectionPlan(context.Background(), "plan-uuid-1234", "test-subnet-id")
	if err != nil {
		t.Fatalf("GetProtectionPlan() error = %v", err)
	}

	if plan.ID != "plan-uuid-1234" {
		t.Errorf("GetProtectionPlan() ID = %v, want plan-uuid-1234", plan.ID)
	}
}

func TestCreateProtectionPlanFormData(t *testing.T) {
	var capturedForm map[string]string
	var capturedContentType string
	var capturedSubnetID string

	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	// Override the POST handler to capture form data
	mockServer.AddHandler("POST", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protection_plans/", func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedSubnetID = r.Header.Get("subnet-id")
		r.ParseForm()
		capturedForm = make(map[string]string)
		for key := range r.PostForm {
			capturedForm[key] = r.PostFormValue(key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Protection plan created and assigned successfully")
	})

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	req := &models.CreateProtectionPlanRequest{
		Name:          "test-plan",
		Description:   "Test plan description",
		Retention:     1,
		RetentionUnit: "DAYS",
		Recurrence:    86400,
		SelectorKey:   "AZ",
		SelectorValue: "S1",
	}

	plan, err := client.CreateProtectionPlan(context.Background(), req, "test-subnet-id")
	if err != nil {
		t.Fatalf("CreateProtectionPlan() error = %v", err)
	}
	if plan == nil {
		t.Fatal("CreateProtectionPlan() returned nil")
	}

	// Verify Content-Type
	if !strings.HasPrefix(capturedContentType, "application/x-www-form-urlencoded") {
		t.Errorf("Content-Type = %v, want application/x-www-form-urlencoded", capturedContentType)
	}

	// Verify subnet-id header was sent
	if capturedSubnetID != "test-subnet-id" {
		t.Errorf("subnet-id header = %v, want test-subnet-id", capturedSubnetID)
	}

	// Verify form data was sent correctly
	expectedFields := map[string]string{
		"name":           "test-plan",
		"description":    "Test plan description",
		"retention":      "1",
		"retention_unit": "DAYS",
		"recurrence":     "86400",
		"selector_key":   "AZ",
		"selector_value": "S1",
	}

	for field, expected := range expectedFields {
		if got, ok := capturedForm[field]; !ok {
			t.Errorf("Form field %q not found in request", field)
		} else if got != expected {
			t.Errorf("Form field %q = %v, want %v", field, got, expected)
		}
	}
}

func TestListProtectionPlans(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 1,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/backups/domain/test-org/project/test-project/backups/protection_plans/", 500, "Internal server error")
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

			plans, err := client.ListProtectionPlans(context.Background(), "test-subnet-id")

			if (err != nil) != tt.wantErr {
				t.Errorf("ListProtectionPlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(plans) != tt.wantCount {
					t.Errorf("ListProtectionPlans() count = %d, want %d", len(plans), tt.wantCount)
				}
			}
		})
	}
}
