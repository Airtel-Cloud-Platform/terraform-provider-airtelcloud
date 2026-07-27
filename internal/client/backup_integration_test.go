//go:build integration

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-airtelcloud/internal/models"
)

// getProtectionTestSubnetID returns the subnet used to scope protection-plan
// API calls (the plan endpoints require the subnet-id header), skipping the
// test when it is not configured.
func getProtectionTestSubnetID(t *testing.T) string {
	subnetID := os.Getenv("AIRTEL_TEST_SUBNET_ID")
	if subnetID == "" {
		t.Skip("AIRTEL_TEST_SUBNET_ID not set, skipping protection integration test")
	}
	return subnetID
}

// protectionTestAvailabilityZone returns the availability zone used as the plan
// selector value, defaulting to S1 when unset.
func protectionTestAvailabilityZone() string {
	if az := os.Getenv("AIRTEL_TEST_AVAILABILITY_ZONE"); az != "" {
		return az
	}
	return "S1"
}

// newProtectionPlanRequest builds a daily/30-day plan request with the given
// unique name.
func newProtectionPlanRequest(name string) *models.CreateProtectionPlanRequest {
	return &models.CreateProtectionPlanRequest{
		Name:          name,
		Description:   "terraform integration test plan",
		Retention:     1,
		RetentionUnit: "DAYS",
		Recurrence:    86400,
		SelectorKey:   "AZ",
		SelectorValue: protectionTestAvailabilityZone(),
	}
}

// TestProtectionPlanIntegration_CreateGetList creates a protection plan against
// the live API and confirms it can be fetched and listed. The plan API has no
// delete endpoint, so the created plan is intentionally left behind as an
// orphan (see the log below).
func TestProtectionPlanIntegration_CreateGetList(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	subnetID := getProtectionTestSubnetID(t)
	ctx := context.Background()

	planName := fmt.Sprintf("tf-int-pp-%d", time.Now().Unix())
	t.Logf("Creating protection plan %s (subnet %s)", planName, subnetID)

	plan, err := client.CreateProtectionPlan(ctx, newProtectionPlanRequest(planName), subnetID)
	if err != nil {
		t.Fatalf("CreateProtectionPlan failed: %v", err)
	}
	if plan == nil || plan.ID == "" {
		t.Fatalf("CreateProtectionPlan returned no plan ID: %+v", plan)
	}
	t.Logf("Protection plan created: ID=%s, Name=%s", plan.ID, plan.Name)
	t.Logf("NOTE: the protection-plan API has no delete endpoint; plan %s (%s) is left as an orphan", plan.ID, plan.Name)

	// Get by ID (implemented as a filtered list) must return the plan.
	got, err := client.GetProtectionPlan(ctx, plan.ID, subnetID)
	if err != nil {
		t.Fatalf("GetProtectionPlan failed: %v", err)
	}
	if got.ID != plan.ID {
		t.Errorf("GetProtectionPlan ID = %q, want %q", got.ID, plan.ID)
	}

	// The full list must include the new plan.
	plans, err := client.ListProtectionPlans(ctx, subnetID)
	if err != nil {
		t.Fatalf("ListProtectionPlans failed: %v", err)
	}
	if !containsProtectionPlanID(plans, plan.ID) {
		t.Errorf("ListProtectionPlans does not contain %s", plan.ID)
	}
}

// TestProtectionPlanIntegration_List lists every protection plan in the project.
func TestProtectionPlanIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	subnetID := getProtectionTestSubnetID(t)
	ctx := context.Background()

	plans, err := client.ListProtectionPlans(ctx, subnetID)
	if err != nil {
		t.Fatalf("ListProtectionPlans failed: %v", err)
	}
	t.Logf("Found %d protection plans in project", len(plans))
	for i, p := range plans {
		t.Logf("Plan %d: ID=%s, Name=%s", i, p.ID, p.Name)
	}
}

