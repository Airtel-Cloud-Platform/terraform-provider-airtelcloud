package client

import (
	"context"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	monitoringDashboardsPath = "/api/monitoring-admin/v2/dashboards"
	monitoringProvidersPath  = "/api/monitoring-admin/v2/providers"
)

// ListMonitoringDashboards lists all monitoring dashboards.
func (c *Client) ListMonitoringDashboards(ctx context.Context) ([]models.MonitoringDashboard, error) {
	var resp models.MonitoringDashboardListResponse
	if err := c.Get(ctx, monitoringDashboardsPath, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ListMonitoringProviders lists all monitoring providers.
func (c *Client) ListMonitoringProviders(ctx context.Context) ([]models.MonitoringProvider, error) {
	var resp models.MonitoringProviderListResponse
	if err := c.Get(ctx, monitoringProvidersPath, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// CreateMonitoringDashboard creates a monitoring dashboard.
func (c *Client) CreateMonitoringDashboard(ctx context.Context, req *models.CreateMonitoringDashboardRequest) (*models.MonitoringDashboard, error) {
	var resp models.MonitoringDashboard
	if err := c.Post(ctx, monitoringDashboardsPath, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMonitoringDashboard retrieves a monitoring dashboard by ID.
func (c *Client) GetMonitoringDashboard(ctx context.Context, id string) (*models.MonitoringDashboard, error) {
	var resp models.MonitoringDashboard
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", monitoringDashboardsPath, id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateMonitoringDashboard updates a monitoring dashboard.
func (c *Client) UpdateMonitoringDashboard(ctx context.Context, id string, req *models.UpdateMonitoringDashboardRequest) (*models.MonitoringDashboard, error) {
	var resp models.MonitoringDashboard
	if err := c.Put(ctx, fmt.Sprintf("%s/%s", monitoringDashboardsPath, id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteMonitoringDashboard deletes a monitoring dashboard.
func (c *Client) DeleteMonitoringDashboard(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s", monitoringDashboardsPath, id))
	if err != nil && IsNotFoundError(err) {
		return nil
	}
	return err
}

// CreateMonitoringProvider creates a monitoring provider.
func (c *Client) CreateMonitoringProvider(ctx context.Context, req *models.CreateMonitoringProviderRequest) (*models.MonitoringProvider, error) {
	var resp models.MonitoringProvider
	if err := c.Post(ctx, monitoringProvidersPath, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMonitoringProvider retrieves a monitoring provider by ID.
func (c *Client) GetMonitoringProvider(ctx context.Context, id string) (*models.MonitoringProvider, error) {
	var resp models.MonitoringProvider
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", monitoringProvidersPath, id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateMonitoringProvider updates a monitoring provider.
func (c *Client) UpdateMonitoringProvider(ctx context.Context, id string, req *models.UpdateMonitoringProviderRequest) (*models.MonitoringProvider, error) {
	var resp models.MonitoringProvider
	if err := c.Put(ctx, fmt.Sprintf("%s/%s", monitoringProvidersPath, id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteMonitoringProvider deletes a monitoring provider.
func (c *Client) DeleteMonitoringProvider(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s", monitoringProvidersPath, id))
	if err != nil && IsNotFoundError(err) {
		return nil
	}
	return err
}
