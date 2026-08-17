package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

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

var _ resource.Resource = &PublicIPPolicyRuleResource{}
var _ resource.ResourceWithImportState = &PublicIPPolicyRuleResource{}
var _ resource.ResourceWithValidateConfig = &PublicIPPolicyRuleResource{}

func NewPublicIPPolicyRuleResource() resource.Resource {
	return &PublicIPPolicyRuleResource{}
}

type PublicIPPolicyRuleResource struct {
	client *client.Client
}

type PublicIPPolicyRuleResourceModel struct {
	ID               types.String `tfsdk:"id"`
	PublicIPID       types.String `tfsdk:"public_ip_id"`
	PublicIPName     types.String `tfsdk:"public_ip_name"`
	DisplayName      types.String `tfsdk:"display_name"`
	Source           types.String `tfsdk:"source"`
	SourceConfig     types.List   `tfsdk:"source_config"`
	Services         types.List   `tfsdk:"services"`
	ServiceConfig    types.List   `tfsdk:"service_config"`
	Action           types.String `tfsdk:"action"`
	ResourceType     types.String `tfsdk:"resource_type"`
	RevisionNote     types.String `tfsdk:"revision_note"`
	TargetVIP        types.String `tfsdk:"target_vip"`
	PublicIP         types.String `tfsdk:"public_ip"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	State            types.String `tfsdk:"state"`
}

type PublicIPPolicyRuleSourceConfigModel struct {
	CreateNew  types.Bool   `tfsdk:"create_new"`
	IPCIDR     types.String `tfsdk:"ip_cidr"`
	SourceType types.String `tfsdk:"source_type"`
}

type PublicIPPolicyRuleServiceConfigModel struct {
	CreateNew types.Bool   `tfsdk:"create_new"`
	Name      types.String `tfsdk:"name"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

func (r *PublicIPPolicyRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip_policy_rule"
}

func (r *PublicIPPolicyRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a policy rule on an Airtel Cloud Public IP (NAT Gateway). Policy rules control traffic allowed or denied through the public IP.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the policy rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the parent public IP resource. Either public_ip_id or public_ip_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ip_name": schema.StringAttribute{
				MarkdownDescription: "The name (object_name) of the parent public IP resource. If set, it is resolved to public_ip_id. Either public_ip_id or public_ip_name must be specified.",
				Optional:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the policy rule.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source IP address/CIDR or `any` for all sources. Optional when `source_config` is used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_config": schema.ListNestedAttribute{
				MarkdownDescription: "Detailed source entries sent to the source-of-truth API. When provided, this takes precedence over `source`.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"create_new": schema.BoolAttribute{
							MarkdownDescription: "Whether to create/use a new source selector. Defaults to `true` when omitted.",
							Optional:            true,
						},
						"ip_cidr": schema.StringAttribute{
							MarkdownDescription: "Source CIDR, for example `1.2.1.0/24`.",
							Optional:            true,
						},
						"source_type": schema.StringAttribute{
							MarkdownDescription: "Source type such as `ip_cidr` or `all`.",
							Optional:            true,
						},
					},
				},
			},
			"services": schema.ListAttribute{
				MarkdownDescription: "List of service names to allow/deny (e.g., `HTTP`, `HTTPS`, `SSH`). Optional when `service_config` is used.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.List{
					// List doesn't have RequiresReplace in the same way,
					// but since there's no update API, changes require replacement
				},
			},
			"service_config": schema.ListNestedAttribute{
				MarkdownDescription: "Detailed service entries sent to the source-of-truth API. When provided, this takes precedence over `services`.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"create_new": schema.BoolAttribute{
							MarkdownDescription: "Whether to create/use a new service selector. Defaults to `false` when omitted.",
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Service name, for example `RDP`.",
							Required:            true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: "Whether the service is marked as default. Defaults to `false` when omitted.",
							Optional:            true,
						},
					},
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action to take: `accept` or `deny`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Resource type sent to policy API. Defaults to `ipam`.",
				Optional:            true,
			},
			"revision_note": schema.StringAttribute{
				MarkdownDescription: "Revision note sent to policy API. Defaults to `creating Policy`.",
				Optional:            true,
			},
			"target_vip": schema.StringAttribute{
				MarkdownDescription: "The target private IP (from the parent public IP resource).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The public IP address (from the parent public IP resource).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone (e.g., `S1`, `S2`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the policy rule.",
				Computed:            true,
			},
		},
	}
}

