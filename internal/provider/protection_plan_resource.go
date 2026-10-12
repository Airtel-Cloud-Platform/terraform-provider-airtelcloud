package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const (
	protectionPlanSelectorKey   = "AZ"
	protectionPlanRetentionUnit = "DAYS"
	secondsPerDay               = 86400
	protectionPlanRecurrenceMin = 1
	protectionPlanRecurrenceMax = 30
	protectionPlanRetentionMin  = 1
	protectionPlanRetentionMax  = 365
)

var _ resource.Resource = &ProtectionPlanResource{}
var _ resource.ResourceWithImportState = &ProtectionPlanResource{}
var _ resource.ResourceWithValidateConfig = &ProtectionPlanResource{}

func NewProtectionPlanResource() resource.Resource {
	return &ProtectionPlanResource{}
}

type ProtectionPlanResource struct {
	client *client.Client
}

type ProtectionPlanResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Recurrence       types.Int64  `tfsdk:"recurrence"`
	RecurrencePeriod types.String `tfsdk:"recurrence_period"`
	Retention        types.Int64  `tfsdk:"retention"`
	VMName           types.String `tfsdk:"vm_name"`
	SelectorValue    types.String `tfsdk:"selector_value"`
	SubnetID         types.String `tfsdk:"subnet_id"`
}

func protectionPlanAPIName(name, period string) string {
	n := strings.TrimSpace(name)
	p := strings.ToLower(strings.TrimSpace(period))
	if n == "" || p == "" {
		return n
	}
	suffix := "-" + p
	if strings.HasSuffix(strings.ToLower(n), suffix) {
		return n
	}
	return n + suffix
}

func protectionPlanRecurrenceSeconds(n int64, period string) int {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "weekly":
		return int(n) * 7 * secondsPerDay
	default:
		return int(n) * secondsPerDay
	}
}

func protectionPlanOutOfRange(got, min, max int64) bool {
	return got < min || got > max
}

func protectionPlanRangeError(attr string, got, min, max int64, unit string) string {
	suffix := ""
	if unit != "" {
		suffix = " " + unit
	}
	return fmt.Sprintf("%s must be between %d and %d%s (got %d). Fix the value; the backup API is not called.", attr, min, max, suffix, got)
}

type protectionPlanIntRangeValidator struct {
	min, max int64
	attr     string
	unit     string
}

func (v protectionPlanIntRangeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("%s must be between %d and %d", v.attr, v.min, v.max)
}

func (v protectionPlanIntRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v protectionPlanIntRangeValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	got := req.ConfigValue.ValueInt64()
	if protectionPlanOutOfRange(got, v.min, v.max) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid "+v.attr, protectionPlanRangeError(v.attr, got, v.min, v.max, v.unit))
	}
}

func protectionPlanVMLocation(vm *models.Compute) (az, subnet string) {
	if vm == nil {
		return "", ""
	}
	az = strings.TrimSpace(vm.AZName)
	if az == "" {
		az = strings.TrimSpace(vm.AvailabilityZone)
	}
	subnet = strings.TrimSpace(vm.NetworkID)
	if subnet == "" {
		subnet = strings.TrimSpace(vm.SubnetID)
	}
	return az, subnet
}

func (r *ProtectionPlanResource) lookupProtectionPlanVM(ctx context.Context, vmName string) (az, subnet string, err error) {
	vm, err := r.client.ResolveComputeNode(ctx, "", strings.TrimSpace(vmName))
	if err != nil {
		return "", "", err
	}
	az, subnet = protectionPlanVMLocation(vm)
	if az == "" {
		return "", "", fmt.Errorf("VM %q has no availability zone (az_name)", vmName)
	}
	if subnet == "" {
		return "", "", fmt.Errorf("VM %q has no subnet (network_id)", vmName)
	}
	return az, subnet, nil
}

func (r *ProtectionPlanResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection_plan"
}

