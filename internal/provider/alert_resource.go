package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &AlertResource{}
var _ resource.ResourceWithImportState = &AlertResource{}
var _ resource.ResourceWithValidateConfig = &AlertResource{}

const (
	alertDefaultEvaluationPeriods = "5"
	alertDefaultDatapointsToAlarm = "3"
)

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

// AlertResource manages airtelcloud_alert.
type AlertResource struct {
	client *client.Client
}

// AlertResourceModel matches the console Create Alert form (required fields only).
type AlertResourceModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Description  types.String  `tfsdk:"description"`
	Emails       types.List    `tfsdk:"emails"`
	ResourceType types.String  `tfsdk:"resource_type"`
	Severity     types.Int64   `tfsdk:"severity"`
	MetricType   types.String  `tfsdk:"metric_type"`
	Function     types.String  `tfsdk:"function"`
	Condition    types.String  `tfsdk:"condition"`
	Value        types.Float64 `tfsdk:"value"`
	State        types.String  `tfsdk:"state"`
	CreatedAt    types.String  `tfsdk:"created_at"`
	UpdatedAt    types.String  `tfsdk:"updated_at"`
}

func (r *AlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud metric alert. Attributes mirror the console Create Alert form.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alert UUID used by GET/PUT/DELETE.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Basic Details
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Alert name.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 100)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description.",
			},
			"emails": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Notification email addresses. At least 1, maximum 6.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(6),
				},
			},

			// Resource
			"resource_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Resource type (for example `Virtual Machine`). Validated against the resource-types catalog.",
			},
			"severity": schema.Int64Attribute{
				Required: true,
				MarkdownDescription: "Alert severity integer matching the console labels: " +
					"`0` Default, `1` Information, `2` Warning, `3` Average, `4` High, `5` Disaster.",
				Validators: []validator.Int64{int64validator.Between(0, 5)},
			},

			// Metrics & Thresholds
			"metric_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Metric type from the metric-types catalog for the chosen resource type (for example `Memory Total`, `CPUAverage`).",
			},
			"function": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Aggregation function from the functions catalog: `Average`, `Minimum`, or `Maximum`.",
			},
			"condition": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Condition from the conditions catalog (for example `Greater than, equal to (>=)` or `Equal to (==)`), or a symbol such as `>=` / `==`.",
			},
			"value": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Threshold value between 0 and 100.",
				Validators:          []validator.Float64{float64validator.Between(0, 100)},
			},

			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alarm state: OK, ALARM, or INSUFFICIENT_DATA.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp.",
			},
		},
	}
}

func (r *AlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *AlertResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.Value.IsNull() && !data.Value.IsUnknown() {
		v := data.Value.ValueFloat64()
		if v < 0 || v > 100 {
			resp.Diagnostics.AddAttributeError(
				path.Root("value"),
				"Invalid Configuration",
				"value must be between 0 and 100.",
			)
		}
	}
}

func (r *AlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolved, diags := r.resolveAlertParams(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := r.buildCreateRequest(ctx, &data, resolved)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.CreateAlert(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert, got error: %s", err))
		return
	}

	if alert.UUIDID == "" {
		resp.Diagnostics.AddError("Client Error", "Alert created but uuid_id could not be resolved from the list API")
		return
	}
	data.ID = types.StringValue(alert.UUIDID)
	mapAlertCreateResponse(&data, alert)
	tflog.Trace(ctx, "created an alert resource", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.GetAlert(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert, got error: %s", err))
		return
	}

	data.Name = types.StringValue(alert.Name)
	data.ResourceType = types.StringValue(alert.ResourceTypeName)
	data.MetricType = types.StringValue(alert.MetricType)
	data.Function = types.StringValue(alert.MetricName)
	data.Value = types.Float64Value(alert.Threshold)
	data.Severity = types.Int64Value(int64(alert.Severity))

	if alert.Description != "" {
		data.Description = types.StringValue(alert.Description)
	}
	if alert.AlertEmails != nil {
		emails, diags := types.ListValueFrom(ctx, types.StringType, alert.AlertEmails)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Emails = emails
		}
	}
	if alert.State != "" {
		data.State = types.StringValue(alert.State)
	}
	if alert.CreatedAt != "" {
		data.CreatedAt = types.StringValue(alert.CreatedAt)
	}
	if alert.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*alert.UpdatedAt)
	}

	// condition is config-owned; GET may normalize it — keep state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolved, diags := r.resolveAlertParams(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := r.buildUpdateRequest(ctx, &data, resolved)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.UpdateAlert(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert, got error: %s", err))
		return
	}

	mapAlertCreateResponse(&data, alert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAlert(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert, got error: %s", err))
		return
	}
}

