package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
)

var _ resource.Resource = &LBVipResource{}
var _ resource.ResourceWithImportState = &LBVipResource{}

func NewLBVipResource() resource.Resource {
	return &LBVipResource{}
}

type LBVipResource struct {
	client *client.Client
}

type LBVipResourceModel struct {
	ID             types.String `tfsdk:"id"`
	LBServiceID    types.String `tfsdk:"lb_service_id"`
	Name           types.String `tfsdk:"name"`
	Status         types.String `tfsdk:"status"`
	FixedIPs       types.String `tfsdk:"fixed_ips"`
	PublicIP       types.String `tfsdk:"public_ip"`
	ProviderPortID types.String `tfsdk:"provider_port_id"`
}

func (r *LBVipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_vip"
}

func (r *LBVipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VIP (Virtual IP) port for an Airtel Cloud Load Balancer Service.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the VIP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lb_service_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the parent LB service.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VIP.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the VIP.",
				Computed:            true,
			},
			"fixed_ips": schema.StringAttribute{
				MarkdownDescription: "The fixed IP addresses assigned to the VIP.",
				Computed:            true,
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The public IP address of the VIP.",
				Computed:            true,
			},
			"provider_port_id": schema.StringAttribute{
				MarkdownDescription: "The provider port ID.",
				Computed:            true,
			},
		},
	}
}

func (r *LBVipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LBVipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LBVipResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up the LB service to get the network_id for the subnet-id header
	lbService, err := r.client.GetLBService(ctx, data.LBServiceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB service: %s", err))
		return
	}
	scopedClient := r.client.WithSubnetID(lbService.NetworkID)

	vip, err := scopedClient.CreateLBVip(ctx, data.LBServiceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create LB VIP, got error: %s", err))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(vip.ID))
	data.Name = types.StringValue(vip.Name)
	data.Status = types.StringValue(vip.Status)
	data.FixedIPs = types.StringValue(strings.Join(vip.FixedIPs, ","))
	data.PublicIP = types.StringValue(vip.PublicIP)
	data.ProviderPortID = types.StringValue(vip.ProviderPortID)

	tflog.Trace(ctx, "created LB VIP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LBVipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vipID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse VIP ID: %s", err))
		return
	}

	// Look up the LB service to get the network_id for the subnet-id header
	lbService, err := r.client.GetLBService(ctx, data.LBServiceID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB service: %s", err))
		return
	}
	scopedClient := r.client.WithSubnetID(lbService.NetworkID)

	// List all VIPs and find by ID (no get-by-ID endpoint)
	vips, err := scopedClient.ListLBVips(ctx, data.LBServiceID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB VIPs, got error: %s", err))
		return
	}

	var found bool
	for _, vip := range vips {
		if vip.ID == vipID {
			data.Name = types.StringValue(vip.Name)
			data.Status = types.StringValue(vip.Status)
			data.FixedIPs = types.StringValue(strings.Join(vip.FixedIPs, ","))
			data.PublicIP = types.StringValue(vip.PublicIP)
			data.ProviderPortID = types.StringValue(vip.ProviderPortID)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "LB VIPs cannot be updated. All changes require replacement.")
}

func (r *LBVipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LBVipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vipID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse VIP ID: %s", err))
		return
	}

	// Look up the LB service to get the network_id for the subnet-id header
	lbService, err := r.client.GetLBService(ctx, data.LBServiceID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			// LB service already gone, VIP is implicitly deleted
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB service: %s", err))
		return
	}
	scopedClient := r.client.WithSubnetID(lbService.NetworkID)

	err = scopedClient.DeleteLBVip(ctx, data.LBServiceID.ValueString(), vipID)
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete LB VIP, got error: %s", err))
		return
	}
}

func (r *LBVipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: lb_service_id/vip_id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: lb_service_id/vip_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lb_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