func (r *ProtectionPlanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Protection Plan (backup schedule and retention). The API does not support deletion; destroying this resource only removes it from Terraform state. Set `vm_name`; AZ (`selector_value`) and subnet-id header are read from that VM. `selector_key` is always `AZ` and `retention_unit` is always `DAYS`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the protection plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Plan name. The API is sent `{name}-daily` or `{name}-weekly` from `recurrence_period`. Do not include that suffix unless you want it unchanged.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the protection plan. Sent as form `description`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recurrence": schema.Int64Attribute{
				MarkdownDescription: "How often the plan runs, as a count of `recurrence_period` units. Must be 1–30. Converted to seconds for the API (`daily` × 86400, `weekly` × 604800).",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					protectionPlanIntRangeValidator{
						min:  protectionPlanRecurrenceMin,
						max:  protectionPlanRecurrenceMax,
						attr: "recurrence",
					},
				},
			},
			"recurrence_period": schema.StringAttribute{
				MarkdownDescription: "Unit for `recurrence`. One of `daily` or `weekly`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("daily", "weekly"),
				},
			},
			"retention": schema.Int64Attribute{
				MarkdownDescription: "Retention in days. Must be 1–365. Sent with `retention_unit=DAYS`.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					protectionPlanIntRangeValidator{
						min:  protectionPlanRetentionMin,
						max:  protectionPlanRetentionMax,
						attr: "retention",
						unit: "days",
					},
				},
			},
			"vm_name": schema.StringAttribute{
				MarkdownDescription: "Existing VM `instance_name`. Its AZ becomes `selector_value` and its subnet is sent as the `subnet-id` header.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"selector_value": schema.StringAttribute{
				MarkdownDescription: "Availability zone taken from the named VM. `selector_key` is always `AZ`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Subnet ID taken from the named VM (`network_id`). Used as the API `subnet-id` header. Not set by the user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProtectionPlanResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProtectionPlanResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ProtectionPlanResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Recurrence.IsNull() && !data.Recurrence.IsUnknown() {
		n := data.Recurrence.ValueInt64()
		if protectionPlanOutOfRange(n, protectionPlanRecurrenceMin, protectionPlanRecurrenceMax) {
			resp.Diagnostics.AddAttributeError(path.Root("recurrence"), "Invalid recurrence",
				protectionPlanRangeError("recurrence", n, protectionPlanRecurrenceMin, protectionPlanRecurrenceMax, ""))
		}
	}
	if !data.Retention.IsNull() && !data.Retention.IsUnknown() {
		n := data.Retention.ValueInt64()
		if protectionPlanOutOfRange(n, protectionPlanRetentionMin, protectionPlanRetentionMax) {
			resp.Diagnostics.AddAttributeError(path.Root("retention"), "Invalid retention",
				protectionPlanRangeError("retention", n, protectionPlanRetentionMin, protectionPlanRetentionMax, "days"))
		}
	}
	if !data.RecurrencePeriod.IsNull() && !data.RecurrencePeriod.IsUnknown() {
		p := strings.ToLower(strings.TrimSpace(data.RecurrencePeriod.ValueString()))
		if p != "daily" && p != "weekly" {
			resp.Diagnostics.AddAttributeError(path.Root("recurrence_period"), "Invalid recurrence_period",
				"recurrence_period must be \"daily\" or \"weekly\".")
		}
	}
}

func (r *ProtectionPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtectionPlanResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if protectionPlanOutOfRange(data.Recurrence.ValueInt64(), protectionPlanRecurrenceMin, protectionPlanRecurrenceMax) {
		resp.Diagnostics.AddAttributeError(path.Root("recurrence"), "Invalid recurrence",
			protectionPlanRangeError("recurrence", data.Recurrence.ValueInt64(), protectionPlanRecurrenceMin, protectionPlanRecurrenceMax, ""))
		return
	}
	if protectionPlanOutOfRange(data.Retention.ValueInt64(), protectionPlanRetentionMin, protectionPlanRetentionMax) {
		resp.Diagnostics.AddAttributeError(path.Root("retention"), "Invalid retention",
			protectionPlanRangeError("retention", data.Retention.ValueInt64(), protectionPlanRetentionMin, protectionPlanRetentionMax, "days"))
		return
	}

	az, subnetID, err := r.lookupProtectionPlanVM(ctx, data.VMName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("VM Lookup Error", fmt.Sprintf("Unable to resolve vm_name %q: %s", data.VMName.ValueString(), err))
		return
	}
	data.SelectorValue = types.StringValue(az)
	data.SubnetID = types.StringValue(subnetID)

	period := strings.ToLower(strings.TrimSpace(data.RecurrencePeriod.ValueString()))
	createReq := &models.CreateProtectionPlanRequest{
		Name:          protectionPlanAPIName(data.Name.ValueString(), period),
		Description:   data.Description.ValueString(),
		SelectorKey:   protectionPlanSelectorKey,
		SelectorValue: az,
		Retention:     int(data.Retention.ValueInt64()),
		RetentionUnit: protectionPlanRetentionUnit,
		Recurrence:    protectionPlanRecurrenceSeconds(data.Recurrence.ValueInt64(), period),
	}

	plan, err := r.client.CreateProtectionPlan(ctx, createReq, subnetID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create protection plan, got error: %s", err))
		return
	}

	data.ID = types.StringValue(plan.ID)

	tflog.Trace(ctx, "created protection plan resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionPlanResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	az := data.SelectorValue.ValueString()
	if az == "" && data.VMName.ValueString() != "" {
		var err error
		az, _, err = r.lookupProtectionPlanVM(ctx, data.VMName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VM Lookup Error", fmt.Sprintf("Unable to resolve vm_name %q: %s", data.VMName.ValueString(), err))
			return
		}
		data.SelectorValue = types.StringValue(az)
	}

	plan, err := r.client.GetProtectionPlan(ctx, data.ID.ValueString(), az)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read protection plan, got error: %s", err))
		return
	}

	data.ID = types.StringValue(plan.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Protection plans cannot be updated. All changes require replacement.")
}

func (r *ProtectionPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "Protection plan deletion is not supported by the API. Removing from Terraform state only.")
}

func (r *ProtectionPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