func (r *AlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type alertResolvedParams struct {
	ResourceTypeName   string
	MetricType         string
	MetricName         string
	ComparisonOperator string
	ResourcesInfo      []models.ResourceInfo
}

func (r *AlertResource) resolveAlertParams(ctx context.Context, data *AlertResourceModel) (*alertResolvedParams, diag.Diagnostics) {
	var diags diag.Diagnostics

	resType, metType, metName, condition, err := r.client.ResolveAlertCreateParams(
		ctx,
		data.ResourceType.ValueString(),
		data.MetricType.ValueString(),
		data.Function.ValueString(),
		data.Condition.ValueString(),
	)
	if err != nil {
		diags.AddError("Alert Catalog Resolution Error", err.Error())
		return nil, diags
	}

	resourcesInfo, err := r.client.ResolveAlertResourcesForType(ctx, resType)
	if err != nil {
		diags.AddError("Alert Resource Resolution Error", err.Error())
		return nil, diags
	}

	tflog.Debug(ctx, "resolved alert create params from static catalogs", map[string]interface{}{
		"resource_type": resType,
		"metric_type":   metType,
		"function":      metName,
		"condition":     condition,
		"resources":     len(resourcesInfo),
	})

	return &alertResolvedParams{
		ResourceTypeName:   resType,
		MetricType:         metType,
		MetricName:         metName,
		ComparisonOperator: condition,
		ResourcesInfo:      resourcesInfo,
	}, diags
}

func (r *AlertResource) buildCreateRequest(ctx context.Context, data *AlertResourceModel, resolved *alertResolvedParams) (*models.CreateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var emails []string
	diags.Append(data.Emails.ElementsAs(ctx, &emails, false)...)
	if diags.HasError() {
		return nil, diags
	}

	eval := alertDefaultEvaluationPeriods
	points := alertDefaultDatapointsToAlarm
	req := &models.CreateAlertRequest{
		Name:               data.Name.ValueString(),
		Organization:       r.client.Organization,
		Project:            r.client.ProjectName,
		ResourceTypeName:   resolved.ResourceTypeName,
		MetricType:         resolved.MetricType,
		MetricName:         resolved.MetricName,
		Threshold:          data.Value.ValueFloat64(),
		ComparisonOperator: resolved.ComparisonOperator,
		EvaluationPeriods:  &eval,
		DatapointsToAlarm:  &points,
		Enabled:            true,
		Severity:           int(data.Severity.ValueInt64()),
		ResourcesInfo:      resolved.ResourcesInfo,
		AlertEmails:        emails,
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}

	return req, diags
}

func (r *AlertResource) buildUpdateRequest(ctx context.Context, data *AlertResourceModel, resolved *alertResolvedParams) (*models.UpdateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var emails []string
	diags.Append(data.Emails.ElementsAs(ctx, &emails, false)...)
	if diags.HasError() {
		return nil, diags
	}

	threshold := data.Value.ValueFloat64()
	op := resolved.ComparisonOperator
	enabled := true
	org := r.client.Organization
	project := r.client.ProjectName

	req := &models.UpdateAlertRequest{
		Name:               data.Name.ValueString(),
		Organization:       &org,
		Project:            &project,
		ResourceTypeName:   resolved.ResourceTypeName,
		MetricType:         resolved.MetricType,
		MetricName:         resolved.MetricName,
		Threshold:          &threshold,
		Severity:           int(data.Severity.ValueInt64()),
		ComparisonOperator: &op,
		EvaluationPeriods:  alertDefaultEvaluationPeriods,
		DatapointsToAlarm:  alertDefaultDatapointsToAlarm,
		Enabled:            &enabled,
		AlertEmails:        emails,
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}

	return req, diags
}

func mapAlertCreateResponse(data *AlertResourceModel, alert *models.Alert) {
	data.Name = types.StringValue(alert.Name)
	data.ResourceType = types.StringValue(alert.ResourceTypeName)
	data.MetricType = types.StringValue(alert.MetricType)
	data.Function = types.StringValue(alert.MetricName)
	data.Value = types.Float64Value(alert.Threshold)
	data.Severity = types.Int64Value(int64(alert.Severity))
	data.State = types.StringValue(alert.State)
	data.CreatedAt = types.StringValue(alert.CreatedAt)

	if alert.Description != nil {
		data.Description = types.StringValue(*alert.Description)
	}
	if alert.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*alert.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}
}
