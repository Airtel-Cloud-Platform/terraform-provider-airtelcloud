package provider

import (
	"context"
	"fmt"

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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSZoneResource{}
var _ resource.ResourceWithImportState = &DNSZoneResource{}

func NewDNSZoneResource() resource.Resource {
	return &DNSZoneResource{}
}

// DNSZoneResource defines the resource implementation.
type DNSZoneResource struct {
	client *client.Client
}

// DNSZoneResourceModel describes the resource data model.
type DNSZoneResourceModel struct {
	ID              types.String  `tfsdk:"id"`
	ZoneName        types.String  `tfsdk:"zone_name"`
	ZoneType        types.String  `tfsdk:"zone_type"`
	Description     types.String  `tfsdk:"description"`
	DNSZoneTemplate types.String  `tfsdk:"dns_zone_template"`
	OrgName         types.String  `tfsdk:"org_name"`
	OrgID           types.String  `tfsdk:"org_id"`
	CreatedBy       types.String  `tfsdk:"created_by"`
	CreatedAt       types.Float64 `tfsdk:"created_at"`
	UpdatedAt       types.Float64 `tfsdk:"updated_at"`
}

func (r *DNSZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (r *DNSZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud DNS Zone (DNSaaS).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the DNS zone.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_name": schema.StringAttribute{
				MarkdownDescription: "The name of the DNS zone (e.g., 'example.com').",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_type": schema.StringAttribute{
				MarkdownDescription: "The type of DNS zone. Valid values: 'forward', 'reverse'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("forward"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description for the DNS zone.",
				Optional:            true,
			},
			"dns_zone_template": schema.StringAttribute{
				MarkdownDescription: "The DNS zone template.",
				Computed:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The organization name associated with the zone.",
				Computed:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID associated with the zone.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user who created the zone.",
				Computed:            true,
			},
			"created_at": schema.Float64Attribute{
				MarkdownDescription: "The timestamp when the zone was created (Unix timestamp).",
				Computed:            true,
			},
			"updated_at": schema.Float64Attribute{
				MarkdownDescription: "The timestamp when the zone was last updated (Unix timestamp).",
				Computed:            true,
			},
		},
	}
}

func (r *DNSZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if a DNS zone with the same name already exists
	existingZones, err := r.client.ListDNSZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking for existing DNS zone",
			fmt.Sprintf("Unable to list DNS zones to check for duplicates: %s", err),
		)
		return
	}
	for _, zone := range existingZones.Items {
		if zone.ZoneName == data.ZoneName.ValueString() {
			resp.Diagnostics.AddError(
				"DNS Zone already exists",
				fmt.Sprintf(
					"A DNS zone with name %q already exists (UUID: %s). To manage this resource with Terraform, import it using:\n\n  terraform import airtelcloud_dns_zone.<resource_name> %s",
					zone.ZoneName, zone.UUID, zone.UUID,
				),
			)
			return
		}
	}

	// Build the create request
	createReq := &models.CreateDNSZoneRequest{
		ZoneName: data.ZoneName.ValueString(),
		ZoneType: data.ZoneType.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		createReq.Description = &desc
	}

	zone, err := r.client.CreateDNSZone(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNS zone, got error: %s", err))
		return
	}

	// Update the model with the created zone data
	data.ID = types.StringValue(zone.UUID)
	data.ZoneName = types.StringValue(zone.ZoneName)
	data.ZoneType = types.StringValue(zone.ZoneType)
	data.DNSZoneTemplate = types.StringValue(zone.DNSZoneTemplate)
	data.OrgName = types.StringValue(zone.OrgName)
	data.OrgID = types.StringValue(zone.OrgID)
	data.CreatedBy = types.StringValue(zone.CreatedBy)
	data.CreatedAt = types.Float64Value(zone.CreatedAt)
	data.UpdatedAt = types.Float64Value(zone.UpdatedAt)

	if zone.Description != nil {
		data.Description = types.StringValue(*zone.Description)
	} else {
		data.Description = types.StringNull()
	}

	tflog.Trace(ctx, "created a DNS zone resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetDNSZone(ctx, data.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS zone, got error: %s", err))
		return
	}

	// Update the model with the current zone data
	data.ID = types.StringValue(zone.UUID)
	data.ZoneName = types.StringValue(zone.ZoneName)
	data.ZoneType = types.StringValue(zone.ZoneType)
	data.DNSZoneTemplate = types.StringValue(zone.DNSZoneTemplate)
	data.OrgName = types.StringValue(zone.OrgName)
	data.OrgID = types.StringValue(zone.OrgID)
	data.CreatedBy = types.StringValue(zone.CreatedBy)
	data.CreatedAt = types.Float64Value(zone.CreatedAt)
	data.UpdatedAt = types.Float64Value(zone.UpdatedAt)

	if zone.Description != nil {
		data.Description = types.StringValue(*zone.Description)
	} else {
		data.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSZoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the update request (only description can be updated)
	updateReq := &models.UpdateDNSZoneRequest{}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		updateReq.Description = &desc
	}

	zone, err := r.client.UpdateDNSZone(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNS zone, got error: %s", err))
		return
	}

	// Update the model with the updated zone data
	data.DNSZoneTemplate = types.StringValue(zone.DNSZoneTemplate)
	data.OrgName = types.StringValue(zone.OrgName)
	data.OrgID = types.StringValue(zone.OrgID)
	data.CreatedBy = types.StringValue(zone.CreatedBy)
	data.CreatedAt = types.Float64Value(zone.CreatedAt)
	data.UpdatedAt = types.Float64Value(zone.UpdatedAt)

	if zone.Description != nil {
		data.Description = types.StringValue(*zone.Description)
	} else {
		data.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSZone(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS zone, got error: %s", err))
		return
	}
}

func (r *DNSZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
