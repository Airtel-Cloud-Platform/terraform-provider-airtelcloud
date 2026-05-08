package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &VPCPeeringResource{}
var _ resource.ResourceWithImportState = &VPCPeeringResource{}
var _ resource.ResourceWithValidateConfig = &VPCPeeringResource{}

func NewVPCPeeringResource() resource.Resource {
	return &VPCPeeringResource{}
}

type VPCPeeringResource struct {
	client *client.Client
}

type VPCPeeringResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Description       types.String   `tfsdk:"description"`
	VPCSourceID       types.String   `tfsdk:"vpc_source_id"`
	VPCSourceName     types.String   `tfsdk:"vpc_source_name"`
	VPCTargetID       types.String   `tfsdk:"vpc_target_id"`
	VPCTargetName     types.String   `tfsdk:"vpc_target_name"`
	AZ                types.String   `tfsdk:"az"`
	Region            types.String   `tfsdk:"region"`
	IsPclEnabled      types.Bool     `tfsdk:"is_pcl_enabled"`
	AllowedSubnetList types.List     `tfsdk:"allowed_subnet_list"`
	BlockedSubnetList types.List     `tfsdk:"blocked_subnet_list"`
	State             types.String   `tfsdk:"state"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (r *VPCPeeringResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peering"
}

func (r *VPCPeeringResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud VPC Peering connection.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the VPC peering connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPC peering connection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the VPC peering connection.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_source_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the source VPC. Either vpc_source_id or vpc_source_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_source_name": schema.StringAttribute{
				MarkdownDescription: "The name of the source VPC. Either vpc_source_id or vpc_source_name must be specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_target_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the target VPC. Either vpc_target_id or vpc_target_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_target_name": schema.StringAttribute{
				MarkdownDescription: "The name of the target VPC. Either vpc_target_id or vpc_target_name must be specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"az": schema.StringAttribute{
				MarkdownDescription: "The availability zone for the VPC peering connection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region for the VPC peering connection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_pcl_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether PCL (Private Connectivity Link) is enabled.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"allowed_subnet_list": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of allowed subnet IDs for the peering connection.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"blocked_subnet_list": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of blocked subnet IDs for the peering connection.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the VPC peering connection.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *VPCPeeringResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *VPCPeeringResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPCPeeringResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve VPC source name to ID if needed
	sourceID := data.VPCSourceID.ValueString()
	if sourceID == "" && !data.VPCSourceName.IsNull() {
		resolved, err := r.client.ResolveVPCID(ctx, data.VPCSourceName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VPC Source Resolution Error", err.Error())
			return
		}
		sourceID = resolved
	}

	// Resolve VPC target name to ID if needed
	targetID := data.VPCTargetID.ValueString()
	if targetID == "" && !data.VPCTargetName.IsNull() {
		resolved, err := r.client.ResolveVPCID(ctx, data.VPCTargetName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VPC Target Resolution Error", err.Error())
			return
		}
		targetID = resolved
	}

	// Validate resolved IDs are not empty
	if sourceID == "" {
		resp.Diagnostics.AddError("VPC Source Resolution Error",
			"Resolved VPC source ID is empty. Provide a valid vpc_source_id or vpc_source_name.")
		return
	}
	if targetID == "" {
		resp.Diagnostics.AddError("VPC Target Resolution Error",
			"Resolved VPC target ID is empty. Provide a valid vpc_target_id or vpc_target_name.")
		return
	}

	tflog.Debug(ctx, "VPC peering Create: resolved VPC IDs", map[string]interface{}{
		"source_id": sourceID,
		"target_id": targetID,
	})

	allowedSubnets := make([]string, 0)
	if !data.AllowedSubnetList.IsNull() {
		resp.Diagnostics.Append(data.AllowedSubnetList.ElementsAs(ctx, &allowedSubnets, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	blockedSubnets := make([]string, 0)
	if !data.BlockedSubnetList.IsNull() {
		resp.Diagnostics.Append(data.BlockedSubnetList.ElementsAs(ctx, &blockedSubnets, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get create timeout (default: 5 minutes)
	createTimeout, diags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateVPCPeeringRequest{
		Name:              data.Name.ValueString(),
		Description:       data.Description.ValueString(),
		VPCSourceID:       sourceID,
		VPCTargetID:       targetID,
		AZ:                data.AZ.ValueString(),
		Region:            data.Region.ValueString(),
		PeerVpcRegion:     data.Region.ValueString(),
		IsPclEnabled:      data.IsPclEnabled.ValueBool(),
		AllowedSubnetList: allowedSubnets,
		BlockedSubnetList: blockedSubnets,
	}

	peering, err := r.client.CreateVPCPeeringWithTimeout(ctx, createReq, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VPC peering, got error: %s", err))
		return
	}

	data.ID = types.StringValue(peering.ID)
	data.VPCSourceID = types.StringValue(sourceID)
	data.VPCTargetID = types.StringValue(targetID)
	data.State = types.StringValue(peering.State)

	tflog.Trace(ctx, "created a VPC peering resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCPeeringResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPCPeeringResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	peering, err := r.client.GetVPCPeering(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VPC peering, got error: %s", err))
		return
	}

	data.Name = types.StringValue(peering.Name)
	data.VPCSourceID = types.StringValue(peering.VPCSourceID)
	data.VPCTargetID = types.StringValue(peering.VPCTargetID)
	data.AZ = types.StringValue(peering.AZ)
	data.Region = types.StringValue(peering.Region)
	data.IsPclEnabled = types.BoolValue(peering.IsPclEnabled)
	data.State = types.StringValue(peering.State)

	if peering.Description != "" {
		data.Description = types.StringValue(peering.Description)
	}

	if len(peering.AllowedSubnetList) > 0 {
		allowedList, diags := types.ListValueFrom(ctx, types.StringType, peering.AllowedSubnetList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AllowedSubnetList = allowedList
	}

	if len(peering.BlockedSubnetList) > 0 {
		blockedList, diags := types.ListValueFrom(ctx, types.StringType, peering.BlockedSubnetList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.BlockedSubnetList = blockedList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCPeeringResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields use RequiresReplace, so Terraform will destroy and recreate.
	// This method should never be called but is required by the interface.
	var data VPCPeeringResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCPeeringResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPCPeeringResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get delete timeout (default: 5 minutes)
	deleteTimeout, diags := data.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVPCPeeringWithTimeout(ctx, data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VPC peering, got error: %s", err))
		return
	}
}

func (r *VPCPeeringResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data VPCPeeringResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate VPC source: exactly one of vpc_source_id or vpc_source_name must be set
	if !data.VPCSourceID.IsNull() && !data.VPCSourceName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of vpc_source_id or vpc_source_name may be specified, not both.")
	}
	if data.VPCSourceID.IsNull() && data.VPCSourceName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of vpc_source_id or vpc_source_name must be specified.")
	}

	// Validate VPC target: exactly one of vpc_target_id or vpc_target_name must be set
	if !data.VPCTargetID.IsNull() && !data.VPCTargetName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of vpc_target_id or vpc_target_name may be specified, not both.")
	}
	if data.VPCTargetID.IsNull() && data.VPCTargetName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of vpc_target_id or vpc_target_name must be specified.")
	}

	// Skip validation if is_pcl_enabled is unknown (e.g., during plan with variables)
	if data.IsPclEnabled.IsUnknown() {
		return
	}

	if !data.IsPclEnabled.ValueBool() {
		if !data.AllowedSubnetList.IsNull() && len(data.AllowedSubnetList.Elements()) > 0 {
			resp.Diagnostics.AddError("Invalid Configuration",
				"allowed_subnet_list can only be set when is_pcl_enabled is true.")
		}
		if !data.BlockedSubnetList.IsNull() && len(data.BlockedSubnetList.Elements()) > 0 {
			resp.Diagnostics.AddError("Invalid Configuration",
				"blocked_subnet_list can only be set when is_pcl_enabled is true.")
		}
	}
}

func (r *VPCPeeringResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
