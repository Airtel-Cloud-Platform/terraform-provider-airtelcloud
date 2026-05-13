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

// vpcTestConfig contains VPC-specific test configuration
type vpcTestConfig struct {
	APIEndpoint  string
	APIKey       string
	APISecret    string
	Region       string
	Organization string
	ProjectName  string
	SubnetID     string
}

func getVPCTestConfig(t *testing.T) *vpcTestConfig {
	config := &vpcTestConfig{
		APIEndpoint:  os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:       os.Getenv("AIRTEL_API_KEY"),
		APISecret:    os.Getenv("AIRTEL_API_SECRET"),
		Region:       os.Getenv("AIRTEL_REGION"),
		Organization: os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:  os.Getenv("AIRTEL_PROJECT_NAME"),
		SubnetID:     os.Getenv("AIRTEL_TEST_SUBNET_ID"),
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
	if config.Region == "" {
		t.Skip("AIRTEL_REGION not set, skipping integration test")
	}
	if config.Organization == "" {
		t.Skip("AIRTEL_ORGANIZATION not set, skipping integration test")
	}
	if config.ProjectName == "" {
		t.Skip("AIRTEL_PROJECT_NAME not set, skipping integration test")
	}

	return config
}

func createVPCTestClient(t *testing.T, config *vpcTestConfig) *Client {
	client, err := NewClient(
		config.APIEndpoint,
		config.APIKey,
		config.APISecret,
		config.Region,
		config.Organization,
		config.ProjectName,
		config.SubnetID,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// TestVPCIntegration_CreateGetDelete tests the full lifecycle of a VPC
func TestVPCIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Generate unique VPC name
	vpcName := fmt.Sprintf("test-vpc-%d", time.Now().Unix())

	// Create VPC request
	createReq := &models.CreateVPCRequest{
		Name:               vpcName,
		CIDRBlock:          "10.200.0.0/16",
		EnableDNSHostnames: true,
		EnableDNSSupport:   true,
		Tags: []models.Tag{
			{Key: "env", Value: "test"},
			{Key: "purpose", Value: "integration-test"},
		},
	}

	t.Logf("Creating VPC: %s", vpcName)

	// Create VPC
	vpc, err := client.CreateVPC(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateVPC failed: %v", err)
	}

	t.Logf("VPC created with ID: %s", vpc.ID)

	// Verify created VPC fields
	if vpc.Name != vpcName {
		t.Errorf("Expected VPC name %s, got %s", vpcName, vpc.Name)
	}
	if vpc.CIDRBlock != createReq.CIDRBlock {
		t.Errorf("Expected CIDR block %s, got %s", createReq.CIDRBlock, vpc.CIDRBlock)
	}
	if vpc.ID == "" {
		t.Error("VPC ID should not be empty")
	}

	// Cleanup: Delete VPC at the end
	defer func() {
		t.Logf("Deleting VPC: %s", vpc.ID)
		err := client.DeleteVPC(ctx, vpc.ID)
		if err != nil {
			t.Errorf("DeleteVPC failed: %v", err)
		} else {
			t.Log("VPC deleted successfully")
		}
	}()

	// Get VPC
	t.Logf("Getting VPC: %s", vpc.ID)
	fetchedVPC, err := client.GetVPC(ctx, vpc.ID)
	if err != nil {
		t.Fatalf("GetVPC failed: %v", err)
	}

	// Verify fetched VPC
	if fetchedVPC.ID != vpc.ID {
		t.Errorf("Expected VPC ID %s, got %s", vpc.ID, fetchedVPC.ID)
	}
	if fetchedVPC.Name != vpcName {
		t.Errorf("Expected VPC name %s, got %s", vpcName, fetchedVPC.Name)
	}
	if fetchedVPC.CIDRBlock != createReq.CIDRBlock {
		t.Errorf("Expected CIDR block %s, got %s", createReq.CIDRBlock, fetchedVPC.CIDRBlock)
	}

	t.Log("GetVPC returned correct data")
}

// TestVPCIntegration_List tests listing VPCs
func TestVPCIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	// List VPCs
	response, err := client.ListVPCs(ctx)
	if err != nil {
		t.Fatalf("ListVPCs failed: %v", err)
	}

	t.Logf("Found %d VPCs", response.Count)

	// Log VPC details
	for i, vpc := range response.Items {
		t.Logf("VPC %d: ID=%s, Name=%s, CIDR=%s, State=%s",
			i+1, vpc.ID, vpc.Name, vpc.CIDRBlock, vpc.State)
	}
}

// TestVPCIntegration_Update tests updating a VPC
func TestVPCIntegration_Update(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	// Generate unique VPC name
	vpcName := fmt.Sprintf("test-vpc-update-%d", time.Now().Unix())
	updatedName := fmt.Sprintf("test-vpc-updated-%d", time.Now().Unix())

	// Create VPC request
	createReq := &models.CreateVPCRequest{
		Name:               vpcName,
		CIDRBlock:          "10.201.0.0/16",
		EnableDNSHostnames: false,
		EnableDNSSupport:   true,
		Tags: []models.Tag{
			{Key: "env", Value: "test"},
		},
	}

	t.Logf("Creating VPC for update test: %s", vpcName)

	// Create VPC
	vpc, err := client.CreateVPC(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateVPC failed: %v", err)
	}

	t.Logf("VPC created with ID: %s", vpc.ID)

	// Cleanup: Delete VPC at the end
	defer func() {
		t.Logf("Deleting VPC: %s", vpc.ID)
		err := client.DeleteVPC(ctx, vpc.ID)
		if err != nil {
			t.Errorf("DeleteVPC failed: %v", err)
		} else {
			t.Log("VPC deleted successfully")
		}
	}()

	// Update VPC
	enableDNSHostnames := true
	updateReq := &models.UpdateVPCRequest{
		Name:               updatedName,
		EnableDNSHostnames: &enableDNSHostnames,
		Tags: []models.Tag{
			{Key: "env", Value: "test"},
			{Key: "updated", Value: "true"},
		},
	}

	t.Logf("Updating VPC: %s", vpc.ID)
	updatedVPC, err := client.UpdateVPC(ctx, vpc.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateVPC failed: %v", err)
	}

	// Verify update
	if updatedVPC.Name != updatedName {
		t.Errorf("Expected name %s, got %s", updatedName, updatedVPC.Name)
	}
	if !updatedVPC.EnableDNSHostnames {
		t.Error("Expected EnableDNSHostnames to be true")
	}

	t.Log("VPC updated successfully")
}

// TestVPCIntegration_GetNonExistent tests getting a non-existent VPC
func TestVPCIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentID := "non-existent-vpc-id-12345"

	t.Logf("Attempting to get non-existent VPC: %s", nonExistentID)

	_, err := client.GetVPC(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent VPC, got nil")
		return
	}

	// Check if it's a 404 error
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent VPC")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}
