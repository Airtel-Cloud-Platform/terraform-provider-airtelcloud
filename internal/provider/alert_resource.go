package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

// AlertResource manages airtelcloud_alert.
type AlertResource struct {
	client *client.Client
}

// AlertResourceModel is the Terraform state model.
type AlertResourceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Description        types.String  `tfsdk:"description"`
	ResourceTypeName   types.String  `tfsdk:"resource_type_name"`
	MetricType         types.String  `tfsdk:"metric_type"`
	MetricName         types.String  `tfsdk:"metric_name"`
	Threshold          types.Float64 `tfsdk:"threshold"`
	ComparisonOperator types.String  `tfsdk:"comparison_operator"`
	EvaluationPeriods  types.String  `tfsdk:"evaluation_periods"`
	DatapointsToAlarm  types.String  `tfsdk:"datapoints_to_alarm"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	Severity           types.Int64   `tfsdk:"severity"`
	AlertEmails        types.List    `tfsdk:"alert_emails"`
	ResourcesInfo      types.List    `tfsdk:"resources_info"`
	State              types.String  `tfsdk:"state"`
	CreatedAt          types.String  `tfsdk:"created_at"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
}

// AlertResourceInfoModel is one resources_info nested object.
type AlertResourceInfoModel struct {
	ResourceID         types.String `tfsdk:"resource_id"`
	ResourceName       types.String `tfsdk:"resource_name"`
	ResourceIP         types.String `tfsdk:"resource_ip"`
	ProviderInstanceID types.String `tfsdk:"provider_instance_id"`
}

func (r *AlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud metric alert (Alert Manager v2.1).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alert UUID used by GET/PUT/DELETE. Create returns an integer id; the provider resolves `uuid_id` from the list API afterward.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Alert name.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 100)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description.",
			},
			"resource_type_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Resource type name (for example `Virtual Machine`).",
			},
			"metric_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Metric type (for example `CPUAverage`).",
			},
			"metric_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Aggregation function: `Average`, `Minimum`, or `Maximum`.",
			},
			"threshold": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Threshold between 0 and 100.",
				Validators:          []validator.Float64{float64validator.Between(0, 100)},
			},
			"comparison_operator": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Comparison operator. Live API accepts labels like `Greater than, equal to (>=)` or symbols `>=`.",
			},
			"evaluation_periods": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Evaluation periods. Required when updating the alert.",
			},
			"datapoints_to_alarm": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Datapoints to alarm. Required when updating the alert.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the alert is enabled. Defaults to true.",
			},
			"severity": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Alert severity.",
			},
			"alert_emails": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Email addresses notified when the alert fires.",
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"resources_info": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Resources to monitor. Force-new; changing this replaces the alert.",
				Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Resource id.",
						},
						"resource_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Resource name.",
						},
						"resource_ip": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Resource IP address.",
						},
						"provider_instance_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Provider instance id.",
						},
					},
				},
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
	if !data.Threshold.IsNull() && !data.Threshold.IsUnknown() {
		v := data.Threshold.ValueFloat64()
		if v < 0 || v > 100 {
			resp.Diagnostics.AddAttributeError(
				path.Root("threshold"),
				"Invalid Configuration",
				"threshold must be between 0 and 100.",
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

	createReq, diags := r.buildCreateRequest(ctx, &data)
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

	// GET uses uuid_id; keep state id (do not overwrite with integer create id).
	data.Name = types.StringValue(alert.Name)
	data.ResourceTypeName = types.StringValue(alert.ResourceTypeName)
	data.MetricType = types.StringValue(alert.MetricType)
	data.MetricName = types.StringValue(alert.MetricName)
	data.Threshold = types.Float64Value(alert.Threshold)
	data.ComparisonOperator = types.StringValue(alert.ComparisonOperator)
	data.Enabled = types.BoolValue(alert.Enabled)
	data.Severity = types.Int64Value(int64(alert.Severity))

	if alert.Description != "" {
		data.Description = types.StringValue(alert.Description)
	}

	if alert.AlertEmails != nil {
		emails, diags := types.ListValueFrom(ctx, types.StringType, alert.AlertEmails)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.AlertEmails = emails
		}
	}

	// fields absent from AlertListShow stay from prior state
	if alert.EvaluationPeriods != nil {
		data.EvaluationPeriods = types.StringValue(*alert.EvaluationPeriods)
	}
	if alert.DatapointsToAlarm != nil {
		data.DatapointsToAlarm = types.StringValue(*alert.DatapointsToAlarm)
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

	// resources_info is create-only; GET does not round-trip it — keep state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.EvaluationPeriods.IsNull() || data.EvaluationPeriods.IsUnknown() || data.EvaluationPeriods.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("evaluation_periods"),
			"Invalid Configuration",
			"evaluation_periods is required when updating an alert.",
		)
	}
	if data.DatapointsToAlarm.IsNull() || data.DatapointsToAlarm.IsUnknown() || data.DatapointsToAlarm.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("datapoints_to_alarm"),
			"Invalid Configuration",
			"datapoints_to_alarm is required when updating an alert.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := r.buildUpdateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.UpdateAlert(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert, got error: %s", err))
		return
	}

	// Keep uuid_id from state; update response returns integer id only.
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

