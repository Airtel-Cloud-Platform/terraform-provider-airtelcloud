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

func TestCreateLBService(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ms *testutil.MockServer)
		request *models.CreateLBServiceRequest
		wantErr bool
	}{
		{
			name: "successful creation",
			request: &models.CreateLBServiceRequest{
				Name:      "test-lb",
				FlavorID:  1,
				NetworkID: "net-1",
				VPCID:     "vpc-1",
				VPCName:   "test-vpc",
			},
			wantErr: false,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("POST", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/", 500, "Internal server error")
			},
			request: &models.CreateLBServiceRequest{
				Name: "test-lb",
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

			svc, err := client.CreateLBService(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLBService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if svc == nil {
					t.Error("CreateLBService() returned nil")
					return
				}
				if svc.ID != "lb-svc-1" {
					t.Errorf("CreateLBService() ID = %v, want lb-svc-1", svc.ID)
				}
				if svc.Status != "Active" {
					t.Errorf("CreateLBService() Status = %v, want Active", svc.Status)
				}
			}
		})
	}
}

func TestGetLBService(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "successful retrieval",
			id:      "lb-svc-1",
			wantErr: false,
		},
		{
			name:    "not found",
			id:      "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := client.GetLBService(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLBService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if svc == nil {
					t.Error("GetLBService() returned nil")
					return
				}
				if svc.Name != "test-lb" {
					t.Errorf("GetLBService() Name = %v, want test-lb", svc.Name)
				}
				if svc.FlavorID != 1 {
					t.Errorf("GetLBService() FlavorID = %v, want 1", svc.FlavorID)
				}
			}
		})
	}
}

func TestListLBServices(t *testing.T) {
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
				ms.AddHandler("GET", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]models.LBService{})
				})
			},
			wantCount: 0,
		},
		{
			name: "server error",
			setup: func(ms *testutil.MockServer) {
				ms.SetErrorResponse("GET", "/api/v2.1/load-balancers/domain/test-org/project/test-project/load-balancers/lb_service/", 500, "Internal server error")
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

			services, err := client.ListLBServices(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListLBServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(services) != tt.wantCount {
					t.Errorf("ListLBServices() count = %d, want %d", len(services), tt.wantCount)
				}
			}
		})
	}
}

func TestDeleteLBService(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeleteLBService(context.Background(), "lb-svc-1")
	if err != nil {
		t.Errorf("DeleteLBService() error = %v", err)
	}
}

func TestCreateLBVip(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	vip, err := client.CreateLBVip(context.Background(), "lb-svc-1")
	if err != nil {
		t.Fatalf("CreateLBVip() error = %v", err)
	}

	if vip.ID != 1 {
		t.Errorf("CreateLBVip() ID = %v, want 1", vip.ID)
	}
	if vip.Status != "Active" {
		t.Errorf("CreateLBVip() Status = %v, want Active", vip.Status)
	}
	if len(vip.FixedIPs) != 1 || vip.FixedIPs[0] != "10.0.0.100" {
		t.Errorf("CreateLBVip() FixedIPs = %v, want [10.0.0.100]", vip.FixedIPs)
	}
}

func TestListLBVips(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	vips, err := client.ListLBVips(context.Background(), "lb-svc-1")
	if err != nil {
		t.Fatalf("ListLBVips() error = %v", err)
	}

	if len(vips) != 1 {
		t.Errorf("ListLBVips() count = %d, want 1", len(vips))
	}
}

func TestDeleteLBVip(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeleteLBVip(context.Background(), "lb-svc-1", 1)
	if err != nil {
		t.Errorf("DeleteLBVip() error = %v", err)
	}
}

func TestCreateLBCertificate(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	req := &models.CreateLBCertificateRequest{
		Name:      "test-cert",
		SSLCert:   "-----BEGIN CERTIFICATE-----\nMIIBxTCCAW...",
		SSLPvtKey: "-----BEGIN PRIVATE KEY-----\nMIIEvwIBAD...",
	}

	cert, err := client.CreateLBCertificate(context.Background(), "lb-svc-1", req)
	if err != nil {
		t.Fatalf("CreateLBCertificate() error = %v", err)
	}

	if cert.ID != 1 {
		t.Errorf("CreateLBCertificate() ID = %v, want 1", cert.ID)
	}
	if cert.Name != "test-cert" {
		t.Errorf("CreateLBCertificate() Name = %v, want test-cert", cert.Name)
	}
}

func TestListLBCertificates(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	certs, err := client.ListLBCertificates(context.Background(), "lb-svc-1")
	if err != nil {
		t.Fatalf("ListLBCertificates() error = %v", err)
	}

	if len(certs) != 1 {
		t.Errorf("ListLBCertificates() count = %d, want 1", len(certs))
	}
}

func TestDeleteLBCertificate(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	baseURL := strings.TrimSuffix(mockServer.URL, "/")
	client, _ := NewClient(baseURL, "test-api-key", "test-api-secret", "south-1", "test-org", "test-project", "")

	err := client.DeleteLBCertificate(context.Background(), "lb-svc-1", 1)
	if err != nil {
		t.Errorf("DeleteLBCertificate() error = %v", err)
	}
}

func TestListLBFlavors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(ms *testutil.MockServer)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful list",
			wantCount: 2,
		},
		{
			name: "server error",
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

			flavors, err := client.ListLBFlavors(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListLBFlavors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(flavors) != tt.wantCount {
					t.Errorf("ListLBFlavors() count = %d, want %d", len(flavors), tt.wantCount)
				}
			}
		})
	}
}
