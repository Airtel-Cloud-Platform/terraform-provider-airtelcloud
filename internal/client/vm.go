package client

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// computeAPIBase returns the v2.1 base path prefix for all compute-related endpoints
func (c *Client) computeAPIBase() string {
	return fmt.Sprintf("/api/v2.1/computes/domain/%s/project/%s",
		c.Organization, c.ProjectName)
}

// computeBasePath returns the v2.1 base path for compute instance endpoints
func (c *Client) computeBasePath() string {
	return c.computeAPIBase() + "/computes"
}

// GetCompute retrieves a compute instance by ID
func (c *Client) GetCompute(ctx context.Context, id string) (*models.Compute, error) {
	var compute models.Compute
	err := c.Get(ctx, fmt.Sprintf("%s/%s/", c.computeBasePath(), id), &compute)
	if err != nil {
		return nil, err
	}
	return &compute, nil
}

// ListComputes retrieves all compute instances
func (c *Client) ListComputes(ctx context.Context) ([]models.Compute, error) {
	var computes []models.Compute
	err := c.Get(ctx, fmt.Sprintf("%s/", c.computeBasePath()), &computes)
	if err != nil {
		return nil, err
	}
	return computes, nil
}

// CreateCompute creates a new compute instance using URL-encoded form data
func (c *Client) CreateCompute(ctx context.Context, req *models.CreateComputeRequest) (*models.Compute, error) {
	// Convert struct to form data map
	formData := structToFormData(req)

	var response []models.Compute
	err := c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/", c.computeBasePath()), formData, &response)
	if err != nil {
		return nil, err
	}

	// API returns array with single compute instance
	if len(response) == 0 {
		return nil, fmt.Errorf("no compute instance returned from API")
	}

	// Return the created compute instance
	return &response[0], nil
}

// UpdateCompute updates an existing compute instance (limited update support)
func (c *Client) UpdateCompute(ctx context.Context, id string, req *models.UpdateComputeRequest) (*models.Compute, error) {
	// Convert struct to form data map
	formData := structToFormData(req)

	var compute models.Compute
	err := c.PutURLEncodedForm(ctx, fmt.Sprintf("%s/%s/", c.computeBasePath(), id), formData, &compute)
	if err != nil {
		return nil, err
	}

	return &compute, nil
}

// DeleteCompute deletes a compute instance
func (c *Client) DeleteCompute(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s/", c.computeBasePath(), id))
	if err != nil {
		return err
	}

	// Wait for compute to be deleted (poll until 404 or deleted status)
	for i := 0; i < 60; i++ {
		compute, err := c.GetCompute(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if compute.Status == "Deleted" || compute.Status == "deleted" || compute.Status == "soft-deleted" {
			return nil
		}
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("compute deletion timed out")
}

// WaitForComputeReady polls until the compute instance reaches Active status
func (c *Client) WaitForComputeReady(ctx context.Context, id string, timeout time.Duration) (*models.Compute, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		compute, err := c.GetCompute(ctx, id)
		if err != nil {
			return nil, err
		}

		switch compute.Status {
		case "ACTIVE", "active", "Active":
			return compute, nil
		case "ERROR", "error", "Error":
			return nil, fmt.Errorf("compute instance entered error state")
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("compute instance did not become ready within %v", timeout)
}

// PerformComputeAction performs an action on a compute instance (start/stop/reboot/etc.)
func (c *Client) PerformComputeAction(ctx context.Context, id string, action models.ComputeAction) error {
	formData := map[string]interface{}{
		"action": string(action),
	}

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/%s", c.computeBasePath(), id, action), formData, nil)
}

// GetComputeConsoleURL gets the console URL for a compute instance
func (c *Client) GetComputeConsoleURL(ctx context.Context, id string) (string, error) {
	var response struct {
		ConsoleURL string `json:"console_url"`
	}

	err := c.Get(ctx, fmt.Sprintf("%s/%s/console_url", c.computeBasePath(), id), &response)
	if err != nil {
		return "", err
	}

	return response.ConsoleURL, nil
}

// ResizeCompute resizes a compute instance to a new flavor
func (c *Client) ResizeCompute(ctx context.Context, id string, flavorID string) error {
	formData := map[string]interface{}{
		"flavor_id": flavorID,
	}

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/resize/%s", c.computeBasePath(), id, flavorID), formData, nil)
}

// RebuildCompute rebuilds a compute instance with a new image
func (c *Client) RebuildCompute(ctx context.Context, id string, imageID string) error {
	formData := map[string]interface{}{
		"image_id": imageID,
	}

	return c.PostURLEncodedForm(ctx, fmt.Sprintf("%s/%s/rebuild-server/", c.computeBasePath(), id), formData, nil)
}

// structToFormData converts a struct with form tags to a map[string]interface{}
func structToFormData(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the form tag
		formTag := fieldType.Tag.Get("form")
		if formTag == "" || formTag == "-" {
			continue
		}

		// Parse form tag (handle omitempty)
		tagParts := strings.Split(formTag, ",")
		fieldName := tagParts[0]
		omitEmpty := len(tagParts) > 1 && tagParts[1] == "omitempty"

		// Skip zero values if omitempty is set
		if omitEmpty && field.IsZero() {
			continue
		}

		// Convert field value to interface{}
		switch field.Kind() {
		case reflect.String:
			if s := field.String(); s != "" || !omitEmpty {
				result[fieldName] = s
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if i := field.Int(); i != 0 || !omitEmpty {
				result[fieldName] = strconv.FormatInt(i, 10)
			}
		case reflect.Bool:
			if b := field.Bool(); b || !omitEmpty {
				result[fieldName] = strconv.FormatBool(b)
			}
		case reflect.Slice:
			if field.Len() > 0 || !omitEmpty {
				slice := make([]interface{}, field.Len())
				for j := 0; j < field.Len(); j++ {
					slice[j] = field.Index(j).Interface()
				}
				result[fieldName] = slice
			}
		default:
			if !field.IsZero() || !omitEmpty {
				result[fieldName] = field.Interface()
			}
		}
	}

	return result
}
