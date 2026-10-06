package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

var _ resource.Resource = &ProtectionResource{}
var _ resource.ResourceWithImportState = &ProtectionResource{}
var _ resource.ResourceWithValidateConfig = &ProtectionResource{}

func NewProtectionResource() resource.Resource {
	return &ProtectionResource{}
}

type ProtectionResource struct {
	client *client.Client
}

func stringValueOrNull(v string) types.String {
	if strings.TrimSpace(v) == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

type ProtectionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	PolicyTypeID    types.String `tfsdk:"policy_type_id"`
	VMName          types.String `tfsdk:"vm_name"`
	ComputeID       types.String `tfsdk:"compute_id"`
	ProtectionPlan  types.String `tfsdk:"protection_plan"`
	EnableScheduler types.String `tfsdk:"enable_scheduler"`
	StartDate       types.String `tfsdk:"start_date"`
	Weekday         types.String `tfsdk:"weekday"`
	EndDate         types.String `tfsdk:"end_date"`
	StartTime       types.String `tfsdk:"start_time"`
	Status          types.String `tfsdk:"status"`
	Region          types.String `tfsdk:"region"`
	AZName          types.String `tfsdk:"az_name"`
	Created         types.String `tfsdk:"created"`
}

func (r *ProtectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protection"
}

func (r *ProtectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a protection plan to a VM. Policy `name` is the VM name. `compute_id` is looked up from `vm_name`. Create matches the console form: `name`, `compute_id`, `protection_plan`, `start_date`, `start_time`, `description`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the protection policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Protection policy name. Always the VM `instance_name` (`vm_name`). Not set by the user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the protection policy.",
				Optional:            true,
			},
			"policy_type_id": schema.StringAttribute{
				MarkdownDescription: "The policy type ID.",
				Optional:            true,
			},
			"vm_name": schema.StringAttribute{
				MarkdownDescription: "Existing VM `instance_name`. Sent as form `name`. The VM UUID is sent as `compute_id`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_id": schema.StringAttribute{
				MarkdownDescription: "VM UUID looked up from `vm_name`. Sent as form `compute_id`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protection_plan": schema.StringAttribute{
				MarkdownDescription: "Protection plan UUID or console name. Names are resolved to UUID at apply. Prefer `airtelcloud_protection_plan.example.id`.",
				Required:            true,
			},
			"enable_scheduler": schema.StringAttribute{
				MarkdownDescription: "Whether to enable the backup scheduler. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("true"),
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date for the protection schedule, in `YYYY-MM-DD` format (e.g. `2026-07-20`). Mutually exclusive with `weekday`. Sent to the API as `MM/DD/YYYY`.",
				Optional:            true,
			},
			"weekday": schema.StringAttribute{
				MarkdownDescription: "Optional weekday convenience input (`monday`..`sunday` or `mon`..`sun`). Mutually exclusive with `start_date`. Converted internally to the next matching `start_date` in IST (`Asia/Kolkata`).",
				Optional:            true,
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The end date for the protection schedule, in `YYYY-MM-DD` format. Sent to the API as `MM/DD/YYYY`.",
				Optional:            true,
			},
			"start_time": schema.StringAttribute{
				MarkdownDescription: "The start time for the protection schedule in IST (`Asia/Kolkata`), in 24-hour `HH:MM` format (e.g. `02:00`, `00:00`). Sent to the API as 12-hour `H:MM AM/PM`.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the protection policy.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the protection policy.",
				Computed:            true,
			},
			"az_name": schema.StringAttribute{
				MarkdownDescription: "The availability zone of the protection policy.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *ProtectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig requires vm_name. start_date and weekday are mutually exclusive.
func (r *ProtectionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.VMName.IsUnknown() {
		return
	}
	if data.VMName.IsNull() || strings.TrimSpace(data.VMName.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(path.Root("vm_name"), "Invalid Configuration", "vm_name is required.")
	}

	hasStartDate := !data.StartDate.IsNull() && strings.TrimSpace(data.StartDate.ValueString()) != ""
	hasWeekday := !data.Weekday.IsNull() && strings.TrimSpace(data.Weekday.ValueString()) != ""
	if hasStartDate && hasWeekday {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of start_date or weekday may be specified, not both.")
	}
}

// formatProtectionStartDate converts a start date to the API's MM/DD/YYYY format.
// It accepts ISO 8601 input (2026-07-20) — the Terraform-friendly form — and also
// passes through input already in MM/DD/YYYY. Empty input stays empty (omitted).
func formatProtectionStartDate(in string) (string, error) {
	if in == "" {
		return "", nil
	}
	for _, layout := range []string{"2006-01-02", "01/02/2006"} {
		if t, err := time.Parse(layout, in); err == nil {
			return t.Format("01/02/2006"), nil
		}
	}
	return "", fmt.Errorf("invalid start_date %q: expected YYYY-MM-DD (e.g. 2026-07-20)", in)
}

// formatProtectionStartTime converts a start time to the API's 12-hour "3:04 PM" format
// (e.g. "12:00 AM"). It accepts 24-hour HH:MM input (02:00) — the Terraform-friendly
// form — and also passes through input already in 12-hour AM/PM. Empty stays empty.
func formatProtectionStartTime(in string) (string, error) {
	if in == "" {
		return "", nil
	}
	normalized := strings.ToUpper(strings.TrimSpace(in))
	for _, layout := range []string{"15:04", "3:04 PM", "03:04 PM", "3:04PM", "03:04PM"} {
		if t, err := time.Parse(layout, normalized); err == nil {
			return t.Format("3:04 PM"), nil
		}
	}
	return "", fmt.Errorf("invalid start_time %q: expected 24-hour HH:MM (e.g. 02:00) or 12-hour (e.g. 2:00 AM)", in)
}

func resolveWeekdayStartDate(in string, now time.Time) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(in))
	if normalized == "" {
		return "", nil
	}

	weekdayMap := map[string]time.Weekday{
		"sun":       time.Sunday,
		"sunday":    time.Sunday,
		"mon":       time.Monday,
		"monday":    time.Monday,
		"tue":       time.Tuesday,
		"tues":      time.Tuesday,
		"tuesday":   time.Tuesday,
		"wed":       time.Wednesday,
		"wednesday": time.Wednesday,
		"thu":       time.Thursday,
		"thur":      time.Thursday,
		"thurs":     time.Thursday,
		"thursday":  time.Thursday,
		"fri":       time.Friday,
		"friday":    time.Friday,
		"sat":       time.Saturday,
		"saturday":  time.Saturday,
	}

	target, ok := weekdayMap[normalized]
	if !ok {
		return "", fmt.Errorf("invalid weekday %q: expected monday..sunday or mon..sun", in)
	}

	istNow := now.In(istLocation)
	base := time.Date(istNow.Year(), istNow.Month(), istNow.Day(), 0, 0, 0, 0, istLocation)
	delta := (int(target) - int(base.Weekday()) + 7) % 7
	return base.AddDate(0, 0, delta).Format("2006-01-02"), nil
}

func resolveProtectionStartDateInput(startDate, weekday string, now time.Time) (string, error) {
	if strings.TrimSpace(startDate) != "" {
		return startDate, nil
	}
	return resolveWeekdayStartDate(weekday, now)
}

func (r *ProtectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.ResolveComputeNode(ctx, "", data.VMName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("VM Lookup Error", fmt.Sprintf("Unable to resolve vm_name %q: %s", data.VMName.ValueString(), err))
		return
	}
	computeID := strings.TrimSpace(vm.ID)
	if computeID == "" {
		resp.Diagnostics.AddError("VM Lookup Error", fmt.Sprintf("VM %q has an empty id", data.VMName.ValueString()))
		return
	}
	az, _ := protectionPlanVMLocation(vm)
	policyName := strings.TrimSpace(vm.InstanceName)
	if policyName == "" {
		policyName = strings.TrimSpace(data.VMName.ValueString())
	}
	data.ComputeID = types.StringValue(computeID)
	data.Name = types.StringValue(policyName)

	planID, err := r.client.ResolveProtectionPlanID(ctx, data.ProtectionPlan.ValueString(), az)
	if err != nil {
		resp.Diagnostics.AddError("Protection Plan Error", fmt.Sprintf("Unable to resolve protection_plan %q: %s", data.ProtectionPlan.ValueString(), err))
		return
	}

	// The API expects start_date as MM/DD/YYYY and start_time as 12-hour AM/PM.
	startDateInput, err := resolveProtectionStartDateInput(data.StartDate.ValueString(), data.Weekday.ValueString(), time.Now())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startDate, err := formatProtectionStartDate(startDateInput)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startTime, err := formatProtectionStartTime(data.StartTime.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	endDate, err := formatProtectionStartDate(data.EndDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", strings.Replace(err.Error(), "start_date", "end_date", 1))
		return
	}

	createReq := &models.CreateProtectionRequest{
		Name:           policyName,
		Description:    data.Description.ValueString(),
		PolicyTypeID:   data.PolicyTypeID.ValueString(),
		ComputeID:      computeID,
		ProtectionPlan: planID,
		StartDate:      startDate,
		EndDate:        endDate,
		StartTime:      startTime,
	}

	protection, err := r.client.CreateProtection(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create protection, got error: %s", err))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(protection.ID))
	if protection.Name != "" {
		data.Name = types.StringValue(protection.Name)
	}
	data.Status = types.StringValue(protection.Status)
	data.Region = stringValueOrNull(protection.Region)
	data.AZName = stringValueOrNull(protection.AZName)
	data.Created = stringValueOrNull(protection.Created)

	tflog.Trace(ctx, "created protection resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	protection, err := r.client.GetProtection(ctx, id)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read protection, got error: %s", err))
		return
	}

	if protection.Name != "" {
		data.Name = types.StringValue(protection.Name)
	}
	if protection.Description != "" {
		data.Description = types.StringValue(protection.Description)
	}
	if protection.ComputeID != "" {
		data.ComputeID = types.StringValue(protection.ComputeID)
	}
	data.Status = types.StringValue(protection.Status)
	data.Region = stringValueOrNull(protection.Region)
	data.AZName = stringValueOrNull(protection.AZName)
	data.Created = stringValueOrNull(protection.Created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	// The API expects start_date as MM/DD/YYYY and start_time as 12-hour AM/PM.
	startDateInput, err := resolveProtectionStartDateInput(data.StartDate.ValueString(), data.Weekday.ValueString(), time.Now())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startDate, err := formatProtectionStartDate(startDateInput)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startTime, err := formatProtectionStartTime(data.StartTime.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	endDate, err := formatProtectionStartDate(data.EndDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", strings.Replace(err.Error(), "start_date", "end_date", 1))
		return
	}

	az := data.AZName.ValueString()
	planID, err := r.client.ResolveProtectionPlanID(ctx, data.ProtectionPlan.ValueString(), az)
	if err != nil {
		resp.Diagnostics.AddError("Protection Plan Error", fmt.Sprintf("Unable to resolve protection_plan %q: %s", data.ProtectionPlan.ValueString(), err))
		return
	}

	updateReq := &models.UpdateProtectionRequest{
		Name:           data.VMName.ValueString(),
		Description:    data.Description.ValueString(),
		PolicyTypeID:   data.PolicyTypeID.ValueString(),
		ProtectionPlan: planID,
		StartDate:      startDate,
		EndDate:        endDate,
		StartTime:      startTime,
	}

	protection, err := r.client.UpdateProtection(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update protection, got error: %s", err))
		return
	}

	data.Name = types.StringValue(protection.Name)
	data.Status = types.StringValue(protection.Status)
	data.Region = stringValueOrNull(protection.Region)
	data.AZName = stringValueOrNull(protection.AZName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse protection ID: %s", err))
		return
	}

	err = r.client.DeleteProtection(ctx, id)
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete protection, got error: %s", err))
		return
	}
}

func (r *ProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
