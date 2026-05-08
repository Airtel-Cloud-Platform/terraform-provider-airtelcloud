package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

// SubnetResource defines the resource implementation.
type SubnetResource struct {
	client *client.Client
}

// SubnetResourceModel describes the resource data model.
type SubnetResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	NetworkID        types.String   `tfsdk:"network_id"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	IPv4AddressSpace types.String   `tfsdk:"ipv4_address_space"`
	SubnetSubRole    types.String   `tfsdk:"subnet_sub_role"`
	Region           types.String   `tfsdk:"region"`
	Labels           types.List     `tfsdk:"labels"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	// Computed fields from API response
	State            types.String `tfsdk:"state"`
	IPv6AddressSpace types.String `tfsdk:"ipv6_address_space"`
	CreatedBy        types.String `tfsdk:"created_by"`
	CreateTime       types.String `tfsdk:"create_time"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud VPC subnet.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network (VPC) to create the subnet in.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the subnet.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the subnet.",
				Optional:            true,
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone for the subnet (e.g., S2).",
				Optional:            true,
				Computed:            true,
			},
			"ipv4_address_space": schema.StringAttribute{
				MarkdownDescription: "The IPv4 CIDR block for the subnet (e.g., 10.10.26.0/24).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_sub_role": schema.StringAttribute{
				MarkdownDescription: "The sub-role of the subnet. Valid values: Private, VIP.",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the subnet (e.g., south).",
				Optional:            true,
				Computed:            true,
			},
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of labels for the subnet.",
				Optional:            true,
			},
			// Computed fields from API response
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the subnet.",
				Computed:            true,
			},
			"ipv6_address_space": schema.StringAttribute{
				MarkdownDescription: "The IPv6 address space.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user who created the subnet.",
				Computed:            true,
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "The creation time of the subnet.",
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

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get create timeout (default: 10 minutes)
	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert labels list to string slice
	var labels []string
	if !data.Labels.IsNull() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &models.CreateSubnetRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
		IPv4AddressSpace: data.IPv4AddressSpace.ValueString(),
		SubnetSubRole:    data.SubnetSubRole.ValueString(),
		Labels:           labels,
	}

	networkID := data.NetworkID.ValueString()

	subnet, err := r.client.CreateSubnetWithTimeout(ctx, networkID, createReq, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subnet, got error: %s", err))
		return
	}

	// Update the model with the created subnet data
	mapSubnetToModel(ctx, subnet, &data, &resp.Diagnostics)

	tflog.Trace(ctx, "created a subnet resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	networkID := data.NetworkID.ValueString()
	subnetID := data.ID.ValueString()

	subnet, err := r.client.GetSubnet(ctx, networkID, subnetID)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subnet, got error: %s", err))
		return
	}

	// Update the model with the current subnet data
	mapSubnetToModel(ctx, subnet, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &models.UpdateSubnetRequest{
		Description: data.Description.ValueString(),
	}

	networkID := data.NetworkID.ValueString()
	subnetID := data.ID.ValueString()

	subnet, err := r.client.UpdateSubnet(ctx, networkID, subnetID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update subnet, got error: %s", err))
		return
	}

	// Update the model with the updated subnet data
	mapSubnetToModel(ctx, subnet, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get delete timeout (default: 10 minutes)
	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkID := data.NetworkID.ValueString()
	subnetID := data.ID.ValueString()

	err := r.client.DeleteSubnetWithTimeout(ctx, networkID, subnetID, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete subnet, got error: %s", err))
		return
	}
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: network_id/subnet_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: network_id/subnet_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// mapSubnetToModel maps the API subnet response to the Terraform model
func mapSubnetToModel(ctx context.Context, subnet *models.Subnet, data *SubnetResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(subnet.SubnetID)
	data.NetworkID = types.StringValue(subnet.NetworkID)
	data.Name = types.StringValue(subnet.Name)
	if subnet.Description != "" {
		data.Description = types.StringValue(subnet.Description)
	} else if data.Description.IsNull() {
		data.Description = types.StringNull()
	}
	data.AvailabilityZone = types.StringValue(subnet.AvailabilityZone)
	data.IPv4AddressSpace = types.StringValue(subnet.IPv4AddressSpace)
	data.SubnetSubRole = types.StringValue(subnet.SubnetSubRole)
	data.Region = types.StringValue(subnet.Region)
	// Computed fields
	data.State = types.StringValue(subnet.State)
	data.IPv6AddressSpace = types.StringValue(subnet.IPv6AddressSpace)
	data.CreatedBy = types.StringValue(subnet.CreatedBy)
	data.CreateTime = types.StringValue(subnet.CreateTime)

	// Convert Labels slice to Terraform list
	if len(subnet.Labels) > 0 {
		labelsValue, d := types.ListValueFrom(ctx, types.StringType, subnet.Labels)
		diags.Append(d...)
		data.Labels = labelsValue
	} else {
		data.Labels = types.ListNull(types.StringType)
	}
}