// TestProtectionIntegration_CreateGetUpdateDelete exercises the full protection
// lifecycle against the live API: create a plan to associate with, create the
// protection, get, update, then delete and confirm removal.
func TestProtectionIntegration_CreateGetUpdateDelete(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	computeID := getSnapshotTestComputeID(t)
	subnetID := getProtectionTestSubnetID(t)
	ctx := context.Background()

	// A protection references a plan by name, so create one first.
	planName := fmt.Sprintf("tf-int-pp-%d", time.Now().Unix())
	plan, err := client.CreateProtectionPlan(ctx, newProtectionPlanRequest(planName), subnetID)
	if err != nil {
		t.Fatalf("CreateProtectionPlan (for protection) failed: %v", err)
	}
	if plan == nil || plan.Name == "" {
		t.Fatalf("CreateProtectionPlan returned no plan name: %+v", plan)
	}
	t.Logf("Associated plan created: ID=%s, Name=%s (orphaned; no delete API)", plan.ID, plan.Name)

	protectionName := fmt.Sprintf("tf-int-prot-%d", time.Now().Unix())
	t.Logf("Creating protection %s for compute %s using plan %s", protectionName, computeID, plan.Name)

	// The protection API associates by the plan name returned from the list. If
	// a live run rejects this, fall back to the raw input name (planName).
	created, err := client.CreateProtection(ctx, &models.CreateProtectionRequest{
		Name:            protectionName,
		Description:     "terraform integration test protection",
		ComputeID:       computeID,
		ProtectionPlan:  plan.Name,
		EnableScheduler: "true",
	})
	if err != nil {
		t.Fatalf("CreateProtection failed: %v", err)
	}
	if created == nil || created.ID <= 0 {
		t.Fatalf("CreateProtection returned no ID: %+v", created)
	}
	protectionID := created.ID
	t.Logf("Protection created: ID=%d", protectionID)

	// Always clean up, even if a later assertion fails.
	deleted := false
	defer func() {
		if deleted {
			return
		}
		t.Logf("Deleting protection (cleanup): %d", protectionID)
		if err := client.DeleteProtection(ctx, protectionID); err != nil {
			t.Errorf("DeleteProtection (cleanup) failed: %v", err)
		}
	}()

	// Get by ID.
	got, err := client.GetProtection(ctx, protectionID)
	if err != nil {
		t.Fatalf("GetProtection failed: %v", err)
	}
	if got.Name != protectionName {
		t.Errorf("GetProtection Name = %q, want %q", got.Name, protectionName)
	}

	// Update name and description.
	updatedName := protectionName + "-upd"
	if _, err := client.UpdateProtection(ctx, protectionID, &models.UpdateProtectionRequest{
		Name:            updatedName,
		Description:     "updated by integration test",
		ProtectionPlan:  plan.Name,
		EnableScheduler: "true",
	}); err != nil {
		t.Fatalf("UpdateProtection failed: %v", err)
	}

	afterUpdate, err := client.GetProtection(ctx, protectionID)
	if err != nil {
		t.Fatalf("GetProtection (post-update) failed: %v", err)
	}
	if afterUpdate.Name != updatedName {
		t.Errorf("post-update Name = %q, want %q", afterUpdate.Name, updatedName)
	}

	// Delete and confirm it is gone from the list.
	t.Logf("Deleting protection: %d", protectionID)
	if err := client.DeleteProtection(ctx, protectionID); err != nil {
		t.Fatalf("DeleteProtection failed: %v", err)
	}
	deleted = true

	protections, err := client.ListProtections(ctx)
	if err != nil {
		t.Fatalf("ListProtections (post-delete) failed: %v", err)
	}
	if containsProtectionID(protections, protectionID) {
		t.Errorf("protection %d still present after delete", protectionID)
	}
}

// TestProtectionIntegration_List lists every protection policy in the project.
func TestProtectionIntegration_List(t *testing.T) {
	config := getVPCTestConfig(t)
	client := createVPCTestClient(t, config)
	ctx := context.Background()

	protections, err := client.ListProtections(ctx)
	if err != nil {
		t.Fatalf("ListProtections failed: %v", err)
	}
	t.Logf("Found %d protections in project", len(protections))
	for i, p := range protections {
		t.Logf("Protection %d: ID=%d, Name=%s, Status=%s", i, p.ID, p.Name, p.Status)
	}
}

func containsProtectionPlanID(plans []models.ProtectionPlan, id string) bool {
	for i := range plans {
		if plans[i].ID == id {
			return true
		}
	}
	return false
}

func containsProtectionID(protections []models.VeritasProtection, id int) bool {
	for i := range protections {
		if protections[i].ID == id {
			return true
		}
	}
	return false
}
