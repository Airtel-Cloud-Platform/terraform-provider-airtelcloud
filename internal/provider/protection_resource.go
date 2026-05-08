package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &ProtectionResource{}
var _ resource.ResourceWithImportState = &ProtectionResource{}

func NewProtectionResource() resource.Resource {
	return &ProtectionResource{}
}

type ProtectionResource struct {
	client *client.Client
}

type ProtectionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	PolicyTypeID    types.String `tfsdk:"policy_type_id"`
	ComputeID       types.String `tfsdk:"compute_id"`
	ProtectionPlan  types.String `tfsdk:"protection_plan"`
	EnableScheduler types.String `tfsdk:"enable_scheduler"`
	StartDate       types.String `tfsdk:"start_date"`
	EndDate         types.String `tfsdk:"end_date"`
	StartTime       types.String `tfsdk:"start_time"`
	Status          types.String `tfsdk:"status"`
	Region          types.String `tfsdk:"region"`
	AZName          types.String `tfsdk:"az_name"`
	Created         types.String `tfsdk:"created"`
}

func (r *ProtectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection"
}

func (r *ProtectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Veritas Backup Protection policy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the protection policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the protection policy.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the protection policy.",
				Optional:            true,
			},
			"policy_type_id": schema.StringAttribute{
				MarkdownDescription: "The policy type ID.",
				Optional:            true,
			},
			"compute_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the compute instance to protect.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protection_plan": schema.StringAttribute{
				MarkdownDescription: "The protection plan to associate with this policy.",
				Required:            true,
			},
			"enable_scheduler": schema.StringAttribute{
				MarkdownDescription: "Whether to enable the backup scheduler. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("true"),
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date for the protection schedule.",
				Optional:            true,
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The end date for the protection schedule.",
				Optional:            true,
			},
			"start_time": schema.StringAttribute{
				MarkdownDescription: "The start time for the protection schedule.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the protection policy.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the protection policy.",
				Computed:            true,
			},
			"az_name": schema.StringAttribute{
				MarkdownDescription: "The availability zone of the protection policy.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *ProtectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProtectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateProtectionRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		PolicyTypeID:    data.PolicyTypeID.ValueString(),
		ComputeID:       data.ComputeID.ValueString(),
		ProtectionPlan:  data.ProtectionPlan.ValueString(),
		EnableScheduler: data.EnableScheduler.ValueString(),
		StartDate:       data.StartDate.ValueString(),
		EndDate:         data.EndDate.ValueString(),
		StartTime:       data.StartTime.ValueString(),
	}

	protection, err := r.client.CreateProtection(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create protection, got error: %s", err))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(protection.ID))
	data.Name = types.StringValue(protection.Name)
	data.Status = types.StringValue(protection.Status)
	if protection.Region != "" {
		data.Region = types.StringValue(protection.Region)
	}
	if protection.AZName != "" {
		data.AZName = types.StringValue(protection.AZName)
	}
	if protection.Created != "" {
		data.Created = types.StringValue(protection.Created)
	}

	tflog.Trace(ctx, "created protection resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	protection, err := r.client.GetProtection(ctx, id)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read protection, got error: %s", err))
		return
	}

	data.Name = types.StringValue(protection.Name)
	if protection.Description != "" {
		data.Description = types.StringValue(protection.Description)
	}
	if protection.ComputeID != "" {
		data.ComputeID = types.StringValue(protection.ComputeID)
	}
	if protection.ProtectionPlan != "" {
		data.ProtectionPlan = types.StringValue(protection.ProtectionPlan)
	}
	data.Status = types.StringValue(protection.Status)
	if protection.Region != "" {
		data.Region = types.StringValue(protection.Region)
	}
	if protection.AZName != "" {
		data.AZName = types.StringValue(protection.AZName)
	}
	if protection.Created != "" {
		data.Created = types.StringValue(protection.Created)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	updateReq := &models.UpdateProtectionRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		PolicyTypeID:    data.PolicyTypeID.ValueString(),
		ProtectionPlan:  data.ProtectionPlan.ValueString(),
		EnableScheduler: data.EnableScheduler.ValueString(),
		StartDate:       data.StartDate.ValueString(),
		EndDate:         data.EndDate.ValueString(),
		StartTime:       data.StartTime.ValueString(),
	}

	protection, err := r.client.UpdateProtection(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update protection, got error: %s", err))
		return
	}

	data.Name = types.StringValue(protection.Name)
	data.Status = types.StringValue(protection.Status)
	if protection.Region != "" {
		data.Region = types.StringValue(protection.Region)
	}
	if protection.AZName != "" {
		data.AZName = types.StringValue(protection.AZName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	err = r.client.DeleteProtection(ctx, id)
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete protection, got error: %s", err))
		return
	}
}

func (r *ProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
