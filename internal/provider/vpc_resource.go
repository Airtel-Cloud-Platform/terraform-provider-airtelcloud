package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VPCResource{}
var _ resource.ResourceWithImportState = &VPCResource{}

func NewVPCResource() resource.Resource {
	return &VPCResource{}
}

// VPCResource defines the resource implementation.
type VPCResource struct {
	client *client.Client
}

// VPCResourceModel describes the resource data model.
type VPCResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	CIDRBlock          types.String `tfsdk:"cidr_block"`
	State              types.String `tfsdk:"state"`
	EnableDNSHostnames types.Bool   `tfsdk:"enable_dns_hostnames"`
	EnableDNSSupport   types.Bool   `tfsdk:"enable_dns_support"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	Tags               types.Map    `tfsdk:"tags"`
}

func (r *VPCResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Virtual Private Cloud (VPC).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPC.",
				Required:            true,
			},
			"cidr_block": schema.StringAttribute{
				MarkdownDescription: "The CIDR block for the VPC.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the VPC.",
				Computed:            true,
			},
			"enable_dns_hostnames": schema.BoolAttribute{
				MarkdownDescription: "Whether DNS hostnames are enabled for the VPC.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"enable_dns_support": schema.BoolAttribute{
				MarkdownDescription: "Whether DNS support is enabled for the VPC.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default VPC.",
				Computed:            true,
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of tags to assign to the VPC.",
				Optional:            true,
			},
		},
	}
}

func (r *VPCResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VPCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPCResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags map to []models.Tag
	var tags []models.Tag
	if !data.Tags.IsNull() {
		var tagsMap map[string]string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags = models.MapToTags(tagsMap)
	}

	createReq := &models.CreateVPCRequest{
		Name:               data.Name.ValueString(),
		CIDRBlock:          data.CIDRBlock.ValueString(),
		EnableDNSHostnames: data.EnableDNSHostnames.ValueBool(),
		EnableDNSSupport:   data.EnableDNSSupport.ValueBool(),
		Tags:               tags,
	}

	vpc, err := r.client.CreateVPC(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VPC, got error: %s", err))
		return
	}

	// Update the model with the created VPC data
	data.ID = types.StringValue(vpc.ID)
	data.State = types.StringValue(vpc.State)
	data.IsDefault = types.BoolValue(vpc.IsDefault)

	tflog.Trace(ctx, "created a VPC resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPCResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.GetVPC(ctx, data.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VPC, got error: %s", err))
		return
	}

	// Update the model with the current VPC data
	data.Name = types.StringValue(vpc.Name)
	data.CIDRBlock = types.StringValue(vpc.CIDRBlock)
	data.State = types.StringValue(vpc.State)
	data.EnableDNSHostnames = types.BoolValue(vpc.EnableDNSHostnames)
	data.EnableDNSSupport = types.BoolValue(vpc.EnableDNSSupport)
	data.IsDefault = types.BoolValue(vpc.IsDefault)

	// Convert tags to Terraform map
	if len(vpc.Tags) > 0 {
		tagsMap := models.TagsToMap(vpc.Tags)
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, tagsMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsValue
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VPCResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags map to []models.Tag
	var tags []models.Tag
	if !data.Tags.IsNull() {
		var tagsMap map[string]string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags = models.MapToTags(tagsMap)
	}

	enableDNSHostnames := data.EnableDNSHostnames.ValueBool()
	enableDNSSupport := data.EnableDNSSupport.ValueBool()

	updateReq := &models.UpdateVPCRequest{
		Name:               data.Name.ValueString(),
		EnableDNSHostnames: &enableDNSHostnames,
		EnableDNSSupport:   &enableDNSSupport,
		Tags:               tags,
	}

	vpc, err := r.client.UpdateVPC(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update VPC, got error: %s", err))
		return
	}

	// Update the model with the updated VPC data
	data.State = types.StringValue(vpc.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPCResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVPC(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VPC, got error: %s", err))
		return
	}
}

func (r *VPCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
