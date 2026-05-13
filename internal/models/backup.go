package models

// VeritasProtection represents a Veritas backup protection policy (API response)
type VeritasProtection struct {
	ID             int    `json:"id"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Status         string `json:"status,omitempty"`
	ComputeID      string `json:"compute_id,omitempty"`
	ProtectionPlan string `json:"protection_plan,omitempty"`
	PolicyTypeID   int    `json:"policy_type_id,omitempty"`
	Region         string `json:"region,omitempty"`
	AZName         string `json:"az_name,omitempty"`
	Created        string `json:"created,omitempty"`
	Updated        string `json:"updated,omitempty"`
}

// CreateProtectionRequest represents the request to create a Veritas protection policy
type CreateProtectionRequest struct {
	Name            string `form:"name"`
	Description     string `form:"description,omitempty"`
	PolicyTypeID    string `form:"policy_type_id,omitempty"`
	ComputeID       string `form:"compute_id,omitempty"`
	ProtectionPlan  string `form:"protection_plan,omitempty"`
	EnableScheduler string `form:"enable_scheduler,omitempty"`
	StartDate       string `form:"start_date,omitempty"`
	EndDate         string `form:"end_date,omitempty"`
	StartTime       string `form:"start_time,omitempty"`
}

// UpdateProtectionRequest represents the request to update a Veritas protection policy
type UpdateProtectionRequest struct {
	Name            string `form:"name,omitempty"`
	Description     string `form:"description,omitempty"`
	PolicyTypeID    string `form:"policy_type_id,omitempty"`
	ProtectionPlan  string `form:"protection_plan,omitempty"`
	EnableScheduler string `form:"enable_scheduler,omitempty"`
	StartDate       string `form:"start_date,omitempty"`
	EndDate         string `form:"end_date,omitempty"`
	StartTime       string `form:"start_time,omitempty"`
}

// ProtectionPlan represents a Veritas protection plan (API response)
type ProtectionPlan struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	Version     string `json:"version,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
}

// ProtectionPlanListResponse wraps the list API response
type ProtectionPlanListResponse struct {
	PolicyAttributeList []ProtectionPlan `json:"policy_attribute_list"`
}

// CreateProtectionPlanRequest represents the request to create a protection plan
type CreateProtectionPlanRequest struct {
	Name          string `form:"name"`
	Description   string `form:"description,omitempty"`
	ScheduleType  string `form:"scheduleType,omitempty"`
	SelectorKey   string `form:"selector_key,omitempty"`
	SelectorValue string `form:"selector_value,omitempty"`
	Retention     int    `form:"retention,omitempty"`
	RetentionUnit string `form:"retention_unit,omitempty"`
	Recurrence    int    `form:"recurrence,omitempty"`
}
