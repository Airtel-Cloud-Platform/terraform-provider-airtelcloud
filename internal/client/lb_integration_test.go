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

// lbTestDefaults holds resolved values needed for LB integration tests
type lbTestDefaults struct {
	FlavorID int
	VPCID    string
	VPCName  string
	SubnetID string // used as network_id in LB service creation
}

// getLBTestDefaults resolves LB flavor, VPC, and subnet info from the API and env vars.
// The client returned is scoped with SubnetID header, matching the provider's pattern.
func getLBTestDefaults(t *testing.T, client *Client) (*lbTestDefaults, *Client) {
	t.Helper()
	ctx := context.Background()

	// VPC ID from env var (required)
	vpcID := os.Getenv("AIRTEL_TEST_NETWORK_ID")
	if vpcID == "" {
		t.Skip("AIRTEL_TEST_NETWORK_ID not set, skipping LB integration test")
	}

	// Subnet ID: prefer LB-specific env var, fall back to general subnet
	subnetID := os.Getenv("AIRTEL_TEST_LB_SUBNET_ID")
	if subnetID == "" {
		subnetID = os.Getenv("AIRTEL_TEST_SUBNET_ID")
	}
	if subnetID == "" {
		t.Skip("AIRTEL_TEST_LB_SUBNET_ID or AIRTEL_TEST_SUBNET_ID not set, skipping LB integration test")
	}

	// Scope client with subnet-id header (matches provider pattern in lb_service_resource.go)
	scopedClient := client.WithSubnetID(subnetID)

	// Resolve VPC name by listing VPCs
	vpcResp, err := scopedClient.ListVPCs(ctx)
	if err != nil {
		t.Fatalf("ListVPCs failed: %v", err)
	}

	var vpcName string
	for _, vpc := range vpcResp.Items {
		if vpc.ID == vpcID {
			vpcName = vpc.Name
			break
		}
	}
	if vpcName == "" {
		t.Fatalf("VPC with ID %s not found in VPC list", vpcID)
	}

	// Resolve LB flavor using scoped client (matches provider pattern)
	flavors, err := scopedClient.ListLBFlavors(ctx)
	if err != nil {
		t.Fatalf("ListLBFlavors failed: %v", err)
	}
	if len(flavors) == 0 {
		t.Fatal("No LB flavors available")
	}

	t.Logf("Resolved LB defaults: FlavorID=%d (%s), VPC=%s (%s), SubnetID=%s",
		flavors[0].ID, flavors[0].Name, vpcID, vpcName, subnetID)

	return &lbTestDefaults{
		FlavorID: flavors[0].ID,
		VPCID:    vpcID,
		VPCName:  vpcName,
		SubnetID: subnetID,
	}, scopedClient
}

// --- LB Flavor Tests ---

func TestLBIntegration_ListFlavors(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	flavors, err := client.ListLBFlavors(ctx)
	if err != nil {
		t.Fatalf("ListLBFlavors failed: %v", err)
	}

	t.Logf("Found %d LB flavors", len(flavors))
	for i, f := range flavors {
		t.Logf("  Flavor %d: ID=%d, Name=%s", i+1, f.ID, f.Name)
	}
}

// --- LB Service Tests ---

func TestLBIntegration_ListServices(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	services, err := client.ListLBServices(ctx)
	if err != nil {
		t.Fatalf("ListLBServices failed: %v", err)
	}

	t.Logf("Found %d LB services", len(services))
	for i, svc := range services {
		t.Logf("  Service %d: ID=%s, Name=%s, Status=%s, FlavorID=%d, AZ=%s",
			i+1, svc.ID, svc.Name, svc.Status, svc.FlavorID, svc.AZName)
	}
}

