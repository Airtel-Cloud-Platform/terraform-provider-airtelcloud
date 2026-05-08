package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &ProtectionPlanResource{}
var _ resource.ResourceWithImportState = &ProtectionPlanResource{}

func NewProtectionPlanResource() resource.Resource {
	return &ProtectionPlanResource{}
}

type ProtectionPlanResource struct {
	client *client.Client
}

type ProtectionPlanResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ScheduleType  types.String `tfsdk:"schedule_type"`
	SelectorKey   types.String `tfsdk:"selector_key"`
	SelectorValue types.String `tfsdk:"selector_value"`
	Retention     types.Int64  `tfsdk:"retention"`
	RetentionUnit types.String `tfsdk:"retention_unit"`
	Recurrence    types.Int64  `tfsdk:"recurrence"`
	SubnetID      types.String `tfsdk:"subnet_id"`
}

func (r *ProtectionPlanResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_plan"
}

func (r *ProtectionPlanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Protection Plan. Protection plans define backup schedules and retention policies. Note: the API does not support deletion of protection plans; destroying this resource will only remove it from Terraform state.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the protection plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the protection plan.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the protection plan.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule_type": schema.StringAttribute{
				MarkdownDescription: "The schedule type for the protection plan.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"selector_key": schema.StringAttribute{
				MarkdownDescription: "The selector key for matching resources (e.g., `AZ` for availability zone).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"selector_value": schema.StringAttribute{
				MarkdownDescription: "The selector value to match (e.g., `S1`, `S2` for availability zone names).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"retention": schema.Int64Attribute{
				MarkdownDescription: "The retention period value.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"retention_unit": schema.StringAttribute{
				MarkdownDescription: "The unit for the retention period. Must be uppercase: `DAYS`, `WEEKS`, or `MONTHS`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recurrence": schema.Int64Attribute{
				MarkdownDescription: "The recurrence interval in seconds (e.g., `86400` for daily, `604800` for weekly).",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "The subnet ID used for routing backup API requests. Required by the backup service.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ProtectionPlanResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ProtectionPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtectionPlanResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateProtectionPlanRequest{
		Name:          data.Name.ValueString(),
		Description:   data.Description.ValueString(),
		ScheduleType:  data.ScheduleType.ValueString(),
		SelectorKey:   data.SelectorKey.ValueString(),
		SelectorValue: data.SelectorValue.ValueString(),
		Retention:     int(data.Retention.ValueInt64()),
		RetentionUnit: data.RetentionUnit.ValueString(),
		Recurrence:    int(data.Recurrence.ValueInt64()),
	}

	subnetID := data.SubnetID.ValueString()

	plan, err := r.client.CreateProtectionPlan(ctx, createReq, subnetID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create protection plan, got error: %s", err))
		return
	}

	// The API returns a UUID id and a server-generated name
	data.ID = types.StringValue(plan.ID)
	// Keep the user-specified input values in state (the API does not echo them back)

	tflog.Trace(ctx, "created protection plan resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionPlanResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnetID := data.SubnetID.ValueString()

	plan, err := r.client.GetProtectionPlan(ctx, data.ID.ValueString(), subnetID)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read protection plan, got error: %s", err))
		return
	}

	data.ID = types.StringValue(plan.ID)
	// The list API only returns id, name, project_id, project_name, version, created_at.
	// Retain user-specified input values for attributes not returned by the API.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Protection plans cannot be updated. All changes require replacement.")
}

func (r *ProtectionPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// API does not support deletion of protection plans.
	// Remove from Terraform state only.
	tflog.Warn(ctx, "Protection plan deletion is not supported by the API. Removing from Terraform state only.")
}

func (r *ProtectionPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
