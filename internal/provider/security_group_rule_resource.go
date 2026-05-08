package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

var _ resource.Resource = &SecurityGroupRuleResource{}
var _ resource.ResourceWithImportState = &SecurityGroupRuleResource{}

func NewSecurityGroupRuleResource() resource.Resource {
	return &SecurityGroupRuleResource{}
}

type SecurityGroupRuleResource struct {
	client *client.Client
}

type SecurityGroupRuleResourceModel struct {
	ID                          types.Int64  `tfsdk:"id"`
	UUID                        types.String `tfsdk:"uuid"`
	SecurityGroupID             types.Int64  `tfsdk:"security_group_id"`
	SecurityGroupUUID           types.String `tfsdk:"security_group_uuid"`
	Direction                   types.String `tfsdk:"direction"`
	Protocol                    types.String `tfsdk:"protocol"`
	PortRangeMin                types.String `tfsdk:"port_range_min"`
	PortRangeMax                types.String `tfsdk:"port_range_max"`
	RemoteIPPrefix              types.String `tfsdk:"remote_ip_prefix"`
	RemoteGroupID               types.String `tfsdk:"remote_group_id"`
	Ethertype                   types.String `tfsdk:"ethertype"`
	Description                 types.String `tfsdk:"description"`
	Status                      types.String `tfsdk:"status"`
	ProviderSecurityGroupRuleID types.String `tfsdk:"provider_security_group_rule_id"`
}

func (r *SecurityGroupRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}

func (r *SecurityGroupRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud security group rule.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the security group rule.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the security group rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the security group this rule belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"security_group_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the security group this rule belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"direction": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The direction of the rule (ingress or egress).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The protocol (tcp, udp, icmp, etc.).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_range_min": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The minimum port number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_range_max": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The maximum port number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_ip_prefix": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The remote IP prefix (CIDR notation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The remote security group ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ethertype": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ethertype (IPv4 or IPv6).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the security group rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the rule.",
			},
			"provider_security_group_rule_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The provider-specific security group rule ID.",
			},
		},
	}
}

func (r *SecurityGroupRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgID := int(data.SecurityGroupID.ValueInt64())

	createReq := &models.CreateSecurityGroupRuleRequest{
		Direction:      data.Direction.ValueString(),
		Protocol:       data.Protocol.ValueString(),
		PortRangeMin:   data.PortRangeMin.ValueString(),
		PortRangeMax:   data.PortRangeMax.ValueString(),
		RemoteIPPrefix: data.RemoteIPPrefix.ValueString(),
		RemoteGroupID:  data.RemoteGroupID.ValueString(),
		Ethertype:      data.Ethertype.ValueString(),
		Description:    data.Description.ValueString(),
	}

	rule, err := r.client.CreateSecurityGroupRule(ctx, sgID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create security group rule, got error: %s", err))
		return
	}

	// Set computed fields from API response
	data.ID = types.Int64Value(int64(rule.ID))
	data.UUID = types.StringValue(rule.UUID)
	data.SecurityGroupUUID = types.StringValue("")
	data.Status = types.StringValue(rule.Status)
	data.ProviderSecurityGroupRuleID = types.StringValue(rule.ProviderSecurityGroupRuleID)

	// For Optional fields, preserve the planned value to avoid null-vs-empty-string mismatches.
	// Only override if the API returned a non-empty value different from plan.
	if rule.RemoteGroupID != "" {
		data.RemoteGroupID = types.StringValue(rule.RemoteGroupID)
	}

	tflog.Trace(ctx, "created a security group rule resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := int(data.ID.ValueInt64())
	sgID := int(data.SecurityGroupID.ValueInt64())

	rule, err := r.client.GetSecurityGroupRule(ctx, sgID, ruleID)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security group rule, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(int64(rule.ID))
	data.UUID = types.StringValue(rule.UUID)
	data.SecurityGroupUUID = types.StringValue("")
	data.Direction = types.StringValue(rule.Direction)
	data.Protocol = types.StringValue(rule.Protocol)
	data.PortRangeMin = types.StringValue(rule.PortRangeMin)
	data.PortRangeMax = types.StringValue(rule.PortRangeMax)
	data.RemoteIPPrefix = types.StringValue(rule.RemoteIPPrefix)
	if rule.RemoteGroupID != "" {
		data.RemoteGroupID = types.StringValue(rule.RemoteGroupID)
	}
	data.Ethertype = types.StringValue(rule.Ethertype)
	data.Description = types.StringValue(rule.Description)
	data.Status = types.StringValue(rule.Status)
	data.ProviderSecurityGroupRuleID = types.StringValue(rule.ProviderSecurityGroupRuleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update API — RequiresReplace on all user-facing attributes ensures this is never called
	resp.Diagnostics.AddError("Update Not Supported", "Security group rules cannot be updated. All changes require replacement.")
}

func (r *SecurityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := int(data.ID.ValueInt64())
	sgID := int(data.SecurityGroupID.ValueInt64())

	err := r.client.DeleteSecurityGroupRule(ctx, sgID, ruleID)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete security group rule, got error: %s", err))
		return
	}
}

func (r *SecurityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: security_group_id/rule_id
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: security_group_id/rule_id. Got: %q", req.ID),
		)
		return
	}

	sgID, err := strconv.ParseInt(idParts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected numeric security_group_id, got: %q", idParts[0]),
		)
		return
	}

	ruleID, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected numeric rule_id, got: %q", idParts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_group_id"), sgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ruleID)...)
}
