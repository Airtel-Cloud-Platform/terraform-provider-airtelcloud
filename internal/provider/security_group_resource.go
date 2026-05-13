package provider

import (
	"context"
	"fmt"
	"strconv"

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

var _ resource.Resource = &SecurityGroupResource{}
var _ resource.ResourceWithImportState = &SecurityGroupResource{}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

type SecurityGroupResource struct {
	client *client.Client
}

type SecurityGroupResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	UUID              types.String `tfsdk:"uuid"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	AvailabilityZone  types.String `tfsdk:"availability_zone"`
	Status            types.String `tfsdk:"status"`
	AZName            types.String `tfsdk:"az_name"`
	AZRegion          types.String `tfsdk:"az_region"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud security group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the security group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the security group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the security group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The availability zone in which to create the security group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the security group.",
			},
			"az_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The availability zone name.",
			},
			"az_region": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The availability zone region.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update timestamp.",
			},
		},
	}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupResource) sgClient(az types.String) *client.Client {
	if !az.IsNull() && az.ValueString() != "" {
		return r.client.WithAvailabilityZone(az.ValueString())
	}
	return r.client
}

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateSecurityGroupRequest{
		SecurityGroupName: data.SecurityGroupName.ValueString(),
	}

	sgClient := r.sgClient(data.AvailabilityZone)

	sg, err := sgClient.CreateSecurityGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create security group, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(int64(sg.ID))
	data.UUID = types.StringValue(sg.UUID)
	data.SecurityGroupName = types.StringValue(sg.SecurityGroupName)
	data.AvailabilityZone = types.StringValue(sg.AZName)
	data.Status = types.StringValue(sg.Status)
	data.AZName = types.StringValue(sg.AZName)
	data.AZRegion = types.StringValue(sg.AZRegion)
	data.CreatedAt = types.StringValue(sg.Created)
	data.UpdatedAt = types.StringValue(sg.Updated)

	tflog.Trace(ctx, "created a security group resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgClient := r.sgClient(data.AvailabilityZone)

	sg, err := sgClient.GetSecurityGroup(ctx, int(data.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security group, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(int64(sg.ID))
	data.UUID = types.StringValue(sg.UUID)
	data.SecurityGroupName = types.StringValue(sg.SecurityGroupName)
	data.AvailabilityZone = types.StringValue(sg.AZName)
	data.Status = types.StringValue(sg.Status)
	data.AZName = types.StringValue(sg.AZName)
	data.AZRegion = types.StringValue(sg.AZRegion)
	data.CreatedAt = types.StringValue(sg.Created)
	data.UpdatedAt = types.StringValue(sg.Updated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update API — RequiresReplace on all user-facing attributes ensures this is never called
	resp.Diagnostics.AddError("Update Not Supported", "Security groups cannot be updated. All changes require replacement.")
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgClient := r.sgClient(data.AvailabilityZone)

	err := sgClient.DeleteSecurityGroup(ctx, int(data.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete security group, got error: %s", err))
		return
	}
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected a numeric ID, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
