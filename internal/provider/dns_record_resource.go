package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client *client.Client
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	ZoneID      types.String  `tfsdk:"zone_id"`
	ZoneName    types.String  `tfsdk:"zone_name"`
	Owner       types.String  `tfsdk:"owner"`
	Data        types.String  `tfsdk:"data"`
	RecordType  types.String  `tfsdk:"record_type"`
	TTL         types.Int64   `tfsdk:"ttl"`
	Description types.String  `tfsdk:"description"`
	Preference  types.Int64   `tfsdk:"preference"`
	OrgName     types.String  `tfsdk:"org_name"`
	OrgID       types.String  `tfsdk:"org_id"`
	CreatedBy   types.String  `tfsdk:"created_by"`
	CreatedAt   types.Float64 `tfsdk:"created_at"`
	UpdatedAt   types.Float64 `tfsdk:"updated_at"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud DNS Record (DNSaaS).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the DNS record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the DNS zone this record belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_name": schema.StringAttribute{
				MarkdownDescription: "The name of the DNS zone this record belongs to.",
				Computed:            true,
			},
			"owner": schema.StringAttribute{
				MarkdownDescription: "The owner/name of the DNS record (e.g., 'www', '@' for apex).",
				Optional:            true,
				Computed:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "The data/value of the DNS record (e.g., IP address for A record, hostname for CNAME).",
				Optional:            true,
				Computed:            true,
			},
			"record_type": schema.StringAttribute{
				MarkdownDescription: "The type of DNS record. Valid values: 'A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV', 'CAA', 'PTR', etc.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The Time-To-Live (TTL) in seconds for the DNS record.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description for the DNS record.",
				Optional:            true,
			},
			"preference": schema.Int64Attribute{
				MarkdownDescription: "The preference/priority value for MX records.",
				Optional:            true,
			},
			"org_name": schema.StringAttribute{
				MarkdownDescription: "The organization name associated with the record.",
				Computed:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID associated with the record.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The user who created the record.",
				Computed:            true,
			},
			"created_at": schema.Float64Attribute{
				MarkdownDescription: "The timestamp when the record was created (Unix timestamp).",
				Computed:            true,
			},
			"updated_at": schema.Float64Attribute{
				MarkdownDescription: "The timestamp when the record was last updated (Unix timestamp).",
				Computed:            true,
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if a matching DNS record already exists
	existingRecords, err := r.client.ListDNSRecords(ctx, data.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking for existing DNS record",
			fmt.Sprintf("Unable to list DNS records to check for duplicates: %s", err),
		)
		return
	}
	planOwner := ""
	if !data.Owner.IsNull() && !data.Owner.IsUnknown() {
		planOwner = data.Owner.ValueString()
	}
	planData := ""
	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		planData = data.Data.ValueString()
	}
	planRecordType := data.RecordType.ValueString()
	for _, record := range existingRecords.Items {
		existingOwner := normalizeOwner(record.Owner, record.ZoneName)
		if existingOwner == planOwner && record.RecordType == planRecordType && record.Data == planData {
			resp.Diagnostics.AddError(
				"DNS Record already exists",
				fmt.Sprintf(
					"A DNS record with owner %q, type %q, and data %q already exists in zone %s (UUID: %s). To manage this resource with Terraform, import it using:\n\n  terraform import airtelcloud_dns_record.<resource_name> %s/%s",
					existingOwner, record.RecordType, record.Data, data.ZoneID.ValueString(), record.UUID,
					data.ZoneID.ValueString(), record.UUID,
				),
			)
			return
		}
	}

	// Build the create request
	createReq := &models.CreateDNSRecordRequest{
		RecordType: data.RecordType.ValueString(),
	}

	if !data.Owner.IsNull() && !data.Owner.IsUnknown() {
		owner := data.Owner.ValueString()
		createReq.Owner = &owner
	}

	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		dataVal := data.Data.ValueString()
		createReq.Data = &dataVal
	}

	if !data.TTL.IsNull() && !data.TTL.IsUnknown() {
		ttl := int(data.TTL.ValueInt64())
		createReq.TTL = &ttl
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		createReq.Description = &desc
	}

	if !data.Preference.IsNull() && !data.Preference.IsUnknown() {
		pref := int(data.Preference.ValueInt64())
		createReq.Preference = &pref
	}

	record, err := r.client.CreateDNSRecord(ctx, data.ZoneID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNS record, got error: %s", err))
		return
	}

	// Update the model with the created record data
	data.ID = types.StringValue(record.UUID)
	data.ZoneName = types.StringValue(record.ZoneName)
	data.Owner = types.StringValue(normalizeOwner(record.Owner, record.ZoneName))
	data.Data = types.StringValue(record.Data)
	data.RecordType = types.StringValue(record.RecordType)
	data.TTL = types.Int64Value(int64(record.TTL))
	data.OrgName = types.StringValue(record.OrgName)
	data.OrgID = types.StringValue(record.OrgID)
	data.CreatedBy = types.StringValue(record.CreatedBy)
	data.CreatedAt = types.Float64Value(record.CreatedAt)
	data.UpdatedAt = types.Float64Value(record.UpdatedAt)

	if record.Description != nil {
		data.Description = types.StringValue(*record.Description)
	} else {
		data.Description = types.StringNull()
	}

	if record.Preference != nil {
		data.Preference = types.Int64Value(int64(*record.Preference))
	} else {
		data.Preference = types.Int64Null()
	}

	tflog.Trace(ctx, "created a DNS record resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.GetDNSRecord(ctx, data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS record, got error: %s", err))
		return
	}

	// Update the model with the current record data
	data.ID = types.StringValue(record.UUID)
	data.ZoneID = types.StringValue(record.ZoneID)
	data.ZoneName = types.StringValue(record.ZoneName)
	data.Owner = types.StringValue(normalizeOwner(record.Owner, record.ZoneName))
	data.Data = types.StringValue(record.Data)
	data.RecordType = types.StringValue(record.RecordType)
	data.TTL = types.Int64Value(int64(record.TTL))
	data.OrgName = types.StringValue(record.OrgName)
	data.OrgID = types.StringValue(record.OrgID)
	data.CreatedBy = types.StringValue(record.CreatedBy)
	data.CreatedAt = types.Float64Value(record.CreatedAt)
	data.UpdatedAt = types.Float64Value(record.UpdatedAt)

	if record.Description != nil {
		data.Description = types.StringValue(*record.Description)
	} else {
		data.Description = types.StringNull()
	}

	if record.Preference != nil {
		data.Preference = types.Int64Value(int64(*record.Preference))
	} else {
		data.Preference = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the update request
	updateReq := &models.UpdateDNSRecordRequest{
		RecordType: data.RecordType.ValueString(),
	}

	if !data.Owner.IsNull() && !data.Owner.IsUnknown() {
		owner := data.Owner.ValueString()
		updateReq.Owner = &owner
	}

	if !data.Data.IsNull() && !data.Data.IsUnknown() {
		dataVal := data.Data.ValueString()
		updateReq.Data = &dataVal
	}

	if !data.TTL.IsNull() && !data.TTL.IsUnknown() {
		ttl := int(data.TTL.ValueInt64())
		updateReq.TTL = &ttl
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		updateReq.Description = &desc
	}

	if !data.Preference.IsNull() && !data.Preference.IsUnknown() {
		pref := int(data.Preference.ValueInt64())
		updateReq.Preference = &pref
	}

	record, err := r.client.UpdateDNSRecord(ctx, data.ZoneID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNS record, got error: %s", err))
		return
	}

	// Update the model with the updated record data
	data.ZoneName = types.StringValue(record.ZoneName)
	data.Owner = types.StringValue(normalizeOwner(record.Owner, record.ZoneName))
	data.Data = types.StringValue(record.Data)
	data.TTL = types.Int64Value(int64(record.TTL))
	data.OrgName = types.StringValue(record.OrgName)
	data.OrgID = types.StringValue(record.OrgID)
	data.CreatedBy = types.StringValue(record.CreatedBy)
	data.CreatedAt = types.Float64Value(record.CreatedAt)
	data.UpdatedAt = types.Float64Value(record.UpdatedAt)

	if record.Description != nil {
		data.Description = types.StringValue(*record.Description)
	} else {
		data.Description = types.StringNull()
	}

	if record.Preference != nil {
		data.Preference = types.Int64Value(int64(*record.Preference))
	} else {
		data.Preference = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSRecord(ctx, data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS record, got error: %s", err))
		return
	}
}

func normalizeOwner(owner, zoneName string) string {
	zoneName = strings.TrimSuffix(zoneName, ".")
	owner = strings.TrimSuffix(owner, ".")
	owner = strings.TrimSuffix(owner, "."+zoneName)
	return owner
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone_id/record_id
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: zone_id/record_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
