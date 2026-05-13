package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &LBServiceResource{}
var _ resource.ResourceWithImportState = &LBServiceResource{}

func NewLBServiceResource() resource.Resource {
	return &LBServiceResource{}
}

type LBServiceResource struct {
	client *client.Client
}

type LBServiceResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Description     types.String   `tfsdk:"description"`
	FlavorID        types.Int64    `tfsdk:"flavor_id"`
	NetworkID       types.String   `tfsdk:"network_id"`
	VPCID           types.String   `tfsdk:"vpc_id"`
	VPCName         types.String   `tfsdk:"vpc_name"`
	HA              types.Bool     `tfsdk:"ha"`
	Status          types.String   `tfsdk:"status"`
	OperatingStatus types.String   `tfsdk:"operating_status"`
	AZName          types.String   `tfsdk:"az_name"`
	Created         types.String   `tfsdk:"created"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (r *LBServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_service"
}

func (r *LBServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Load Balancer Service.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the LB service.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the load balancer service.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the load balancer service.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flavor_id": schema.Int64Attribute{
				MarkdownDescription: "The LB flavor ID, automatically resolved by the provider.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The network (subnet) ID for the load balancer.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The VPC ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "The VPC name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ha": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable high availability. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the LB service.",
				Computed:            true,
			},
			"operating_status": schema.StringAttribute{
				MarkdownDescription: "The operating status of the LB service.",
				Computed:            true,
			},
			"az_name": schema.StringAttribute{
				MarkdownDescription: "The availability zone name.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp.",
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

func (r *LBServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LBServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LBServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch LB flavor automatically using network_id as subnet-id header
	scopedClient := r.client.WithSubnetID(data.NetworkID.ValueString())
	lbFlavors, err := scopedClient.ListLBFlavors(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list LB flavors: %s", err))
		return
	}
	if len(lbFlavors) == 0 {
		resp.Diagnostics.AddError("Configuration Error", "No LB flavors available for the specified network")
		return
	}
	flavorID := lbFlavors[0].ID

	createReq := &models.CreateLBServiceRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		FlavorID:    flavorID,
		NetworkID:   data.NetworkID.ValueString(),
		VPCID:       data.VPCID.ValueString(),
		VPCName:     data.VPCName.ValueString(),
		HA:          data.HA.ValueBool(),
	}

	lbService, err := scopedClient.CreateLBService(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create LB service, got error: %s", err))
		return
	}

	// Wait for LB service to become ready (async creation returns 202)
	readyService, err := r.client.WaitForLBServiceReady(ctx, lbService.ID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for LB service to be ready: %s", err))
		return
	}

	data.ID = types.StringValue(readyService.ID)
	data.FlavorID = types.Int64Value(int64(flavorID))
	data.Status = types.StringValue(readyService.Status)
	data.OperatingStatus = types.StringValue(readyService.OperatingStatus)
	if readyService.AZName != "" {
		data.AZName = types.StringValue(readyService.AZName)
	}
	if readyService.Created != "" {
		data.Created = types.StringValue(readyService.Created)
	}

	tflog.Trace(ctx, "created LB service resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LBServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbService, err := r.client.GetLBService(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB service, got error: %s", err))
		return
	}

	if lbService.Name != "" {
		data.Name = types.StringValue(lbService.Name)
	}
	if lbService.Description != "" {
		data.Description = types.StringValue(lbService.Description)
	}
	data.FlavorID = types.Int64Value(int64(lbService.FlavorID))
	if lbService.NetworkID != "" {
		data.NetworkID = types.StringValue(lbService.NetworkID)
	}
	if lbService.VPCID != "" {
		data.VPCID = types.StringValue(lbService.VPCID)
	}
	if lbService.VPCName != "" {
		data.VPCName = types.StringValue(lbService.VPCName)
	}
	data.Status = types.StringValue(lbService.Status)
	data.OperatingStatus = types.StringValue(lbService.OperatingStatus)
	if lbService.AZName != "" {
		data.AZName = types.StringValue(lbService.AZName)
	}
	if lbService.Created != "" {
		data.Created = types.StringValue(lbService.Created)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update API — all fields use RequiresReplace
	resp.Diagnostics.AddError("Update Not Supported", "LB services cannot be updated in place. All changes require replacement.")
}

func (r *LBServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LBServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLBService(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete LB service, got error: %s", err))
		return
	}

	err = r.client.WaitForLBServiceDeleted(ctx, data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for LB service deletion: %s", err))
		return
	}
}

func (r *LBServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
