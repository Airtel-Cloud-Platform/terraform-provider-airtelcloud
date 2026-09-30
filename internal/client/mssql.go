package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	mssqlSLOPageSize = 100
	mssqlSLOMaxPages = 20
)

// mssqlPollInterval is the delay between GET polls. Tests may shorten it.
var mssqlPollInterval = 10 * time.Second

func (c *Client) mssqlBasePath() string {
	return fmt.Sprintf("/api/v1/dbaas/domain/%s/project/%s/mssql", c.Organization, c.ProjectName)
}

// CreateMSSQL creates an MSSQL instance. The response includes the uuid.
func (c *Client) CreateMSSQL(ctx context.Context, req *models.CreateMSSQLRequest) (*models.MSSQLInstance, error) {
	var instance models.MSSQLInstance
	if err := c.Post(ctx, c.mssqlBasePath(), req, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetMSSQL retrieves an MSSQL instance by uuid.
func (c *Client) GetMSSQL(ctx context.Context, id string) (*models.MSSQLInstance, error) {
	var instance models.MSSQLInstance
	if err := c.Get(ctx, fmt.Sprintf("%s/%s", c.mssqlBasePath(), id), &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

// DeleteMSSQL deletes an MSSQL instance. A 404 is treated as success.
func (c *Client) DeleteMSSQL(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/%s", c.mssqlBasePath(), id))
	if err != nil && !IsNotFoundError(err) {
		return err
	}
	return nil
}

// ListMSSQLEditions lists database editions and the versions available for each.
func (c *Client) ListMSSQLEditions(ctx context.Context) ([]models.MSSQLEdition, error) {
	var editions []models.MSSQLEdition
	if err := c.Get(ctx, c.mssqlBasePath()+"/editions", &editions); err != nil {
		return nil, err
	}
	return editions, nil
}

// ListMSSQLCollations lists collations for one edition.
func (c *Client) ListMSSQLCollations(ctx context.Context, edition string) ([]models.MSSQLCollation, error) {
	var collations []models.MSSQLCollation
	path := fmt.Sprintf("%s/collations/%s", c.mssqlBasePath(), url.PathEscape(edition))
	if err := c.Get(ctx, path, &collations); err != nil {
		return nil, err
	}
	return collations, nil
}

// ListMSSQLFlavors lists flavors for one edition. The flavors API requires the availability-zone header.
func (c *Client) ListMSSQLFlavors(ctx context.Context, edition string) ([]models.MSSQLFlavor, error) {
	var flavors []models.MSSQLFlavor
	path := fmt.Sprintf("%s/flavors/%s", c.mssqlBasePath(), url.PathEscape(edition))
	if err := c.Get(ctx, path, &flavors); err != nil {
		return nil, err
	}
	return flavors, nil
}

// ListMSSQLVolumeTypes lists BLOCK_STORAGE volume types. The API requires the availability-zone header.
func (c *Client) ListMSSQLVolumeTypes(ctx context.Context) ([]models.MSSQLVolumeType, error) {
	var types []models.MSSQLVolumeType
	path := c.mssqlBasePath() + "/volumetypes?group=" + url.QueryEscape("BLOCK_STORAGE")
	if err := c.Get(ctx, path, &types); err != nil {
		return nil, err
	}
	return types, nil
}

// ListMSSQLProtectionPlans lists netbackup SLOs for a region and availability zone.
func (c *Client) ListMSSQLProtectionPlans(ctx context.Context, az string) ([]models.MSSQLSLO, error) {
	var all []models.MSSQLSLO
	for page := 0; page < mssqlSLOMaxPages; page++ {
		q := url.Values{}
		q.Set("region", c.Region)
		q.Set("az", az)
		q.Set("offset", strconv.Itoa(page*mssqlSLOPageSize))
		q.Set("limit", strconv.Itoa(mssqlSLOPageSize))
		q.Set("sort", "name")

		var response models.MSSQLSLOListResponse
		path := c.mssqlBasePath() + "/netbackup/slos?" + q.Encode()
		if err := c.Get(ctx, path, &response); err != nil {
			return nil, err
		}
		all = append(all, response.Data...)
		if len(response.Data) < mssqlSLOPageSize {
			return all, nil
		}
	}
	return nil, fmt.Errorf("protection plan catalog exceeded %d plans", mssqlSLOPageSize*mssqlSLOMaxPages)
}

// ResolveMSSQLEdition finds the edition and its single catalog version.
func (c *Client) ResolveMSSQLEdition(ctx context.Context, edition string) (*models.MSSQLEdition, string, error) {
	editions, err := c.ListMSSQLEditions(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list mssql editions: %w", err)
	}
	for i := range editions {
		if editions[i].Edition != edition {
			continue
		}
		switch len(editions[i].Versions) {
		case 0:
			return nil, "", fmt.Errorf("mssql edition %q has no versions", edition)
		case 1:
			found := editions[i]
			return &found, found.Versions[0], nil
		default:
			return nil, "", fmt.Errorf("mssql edition %q has more than one version (%s); a single version is required", edition, editions[i].Versions)
		}
	}
	return nil, "", fmt.Errorf("mssql edition %q not found", edition)
}

// ResolveMSSQLCollation checks the collation and feature values for an edition.
func (c *Client) ResolveMSSQLCollation(ctx context.Context, edition, collation string, features []string) error {
	collations, err := c.ListMSSQLCollations(ctx, edition)
	if err != nil {
		return fmt.Errorf("failed to list mssql collations: %w", err)
	}
	var match *models.MSSQLCollation
	for i := range collations {
		if collations[i].Collation == collation {
			match = &collations[i]
			break
		}
	}
	if match == nil {
		return fmt.Errorf("mssql collation %q not found for edition %q", collation, edition)
	}

	allowed := map[string]struct{}{}
	for _, feature := range match.Features {
		if !mssqlEditionListed(feature.Editions, edition) {
			continue
		}
		allowed[feature.Value] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, feature := range features {
		if _, dup := seen[feature]; dup {
			return fmt.Errorf("mssql feature %q is listed more than once", feature)
		}
		seen[feature] = struct{}{}
		if _, ok := allowed[feature]; !ok {
			return fmt.Errorf("mssql feature %q is not available for collation %q and edition %q", feature, collation, edition)
		}
	}
	return nil
}

// ResolveMSSQLFlavor finds a flavor by name and converts RAM from MB to GB.
func (c *Client) ResolveMSSQLFlavor(ctx context.Context, az, edition, computeSize string) (*models.MSSQLFlavor, int, error) {
	flavors, err := c.WithAvailabilityZone(az).ListMSSQLFlavors(ctx, edition)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list mssql flavors: %w", err)
	}
	for i := range flavors {
		if flavors[i].Name != computeSize {
			continue
		}
		ramGB, err := flavors[i].RAMGiB()
		if err != nil {
			return nil, 0, err
		}
		flavor := flavors[i]
		return &flavor, ramGB, nil
	}
	return nil, 0, fmt.Errorf("mssql flavor %q not found for edition %q in availability zone %q", computeSize, edition, az)
}

// ResolveMSSQLStorage finds a volume type by label.
func (c *Client) ResolveMSSQLStorage(ctx context.Context, az, storageType string) (*models.MSSQLVolumeType, error) {
	types, err := c.WithAvailabilityZone(az).ListMSSQLVolumeTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mssql volume types: %w", err)
	}
	for i := range types {
		if types[i].Label != storageType {
			continue
		}
		if types[i].Name == "" || types[i].ID == 0 || types[i].ProviderVTID == "" {
			return nil, fmt.Errorf("mssql storage type %q is missing name, id, or provider id", storageType)
		}
		vt := types[i]
		return &vt, nil
	}
	return nil, fmt.Errorf("mssql storage type %q not found in availability zone %q", storageType, az)
}

// ResolveMSSQLProtectionPlan finds a netbackup plan by attributes.name.
func (c *Client) ResolveMSSQLProtectionPlan(ctx context.Context, az, name string) (*models.MSSQLSLO, error) {
	plans, err := c.ListMSSQLProtectionPlans(ctx, az)
	if err != nil {
		return nil, fmt.Errorf("failed to list mssql protection plans: %w", err)
	}
	var match *models.MSSQLSLO
	for i := range plans {
		if plans[i].Attributes.Name != name {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("mssql protection plan %q matched more than once", name)
		}
		plan := plans[i]
		match = &plan
	}
	if match == nil {
		return nil, fmt.Errorf("mssql protection plan %q not found in availability zone %q", name, az)
	}
	if match.ID == "" {
		return nil, fmt.Errorf("mssql protection plan %q has no id", name)
	}
	return match, nil
}

// WaitForMSSQLReady polls GET until status is Active.
// Requested, Initiated, and Resizing keep polling. Failed, Deleting, and any other status fail.
func (c *Client) WaitForMSSQLReady(ctx context.Context, id string, timeout time.Duration) (*models.MSSQLInstance, error) {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		instance, err := c.GetMSSQL(ctx, id)
		if err != nil {
			return nil, err
		}

		switch instance.Status {
		case models.MSSQLStatusActive:
			return instance, nil
		case models.MSSQLStatusRequested, models.MSSQLStatusInitiated, models.MSSQLStatusResizing:
		default:
			return nil, mssqlStatusError(instance)
		}

		if !time.Now().Add(mssqlPollInterval).Before(deadline) {
			return nil, fmt.Errorf("mssql instance did not become ready within %v (status %q)", timeout, instance.Status)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(mssqlPollInterval):
		}
	}
}

// WaitForMSSQLDeleted polls GET until the instance returns 404.
func (c *Client) WaitForMSSQLDeleted(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		instance, err := c.GetMSSQL(ctx, id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return err
		}
		if instance.Status == models.MSSQLStatusFailed {
			return mssqlStatusError(instance)
		}

		if !time.Now().Add(mssqlPollInterval).Before(deadline) {
			return fmt.Errorf("mssql instance deletion timed out after %v", timeout)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(mssqlPollInterval):
		}
	}
}

func mssqlStatusError(instance *models.MSSQLInstance) error {
	msg := instance.Status
	if instance.Message != "" {
		msg = instance.Message
	}
	return fmt.Errorf("mssql instance entered %s state: %s", instance.Status, msg)
}

func mssqlEditionListed(editions []string, edition string) bool {
	for _, item := range editions {
		if item == edition {
			return true
		}
	}
	return false
}
