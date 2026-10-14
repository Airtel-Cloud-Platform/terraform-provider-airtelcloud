package models

// MonitoringDashboardListResponse wraps GET /api/monitoring-admin/v2/dashboards.
type MonitoringDashboardListResponse struct {
	Data  []MonitoringDashboard `json:"data"`
	Total int                   `json:"total"`
}

// MonitoringProviderListResponse wraps GET /api/monitoring-admin/v2/providers.
type MonitoringProviderListResponse struct {
	Data  []MonitoringProvider `json:"data"`
	Total int                  `json:"total"`
}

// MonitoringDashboard is the API response for a monitoring dashboard.
type MonitoringDashboard struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	ResourceType string                     `json:"resource_type"`
	Active       bool                       `json:"active"`
	Description  *string                    `json:"description,omitempty"`
	Panels       []MonitoringDashboardPanel `json:"panels,omitempty"`
	CreatedAt    *string                    `json:"created_at,omitempty"`
	UpdatedAt    *string                    `json:"updated_at,omitempty"`
	CreatedBy    *string                    `json:"created_by,omitempty"`
	UpdatedBy    *string                    `json:"updated_by,omitempty"`
}

// MonitoringDashboardPanel is a panel within a dashboard.
type MonitoringDashboardPanel struct {
	ID            *string                    `json:"id,omitempty"`
	Title         string                     `json:"title"`
	ChartType     string                     `json:"chart_type"`
	Description   *string                    `json:"description,omitempty"`
	X             int64                      `json:"x"`
	Y             int64                      `json:"y"`
	W             int64                      `json:"w"`
	H             int64                      `json:"h"`
	MinW          *int64                     `json:"min_w,omitempty"`
	MinH          *int64                     `json:"min_h,omitempty"`
	DisplayOrder  int64                      `json:"display_order"`
	DisplayConfig map[string]interface{}     `json:"display_config,omitempty"`
	Queries       []MonitoringDashboardQuery `json:"queries,omitempty"`
}

// MonitoringDashboardQuery is a query within a panel.
type MonitoringDashboardQuery struct {
	ID                *string                `json:"id,omitempty"`
	QueryTemplate     string                 `json:"query_template"`
	QueryType         string                 `json:"query_type"`
	RefID             string                 `json:"ref_id"`
	DisplayOrder      *int64                 `json:"display_order,omitempty"`
	PlaceholderConfig map[string]interface{} `json:"placeholder_config,omitempty"`
	SeriesConfig      map[string]interface{} `json:"series_config,omitempty"`
}

// CreateMonitoringDashboardRequest is the create/update body for dashboards.
type CreateMonitoringDashboardRequest struct {
	Name         string                     `json:"name"`
	ResourceType string                     `json:"resource_type"`
	Active       *bool                      `json:"active,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Panels       []MonitoringDashboardPanel `json:"panels,omitempty"`
}

// UpdateMonitoringDashboardRequest is the PUT body for dashboards.
type UpdateMonitoringDashboardRequest struct {
	Name         string                     `json:"name"`
	ResourceType string                     `json:"resource_type"`
	Active       *bool                      `json:"active,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Panels       []MonitoringDashboardPanel `json:"panels,omitempty"`
}

// MonitoringProvider is the API response for a monitoring provider.
// api_token is never returned by the API.
type MonitoringProvider struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ResourceType     string  `json:"resource_type"`
	ProviderType     string  `json:"provider_type"`
	DashboardAuthURL string  `json:"dashboard_auth_url"`
	APIKeyName       string  `json:"api_key_name"`
	DataSourceUID    string  `json:"data_source_uid"`
	DataSourceID     string  `json:"data_source_id"`
	AZName           *string `json:"az_name,omitempty"`
	Active           bool    `json:"active"`
	Description      *string `json:"description,omitempty"`
	CreatedAt        *string `json:"created_at,omitempty"`
	UpdatedAt        *string `json:"updated_at,omitempty"`
	CreatedBy        *string `json:"created_by,omitempty"`
	UpdatedBy        *string `json:"updated_by,omitempty"`
}

// CreateMonitoringProviderRequest is the create body for providers.
type CreateMonitoringProviderRequest struct {
	Name             string  `json:"name"`
	ResourceType     string  `json:"resource_type"`
	ProviderType     *string `json:"provider_type,omitempty"`
	DashboardAuthURL string  `json:"dashboard_auth_url"`
	APIKeyName       string  `json:"api_key_name"`
	APIToken         string  `json:"api_token"`
	DataSourceUID    string  `json:"data_source_uid"`
	DataSourceID     string  `json:"data_source_id"`
	AZName           *string `json:"az_name,omitempty"`
	Active           *bool   `json:"active,omitempty"`
	Description      *string `json:"description,omitempty"`
}

// UpdateMonitoringProviderRequest is the PUT body for providers.
// APIToken is omitted when empty so the API keeps the existing token.
type UpdateMonitoringProviderRequest struct {
	Name             string  `json:"name"`
	ResourceType     string  `json:"resource_type"`
	ProviderType     *string `json:"provider_type,omitempty"`
	DashboardAuthURL string  `json:"dashboard_auth_url"`
	APIKeyName       string  `json:"api_key_name"`
	APIToken         *string `json:"api_token,omitempty"`
	DataSourceUID    string  `json:"data_source_uid"`
	DataSourceID     string  `json:"data_source_id"`
	AZName           *string `json:"az_name,omitempty"`
	Active           *bool   `json:"active,omitempty"`
	Description      *string `json:"description,omitempty"`
}