func TestLBIntegration_ServiceLifecycle(t *testing.T) {
	config := getVPCTestConfig(t)
	baseClient := createVPCTestClient(t, config)
	ctx := context.Background()
	defaults, scopedClient := getLBTestDefaults(t, baseClient)

	svcName := fmt.Sprintf("test-lb-%d", time.Now().Unix())

	// Create LB service using scoped client (with subnet-id header)
	createReq := &models.CreateLBServiceRequest{
		Name:      svcName,
		FlavorID:  defaults.FlavorID,
		NetworkID: defaults.SubnetID,
		VPCID:     defaults.VPCID,
		VPCName:   defaults.VPCName,
	}

	t.Logf("Creating LB service: %s (flavor=%d, vpc=%s, subnet=%s)",
		svcName, defaults.FlavorID, defaults.VPCID, defaults.SubnetID)

	svc, err := scopedClient.CreateLBService(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateLBService failed: %v", err)
	}

	t.Logf("LB service created: ID=%s, Status=%s", svc.ID, svc.Status)

	// Cleanup: delete at end
	defer func() {
		t.Logf("Deleting LB service: %s", svc.ID)
		err := scopedClient.DeleteLBService(ctx, svc.ID)
		if err != nil {
			t.Errorf("DeleteLBService failed: %v", err)
			return
		}
		t.Log("Waiting for LB service deletion...")
		err = scopedClient.WaitForLBServiceDeleted(ctx, svc.ID, 10*time.Minute)
		if err != nil {
			t.Errorf("WaitForLBServiceDeleted failed: %v", err)
		} else {
			t.Log("LB service deleted successfully")
		}
	}()

	// Wait for Active status
	t.Log("Waiting for LB service to become Active...")
	readySvc, err := scopedClient.WaitForLBServiceReady(ctx, svc.ID, 15*time.Minute)
	if err != nil {
		t.Fatalf("WaitForLBServiceReady failed: %v", err)
	}

	t.Logf("LB service is Active: ID=%s, Name=%s, Status=%s", readySvc.ID, readySvc.Name, readySvc.Status)

	// Verify fields
	if readySvc.Name != svcName {
		t.Errorf("Expected name %s, got %s", svcName, readySvc.Name)
	}
	if readySvc.ID == "" {
		t.Error("LB service ID should not be empty")
	}

	// Get by ID
	t.Logf("Getting LB service: %s", svc.ID)
	fetchedSvc, err := scopedClient.GetLBService(ctx, svc.ID)
	if err != nil {
		t.Fatalf("GetLBService failed: %v", err)
	}
	if fetchedSvc.ID != svc.ID {
		t.Errorf("Expected ID %s, got %s", svc.ID, fetchedSvc.ID)
	}
	if fetchedSvc.Name != svcName {
		t.Errorf("Expected name %s, got %s", svcName, fetchedSvc.Name)
	}
	t.Log("GetLBService returned correct data")

	// Verify in list
	services, err := scopedClient.ListLBServices(ctx)
	if err != nil {
		t.Fatalf("ListLBServices failed: %v", err)
	}

	found := false
	for _, s := range services {
		if s.ID == svc.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created LB service not found in list")
	} else {
		t.Log("LB service found in list")
	}
}

func TestLBIntegration_GetNonExistent(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	nonExistentID := "non-existent-lb-id-12345"

	t.Logf("Attempting to get non-existent LB service: %s", nonExistentID)

	_, err := client.GetLBService(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent LB service, got nil")
		return
	}

	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != 404 {
			t.Errorf("Expected 404 status code, got %d", apiErr.StatusCode)
		} else {
			t.Log("Correctly received 404 for non-existent LB service")
		}
	} else {
		t.Logf("Received error (non-API error): %v", err)
	}
}

// --- LB VIP Tests ---

