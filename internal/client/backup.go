package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		return nil, fmt.Errorf("protection plan %q was created on the backend but could not be retrieved after retries "+
			"(the backend list endpoint timed out); it may need to be imported or removed manually: %w", req.Name, err)
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

// GetProtectionPlan retrieves a single protection plan by ID via the dedicated
// GET endpoint (.../protection_plan/{id}, singular). This avoids the expensive and
// intermittently slow full-list call on every read. A missing plan returns HTTP 404,
// which surfaces as a not-found *APIError.
func (c *Client) GetProtectionPlan(ctx context.Context, id string, subnetID string) (*models.ProtectionPlan, error) {
	scopedClient := c.WithSubnetID(subnetID)
	path := fmt.Sprintf("%s/protection_plan/%s", c.backupBasePath(), id)

	var resp models.ProtectionPlanDetailResponse
	err := c.doProtectionPlanRequestWithRetry(ctx, func() error {
		return scopedClient.Get(ctx, path, &resp)
	})
	if err != nil {
		return nil, err
	}
	return &resp.PolicyAttribute, nil
}

// Retry parameters for the protection plan backend endpoints, which are intermittently
// slow and return transient 5xx/timeout responses.
const protectionPlanRetryMaxAttempts = 3

// protectionPlanRetryBaseBackoff is the initial backoff between retries. It is a var
// (not a const) so tests can shrink it to keep retry-path tests fast.
var protectionPlanRetryBaseBackoff = 2 * time.Second

// doProtectionPlanRequestWithRetry runs fn, retrying transient backend 5xx/timeout
// failures with exponential backoff. Non-retryable errors (4xx) and context
// cancellation return immediately.
func (c *Client) doProtectionPlanRequestWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= protectionPlanRetryMaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableError(err) || attempt == protectionPlanRetryMaxAttempts {
			break
		}

		// Exponential backoff: 2s, 4s, ... (first attempt is immediate).
		backoff := protectionPlanRetryBaseBackoff * time.Duration(1<<(attempt-1))
		tflog.Warn(ctx, "Protection plan request failed, retrying", map[string]interface{}{
			"attempt":      attempt,
			"max_attempts": protectionPlanRetryMaxAttempts,
			"backoff":      backoff.String(),
			"error":        err.Error(),
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
	return lastErr
}

// ListProtectionPlans retrieves all protection plans. This is still used for
// create-time ID recovery (the create POST returns only a success message, so the new
// plan's ID must be discovered by name from the list). The backend list endpoint
// intermittently times out (HTTP 500), so transient failures are retried with backoff.
func (c *Client) ListProtectionPlans(ctx context.Context, subnetID string) ([]models.ProtectionPlan, error) {
	scopedClient := c.WithSubnetID(subnetID)
	path := fmt.Sprintf("%s/protection_plans/", c.backupBasePath())

	var resp models.ProtectionPlanListResponse
	err := c.doProtectionPlanRequestWithRetry(ctx, func() error {
		return scopedClient.Get(ctx, path, &resp)
	})
	if err != nil {
		return nil, err
	}
	return resp.PolicyAttributeList, nil
}

// isRetryableError reports whether an error from an API call is worth retrying.
// Server errors (5xx) and transport/network failures are transient; 4xx client errors
// (e.g. 404 not found) are not.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	// Non-APIError failures are transport-level (connection reset, timeout, etc.).
	return true
}

// ResolveProtectionPlanID returns the UUID for the given protection plan name or UUID.
// If nameOrID already looks like a UUID it is returned unchanged.
// Otherwise the plans list is searched for a plan whose name matches nameOrID (exact,
// case-insensitive first, then substring).
func (c *Client) ResolveProtectionPlanID(ctx context.Context, nameOrID, subnetID string) (string, error) {
	if isUUID(nameOrID) {
		return nameOrID, nil
	}
	plans, err := c.ListProtectionPlans(ctx, subnetID)
	if err != nil {
		return "", fmt.Errorf("listing protection plans to resolve %q: %w", nameOrID, err)
	}
	for _, plan := range plans {
		if strings.EqualFold(plan.Name, nameOrID) {
			return plan.ID, nil
		}
	}
	for _, plan := range plans {
		if containsIgnoreCase(plan.Name, nameOrID) {
			return plan.ID, nil
		}
	}
	return "", fmt.Errorf("protection plan %q not found", nameOrID)
}

// isUUID reports whether s is a standard UUID (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
	}
	return true
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
