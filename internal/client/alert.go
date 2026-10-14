package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

func (c *Client) alertBasePath() string {
	return fmt.Sprintf("/api/v2.1/alert/domain/%s/project/%s/alerts", c.Organization, c.ProjectName)
}

func (c *Client) alertPath(id string) string {
	return fmt.Sprintf("%s/%s", c.alertBasePath(), id)
}

// ListAlertResourceTypes returns static resource type labels (e.g. "Virtual Machine").
func (c *Client) ListAlertResourceTypes(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.Get(ctx, models.AlertStaticResourceTypesPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAlertFunctions returns aggregation function labels (Average, Minimum, Maximum).
func (c *Client) ListAlertFunctions(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.Get(ctx, models.AlertStaticFunctionsPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAlertConditions returns comparison operator labels from the catalog.
func (c *Client) ListAlertConditions(ctx context.Context) ([]string, error) {
	var out []string
	if err := c.Get(ctx, models.AlertStaticConditionsPath, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAlertMetricTypes returns metric type labels for a resource type.
func (c *Client) ListAlertMetricTypes(ctx context.Context, resourceType string) ([]string, error) {
	path := models.AlertStaticMetricTypesPath + "?resource_type=" + url.QueryEscape(resourceType)
	var out []string
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResolveAlertResourceType confirms name exists in the resource-types catalog.
func (c *Client) ResolveAlertResourceType(ctx context.Context, name string) (string, error) {
	types, err := c.ListAlertResourceTypes(ctx)
	if err != nil {
		return "", fmt.Errorf("list alert resource types: %w", err)
	}
	if got := matchAlertCatalogValue(types, name); got != "" {
		return got, nil
	}
	return "", fmt.Errorf("resource_type_name %q not found in catalog (available: %s)", name, strings.Join(types, ", "))
}

// ResolveAlertMetricType confirms metricType exists for resourceType.
func (c *Client) ResolveAlertMetricType(ctx context.Context, resourceType, metricType string) (string, error) {
	types, err := c.ListAlertMetricTypes(ctx, resourceType)
	if err != nil {
		return "", fmt.Errorf("list alert metric types: %w", err)
	}
	if got := matchAlertCatalogValue(types, metricType); got != "" {
		return got, nil
	}
	return "", fmt.Errorf("metric_type %q not found for resource_type %q (available: %s)", metricType, resourceType, strings.Join(types, ", "))
}

// ResolveAlertFunction confirms metric_name (Average/Minimum/Maximum) in the functions catalog.
func (c *Client) ResolveAlertFunction(ctx context.Context, name string) (string, error) {
	fns, err := c.ListAlertFunctions(ctx)
	if err != nil {
		return "", fmt.Errorf("list alert functions: %w", err)
	}
	if got := matchAlertCatalogValue(fns, name); got != "" {
		return got, nil
	}
	return "", fmt.Errorf("metric_name %q not found in functions catalog (available: %s)", name, strings.Join(fns, ", "))
}

// ResolveAlertCondition maps a configured operator to a catalog condition label.
// Accepts full labels (e.g. "Greater than, equal to (>=)") or symbols (e.g. ">=").
func (c *Client) ResolveAlertCondition(ctx context.Context, op string) (string, error) {
	conds, err := c.ListAlertConditions(ctx)
	if err != nil {
		return "", fmt.Errorf("list alert conditions: %w", err)
	}
	if got := matchAlertCatalogValue(conds, op); got != "" {
		return got, nil
	}
	if got := matchAlertConditionSymbol(conds, op); got != "" {
		return got, nil
	}
	return "", fmt.Errorf("comparison_operator %q not found in conditions catalog (available: %s)", op, strings.Join(conds, ", "))
}

// ResolveAlertCreateParams validates static catalogs and returns canonical API values.
func (c *Client) ResolveAlertCreateParams(ctx context.Context, resourceType, metricType, metricName, comparisonOperator string) (resType, metType, metName, condition string, err error) {
	resType, err = c.ResolveAlertResourceType(ctx, resourceType)
	if err != nil {
		return "", "", "", "", err
	}
	metType, err = c.ResolveAlertMetricType(ctx, resType, metricType)
	if err != nil {
		return "", "", "", "", err
	}
	metName, err = c.ResolveAlertFunction(ctx, metricName)
	if err != nil {
		return "", "", "", "", err
	}
	condition, err = c.ResolveAlertCondition(ctx, comparisonOperator)
	if err != nil {
		return "", "", "", "", err
	}
	return resType, metType, metName, condition, nil
}

// alertVMListItem is the subset of the compute list response needed for alert resources_info.
type alertVMListItem struct {
	ID                 string   `json:"id"`
	InstanceName       string   `json:"instance_name"`
	Status             string   `json:"status"`
	ProviderInstanceID string   `json:"provider_instance_id"`
	FixedIPs           []string `json:"fixed_ips"`
	FloatingIP         string   `json:"floating_ip"`
	Ports              []struct {
		FixedIPs []string `json:"fixed_ips"`
	} `json:"ports"`
}

func (v alertVMListItem) ip() string {
	if len(v.FixedIPs) > 0 {
		return v.FixedIPs[0]
	}
	if len(v.Ports) > 0 && len(v.Ports[0].FixedIPs) > 0 {
		return v.Ports[0].FixedIPs[0]
	}
	return v.FloatingIP
}

// ResolveAlertResourcesForType builds API resources_info without customer input.
// For Virtual Machine it lists Active computes in the project and fills id/name/ip.
func (c *Client) ResolveAlertResourcesForType(ctx context.Context, resourceType string) ([]models.ResourceInfo, error) {
	if !strings.EqualFold(resourceType, "Virtual Machine") {
		return nil, fmt.Errorf("automatic resource discovery is only supported for Virtual Machine (got %q)", resourceType)
	}

	// Read-only GET of the compute list. Uses a local struct (alert-only) so the
	// shared VM models stay untouched; the list API returns IPs as top-level fixed_ips.
	var computes []alertVMListItem
	if err := c.Get(ctx, fmt.Sprintf("%s/", c.computeBasePath()), &computes); err != nil {
		return nil, fmt.Errorf("list computes for alert resources_info: %w", err)
	}

	out := make([]models.ResourceInfo, 0, len(computes))
	for _, comp := range computes {
		status := strings.TrimSpace(comp.Status)
		if status != "" && !strings.EqualFold(status, "Active") {
			continue
		}
		if comp.ID == "" {
			continue
		}
		ip := comp.ip()
		if ip == "" {
			// API rejects resources without resource_ip.
			continue
		}
		providerID := comp.ProviderInstanceID
		if providerID == "" {
			providerID = comp.ID
		}
		out = append(out, models.ResourceInfo{
			ResourceID:         comp.ID,
			ResourceName:       comp.InstanceName,
			ResourceIP:         ip,
			ProviderInstanceID: providerID,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no Active virtual machines with an IP address found in project %q to attach to the alert", c.ProjectName)
	}
	return out, nil
}

func matchAlertCatalogValue(values []string, want string) string {
	want = strings.TrimSpace(want)
	for _, v := range values {
		if strings.EqualFold(strings.TrimSpace(v), want) {
			return v
		}
	}
	return ""
}

// matchAlertConditionSymbol matches operators like ">=" against labels ending in "(>=)".
func matchAlertConditionSymbol(conds []string, op string) string {
	op = strings.TrimSpace(op)
	if op == "" {
		return ""
	}
	for _, c := range conds {
		sym := alertConditionSymbol(c)
		if sym != "" && sym == op {
			return c
		}
	}
	return ""
}

func alertConditionSymbol(label string) string {
	start := strings.LastIndex(label, "(")
	end := strings.LastIndex(label, ")")
	if start < 0 || end <= start {
		return ""
	}
	return strings.TrimSpace(label[start+1 : end])
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
