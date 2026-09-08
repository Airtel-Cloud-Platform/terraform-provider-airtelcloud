package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const postgresPollInterval = 10 * time.Second

func (c *Client) postgresBasePath() string {
	return fmt.Sprintf("/api/v1/dbaas/domain/%s/project/%s/postgres", c.Organization, c.ProjectName)
}

func (c *Client) postgresVolumeTypesPath() string {
	return fmt.Sprintf("/api/v1/dbaas/domain/%s/project/%s/mssql/volumetypes", c.Organization, c.ProjectName)
}

// CreatePostgresCluster creates a PostgreSQL cluster. The create response includes
// the uuid immediately; status starts as Creating.
func (c *Client) CreatePostgresCluster(ctx context.Context, req *models.CreatePostgresClusterRequest) (*models.PostgresCluster, error) {
	var cluster models.PostgresCluster
	if err := c.Post(ctx, c.postgresBasePath(), req, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// GetPostgresCluster retrieves a PostgreSQL cluster by uuid.
func (c *Client) GetPostgresCluster(ctx context.Context, id string) (*models.PostgresCluster, error) {
	var cluster models.PostgresCluster
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", c.postgresBasePath(), id), &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// DeletePostgresCluster deletes a PostgreSQL cluster. A 404 is treated as success.
func (c *Client) DeletePostgresCluster(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s", c.postgresBasePath(), id))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// ListPostgresFlavors lists postgres flavors. Callers that need the AZ header
// should use WithAvailabilityZone first.
func (c *Client) ListPostgresFlavors(ctx context.Context) ([]models.PostgresFlavor, error) {
	var flavors []models.PostgresFlavor
	if err := c.Get(ctx, c.postgresBasePath()+"/flavors", &flavors); err != nil {
		return nil, err
	}
	return flavors, nil
}

// ListPostgresTopologies lists postgres topologies.
func (c *Client) ListPostgresTopologies(ctx context.Context) ([]models.PostgresTopology, error) {
	var topologies []models.PostgresTopology
	if err := c.Get(ctx, c.postgresBasePath()+"/topologies", &topologies); err != nil {
		return nil, err
	}
	return topologies, nil
}

// ListPostgresVolumeTypes lists BLOCK_STORAGE volume types for postgres.
func (c *Client) ListPostgresVolumeTypes(ctx context.Context) ([]models.PostgresVolumeType, error) {
	var types []models.PostgresVolumeType
	path := c.postgresVolumeTypesPath() + "?group=" + url.QueryEscape("BLOCK_STORAGE")
	if err := c.Get(ctx, path, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// ListPostgresZones lists availability zones for the client's region.
func (c *Client) ListPostgresZones(ctx context.Context) (*models.PostgresZoneListResponse, error) {
	var response models.PostgresZoneListResponse
	path := "/api/auth-mgmt/v1/zones?regionCode=" + url.QueryEscape(c.Region)
	if err := c.Get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ListPostgresProtectionPlans lists DBaaS backup protection plans.
func (c *Client) ListPostgresProtectionPlans(ctx context.Context) (*models.PostgresProtectionPlanListResponse, error) {
	var response models.PostgresProtectionPlanListResponse
	if err := c.Get(ctx, c.postgresBasePath()+"/protection-plans", &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ResolvePostgresAvailabilityZone confirms azCode exists in the zones catalog
// and returns the matching zone.
func (c *Client) ResolvePostgresAvailabilityZone(ctx context.Context, azCode string) (*models.PostgresZone, error) {
	zones, err := c.ListPostgresZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list availability zones: %w", err)
	}
	for i := range zones.Items {
		if zones.Items[i].AZCode == azCode {
			return &zones.Items[i], nil
		}
	}
	return nil, fmt.Errorf("availability zone %q not found", azCode)
}

// ResolvePostgresFlavor finds an active-catalog flavor whose name matches computeSize.
// The flavors API requires the availability-zone header.
func (c *Client) ResolvePostgresFlavor(ctx context.Context, az, computeSize string) (*models.PostgresFlavor, error) {
	flavors, err := c.WithAvailabilityZone(az).ListPostgresFlavors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list postgres flavors: %w", err)
	}
	for i := range flavors {
		if flavors[i].Name == computeSize {
			flavor := flavors[i]
			return &flavor, nil
		}
	}
	return nil, fmt.Errorf("postgres flavor %q not found in availability zone %q", computeSize, az)
}

// ResolvePostgresTopology selects standalone or primary-standby from the catalog.
func (c *Client) ResolvePostgresTopology(ctx context.Context, highAvailability bool) (string, error) {
	wanted := models.PostgresTopologyStandalone
	if highAvailability {
		wanted = models.PostgresTopologyPrimaryStandby
	}

	topologies, err := c.ListPostgresTopologies(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list postgres topologies: %w", err)
	}
	for _, topology := range topologies {
		if topology.Topology == wanted {
			return wanted, nil
		}
	}
	return "", fmt.Errorf("postgres topology %q not found", wanted)
}

// ResolvePostgresStorage finds a volume type whose label matches storageType.
// The volume-types API requires the availability-zone header.
func (c *Client) ResolvePostgresStorage(ctx context.Context, az, storageType string) (*models.PostgresVolumeType, error) {
	types, err := c.WithAvailabilityZone(az).ListPostgresVolumeTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list postgres storage types: %w", err)
	}
	for i := range types {
		if types[i].Label == storageType {
			vt := types[i]
			return &vt, nil
		}
	}
	return nil, fmt.Errorf("postgres storage type %q not found in availability zone %q", storageType, az)
}

// ResolvePostgresProtectionPlan confirms a protection plan value exists in the catalog.
func (c *Client) ResolvePostgresProtectionPlan(ctx context.Context, value string) (*models.PostgresProtectionPlan, error) {
	plans, err := c.ListPostgresProtectionPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list postgres protection plans: %w", err)
	}
	for i := range plans.ProtectionPlans {
		if plans.ProtectionPlans[i].Value == value {
			plan := plans.ProtectionPlans[i]
			return &plan, nil
		}
	}
	return nil, fmt.Errorf("postgres protection plan %q not found", value)
}

// WaitForPostgresReady polls GET until status is Active and connection_string is set.
func (c *Client) WaitForPostgresReady(ctx context.Context, id string, timeout time.Duration) (*models.PostgresCluster, error) {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		cluster, err := c.GetPostgresCluster(ctx, id)
		if err != nil {
			return nil, err
		}

		switch cluster.Status {
		case models.PostgresStatusActive:
			if cluster.ConnectionString != nil && strings.TrimSpace(*cluster.ConnectionString) != "" {
				return cluster, nil
			}
		case models.PostgresStatusFailed, models.PostgresStatusDeleteFailed:
			msg := cluster.Status
			if cluster.Message != "" {
				msg = cluster.Message
			}
			return nil, fmt.Errorf("postgres cluster entered %s state: %s", cluster.Status, msg)
		case models.PostgresStatusDeleted:
			return nil, fmt.Errorf("postgres cluster was deleted while waiting to become ready")
		}

		if !time.Now().Add(postgresPollInterval).Before(deadline) {
			return nil, fmt.Errorf("postgres cluster did not become ready within %v (status %q)", timeout, cluster.Status)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(postgresPollInterval):
		}
	}
}

// WaitForPostgresDeleted polls until the cluster is gone (404) or status is Deleted.
func (c *Client) WaitForPostgresDeleted(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cluster, err := c.GetPostgresCluster(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if cluster.Status == models.PostgresStatusDeleted {
			return nil
		}
		if cluster.Status == models.PostgresStatusDeleteFailed {
			msg := cluster.Status
			if cluster.Message != "" {
				msg = cluster.Message
			}
			return fmt.Errorf("postgres cluster delete failed: %s", msg)
		}

		if !time.Now().Add(postgresPollInterval).Before(deadline) {
			return fmt.Errorf("postgres cluster deletion timed out after %v (status %q)", timeout, cluster.Status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(postgresPollInterval):
		}
	}
}
