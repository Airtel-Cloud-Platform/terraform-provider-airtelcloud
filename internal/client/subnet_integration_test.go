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

// Integration test configuration from environment variables
type testConfig struct {
	APIEndpoint  string
	APIKey       string
	APISecret    string
	Region       string
	Organization string
	ProjectName  string
	NetworkID    string // VPC/Network ID to create subnets in
}

func getTestConfig(t *testing.T) *testConfig {
	config := &testConfig{
		APIEndpoint:  os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:       os.Getenv("AIRTEL_API_KEY"),
		APISecret:    os.Getenv("AIRTEL_API_SECRET"),
		Region:       os.Getenv("AIRTEL_REGION"),
		Organization: os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:  os.Getenv("AIRTEL_PROJECT_NAME"),
		NetworkID:    os.Getenv("AIRTEL_TEST_NETWORK_ID"),
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
	if config.NetworkID == "" {
		t.Skip("AIRTEL_TEST_NETWORK_ID not set, skipping integration test")
	}

	// Set defaults
	if config.Region == "" {
		config.Region = "south"
	}

	return config
}

func createTestClient(t *testing.T, config *testConfig) *Client {
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

// TestSubnetIntegration_CreateGetDelete tests the full lifecycle of a subnet
func TestSubnetIntegration_CreateGetDelete(t *testing.T) {
	config := getTestConfig(t)
	client := createTestClient(t, config)
	ctx := context.Background()

	// Generate unique subnet name
	subnetName := fmt.Sprintf("test-subnet-%d", time.Now().Unix())

	// Create subnet request
	createReq := &models.CreateSubnetRequest{
		Name:             subnetName,
		Description:      "Integration test subnet",
		AvailabilityZone: "S2",
		IPv4AddressSpace: "10.100.1.0/24",
		SubnetSubRole:    "Private",
	}

	t.Logf("Creating subnet: %s", subnetName)

	// Create subnet
	subnet, err := client.CreateSubnet(ctx, config.NetworkID, createReq)
	if err != nil {
		t.Fatalf("CreateSubnet failed: %v", err)
	}

	t.Logf("Subnet created with ID: %s", subnet.SubnetID)

	// Verify created subnet fields
	if subnet.Name != subnetName {
		t.Errorf("Expected subnet name %s, got %s", subnetName, subnet.Name)
	}
	if subnet.IPv4AddressSpace != createReq.IPv4AddressSpace {
		t.Errorf("Expected IPv4AddressSpace %s, got %s", createReq.IPv4AddressSpace, subnet.IPv4AddressSpace)
	}
	if subnet.SubnetID == "" {
		t.Error("SubnetID should not be empty")
	}

	// Cleanup: Delete subnet at the end
	defer func() {
		t.Logf("Deleting subnet: %s", subnet.SubnetID)
		err := client.DeleteSubnet(ctx, config.NetworkID, subnet.SubnetID)
		if err != nil {
			t.Errorf("DeleteSubnet failed: %v", err)
		} else {
			t.Log("Subnet deleted successfully")
		}
	}()

	// Get subnet
	t.Logf("Getting subnet: %s", subnet.SubnetID)
	fetchedSubnet, err := client.GetSubnet(ctx, config.NetworkID, subnet.SubnetID)
	if err != nil {
		t.Fatalf("GetSubnet failed: %v", err)
	}

	// Verify fetched subnet
	if fetchedSubnet.SubnetID != subnet.SubnetID {
		t.Errorf("Expected SubnetID %s, got %s", subnet.SubnetID, fetchedSubnet.SubnetID)
	}
	if fetchedSubnet.Name != subnetName {
		t.Errorf("Expected subnet name %s, got %s", subnetName, fetchedSubnet.Name)
	}

	t.Log("GetSubnet returned correct data")
}

// TestSubnetIntegration_ListSubnets tests listing subnets in a network
func TestSubnetIntegration_ListSubnets(t *testing.T) {
	config := getTestConfig(t)
	client := createTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization (domain): %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)
	t.Logf("Network ID: %s", config.NetworkID)
	t.Logf("Expected URL path: /api/network-manager/v1/domain/%s/project/%s/network/%s/subnets?limit=1000",
		config.Organization, config.ProjectName, config.NetworkID)

	// List subnets
	response, err := client.ListSubnets(ctx, config.NetworkID)
	if err != nil {
		t.Fatalf("ListSubnets failed: %v", err)
	}

	t.Logf("Found %d subnets", response.Count)

	// Log subnet details
	for i, subnet := range response.Items {
		t.Logf("Subnet %d: ID=%s, Name=%s, CIDR=%s, State=%s",
			i+1, subnet.SubnetID, subnet.Name, subnet.IPv4AddressSpace, subnet.State)
	}
}

// TestSubnetIntegration_UpdateSubnet tests updating a subnet
func TestSubnetIntegration_UpdateSubnet(t *testing.T) {
	config := getTestConfig(t)
	client := createTestClient(t, config)
	ctx := context.Background()

	// Generate unique subnet name
	subnetName := fmt.Sprintf("test-subnet-update-%d", time.Now().Unix())
	updatedDescription := "Updated integration test subnet"

	// Create subnet request
	createReq := &models.CreateSubnetRequest{
		Name:             subnetName,
		Description:      "Original description",
		AvailabilityZone: "S2",
		IPv4AddressSpace: "10.100.2.0/24",
		SubnetSubRole:    "Private",
	}

	t.Logf("Creating subnet for update test: %s", subnetName)

	// Create subnet
	subnet, err := client.CreateSubnet(ctx, config.NetworkID, createReq)
	if err != nil {
		t.Fatalf("CreateSubnet failed: %v", err)
	}

	t.Logf("Subnet created with ID: %s", subnet.SubnetID)

	// Cleanup: Delete subnet at the end
	defer func() {
		t.Logf("Deleting subnet: %s", subnet.SubnetID)
		err := client.DeleteSubnet(ctx, config.NetworkID, subnet.SubnetID)
		if err != nil {
			t.Errorf("DeleteSubnet failed: %v", err)
		} else {
			t.Log("Subnet deleted successfully")
		}
	}()

	// Update subnet
	updateReq := &models.UpdateSubnetRequest{
		Description: updatedDescription,
	}

	t.Logf("Updating subnet: %s", subnet.SubnetID)
	updatedSubnet, err := client.UpdateSubnet(ctx, config.NetworkID, subnet.SubnetID, updateReq)
	if err != nil {
		t.Fatalf("UpdateSubnet failed: %v", err)
	}

	// Verify update
	if updatedSubnet.Description != updatedDescription {
		t.Errorf("Expected description %s, got %s", updatedDescription, updatedSubnet.Description)
	}

	t.Log("Subnet updated successfully")
}

// TestSubnetIntegration_GetNonExistent tests getting a non-existent subnet
func TestSubnetIntegration_GetNonExistent(t *testing.T) {
	config := getTestConfig(t)
	client := createTestClient(t, config)
	ctx := context.Background()

	nonExistentID := "non-existent-subnet-id-12345"

	t.Logf("Attempting to get non-existent subnet: %s", nonExistentID)

	_, err := client.GetSubnet(ctx, config.NetworkID, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent subnet, got nil")
		return
	}

	// Check if it's a 404 error
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent subnet")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// TestSubnetIntegration_CreateWithLabels tests creating a subnet with labels
func TestSubnetIntegration_CreateWithLabels(t *testing.T) {
	config := getTestConfig(t)
	client := createTestClient(t, config)
	ctx := context.Background()

	// Generate unique subnet name
	subnetName := fmt.Sprintf("test-subnet-labels-%d", time.Now().Unix())

	// Create subnet request with labels
	createReq := &models.CreateSubnetRequest{
		Name:             subnetName,
		Description:      "Integration test subnet with labels",
		AvailabilityZone: "S2",
		IPv4AddressSpace: "10.100.3.0/24",
		SubnetSubRole:    "Private",
		Labels:           []string{"environment:test", "team:integration"},
	}

	t.Logf("Creating subnet with labels: %s", subnetName)

	// Create subnet
	subnet, err := client.CreateSubnet(ctx, config.NetworkID, createReq)
	if err != nil {
		t.Fatalf("CreateSubnet failed: %v", err)
	}

	t.Logf("Subnet created with ID: %s", subnet.SubnetID)

	// Cleanup: Delete subnet at the end
	defer func() {
		t.Logf("Deleting subnet: %s", subnet.SubnetID)
		err := client.DeleteSubnet(ctx, config.NetworkID, subnet.SubnetID)
		if err != nil {
			t.Errorf("DeleteSubnet failed: %v", err)
		} else {
			t.Log("Subnet deleted successfully")
		}
	}()

	// Verify labels
	if len(subnet.Labels) != len(createReq.Labels) {
		t.Errorf("Expected %d labels, got %d", len(createReq.Labels), len(subnet.Labels))
	}

	t.Logf("Subnet created with %d labels", len(subnet.Labels))
}
