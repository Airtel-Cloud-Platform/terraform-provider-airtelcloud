package provider

import (
	"context"
	"errors"
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
	RuleName         types.String `tfsdk:"rule_name"`
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
	CreateNew  types.Bool                         `tfsdk:"create_new"`
	IPCIDR     types.String                       `tfsdk:"ip_cidr"`
	SourceType types.String                       `tfsdk:"source_type"`
	Geographic *PublicIPPolicyRuleGeographicModel `tfsdk:"geographic"`
}

type PublicIPPolicyRuleGeographicModel struct {
	CountryCode types.String `tfsdk:"country_code"`
	CountryName types.String `tfsdk:"country_name"`
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
		MarkdownDescription: "Manages a policy rule on an Airtel Cloud Public IP (NAT Gateway). The parent public IP must already be attached to a VM, load balancer, or baremetal server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the policy rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the parent public IP, resolved from `public_ip_name`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip_name": schema.StringAttribute{
				MarkdownDescription: "The object name of the parent public IP resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_name": schema.StringAttribute{
				MarkdownDescription: "The name of the policy rule. Sent to the API as `rule_name`.",
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
							MarkdownDescription: "Whether to create a new source selector. Defaults to `false` when omitted.",
							Optional:            true,
						},
						"ip_cidr": schema.StringAttribute{
							MarkdownDescription: "Source CIDR, for example `182.77.78.18/32`.",
							Optional:            true,
						},
						"source_type": schema.StringAttribute{
							MarkdownDescription: "Source type: `ip_cidr`, `geographic`, or `all`.",
							Optional:            true,
						},
						"geographic": schema.SingleNestedAttribute{
							MarkdownDescription: "Country selector when `source_type` is `geographic`.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"country_code": schema.StringAttribute{
									MarkdownDescription: "ISO country code, for example `IN`.",
									Optional:            true,
								},
								"country_name": schema.StringAttribute{
									MarkdownDescription: "Country name, for example `India`.",
									Optional:            true,
								},
							},
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
				MarkdownDescription: "The target private IP. Read from the parent public IP when omitted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The public IP address. Read from the parent public IP when omitted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone (e.g., `S1`, `S2`). Read from the parent public IP when omitted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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

	if !data.PublicIPName.IsNull() && !data.PublicIPName.IsUnknown() && strings.TrimSpace(data.PublicIPName.ValueString()) == "" {
		resp.Diagnostics.AddError("Invalid Configuration",
			"public_ip_name must be set.")
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

	// Resolve public_ip_name -> UUID, then refuse policy create unless attached.
	resolved, err := r.client.ResolvePublicIPID(ctx, data.PublicIPName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve public IP name %q: %s", data.PublicIPName.ValueString(), err))
		return
	}
	pip, err := r.client.RequirePublicIPAttached(ctx, resolved)
	if err != nil {
		status := ""
		if pip != nil {
			status = pip.Status
		}
		if errors.Is(err, client.ErrPublicIPNotAttached) {
			resp.Diagnostics.AddError(
				"Public IP Not Attached",
				fmt.Sprintf("Cannot add a policy on public IP %q until it is attached to a VM, load balancer, or baremetal server (current status %q). Create airtelcloud_public_ip_attachment first, then add the policy.", data.PublicIPName.ValueString(), status),
			)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to verify public IP %q is attached: %s", data.PublicIPName.ValueString(), err))
		return
	}
	data.PublicIPID = types.StringValue(pip.UUID)

	// target_vip, public_ip, and availability_zone default to the parent public IP.
	if data.TargetVIP.IsNull() || data.TargetVIP.IsUnknown() || data.TargetVIP.ValueString() == "" {
		data.TargetVIP = types.StringValue(pip.TargetVIP)
	}
	if data.PublicIP.IsNull() || data.PublicIP.IsUnknown() || data.PublicIP.ValueString() == "" {
		data.PublicIP = types.StringValue(getPublicIPAddr(pip))
	}
	if data.AvailabilityZone.IsNull() || data.AvailabilityZone.IsUnknown() || data.AvailabilityZone.ValueString() == "" {
		data.AvailabilityZone = types.StringValue(getPublicIPAZName(pip))
	}

	if data.TargetVIP.ValueString() == "" || data.PublicIP.ValueString() == "" || data.AvailabilityZone.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Incomplete Public IP Details",
			fmt.Sprintf("Public IP %q did not report target_vip, public_ip, and availability zone. Set them explicitly on the policy rule.", data.PublicIPName.ValueString()),
		)
		return
	}

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
		if item.Geographic != nil {
			geo := &models.PublicIPPolicyRuleGeographicInput{}
			if !item.Geographic.CountryCode.IsNull() && !item.Geographic.CountryCode.IsUnknown() {
				geo.CountryCode = item.Geographic.CountryCode.ValueString()
			}
			if !item.Geographic.CountryName.IsNull() && !item.Geographic.CountryName.IsUnknown() {
				geo.CountryName = item.Geographic.CountryName.ValueString()
			}
			if geo.CountryCode != "" || geo.CountryName != "" {
				entry.Geographic = geo
			}
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
		case "geographic":
			if first.Geographic != nil && first.Geographic.CountryCode != "" {
				sourceValue = first.Geographic.CountryCode
			}
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
		DisplayName:   data.RuleName.ValueString(),
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
		data.RuleName = types.StringValue(readyRule.DisplayName)
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

	if rule.DisplayName != "" {
		data.RuleName = types.StringValue(rule.DisplayName)
	}
	if rule.Action != "" {
		data.Action = types.StringValue(rule.Action)
	}
	data.State = types.StringValue(rule.State)

	// source/services are only refreshed when the rule is managed through the
	// flat attributes. When source_config/service_config drive the payload, the
	// API shape does not round-trip into them and refreshing would show a
	// permanent diff against a null configuration value.
	usesSourceConfig := !data.SourceConfig.IsNull() && len(data.SourceConfig.Elements()) > 0
	if !usesSourceConfig && rule.SourceIP != "" {
		data.Source = types.StringValue(rule.SourceIP)
	}

	usesServiceConfig := !data.ServiceConfig.IsNull() && len(data.ServiceConfig.Elements()) > 0
	if !usesServiceConfig && len(rule.Services) > 0 {
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
