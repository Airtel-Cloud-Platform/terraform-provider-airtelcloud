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

// vpcPeeringTestConfig contains VPC peering-specific test configuration
type vpcPeeringTestConfig struct {
	APIEndpoint  string
	APIKey       string
	APISecret    string
	Region       string
	Organization string
	ProjectName  string
	SourceVPCID  string
	TargetVPCID  string
}

func getVPCPeeringTestConfig(t *testing.T) *vpcPeeringTestConfig {
	config := &vpcPeeringTestConfig{
		APIEndpoint:  os.Getenv("AIRTEL_API_ENDPOINT"),
		APIKey:       os.Getenv("AIRTEL_API_KEY"),
		APISecret:    os.Getenv("AIRTEL_API_SECRET"),
		Region:       os.Getenv("AIRTEL_REGION"),
		Organization: os.Getenv("AIRTEL_ORGANIZATION"),
		ProjectName:  os.Getenv("AIRTEL_PROJECT_NAME"),
		SourceVPCID:  os.Getenv("AIRTEL_TEST_SOURCE_VPC_ID"),
		TargetVPCID:  os.Getenv("AIRTEL_TEST_TARGET_VPC_ID"),
	}

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
	if config.SourceVPCID == "" {
		t.Skip("AIRTEL_TEST_SOURCE_VPC_ID not set, skipping integration test")
	}
	if config.TargetVPCID == "" {
		t.Skip("AIRTEL_TEST_TARGET_VPC_ID not set, skipping integration test")
	}

	return config
}

func createVPCPeeringTestClient(t *testing.T, config *vpcPeeringTestConfig) *Client {
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

// TestVPCPeeringIntegration_CreateGetDelete tests the full lifecycle of a VPC peering
func TestVPCPeeringIntegration_CreateGetDelete(t *testing.T) {
	config := getVPCPeeringTestConfig(t)
	client := createVPCPeeringTestClient(t, config)
	ctx := context.Background()

	peeringName := fmt.Sprintf("test-peering-%d", time.Now().Unix())

	createReq := &models.CreateVPCPeeringRequest{
		Name:        peeringName,
		Description: "Integration test peering",
		VPCSourceID: config.SourceVPCID,
		VPCTargetID: config.TargetVPCID,
		AZ:          config.Region + "a",
		Region:      config.Region,
	}

	t.Logf("Creating VPC peering: %s", peeringName)

	peering, err := client.CreateVPCPeering(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateVPCPeering failed: %v", err)
	}

	t.Logf("VPC peering created with ID: %s", peering.ID)

	if peering.Name != peeringName {
		t.Errorf("Expected name %s, got %s", peeringName, peering.Name)
	}
	if peering.ID == "" {
		t.Error("VPC peering ID should not be empty")
	}

	// Cleanup: Delete peering at the end
	defer func() {
		t.Logf("Deleting VPC peering: %s", peering.ID)
		err := client.DeleteVPCPeering(ctx, peering.ID)
		if err != nil {
			t.Errorf("DeleteVPCPeering failed: %v", err)
		} else {
			t.Log("VPC peering deleted successfully")
		}
	}()

	// Get VPC peering
	t.Logf("Getting VPC peering: %s", peering.ID)
	fetched, err := client.GetVPCPeering(ctx, peering.ID)
	if err != nil {
		t.Fatalf("GetVPCPeering failed: %v", err)
	}

	if fetched.ID != peering.ID {
		t.Errorf("Expected ID %s, got %s", peering.ID, fetched.ID)
	}
	if fetched.Name != peeringName {
		t.Errorf("Expected name %s, got %s", peeringName, fetched.Name)
	}
	if fetched.VPCSourceID != config.SourceVPCID {
		t.Errorf("Expected source VPC ID %s, got %s", config.SourceVPCID, fetched.VPCSourceID)
	}
	if fetched.VPCTargetID != config.TargetVPCID {
		t.Errorf("Expected target VPC ID %s, got %s", config.TargetVPCID, fetched.VPCTargetID)
	}

	t.Log("GetVPCPeering returned correct data")
}

// TestVPCPeeringIntegration_List tests listing VPC peerings
func TestVPCPeeringIntegration_List(t *testing.T) {
	config := getVPCPeeringTestConfig(t)
	client := createVPCPeeringTestClient(t, config)
	ctx := context.Background()

	t.Logf("API Endpoint: %s", config.APIEndpoint)
	t.Logf("Organization: %s", config.Organization)
	t.Logf("Project: %s", config.ProjectName)

	response, err := client.ListVPCPeerings(ctx)
	if err != nil {
		t.Fatalf("ListVPCPeerings failed: %v", err)
	}

	t.Logf("Found %d VPC peerings", response.Count)

	for i, p := range response.Items {
		t.Logf("Peering %d: ID=%s, Name=%s, Source=%s, Target=%s, State=%s",
			i+1, p.ID, p.Name, p.VPCSourceID, p.VPCTargetID, p.State)
	}
}

// TestVPCPeeringIntegration_GetNonExistent tests getting a non-existent VPC peering
func TestVPCPeeringIntegration_GetNonExistent(t *testing.T) {
	config := getVPCPeeringTestConfig(t)
	client := createVPCPeeringTestClient(t, config)
	ctx := context.Background()

	nonExistentID := "non-existent-peering-id-12345"

	t.Logf("Attempting to get non-existent VPC peering: %s", nonExistentID)

	_, err := client.GetVPCPeering(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent VPC peering, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent VPC peering")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}