func TestLBIntegration_VipLifecycle(t *testing.T) {
	config := getVPCTestConfig(t)
	baseClient := createVPCTestClient(t, config)
	ctx := context.Background()
	defaults, scopedClient := getLBTestDefaults(t, baseClient)

	// Create LB service first
	svcName := fmt.Sprintf("test-lb-vip-%d", time.Now().Unix())
	createReq := &models.CreateLBServiceRequest{
		Name:      svcName,
		FlavorID:  defaults.FlavorID,
		NetworkID: defaults.SubnetID,
		VPCID:     defaults.VPCID,
		VPCName:   defaults.VPCName,
	}

	t.Logf("Creating LB service for VIP test: %s", svcName)
	svc, err := scopedClient.CreateLBService(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateLBService failed: %v", err)
	}

	defer func() {
		t.Logf("Cleaning up LB service: %s", svc.ID)
		err := scopedClient.DeleteLBService(ctx, svc.ID)
		if err != nil {
			t.Errorf("DeleteLBService failed: %v", err)
			return
		}
		scopedClient.WaitForLBServiceDeleted(ctx, svc.ID, 10*time.Minute)
		t.Log("LB service cleaned up")
	}()

	t.Log("Waiting for LB service to become Active...")
	_, err = scopedClient.WaitForLBServiceReady(ctx, svc.ID, 15*time.Minute)
	if err != nil {
		t.Fatalf("WaitForLBServiceReady failed: %v", err)
	}

	// Create VIP
	t.Logf("Creating VIP for LB service: %s", svc.ID)
	vip, err := scopedClient.CreateLBVip(ctx, svc.ID)
	if err != nil {
		t.Fatalf("CreateLBVip failed: %v", err)
	}

	t.Logf("VIP created: ID=%d, Status=%s, FixedIPs=%s", vip.ID, vip.Status, vip.FixedIPs)

	// Cleanup VIP
	defer func() {
		t.Logf("Deleting VIP: %d", vip.ID)
		err := scopedClient.DeleteLBVip(ctx, svc.ID, vip.ID)
		if err != nil {
			t.Errorf("DeleteLBVip failed: %v", err)
		} else {
			t.Log("VIP deleted successfully")
		}
	}()

	// List VIPs
	vips, err := scopedClient.ListLBVips(ctx, svc.ID)
	if err != nil {
		t.Fatalf("ListLBVips failed: %v", err)
	}

	t.Logf("Found %d VIPs for LB service %s", len(vips), svc.ID)

	found := false
	for _, v := range vips {
		t.Logf("  VIP: ID=%d, Status=%s, FixedIPs=%s, NetworkID=%s",
			v.ID, v.Status, v.FixedIPs, v.NetworkID)
		if v.ID == vip.ID {
			found = true
		}
	}
	if !found {
		t.Error("Created VIP not found in list")
	} else {
		t.Log("VIP found in list")
	}
}

// --- LB Virtual Server Tests ---

