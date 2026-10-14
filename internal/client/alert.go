package client

import (
	"context"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) alertBasePath() string {
	return fmt.Sprintf("/api/v2.1/alert/domain/%s/project/%s/alerts", c.Organization, c.ProjectName)
}

func (c *Client) alertPath(id string) string {
	return fmt.Sprintf("%s/%s", c.alertBasePath(), id)
}

// ListAlerts lists alerts for the configured domain/project.
func (c *Client) ListAlerts(ctx context.Context) (*models.AlertListResponse, error) {
	var resp models.AlertListResponse
	if err := c.Get(ctx, c.alertBasePath(), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResolveAlertUUID finds the list uuid_id for an alert name.
// Create returns only an integer id; GET/PUT/DELETE require uuid_id.
func (c *Client) ResolveAlertUUID(ctx context.Context, name string) (string, error) {
	list, err := c.ListAlerts(ctx)
	if err != nil {
		return "", fmt.Errorf("list alerts to resolve uuid: %w", err)
	}
	for _, item := range list.AlertData {
		if item.Name == name && item.UUIDID != "" {
			return item.UUIDID, nil
		}
	}
	return "", fmt.Errorf("alert %q not found in list after create", name)
}

// CreateAlert creates a metric alert, then resolves uuid_id from the list API
// and stores it on Alert.UUIDID for Terraform state / subsequent CRUD.
func (c *Client) CreateAlert(ctx context.Context, req *models.CreateAlertRequest) (*models.Alert, error) {
	var alert models.Alert
	if err := c.Post(ctx, c.alertBasePath(), req, &alert); err != nil {
		return nil, err
	}
	uuid, err := c.ResolveAlertUUID(ctx, alert.Name)
	if err != nil {
		return nil, err
	}
	alert.UUIDID = uuid
	return &alert, nil
}

// GetAlert retrieves an alert by uuid_id (not the integer create id).
func (c *Client) GetAlert(ctx context.Context, id string) (*models.AlertGetResponse, error) {
	var alert models.AlertGetResponse
	if err := c.Get(ctx, c.alertPath(id), &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// UpdateAlert updates an existing alert by uuid_id.
func (c *Client) UpdateAlert(ctx context.Context, id string, req *models.UpdateAlertRequest) (*models.Alert, error) {
	var alert models.Alert
	if err := c.Put(ctx, c.alertPath(id), req, &alert); err != nil {
		return nil, err
	}
	alert.UUIDID = id
	return &alert, nil
}

// DeleteAlert deletes an alert by uuid_id. A 404 is treated as success.
func (c *Client) DeleteAlert(ctx context.Context, id string) error {
	err := c.Delete(ctx, c.alertPath(id))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// CreateAlertNotification creates a notification channel for an alert.
func (c *Client) CreateAlertNotification(ctx context.Context, req *models.CreateAlertNotificationRequest) (*models.AlertNotification, error) {
	var n models.AlertNotification
	if err := c.Post(ctx, "/api/v1/notifications/", req, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetAlertNotification retrieves a notification by id (string form of Notification.id).
func (c *Client) GetAlertNotification(ctx context.Context, id string) (*models.AlertNotification, error) {
	var n models.AlertNotification
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/notifications/%s", id), &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// UpdateAlertNotification updates a notification by id.
func (c *Client) UpdateAlertNotification(ctx context.Context, id string, req *models.UpdateAlertNotificationRequest) (*models.AlertNotification, error) {
	var n models.AlertNotification
	if err := c.Put(ctx, fmt.Sprintf("/api/v1/notifications/%s", id), req, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteAlertNotification deletes a notification. A 404 is treated as success.
func (c *Client) DeleteAlertNotification(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("/api/v1/notifications/%s", id))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}