func (r *PublicIPPolicyRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces that exactly one of public_ip_id or public_ip_name is set.
func (r *PublicIPPolicyRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data PublicIPPolicyRuleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.PublicIPID.IsNull() && !data.PublicIPName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of public_ip_id or public_ip_name may be specified, not both.")
	}
	if data.PublicIPID.IsNull() && data.PublicIPName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of public_ip_id or public_ip_name must be specified.")
	}

	hasSource := !data.Source.IsNull() && strings.TrimSpace(data.Source.ValueString()) != ""
	hasSourceConfig := !data.SourceConfig.IsNull() && !data.SourceConfig.IsUnknown() && len(data.SourceConfig.Elements()) > 0
	if !hasSource && !hasSourceConfig {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of source or source_config must be specified.")
	}

	hasServices := !data.Services.IsNull() && !data.Services.IsUnknown() && len(data.Services.Elements()) > 0
	hasServiceConfig := !data.ServiceConfig.IsNull() && !data.ServiceConfig.IsUnknown() && len(data.ServiceConfig.Elements()) > 0
	if !hasServices && !hasServiceConfig {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of services or service_config must be specified.")
	}
}

func (r *PublicIPPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PublicIPPolicyRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve public_ip_name -> public_ip_id (UUID) and persist into the Computed attribute.
	publicIPID := data.PublicIPID.ValueString()
	if publicIPID == "" && !data.PublicIPName.IsNull() {
		resolved, err := r.client.ResolvePublicIPID(ctx, data.PublicIPName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Public IP Resolution Error", err.Error())
			return
		}
		publicIPID = resolved
	}
	data.PublicIPID = types.StringValue(publicIPID)

	// Get service names from the plan
	var serviceNames []string
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &serviceNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var sourceConfigPlan []PublicIPPolicyRuleSourceConfigModel
	if !data.SourceConfig.IsNull() && !data.SourceConfig.IsUnknown() {
		resp.Diagnostics.Append(data.SourceConfig.ElementsAs(ctx, &sourceConfigPlan, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var serviceConfigPlan []PublicIPPolicyRuleServiceConfigModel
	if !data.ServiceConfig.IsNull() && !data.ServiceConfig.IsUnknown() {
		resp.Diagnostics.Append(data.ServiceConfig.ElementsAs(ctx, &serviceConfigPlan, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	sourceConfig := make([]models.PublicIPPolicyRuleSourceInput, 0, len(sourceConfigPlan))
	for _, item := range sourceConfigPlan {
		entry := models.PublicIPPolicyRuleSourceInput{}
		if !item.CreateNew.IsNull() && !item.CreateNew.IsUnknown() {
			v := item.CreateNew.ValueBool()
			entry.CreateNew = &v
		}
		if !item.IPCIDR.IsNull() && !item.IPCIDR.IsUnknown() {
			entry.IPCIDR = item.IPCIDR.ValueString()
		}
		if !item.SourceType.IsNull() && !item.SourceType.IsUnknown() {
			entry.SourceType = item.SourceType.ValueString()
		}
		sourceConfig = append(sourceConfig, entry)
	}

	serviceConfig := make([]models.PublicIPPolicyRuleServiceInput, 0, len(serviceConfigPlan))
	for _, item := range serviceConfigPlan {
		entry := models.PublicIPPolicyRuleServiceInput{}
		if !item.CreateNew.IsNull() && !item.CreateNew.IsUnknown() {
			v := item.CreateNew.ValueBool()
			entry.CreateNew = &v
		}
		if !item.Name.IsNull() && !item.Name.IsUnknown() {
			entry.Name = item.Name.ValueString()
		}
		if !item.IsDefault.IsNull() && !item.IsDefault.IsUnknown() {
			v := item.IsDefault.ValueBool()
			entry.IsDefault = &v
		}
		serviceConfig = append(serviceConfig, entry)
	}

	sourceValue := strings.TrimSpace(data.Source.ValueString())
	if sourceValue == "" && len(sourceConfig) > 0 {
		first := sourceConfig[0]
		sourceType := strings.ToLower(strings.TrimSpace(first.SourceType))
		switch sourceType {
		case "all", "any":
			sourceValue = "any"
		default:
			sourceValue = strings.TrimSpace(first.IPCIDR)
		}
	}

	if len(serviceNames) == 0 && len(serviceConfig) > 0 {
		for _, s := range serviceConfig {
			name := strings.TrimSpace(s.Name)
			if name == "" {
				continue
			}
			serviceNames = append(serviceNames, name)
		}
	}

	az := data.AvailabilityZone.ValueString()

	createReq := &models.CreatePublicIPPolicyRuleRequest{
		DisplayName:   data.DisplayName.ValueString(),
		Source:        sourceValue,
		SourceConfig:  sourceConfig,
		ServiceList:   serviceNames,
		ServiceConfig: serviceConfig,
		Action:        data.Action.ValueString(),
		ResourceType:  strings.TrimSpace(data.ResourceType.ValueString()),
		RevisionNote:  strings.TrimSpace(data.RevisionNote.ValueString()),
		TargetVIP:     data.TargetVIP.ValueString(),
		PublicIP:      data.PublicIP.ValueString(),
		UUID:          data.PublicIPID.ValueString(),
	}

	createdPolicyID, err := r.client.CreatePublicIPPolicyRule(ctx, createReq, az)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create policy rule, got error: %s", err))
		return
	}

	policyWaitTimeout := 10 * time.Minute
	readyRule, err := r.client.WaitForPublicIPPolicyRuleReady(ctx, data.PublicIPID.ValueString(), data.TargetVIP.ValueString(), data.PublicIP.ValueString(), createdPolicyID, policyWaitTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for public IP policy rule to be ready: %s", err))
		return
	}

	data.ID = types.StringValue(createdPolicyID)
	if readyRule != nil && readyRule.State != "" {
		data.State = types.StringValue(readyRule.State)
	}
	if readyRule != nil && readyRule.DisplayName != "" {
		data.DisplayName = types.StringValue(readyRule.DisplayName)
	}

	tflog.Trace(ctx, "created public IP policy rule resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PublicIPPolicyRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetPublicIPPolicyRule(ctx,
		data.PublicIPID.ValueString(),
		data.TargetVIP.ValueString(),
		data.PublicIP.ValueString(),
		data.ID.ValueString(),
	)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read policy rule, got error: %s", err))
		return
	}

	data.DisplayName = types.StringValue(rule.DisplayName)
	if rule.SourceIP != "" {
		data.Source = types.StringValue(rule.SourceIP)
	}
	data.Action = types.StringValue(rule.Action)
	data.State = types.StringValue(rule.State)

	// Update services from the API response
	if len(rule.Services) > 0 {
		servicesList, diags := types.ListValueFrom(ctx, types.StringType, rule.Services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Services = servicesList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Public IP policy rules cannot be updated in place. All changes require replacement.")
}

func (r *PublicIPPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PublicIPPolicyRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting public IP policy rule resource", map[string]interface{}{
		"public_ip_id": data.PublicIPID.ValueString(),
		"policy_uuid":  data.ID.ValueString(),
	})
	err := r.client.DeletePublicIPPolicyRuleWithWait(
		ctx,
		data.PublicIPID.ValueString(),
		data.TargetVIP.ValueString(),
		data.PublicIP.ValueString(),
		data.ID.ValueString(),
		7*time.Minute,
	)
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete policy rule, got error: %s", err))
		return
	}
}

func (r *PublicIPPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: public_ip_id/target_vip/public_ip/rule_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: public_ip_id/target_vip/public_ip/rule_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("public_ip_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_vip"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("public_ip"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[3])...)
}