func TestLBIntegration_VirtualServerLifecycle(t *testing.T) {
	config := getVPCTestConfig(t)
	baseClient := createVPCTestClient(t, config)
	ctx := context.Background()
	defaults, scopedClient := getLBTestDefaults(t, baseClient)

	// Create LB service
	svcName := fmt.Sprintf("test-lb-vs-%d", time.Now().Unix())
	createReq := &models.CreateLBServiceRequest{
		Name:      svcName,
		FlavorID:  defaults.FlavorID,
		NetworkID: defaults.SubnetID,
		VPCID:     defaults.VPCID,
		VPCName:   defaults.VPCName,
	}

	t.Logf("Creating LB service for virtual server test: %s", svcName)
	svc, err := scopedClient.CreateLBService(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateLBService failed: %v", err)
	}

	defer func() {
		t.Logf("Cleaning up LB service: %s", svc.ID)
		scopedClient.DeleteLBService(ctx, svc.ID)
		scopedClient.WaitForLBServiceDeleted(ctx, svc.ID, 10*time.Minute)
		t.Log("LB service cleaned up")
	}()

	t.Log("Waiting for LB service to become Active...")
	_, err = scopedClient.WaitForLBServiceReady(ctx, svc.ID, 15*time.Minute)
	if err != nil {
		t.Fatalf("WaitForLBServiceReady failed: %v", err)
	}

	// Create VIP (needed for virtual server)
	t.Logf("Creating VIP for LB service: %s", svc.ID)
	vip, err := scopedClient.CreateLBVip(ctx, svc.ID)
	if err != nil {
		t.Fatalf("CreateLBVip failed: %v", err)
	}

	t.Logf("VIP created: ID=%d, FixedIPs=%s", vip.ID, vip.FixedIPs)

	defer func() {
		t.Logf("Deleting VIP: %d", vip.ID)
		scopedClient.DeleteLBVip(ctx, svc.ID, vip.ID)
	}()

	// Create virtual server
	vsName := fmt.Sprintf("test-vs-%d", time.Now().Unix())
	t.Logf("Creating virtual server: %s (vip_port_id=%d)", vsName, vip.ID)

	// Find a compute instance for the backend node
	computes, err := scopedClient.ListComputes(ctx)
	if err != nil {
		t.Fatalf("ListComputes failed: %v", err)
	}
	if len(computes) == 0 {
		t.Skip("No compute instances available for virtual server backend nodes, skipping")
	}

	// Use first active compute as backend node
	var backendNode models.VirtualServerNode
	for _, c := range computes {
		if c.Status == "ACTIVE" || c.Status == "Active" || c.Status == "active" {
			privateIP := c.PrivateIP()
			if privateIP != "" {
				backendNode = models.VirtualServerNode{
					ComputeID: func() int {
						if len(c.Ports) > 0 {
							return c.Ports[0].ID
						}
						return 0
					}(),
					ComputeIP: privateIP,
					Port:      8080,
					Weight:    1,
				}
				t.Logf("Using compute %s (%s) as backend node", c.ID, privateIP)
				break
			}
		}
	}
	if backendNode.ComputeIP == "" {
		t.Skip("No active compute with private IP found for backend node, skipping")
	}

	params := BuildVirtualServerParams(
		vsName,
		"HTTP",
		defaults.VPCID,
		"ROUND_ROBIN",
		"", // monitor_protocol
		"", // certificate_id
		vip.ID,
		80,    // port
		30,    // interval
		false, // persistence_enabled
		true,  // x_forwarded_for
		false, // redirect_https
		"",    // persistence_type
		[]models.VirtualServerNode{backendNode},
	)

	vs, err := scopedClient.CreateVirtualServer(ctx, svc.ID, params)
	if err != nil {
		t.Fatalf("CreateVirtualServer failed: %v", err)
	}

	t.Logf("Virtual server created: ID=%s, Name=%s, Status=%s, Protocol=%s, Port=%d",
		vs.ID, vs.Name, vs.Status, vs.Protocol, vs.Port)

	// Cleanup virtual server
	defer func() {
		t.Logf("Deleting virtual server: %s", vs.ID)
		err := scopedClient.DeleteVirtualServer(ctx, svc.ID, vs.ID)
		if err != nil {
			t.Errorf("DeleteVirtualServer failed: %v", err)
		} else {
			t.Log("Virtual server deleted")
		}
	}()

	// Wait for virtual server to become Active
	t.Log("Waiting for virtual server to become Active...")
	readyVS, err := scopedClient.WaitForVirtualServerReady(ctx, svc.ID, vs.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("WaitForVirtualServerReady failed: %v", err)
	}

	t.Logf("Virtual server Active: ID=%s, Protocol=%s, Port=%d, RoutingAlgorithm=%s",
		readyVS.ID, readyVS.Protocol, readyVS.Port, readyVS.RoutingAlgorithm)

	// Get by ID
	fetchedVS, err := scopedClient.GetVirtualServer(ctx, svc.ID, vs.ID)
	if err != nil {
		t.Fatalf("GetVirtualServer failed: %v", err)
	}
	if fetchedVS.ID != vs.ID {
		t.Errorf("Expected ID %s, got %s", vs.ID, fetchedVS.ID)
	}
	if fetchedVS.Name != vsName {
		t.Errorf("Expected name %s, got %s", vsName, fetchedVS.Name)
	}
	t.Log("GetVirtualServer returned correct data")

	// List virtual servers
	servers, err := scopedClient.ListVirtualServers(ctx, svc.ID)
	if err != nil {
		t.Fatalf("ListVirtualServers failed: %v", err)
	}

	found := false
	for _, s := range servers {
		t.Logf("  VirtualServer: ID=%s, Name=%s, Protocol=%s, Port=%d, Status=%s",
			s.ID, s.Name, s.Protocol, s.Port, s.Status)
		if s.ID == vs.ID {
			found = true
		}
	}
	if !found {
		t.Error("Created virtual server not found in list")
	} else {
		t.Log("Virtual server found in list")
	}

	// Update virtual server (change routing algorithm)
	t.Log("Updating virtual server routing algorithm to LEAST_CONNECTIONS...")
	updateParams := BuildVirtualServerParams(
		vsName,
		"HTTP",
		defaults.VPCID,
		"LEAST_CONNECTIONS",
		"", // monitor_protocol
		"", // certificate_id
		vip.ID,
		80,
		30,
		false,
		true,
		false,
		"",
		[]models.VirtualServerNode{backendNode},
	)

	updatedVS, err := scopedClient.UpdateVirtualServer(ctx, svc.ID, vs.ID, updateParams)
	if err != nil {
		t.Fatalf("UpdateVirtualServer failed: %v", err)
	}

	t.Logf("Virtual server updated: ID=%s, RoutingAlgorithm=%s", updatedVS.ID, updatedVS.RoutingAlgorithm)
}
