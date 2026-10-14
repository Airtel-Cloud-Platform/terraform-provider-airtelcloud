package models

// ResourceInfo is one monitored resource on alert create.
type ResourceInfo struct {
	ResourceID         string `json:"resource_id"`
	ResourceName       string `json:"resource_name"`
	ResourceIP         string `json:"resource_ip"`
	ProviderInstanceID string `json:"provider_instance_id"`
}

// CreateAlertRequest is the POST body for AlertCreate.
type CreateAlertRequest struct {
	Name               string         `json:"name"`
	Description        *string        `json:"description,omitempty"`
	Organization       string         `json:"organization,omitempty"`
	Project            string         `json:"project,omitempty"`
	ResourceTypeName   string         `json:"resource_type_name"`
	MetricType         string         `json:"metric_type"`
	MetricName         string         `json:"metric_name"`
	Threshold          float64        `json:"threshold"`
	ComparisonOperator string         `json:"comparison_operator"`
	EvaluationPeriods  *string        `json:"evaluation_periods,omitempty"`
	DatapointsToAlarm  *string        `json:"datapoints_to_alarm,omitempty"`
	Enabled            bool           `json:"enabled"`
	Severity           int            `json:"severity"`
	ResourcesInfo      []ResourceInfo `json:"resources_info"`
	AlertEmails        []string       `json:"alert_emails"`
}

// UpdateAlertRequest is the PUT body for AlertUpdate.
type UpdateAlertRequest struct {
	Name               string   `json:"name"`
	Description        *string  `json:"description,omitempty"`
	Organization       *string  `json:"organization,omitempty"`
	Project            *string  `json:"project,omitempty"`
	ResourceTypeName   string   `json:"resource_type_name"`
	MetricType         string   `json:"metric_type"`
	MetricName         string   `json:"metric_name"`
	Threshold          *float64 `json:"threshold,omitempty"`
	Severity           int      `json:"severity"`
	ComparisonOperator *string  `json:"comparison_operator,omitempty"`
	EvaluationPeriods  string   `json:"evaluation_periods"`
	DatapointsToAlarm  string   `json:"datapoints_to_alarm"`
	Enabled            *bool    `json:"enabled,omitempty"`
	AlertEmails        []string `json:"alert_emails,omitempty"`
}

// AlertResource is a resource association returned on create/update Alert.
type AlertResource struct {
	ID         int    `json:"id"`
	AlertID    int    `json:"alert_id"`
	ResourceID string `json:"resource_id"`
	CreatedAt  string `json:"created_at"`
}

// Alert is the create/update response schema (id is integer).
// UUIDID is not returned on create; resolve it via ListAlerts after create.
// GET/PUT/DELETE paths require UUIDID, not the integer ID.
type Alert struct {
	ID                 int             `json:"id"`
	UUIDID             string          `json:"uuid_id,omitempty"`
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	Organization       *string         `json:"organization"`
	Project            *string         `json:"project"`
	ResourceTypeName   string          `json:"resource_type_name"`
	MetricType         string          `json:"metric_type"`
	MetricName         string          `json:"metric_name"`
	Threshold          float64         `json:"threshold"`
	ComparisonOperator string          `json:"comparison_operator"`
	EvaluationPeriods  *string         `json:"evaluation_periods"`
	DatapointsToAlarm  *string         `json:"datapoints_to_alarm"`
	Enabled            bool            `json:"enabled"`
	Severity           int             `json:"severity"`
	State              string          `json:"state"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          *string         `json:"updated_at"`
	Resources          []AlertResource `json:"resources"`
}

// AlertListItem is one entry from GET .../alerts list (uses uuid_id).
type AlertListItem struct {
	UUIDID           string   `json:"uuid_id"`
	Name             string   `json:"name"`
	MetricType       string   `json:"metric_type"`
	ResourceTypeName string   `json:"resource_type_name"`
	Organisation     string   `json:"organisation"`
	Cell             string   `json:"cell"`
	Value            float64  `json:"value"`
	Severity         int      `json:"severity"`
	Enabled          bool     `json:"enabled"`
	AlertEmails      []string `json:"alert_emails"`
}

// AlertListResponse wraps list alerts.
type AlertListResponse struct {
	AlertData []AlertListItem `json:"alert_data"`
	Total     int             `json:"total"`
	Offset    int             `json:"offset"`
	Size      int             `json:"size"`
}

// AlertGetResponse models GET .../alerts/{id}. OpenAPI documents AlertListShow
// (no id); some deployments may return a fuller Alert. Absent id stays 0.
type AlertGetResponse struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Organization       *string  `json:"organization"`
	Project            *string  `json:"project"`
	ResourceTypeName   string   `json:"resource_type_name"`
	MetricType         string   `json:"metric_type"`
	MetricName         string   `json:"metric_name"`
	Threshold          float64  `json:"threshold"`
	ComparisonOperator string   `json:"comparison_operator"`
	EvaluationPeriods  *string  `json:"evaluation_periods"`
	DatapointsToAlarm  *string  `json:"datapoints_to_alarm"`
	Enabled            bool     `json:"enabled"`
	Severity           int      `json:"severity"`
	State              string   `json:"state"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          *string  `json:"updated_at"`
	AlertEmails        []string `json:"alert_emails"`
}

// CreateAlertNotificationRequest is the POST body for NotificationCreate.
type CreateAlertNotificationRequest struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
	AlertID int    `json:"alert_id"`
}

// UpdateAlertNotificationRequest is the PUT body for NotificationUpdate.
type UpdateAlertNotificationRequest struct {
	Type    *string `json:"type,omitempty"`
	Target  *string `json:"target,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// AlertNotification is the notification create/get/update response.
type AlertNotification struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	Enabled   bool   `json:"enabled"`
	AlertID   int    `json:"alert_id"`
	CreatedAt string `json:"created_at"`
}