func (r *AlertResource) buildCreateRequest(ctx context.Context, data *AlertResourceModel) (*models.CreateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var emails []string
	diags.Append(data.AlertEmails.ElementsAs(ctx, &emails, false)...)

	var resources []AlertResourceInfoModel
	diags.Append(data.ResourcesInfo.ElementsAs(ctx, &resources, false)...)
	if diags.HasError() {
		return nil, diags
	}

	info := make([]models.ResourceInfo, 0, len(resources))
	for _, item := range resources {
		info = append(info, models.ResourceInfo{
			ResourceID:         item.ResourceID.ValueString(),
			ResourceName:       item.ResourceName.ValueString(),
			ResourceIP:         item.ResourceIP.ValueString(),
			ProviderInstanceID: item.ProviderInstanceID.ValueString(),
		})
	}

	req := &models.CreateAlertRequest{
		Name:               data.Name.ValueString(),
		Organization:       r.client.Organization,
		Project:            r.client.ProjectName,
		ResourceTypeName:   data.ResourceTypeName.ValueString(),
		MetricType:         data.MetricType.ValueString(),
		MetricName:         data.MetricName.ValueString(),
		Threshold:          data.Threshold.ValueFloat64(),
		ComparisonOperator: data.ComparisonOperator.ValueString(),
		Enabled:            data.Enabled.ValueBool(),
		Severity:           int(data.Severity.ValueInt64()),
		ResourcesInfo:      info,
		AlertEmails:        emails,
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}
	if !data.EvaluationPeriods.IsNull() && !data.EvaluationPeriods.IsUnknown() {
		v := data.EvaluationPeriods.ValueString()
		req.EvaluationPeriods = &v
	}
	if !data.DatapointsToAlarm.IsNull() && !data.DatapointsToAlarm.IsUnknown() {
		v := data.DatapointsToAlarm.ValueString()
		req.DatapointsToAlarm = &v
	}

	return req, diags
}

func (r *AlertResource) buildUpdateRequest(ctx context.Context, data *AlertResourceModel) (*models.UpdateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var emails []string
	diags.Append(data.AlertEmails.ElementsAs(ctx, &emails, false)...)
	if diags.HasError() {
		return nil, diags
	}

	threshold := data.Threshold.ValueFloat64()
	op := data.ComparisonOperator.ValueString()
	enabled := data.Enabled.ValueBool()
	org := r.client.Organization
	project := r.client.ProjectName

	req := &models.UpdateAlertRequest{
		Name:               data.Name.ValueString(),
		Organization:       &org,
		Project:            &project,
		ResourceTypeName:   data.ResourceTypeName.ValueString(),
		MetricType:         data.MetricType.ValueString(),
		MetricName:         data.MetricName.ValueString(),
		Threshold:          &threshold,
		Severity:           int(data.Severity.ValueInt64()),
		ComparisonOperator: &op,
		EvaluationPeriods:  data.EvaluationPeriods.ValueString(),
		DatapointsToAlarm:  data.DatapointsToAlarm.ValueString(),
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
	data.ResourceTypeName = types.StringValue(alert.ResourceTypeName)
	data.MetricType = types.StringValue(alert.MetricType)
	data.MetricName = types.StringValue(alert.MetricName)
	data.Threshold = types.Float64Value(alert.Threshold)
	data.ComparisonOperator = types.StringValue(alert.ComparisonOperator)
	data.Enabled = types.BoolValue(alert.Enabled)
	data.Severity = types.Int64Value(int64(alert.Severity))
	data.State = types.StringValue(alert.State)
	data.CreatedAt = types.StringValue(alert.CreatedAt)

	if alert.Description != nil {
		data.Description = types.StringValue(*alert.Description)
	}
	if alert.EvaluationPeriods != nil {
		data.EvaluationPeriods = types.StringValue(*alert.EvaluationPeriods)
	}
	if alert.DatapointsToAlarm != nil {
		data.DatapointsToAlarm = types.StringValue(*alert.DatapointsToAlarm)
	}
	if alert.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*alert.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}
}
