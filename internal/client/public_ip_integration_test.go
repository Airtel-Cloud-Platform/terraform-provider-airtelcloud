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

// TestPublicIPIntegration_List tests listing all public IPs
func TestPublicIPIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	t.Logf("IPAM base path: %s (will be auto-prefixed with /api)", client.ipamBasePath())

	resp, err := client.ListPublicIPs(ctx)
	if err != nil {
		t.Fatalf("ListPublicIPs failed: %v", err)
	}

	t.Logf("Found %d public IPs", len(resp.Items))

	for i, ip := range resp.Items {
		t.Logf("PublicIP %d: UUID=%s, IP=%s, PublicIP=%s, ObjectName=%s, Status=%s, AZ=%s, TargetVIP=%s",
			i+1, ip.UUID, ip.IP, ip.PublicIP, ip.ObjectName, ip.Status, ip.AZName, ip.TargetVIP)
	}
}

// TestPublicIPIntegration_CreateGetDelete tests the full lifecycle of a public IP
func TestPublicIPIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	az := os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE")
	if az == "" {
		az = "S1"
	}

	// We need a private IP (VIP) to NAT against. Use env var or a placeholder.
	vip := os.Getenv("AIRTEL_TEST_VIP")
	if vip == "" {
		// List existing computes to find a usable private IP
		computes, err := client.ListComputes(ctx)
		if err != nil {
			t.Fatalf("ListComputes failed (needed to find a VIP): %v", err)
		}
		if len(computes) == 0 {
			t.Skip("No compute instances available to get a private IP for public IP allocation")
		}
		// Use the first compute's private IP
		for _, c := range computes {
			if pip := c.PrivateIP(); pip != "" {
				vip = pip
				t.Logf("Using private IP from compute %s (%s): %s", c.ID, c.InstanceName, vip)
				break
			}
		}
		if vip == "" {
			t.Skip("No compute with a private IP found, skipping public IP create test")
		}
	}

	objectName := fmt.Sprintf("test-pip-%d", time.Now().Unix())

	createReq := &models.CreatePublicIPRequest{
		ObjectName: objectName,
		VIP:        vip,
	}

	t.Logf("Creating public IP: name=%s, vip=%s, az=%s", objectName, vip, az)

	created, err := client.CreatePublicIP(ctx, createReq, az)
	if err != nil {
		t.Fatalf("CreatePublicIP failed: %v", err)
	}

	t.Logf("Public IP created: UUID=%s, PublicIP=%s", created.UUID, created.PublicIP)

	if created.UUID == "" {
		t.Fatal("Created public IP UUID should not be empty")
	}

	// Cleanup: delete at the end
	defer func() {
		t.Logf("Deleting public IP: %s", created.UUID)
		err := client.DeletePublicIP(ctx, created.UUID)
		if err != nil {
			t.Logf("DeletePublicIP returned error: %v", err)
		} else {
			t.Log("Public IP deleted successfully")
		}
	}()

	// Wait for ready
	t.Logf("Waiting for public IP %s to become ready...", created.UUID)
	readyIP, err := client.WaitForPublicIPReady(ctx, created.UUID, 5*time.Minute)
	if err != nil {
		t.Fatalf("WaitForPublicIPReady failed: %v", err)
	}

	t.Logf("Public IP ready: UUID=%s, IP=%s, Status=%s, AZ=%s, Domain=%s",
		readyIP.UUID, readyIP.IP, readyIP.Status, readyIP.AZName, readyIP.Domain)

	// Get by UUID
	t.Logf("Getting public IP: %s", created.UUID)
	fetched, err := client.GetPublicIP(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetPublicIP failed: %v", err)
	}

	if fetched.UUID != created.UUID {
		t.Errorf("Expected UUID %s, got %s", created.UUID, fetched.UUID)
	}
	if fetched.ObjectName != objectName {
		t.Errorf("Expected ObjectName %s, got %s", objectName, fetched.ObjectName)
	}
	t.Logf("GetPublicIP returned: UUID=%s, IP=%s, ObjectName=%s, Status=%s, TargetVIP=%s",
		fetched.UUID, fetched.IP, fetched.ObjectName, fetched.Status, fetched.TargetVIP)

	// Verify it appears in list
	t.Log("Verifying public IP appears in list")
	listResp, err := client.ListPublicIPs(ctx)
	if err != nil {
		t.Fatalf("ListPublicIPs failed: %v", err)
	}

	found := false
	for _, ip := range listResp.Items {
		if ip.UUID == created.UUID {
			found = true
			t.Logf("Public IP found in list: UUID=%s, IP=%s, Status=%s", ip.UUID, ip.IP, ip.Status)
			break
		}
	}
	if !found {
		t.Errorf("Public IP %s not found in list of %d IPs", created.UUID, len(listResp.Items))
	}

	t.Log("Public IP create, get, list verification passed")
}

// TestPublicIPIntegration_GetNonExistent tests getting a non-existent public IP
func TestPublicIPIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentUUID := "non-existent-public-ip-uuid-99999"

	t.Logf("Attempting to get non-existent public IP: %s", nonExistentUUID)

	_, err := client.GetPublicIP(ctx, nonExistentUUID)
	if err == nil {
		t.Error("Expected error for non-existent public IP, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d: %s", apiErr.StatusCode, apiErr.Message)
		} else {
			t.Log("Correctly received 404 for non-existent public IP")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestPublicIPIntegration_ListIPAMServices tests listing IPAM services
func TestPublicIPIntegration_ListIPAMServices(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	az := os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE")
	if az == "" {
		az = "S1"
	}

	t.Logf("Listing IPAM services for AZ: %s", az)

	services, err := client.ListIPAMServices(ctx, az)
	if err != nil {
		t.Fatalf("ListIPAMServices failed: %v", err)
	}

	t.Logf("Found %d IPAM services", len(services))

	for i, svc := range services {
		t.Logf("Service %d: UUID=%s, Name=%s, PortRange=%s, ProtoType=%v, IsDefault=%v",
			i+1, svc.UUID, svc.Name, svc.PortRange, svc.ProtoType, svc.IsDefault)
	}
}
