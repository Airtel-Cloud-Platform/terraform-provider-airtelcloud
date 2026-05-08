package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &PublicIPResource{}
var _ resource.ResourceWithImportState = &PublicIPResource{}

func NewPublicIPResource() resource.Resource {
	return &PublicIPResource{}
}

type PublicIPResource struct {
	client *client.Client
}

type PublicIPResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	ObjectName       types.String   `tfsdk:"object_name"`
	VIP              types.String   `tfsdk:"vip"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	PublicIP         types.String   `tfsdk:"public_ip"`
	Domain           types.String   `tfsdk:"domain"`
	Status           types.String   `tfsdk:"status"`
	AllocatedTime    types.String   `tfsdk:"allocated_time"`
	AZName           types.String   `tfsdk:"az_name"`
	Region           types.String   `tfsdk:"region"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func (r *PublicIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip"
}

func (r *PublicIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Public IP. Public IPs are allocated via NAT against a Virtual Machine or Load Balancer private IP and are availability zone specific.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the public IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_name": schema.StringAttribute{
				MarkdownDescription: "The name for the public IP allocation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vip": schema.StringAttribute{
				MarkdownDescription: "The target private IP (VM or Load Balancer IP) to NAT against.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone for the public IP (e.g., `S1`, `S2`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The allocated public IP address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain of the public IP.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the public IP.",
				Computed:            true,
			},
			"allocated_time": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the public IP was allocated.",
				Computed:            true,
			},
			"az_name": schema.StringAttribute{
				MarkdownDescription: "The availability zone name.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the public IP.",
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

func (r *PublicIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PublicIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PublicIPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreatePublicIPRequest{
		ObjectName: data.ObjectName.ValueString(),
		VIP:        data.VIP.ValueString(),
	}

	created, err := r.client.CreatePublicIP(ctx, createReq, data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create public IP, got error: %s", err))
		return
	}

	// Create response only returns uuid and public_ip; poll until status is "Created" to get full details
	uuid := created.UUID
	readyIP, err := r.client.WaitForPublicIPReady(ctx, uuid, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for public IP to be ready: %s", err))
		return
	}

	// Look up port_id from compute instances
	portID, err := r.client.FindPortIDByVIP(ctx, data.VIP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find port ID for VIP %s: %s", data.VIP.ValueString(), err))
		return
	}

	// Map the public IP with the internal VIP
	mapReq := &models.MapPublicIPRequest{
		TargetVIP: data.VIP.ValueString(),
		PublicIP:  readyIP.IP,
		UUID:      readyIP.UUID,
		PortID:    portID,
	}
	err = r.client.MapPublicIP(ctx, mapReq, data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to map public IP to VIP: %s", err))
		return
	}

	data.ID = types.StringValue(readyIP.UUID)
	data.PublicIP = types.StringValue(readyIP.IP)
	data.Status = types.StringValue(readyIP.Status)
	if readyIP.Domain != "" {
		data.Domain = types.StringValue(readyIP.Domain)
	}
	if readyIP.AllocatedTime != "" {
		data.AllocatedTime = types.StringValue(readyIP.AllocatedTime)
	}
	if readyIP.AZName != "" {
		data.AZName = types.StringValue(readyIP.AZName)
	}
	if readyIP.Region != "" {
		data.Region = types.StringValue(readyIP.Region)
	}

	tflog.Trace(ctx, "created public IP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PublicIPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicIP, err := r.client.GetPublicIP(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read public IP, got error: %s", err))
		return
	}

	data.ID = types.StringValue(publicIP.UUID)
	if publicIP.ObjectName != "" {
		data.ObjectName = types.StringValue(publicIP.ObjectName)
	}
	if publicIP.TargetVIP != "" {
		data.VIP = types.StringValue(publicIP.TargetVIP)
	}
	if publicIP.IP != "" {
		data.PublicIP = types.StringValue(publicIP.IP)
	}
	data.Status = types.StringValue(publicIP.Status)
	if publicIP.Domain != "" {
		data.Domain = types.StringValue(publicIP.Domain)
	}
	if publicIP.AllocatedTime != "" {
		data.AllocatedTime = types.StringValue(publicIP.AllocatedTime)
	}
	if publicIP.AZName != "" {
		data.AZName = types.StringValue(publicIP.AZName)
	}
	if publicIP.Region != "" {
		data.Region = types.StringValue(publicIP.Region)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Public IPs cannot be updated in place. All changes require replacement.")
}

func (r *PublicIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PublicIPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePublicIP(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete public IP, got error: %s", err))
		return
	}
}

func (r *PublicIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
