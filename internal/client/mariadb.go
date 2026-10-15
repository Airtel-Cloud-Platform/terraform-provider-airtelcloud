package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const mariadbPollInterval = 10 * time.Second

func (c *Client) mariadbBasePath() string {
	return fmt.Sprintf("/api/v1/dbaas/domain/%s/project/%s/mariadb", c.Organization, c.ProjectName)
}

func (c *Client) mariadbVolumeTypesPath() string {
	return fmt.Sprintf("/api/v1/dbaas/domain/%s/project/%s/mssql/volumetypes", c.Organization, c.ProjectName)
}

// CreateMariaDBCluster creates a MariaDB cluster. The create response includes
// the uuid immediately; status starts as Creating.
func (c *Client) CreateMariaDBCluster(ctx context.Context, req *models.CreateMariaDBClusterRequest) (*models.MariaDBCluster, error) {
	var cluster models.MariaDBCluster
	if err := c.Post(ctx, c.mariadbBasePath(), req, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// GetMariaDBCluster retrieves a MariaDB cluster by uuid.
func (c *Client) GetMariaDBCluster(ctx context.Context, id string) (*models.MariaDBCluster, error) {
	var cluster models.MariaDBCluster
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", c.mariadbBasePath(), id), &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// DeleteMariaDBCluster deletes a MariaDB cluster. A 404 is treated as success.
func (c *Client) DeleteMariaDBCluster(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s", c.mariadbBasePath(), id))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// ListMariaDBFlavors lists mariadb flavors. Callers that need the AZ header
// should use WithAvailabilityZone first.
func (c *Client) ListMariaDBFlavors(ctx context.Context) ([]models.MariaDBFlavor, error) {
	var flavors []models.MariaDBFlavor
	if err := c.Get(ctx, c.mariadbBasePath()+"/flavors", &flavors); err != nil {
		return nil, err
	}
	return flavors, nil
}

// ListMariaDBTopologies lists mariadb topologies.
func (c *Client) ListMariaDBTopologies(ctx context.Context) ([]models.MariaDBTopology, error) {
	var topologies []models.MariaDBTopology
	if err := c.Get(ctx, c.mariadbBasePath()+"/topologies", &topologies); err != nil {
		return nil, err
	}
	return topologies, nil
}

// ListMariaDBVolumeTypes lists BLOCK_STORAGE volume types (shared mssql volumetypes path).
func (c *Client) ListMariaDBVolumeTypes(ctx context.Context) ([]models.MariaDBVolumeType, error) {
	var types []models.MariaDBVolumeType
	path := c.mariadbVolumeTypesPath() + "?group=" + url.QueryEscape("BLOCK_STORAGE")
	if err := c.Get(ctx, path, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// ListMariaDBZones lists availability zones for the client's region.
func (c *Client) ListMariaDBZones(ctx context.Context) (*models.MariaDBZoneListResponse, error) {
	var response models.MariaDBZoneListResponse
	path := "/api/auth-mgmt/v1/zones?regionCode=" + url.QueryEscape(c.Region)
	if err := c.Get(ctx, path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ListMariaDBProtectionPlans lists DBaaS backup protection plans.
func (c *Client) ListMariaDBProtectionPlans(ctx context.Context) (*models.MariaDBProtectionPlanListResponse, error) {
	var response models.MariaDBProtectionPlanListResponse
	if err := c.Get(ctx, c.mariadbBasePath()+"/protection-plans", &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ListMariaDBVersions lists supported MariaDB engine versions.
func (c *Client) ListMariaDBVersions(ctx context.Context) ([]models.MariaDBVersion, error) {
	var versions []models.MariaDBVersion
	if err := c.Get(ctx, c.mariadbBasePath()+"/versions", &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// ResolveMariaDBAvailabilityZone confirms azCode exists in the zones catalog
// and returns the matching zone.
func (c *Client) ResolveMariaDBAvailabilityZone(ctx context.Context, azCode string) (*models.MariaDBZone, error) {
	zones, err := c.ListMariaDBZones(ctx)
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

// ResolveMariaDBFlavor finds a catalog flavor whose name matches computeSize.
// The flavors API requires the availability-zone header. RAM is already in GB.
func (c *Client) ResolveMariaDBFlavor(ctx context.Context, az, computeSize string) (*models.MariaDBFlavor, error) {
	flavors, err := c.WithAvailabilityZone(az).ListMariaDBFlavors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mariadb flavors: %w", err)
	}
	for i := range flavors {
		if flavors[i].Name == computeSize {
			flavor := flavors[i]
			return &flavor, nil
		}
	}
	return nil, fmt.Errorf("mariadb flavor %q not found in availability zone %q", computeSize, az)
}

// ResolveMariaDBTopology selects master or master-slave from the catalog.
func (c *Client) ResolveMariaDBTopology(ctx context.Context, highAvailability bool) (string, error) {
	wanted := models.MariaDBTopologyMaster
	if highAvailability {
		wanted = models.MariaDBTopologyMasterSlave
	}

	topologies, err := c.ListMariaDBTopologies(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list mariadb topologies: %w", err)
	}
	for _, topology := range topologies {
		if topology.Topology == wanted {
			return wanted, nil
		}
	}
	return "", fmt.Errorf("mariadb topology %q not found", wanted)
}

// ResolveMariaDBStorage finds a volume type whose label matches storageType.
// The volume-types API requires the availability-zone header.
func (c *Client) ResolveMariaDBStorage(ctx context.Context, az, storageType string) (*models.MariaDBVolumeType, error) {
	types, err := c.WithAvailabilityZone(az).ListMariaDBVolumeTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mariadb storage types: %w", err)
	}
	for i := range types {
		if types[i].Label == storageType {
			vt := types[i]
			return &vt, nil
		}
	}
	return nil, fmt.Errorf("mariadb storage type %q not found in availability zone %q", storageType, az)
}

// ResolveMariaDBProtectionPlan confirms a protection plan value exists in the catalog.
func (c *Client) ResolveMariaDBProtectionPlan(ctx context.Context, value string) (*models.MariaDBProtectionPlan, error) {
	plans, err := c.ListMariaDBProtectionPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mariadb protection plans: %w", err)
	}
	for i := range plans.ProtectionPlans {
		if plans.ProtectionPlans[i].Value == value {
			plan := plans.ProtectionPlans[i]
			return &plan, nil
		}
	}
	return nil, fmt.Errorf("mariadb protection plan %q not found", value)
}

// ResolveMariaDBVersion confirms the version exists and is supported in the catalog.
func (c *Client) ResolveMariaDBVersion(ctx context.Context, version string) (*models.MariaDBVersion, error) {
	versions, err := c.ListMariaDBVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mariadb versions: %w", err)
	}
	for i := range versions {
		if versions[i].Version == version && versions[i].Supported {
			v := versions[i]
			return &v, nil
		}
	}
	return nil, fmt.Errorf("mariadb version %q not found or not supported", version)
}

// WaitForMariaDBReady polls GET until status is Active and connection_string is set.
func (c *Client) WaitForMariaDBReady(ctx context.Context, id string, timeout time.Duration) (*models.MariaDBCluster, error) {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		cluster, err := c.GetMariaDBCluster(ctx, id)
		if err != nil {
			return nil, err
		}

		switch cluster.Status {
		case models.MariaDBStatusActive:
			if cluster.ConnectionString != nil && strings.TrimSpace(*cluster.ConnectionString) != "" {
				return cluster, nil
			}
		case models.MariaDBStatusFailed, models.MariaDBStatusDeleteFailed:
			msg := cluster.Status
			if cluster.Message != "" {
				msg = cluster.Message
			}
			return nil, fmt.Errorf("mariadb cluster entered %s state: %s", cluster.Status, msg)
		case models.MariaDBStatusDeleted:
			return nil, fmt.Errorf("mariadb cluster was deleted while waiting to become ready")
		}

		if !time.Now().Add(mariadbPollInterval).Before(deadline) {
			return nil, fmt.Errorf("mariadb cluster did not become ready within %v (status %q)", timeout, cluster.Status)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(mariadbPollInterval):
		}
	}
}

// WaitForMariaDBDeleted polls until the cluster is gone (404) or status is Deleted.
func (c *Client) WaitForMariaDBDeleted(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cluster, err := c.GetMariaDBCluster(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if cluster.Status == models.MariaDBStatusDeleted {
			return nil
		}
		if cluster.Status == models.MariaDBStatusDeleteFailed {
			msg := cluster.Status
			if cluster.Message != "" {
				msg = cluster.Message
			}
			return fmt.Errorf("mariadb cluster delete failed: %s", msg)
		}

		if !time.Now().Add(mariadbPollInterval).Before(deadline) {
			return fmt.Errorf("mariadb cluster deletion timed out after %v (status %q)", timeout, cluster.Status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(mariadbPollInterval):
		}
	}
}
