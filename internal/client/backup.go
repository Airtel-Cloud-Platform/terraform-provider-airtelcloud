package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// backupBasePath returns the base path for backup/protection endpoints
func (c *Client) backupBasePath() string {
	return fmt.Sprintf("/api/v2.1/backups/domain/%s/project/%s/backups",
		c.Organization, c.ProjectName)
}

// --- Veritas Protection CRUD ---

// CreateProtection creates a new Veritas backup protection policy
func (c *Client) CreateProtection(ctx context.Context, req *models.CreateProtectionRequest) (*models.VeritasProtection, error) {
	formData := structToFormData(req)

	var protection models.VeritasProtection
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/protections/", c.backupBasePath()), formData, &protection)
	if err != nil {
		return nil, err
	}
	return &protection, nil
}

// GetProtection retrieves a Veritas protection by ID
func (c *Client) GetProtection(ctx context.Context, id int) (*models.VeritasProtection, error) {
	var protection models.VeritasProtection
	err := c.Get(ctx, fmt.Sprintf("%s/protections/%s/", c.backupBasePath(), strconv.Itoa(id)), &protection)
	if err != nil {
		return nil, err
	}
	return &protection, nil
}

// ListProtections retrieves all Veritas protection policies
func (c *Client) ListProtections(ctx context.Context) ([]models.VeritasProtection, error) {
	var protections []models.VeritasProtection
	err := c.Get(ctx, fmt.Sprintf("%s/protections/", c.backupBasePath()), &protections)
	if err != nil {
		return nil, err
	}
	return protections, nil
}

// UpdateProtection updates a Veritas protection policy
func (c *Client) UpdateProtection(ctx context.Context, id int, req *models.UpdateProtectionRequest) (*models.VeritasProtection, error) {
	formData := structToFormData(req)

	var protection models.VeritasProtection
	err := c.PutURLEncodedForm(ctx, fmt.Sprintf("%s/protections/%s/", c.backupBasePath(), strconv.Itoa(id)), formData, &protection)
	if err != nil {
		return nil, err
	}
	return &protection, nil
}

// DeleteProtection deletes a Veritas protection policy
func (c *Client) DeleteProtection(ctx context.Context, id int) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/protections/%s/", c.backupBasePath(), strconv.Itoa(id)))
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return err
	}
	return nil
}

// DisableProtectionScheduler disables the scheduler for a protection policy
func (c *Client) DisableProtectionScheduler(ctx context.Context, computeID int) error {
	return c.PutURLEncodedForm(ctx, fmt.Sprintf("%s/protections/%s/disable-scheduler", c.backupBasePath(), strconv.Itoa(computeID)), nil, nil)
}

// --- Protection Plan CRUD ---

// CreateProtectionPlan creates a new protection plan.
// The API returns a success message string, so after creation we list plans to find the newly created one by name.
func (c *Client) CreateProtectionPlan(ctx context.Context, req *models.CreateProtectionPlanRequest, subnetID string) (*models.ProtectionPlan, error) {
	scopedClient := c.WithSubnetID(subnetID)
	formData := structToFormData(req)

	// The API returns a success message string, not a plan object — pass nil to skip unmarshal
	err := scopedClient.PostURLEncodedForm(ctx, fmt.Sprintf("%s/protection_plans/", c.backupBasePath()), formData, nil)
	if err != nil {
		return nil, err
	}

	// After successful creation, list plans and find the one matching the requested name
	plans, err := c.ListProtectionPlans(ctx, subnetID)
	if err != nil {
		return nil, fmt.Errorf("protection plan created but failed to retrieve it: %w", err)
	}

	for _, plan := range plans {
		// The API transforms the name into a pattern like S1-PERFTEST-CELL-1-{NAME}-BKP-PP
		// but also stores the original input. Match by suffix containing the input name (case-insensitive).
		if plan.Name != "" && containsIgnoreCase(plan.Name, req.Name) {
			return &plan, nil
		}
	}

	// If exact match not found, return the most recently created plan
	if len(plans) > 0 {
		return &plans[len(plans)-1], nil
	}

	return nil, fmt.Errorf("protection plan created but could not find it in the list")
}

// GetProtectionPlan retrieves a protection plan by ID.
// The single-plan GET endpoint is not available, so we list all plans and filter by ID.
func (c *Client) GetProtectionPlan(ctx context.Context, id string, subnetID string) (*models.ProtectionPlan, error) {
	plans, err := c.ListProtectionPlans(ctx, subnetID)
	if err != nil {
		return nil, err
	}

	for _, plan := range plans {
		if plan.ID == id {
			return &plan, nil
		}
	}

	return nil, &APIError{StatusCode: 404, Message: "protection plan not found"}
}

// ListProtectionPlans retrieves all protection plans
func (c *Client) ListProtectionPlans(ctx context.Context, subnetID string) ([]models.ProtectionPlan, error) {
	scopedClient := c.WithSubnetID(subnetID)

	var resp models.ProtectionPlanListResponse
	err := scopedClient.Get(ctx, fmt.Sprintf("%s/protection_plans/", c.backupBasePath()), &resp)
	if err != nil {
		return nil, err
	}
	return resp.PolicyAttributeList, nil
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && (findIgnoreCase(s, substr) >= 0))
}

func findIgnoreCase(s, substr string) int {
	s = strings.ToUpper(s)
	substr = strings.ToUpper(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
