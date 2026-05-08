package provider

import (
	"context"
	"fmt"
	"strings"

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

func NewPublicIPPolicyRuleResource() resource.Resource {
	return &PublicIPPolicyRuleResource{}
}

type PublicIPPolicyRuleResource struct {
	client *client.Client
}

type PublicIPPolicyRuleResourceModel struct {
	ID               types.String `tfsdk:"id"`
	PublicIPID       types.String `tfsdk:"public_ip_id"`
	DisplayName      types.String `tfsdk:"display_name"`
	Source           types.String `tfsdk:"source"`
	Services         types.List   `tfsdk:"services"`
	Action           types.String `tfsdk:"action"`
	TargetVIP        types.String `tfsdk:"target_vip"`
	PublicIP         types.String `tfsdk:"public_ip"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	State            types.String `tfsdk:"state"`
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
				MarkdownDescription: "The UUID of the parent public IP resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the policy rule.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source IP address or `any` for all sources.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"services": schema.ListAttribute{
				MarkdownDescription: "List of service names to allow/deny (e.g., `HTTP`, `HTTPS`, `SSH`). Available services can be queried from the IPAM service API.",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.List{
					// List doesn't have RequiresReplace in the same way,
					// but since there's no update API, changes require replacement
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action to take: `accept` or `deny`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *PublicIPPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PublicIPPolicyRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get service names from the plan
	var serviceNames []string
	resp.Diagnostics.Append(data.Services.ElementsAs(ctx, &serviceNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	az := data.AvailabilityZone.ValueString()

	// Resolve service names to UUIDs
	availableServices, err := r.client.ListIPAMServices(ctx, az)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list IPAM services: %s", err))
		return
	}

	serviceUUIDs, err := resolveServiceNamesToUUIDs(serviceNames, availableServices)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	createReq := &models.CreatePublicIPPolicyRuleRequest{
		DisplayName: data.DisplayName.ValueString(),
		Source:      data.Source.ValueString(),
		ServiceList: serviceUUIDs,
		Action:      data.Action.ValueString(),
		TargetVIP:   data.TargetVIP.ValueString(),
		PublicIP:    data.PublicIP.ValueString(),
		UUID:        data.PublicIPID.ValueString(),
	}

	err = r.client.CreatePublicIPPolicyRule(ctx, createReq, az)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create policy rule, got error: %s", err))
		return
	}

	// Create response doesn't return rule ID — list rules to find the newly created one
	rulesResp, err := r.client.ListPublicIPPolicyRules(ctx,
		data.PublicIPID.ValueString(),
		data.TargetVIP.ValueString(),
		data.PublicIP.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list policy rules after creation: %s", err))
		return
	}

	// Find the rule by display_name
	var createdRule *models.PublicIPPolicyRule
	for _, rule := range rulesResp.Items {
		if rule.DisplayName == data.DisplayName.ValueString() {
			createdRule = &rule
			break
		}
	}

	if createdRule == nil {
		resp.Diagnostics.AddError("Client Error", "Policy rule created but not found in list response")
		return
	}

	data.ID = types.StringValue(createdRule.UUID)
	data.State = types.StringValue(createdRule.State)

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

	err := r.client.DeletePublicIPPolicyRule(ctx, data.ID.ValueString())
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

// resolveServiceNamesToUUIDs maps service names to their UUIDs
func resolveServiceNamesToUUIDs(names []string, available []models.IPAMService) ([]string, error) {
	nameToUUID := make(map[string]string, len(available))
	for _, svc := range available {
		nameToUUID[strings.ToUpper(svc.Name)] = svc.UUID
	}

	uuids := make([]string, 0, len(names))
	var missing []string
	for _, name := range names {
		uuid, ok := nameToUUID[strings.ToUpper(name)]
		if !ok {
			missing = append(missing, name)
			continue
		}
		uuids = append(uuids, uuid)
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("unknown services: %s", strings.Join(missing, ", "))
	}

	return uuids, nil
}
