package provider

import (
	"context"
	"fmt"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}
var _ resource.ResourceWithValidateConfig = &VolumeResource{}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// VolumeResource defines the resource implementation.
type VolumeResource struct {
	client *client.Client
}

// VolumeResourceModel describes the resource data model.
type VolumeResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	UUID             types.String `tfsdk:"uuid"`
	ProviderVolumeID types.String `tfsdk:"provider_volume_id"`
	Name             types.String `tfsdk:"name"`
	Size             types.Int64  `tfsdk:"size"`
	Type             types.String `tfsdk:"type"`
	Status           types.String `tfsdk:"status"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	VPCID            types.String `tfsdk:"vpc_id"`
	VPCName          types.String `tfsdk:"vpc_name"`
	SubnetID         types.String `tfsdk:"subnet_id"`
	SubnetName       types.String `tfsdk:"subnet_name"`
	ComputeID        types.String `tfsdk:"compute_id"`
	ComputeName      types.String `tfsdk:"compute_name"`
	IsEncrypted      types.Bool   `tfsdk:"is_encrypted"`
	Bootable         types.Bool   `tfsdk:"bootable"`
	EnableBackup     types.Bool   `tfsdk:"enable_backup"`
	AttachedTo       types.String `tfsdk:"attached_to"`
	AttachmentDevice types.String `tfsdk:"attachment_device"`
}

// computeIDPlanModifier computes the plan value for the Optional+Computed compute_id
// attribute, accounting for the sibling compute_name attribute:
//   - compute_id set in config: keep the configured value.
//   - neither compute_id nor compute_name set: null (detach).
//   - compute_name set, on create: unknown (resolved during apply).
//   - compute_name set, on update with unchanged name: keep the prior resolved id.
//   - compute_name set, on update with changed name: unknown (re-resolved on apply).
type computeIDPlanModifier struct{}

func (m computeIDPlanModifier) Description(_ context.Context) string {
	return "Derives compute_id from config, or from a resolved compute_name, otherwise null when both are removed."
}

func (m computeIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m computeIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// compute_id explicitly set in config: keep it.
	if !req.ConfigValue.IsNull() {
		return
	}

	var configName types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("compute_name"), &configName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Neither compute_id nor compute_name in config: detach.
	if configName.IsNull() {
		resp.PlanValue = types.StringNull()
		return
	}

	// compute_name is set. On create there is no prior state to reuse.
	if req.State.Raw.IsNull() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	// On update, reuse the prior resolved id when the name is unchanged so re-plans
	// are clean; mark unknown when the name changed so it is re-resolved during apply.
	var stateName types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("compute_name"), &stateName)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if configName.Equal(stateName) {
		resp.PlanValue = req.StateValue
	} else {
		resp.PlanValue = types.StringUnknown()
	}
}

// attachmentPlanModifier marks attachment fields as unknown when compute_id changes,
// so Terraform accepts the new value (including null) after apply.
type attachmentPlanModifier struct{}

func (m attachmentPlanModifier) Description(_ context.Context) string {
	return "Marks value as unknown when compute_id changes."
}

func (m attachmentPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m attachmentPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No custom logic needed during create or destroy
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Compare state against config (not plan) because plan modifier execution
	// order across attributes is non-deterministic — computeIDPlanModifier
	// may not have run yet when this modifier reads compute_id from the plan.
	var stateComputeID, configComputeID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("compute_id"), &stateComputeID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("compute_id"), &configComputeID)...)

	if !stateComputeID.Equal(configComputeID) {
		resp.PlanValue = types.StringUnknown()
	}
}

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud block storage volume.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the volume.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the volume, used for API operations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_volume_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The provider-specific volume ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the volume.",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the volume in GB.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the volume.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the volume.",
				Computed:            true,
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone where the volume is placed.",
				Optional:            true,
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The VPC network ID for the volume. Mutually exclusive with vpc_name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPC for the volume. If set, it is resolved to vpc_id. Mutually exclusive with vpc_id.",
				Optional:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "The subnet ID for the volume. Mutually exclusive with subnet_name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_name": schema.StringAttribute{
				MarkdownDescription: "The name of the subnet for the volume. If set, it is resolved to subnet_id (requires vpc_id or vpc_name). Mutually exclusive with subnet_id.",
				Optional:            true,
			},
			"compute_id": schema.StringAttribute{
				MarkdownDescription: "The compute instance ID to attach the volume to. Mutually exclusive with compute_name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					computeIDPlanModifier{},
				},
			},
			"compute_name": schema.StringAttribute{
				MarkdownDescription: "The name of the compute instance to attach the volume to. If set, it is resolved to compute_id. Mutually exclusive with compute_id.",
				Optional:            true,
			},
			"is_encrypted": schema.BoolAttribute{
				MarkdownDescription: "Whether the volume is encrypted. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"bootable": schema.BoolAttribute{
				MarkdownDescription: "Whether the volume is bootable. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_backup": schema.BoolAttribute{
				MarkdownDescription: "Whether backup is enabled for the volume. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"attached_to": schema.StringAttribute{
				MarkdownDescription: "The ID of the compute instance the volume is attached to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					attachmentPlanModifier{},
				},
			},
			"attachment_device": schema.StringAttribute{
				MarkdownDescription: "The device name when attached to a compute instance.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					attachmentPlanModifier{},
				},
			},
		},
	}
}

func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces that compute_id and compute_name are not both set. Both
// may be omitted — an unattached volume is valid.
func (r *VolumeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ComputeID.IsNull() && !data.ComputeName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of compute_id or compute_name may be specified, not both.")
	}

	// VPC and subnet each accept an id or a name (both optional), but not both.
	if !data.VPCID.IsNull() && !data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of vpc_id or vpc_name may be specified, not both.")
	}
	if !data.SubnetID.IsNull() && !data.SubnetName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of subnet_id or subnet_name may be specified, not both.")
	}

	// Resolving subnet_name requires a VPC to scope the lookup.
	if !data.SubnetName.IsNull() && data.VPCID.IsNull() && data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"subnet_name requires one of vpc_id or vpc_name to be specified.")
	}
}

func (r *VolumeResource) volumeClient(subnetID types.String) *client.Client {
	if !subnetID.IsNull() && subnetID.ValueString() != "" {
		return r.client.WithSubnetID(subnetID.ValueString())
	}
	return r.client
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve vpc_name -> vpc_id (needed before subnet resolution), then
	// subnet_name -> subnet_id. Persist the resolved ids into the Computed
	// attributes so they are known in state.
	vpcID := data.VPCID.ValueString()
	if vpcID == "" && !data.VPCName.IsNull() {
		resolved, err := r.client.ResolveVPCID(ctx, data.VPCName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VPC Resolution Error", err.Error())
			return
		}
		vpcID = resolved
	}

	subnetID := data.SubnetID.ValueString()
	if subnetID == "" && !data.SubnetName.IsNull() {
		resolved, err := r.client.ResolveSubnetID(ctx, vpcID, data.SubnetName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Subnet Resolution Error", err.Error())
			return
		}
		subnetID = resolved
	}

	if vpcID == "" {
		data.VPCID = types.StringNull()
	} else {
		data.VPCID = types.StringValue(vpcID)
	}
	if subnetID == "" {
		data.SubnetID = types.StringNull()
	} else {
		data.SubnetID = types.StringValue(subnetID)
	}

	volumeClient := r.volumeClient(data.SubnetID)

	// Resolve compute_name to compute_id when configured by name, and persist the
	// resolved id into the Computed compute_id attribute so it is known in state.
	computeID := data.ComputeID.ValueString()
	if computeID == "" && !data.ComputeName.IsNull() {
		resolved, err := volumeClient.ResolveComputeID(ctx, data.ComputeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Compute Resolution Error", fmt.Sprintf("Unable to resolve compute name %q: %s", data.ComputeName.ValueString(), err))
			return
		}
		computeID = resolved
	}
	// Mirror the plan modifier: an empty compute id means the volume is not
	// attached, which the plan represents as null. Setting StringValue("") here
	// would conflict with that null plan value ("was null, but now \"\"").
	if computeID == "" {
		data.ComputeID = types.StringNull()
	} else {
		data.ComputeID = types.StringValue(computeID)
	}

	// Validate volume type name against the API and resolve its ID
	volumeTypeName := data.Type.ValueString()
	var resolvedVolumeTypeID string
	if volumeTypeName != "" {
		volumeTypes, err := volumeClient.GetVolumeTypes(ctx, "BLOCK_STORAGE")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch volume types for validation: %s", err))
			return
		}

		// Filter to active types only
		var activeTypes []models.VolumeType
		for _, vt := range volumeTypes {
			if vt.IsActive {
				activeTypes = append(activeTypes, vt)
			}
		}

		found := false
		var validNames []string
		for _, vt := range activeTypes {
			validNames = append(validNames, vt.Name)
			if vt.Name == volumeTypeName {
				found = true
				resolvedVolumeTypeID = fmt.Sprintf("%d", vt.ID)
				break
			}
		}
		if !found {
			resp.Diagnostics.AddAttributeError(
				path.Root("volume_type"),
				"Invalid Volume Type",
				fmt.Sprintf("Volume type %q is not valid. Available types: %v", volumeTypeName, validNames),
			)
			return
		}
	}

	// Map is_encrypted bool to API string value
	isEncrypted := ""
	if !data.IsEncrypted.IsNull() && data.IsEncrypted.ValueBool() {
		isEncrypted = "encrypted"
	}

	createReq := &models.CreateVolumeRequest{
		VolumeName:       data.Name.ValueString(),
		VolumeSize:       int(data.Size.ValueInt64()),
		VolumeType:       volumeTypeName,
		VolumeTypeID:     resolvedVolumeTypeID,
		BillingUnit:      "MRC",
		VPCID:            vpcID,
		Network:          vpcID,
		SubnetID:         subnetID,
		ComputeID:        computeID,
		IsEncrypted:      isEncrypted,
		Bootable:         data.Bootable.ValueBool(),
		EnableBackup:     data.EnableBackup.ValueBool(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	volume, err := volumeClient.CreateVolume(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create volume, got error: %s", err))
		return
	}

	// Update the model with the created volume data
	data.ID = types.Int64Value(int64(volume.ID))
	data.UUID = types.StringValue(volume.UUID)
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
	data.Status = types.StringValue(volume.Status)
	az := volume.AZName
	if az == "" {
		az = volume.AvailabilityZone
	}
	data.AvailabilityZone = types.StringValue(az)

	// Handle volume attachments
	if len(volume.VolumeAttachments) > 0 {
		attachment := volume.VolumeAttachments[0]
		data.AttachedTo = types.StringValue(attachment.ComputeID)
		data.AttachmentDevice = types.StringValue(attachment.VolumeAttachmentDeviceName)
	} else {
		data.AttachedTo = types.StringNull()
		data.AttachmentDevice = types.StringNull()
	}

	tflog.Trace(ctx, "created a volume resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volumeClient := r.volumeClient(data.SubnetID)

	volume, err := volumeClient.GetVolume(ctx, data.UUID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume, got error: %s", err))
		return
	}

	// Update the model with the current volume data
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
	data.Name = types.StringValue(volume.VolumeName)
	data.Size = types.Int64Value(int64(volume.VolumeSize))
	data.Status = types.StringValue(volume.Status)
	az := volume.AZName
	if az == "" {
		az = volume.AvailabilityZone
	}
	data.AvailabilityZone = types.StringValue(az)
	data.Bootable = types.BoolValue(volume.Bootable)
	data.EnableBackup = types.BoolValue(volume.EnableBackup)

	if volume.VPCID != "" {
		data.VPCID = types.StringValue(volume.VPCID)
	}

	// Handle volume attachments and reconcile compute_id with API state
	if len(volume.VolumeAttachments) > 0 {
		attachment := volume.VolumeAttachments[0]
		data.ComputeID = types.StringValue(attachment.ComputeID)
		data.AttachedTo = types.StringValue(attachment.ComputeID)
		data.AttachmentDevice = types.StringValue(attachment.VolumeAttachmentDeviceName)
	} else {
		if !data.ComputeID.IsNull() {
			data.ComputeID = types.StringNull()
		}
		data.AttachedTo = types.StringNull()
		data.AttachmentDevice = types.StringNull()
	}

	// Set volume type name if available
	if volume.VolumeType.Name != "" {
		data.Type = types.StringValue(volume.VolumeType.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeResourceModel
	var state VolumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate: size can only increase
	if plan.Size.ValueInt64() < state.Size.ValueInt64() {
		resp.Diagnostics.AddAttributeError(
			path.Root("size"),
			"Invalid Volume Size",
			fmt.Sprintf("Volume size can only be increased. Current: %d GB, requested: %d GB.",
				state.Size.ValueInt64(), plan.Size.ValueInt64()),
		)
		return
	}

	volumeClient := r.volumeClient(plan.SubnetID)

	// Read current volume to get volume_type name and compute_id
	currentVolume, err := volumeClient.GetVolume(ctx, state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume for update: %s", err))
		return
	}

	// Resolve the target compute id for this apply. When configured by name the
	// Computed compute_id is unknown in the plan (its plan modifier only reuses the
	// prior id when the name is unchanged), so resolve compute_name to a concrete id
	// before diffing — otherwise a changed name would look like a detach.
	oldComputeID := state.ComputeID.ValueString()
	newComputeID := plan.ComputeID.ValueString()
	if newComputeID == "" && !plan.ComputeName.IsNull() {
		resolved, err := volumeClient.ResolveComputeID(ctx, plan.ComputeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Compute Resolution Error", fmt.Sprintf("Unable to resolve compute name %q: %s", plan.ComputeName.ValueString(), err))
			return
		}
		newComputeID = resolved
	}
	if oldComputeID != newComputeID {
		// Use API as source of truth for current attachment state
		actualAttachedID := ""
		if len(currentVolume.VolumeAttachments) > 0 {
			actualAttachedID = currentVolume.VolumeAttachments[0].ComputeID
		}

		// Detach from current compute if attached
		if actualAttachedID != "" {
			tflog.Info(ctx, "Detaching volume", map[string]any{
				"volume_uuid": state.UUID.ValueString(),
				"compute_id":  actualAttachedID,
			})
			detachReq := &models.VolumeDetachRequest{
				ComputeID: actualAttachedID,
				VolumeID:  int(state.ID.ValueInt64()),
			}
			if err := volumeClient.DetachVolume(ctx, state.UUID.ValueString(), detachReq); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to detach volume: %s", err))
				return
			}
			if err := volumeClient.WaitForVolumeDetached(ctx, state.UUID.ValueString()); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Volume detach did not complete: %s", err))
				return
			}
		}

		// Attach to new compute if specified
		if newComputeID != "" {
			tflog.Info(ctx, "Attaching volume", map[string]any{
				"volume_uuid": state.UUID.ValueString(),
				"compute_id":  newComputeID,
			})
			attachReq := &models.VolumeAttachRequest{
				ComputeID: newComputeID,
				VolumeID:  int(state.ID.ValueInt64()),
			}
			if err := volumeClient.AttachVolume(ctx, state.UUID.ValueString(), attachReq); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach volume: %s", err))
				return
			}
			if err := volumeClient.WaitForVolumeAttached(ctx, state.UUID.ValueString(), newComputeID); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Volume attach did not complete: %s", err))
				return
			}
		}
	}

	isEncrypted := ""
	if !plan.IsEncrypted.IsNull() && plan.IsEncrypted.ValueBool() {
		isEncrypted = "encrypted"
	}

	computeID := newComputeID

	// Persist the resolved compute id into the Computed compute_id attribute. It may
	// be unknown in the plan (configured by a changed name); writing an unknown value
	// to state is an error, so settle it here to the resolved id (or null if detached).
	if computeID != "" {
		plan.ComputeID = types.StringValue(computeID)
	} else {
		plan.ComputeID = types.StringNull()
	}

	updateReq := &models.UpdateVolumeRequest{
		VolumeName:   plan.Name.ValueString(),
		VolumeSize:   int(plan.Size.ValueInt64()),
		Bootable:     plan.Bootable.ValueBool(),
		EnableBackup: plan.EnableBackup.ValueBool(),
		VolumeType:   currentVolume.VolumeType.Name,
		BillingUnit:  "MRC",
		VolumeRate:   0,
		ComputeID:    computeID,
		IsEncrypted:  isEncrypted,
	}

	err = volumeClient.UpdateVolume(ctx, state.UUID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update volume, got error: %s", err))
		return
	}

	// Read the updated volume to get the latest state
	updatedVolume, err := volumeClient.GetVolume(ctx, state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated volume, got error: %s", err))
		return
	}

	// Preserve stable identifiers from state (they don't change during updates)
	plan.ID = state.ID
	plan.UUID = state.UUID
	plan.ProviderVolumeID = state.ProviderVolumeID

	// Update status from API response
	plan.Status = types.StringValue(updatedVolume.Status)

	// Size: use the plan value, not the API response, because resize is async
	// and GetVolume may return the old size before the backend finishes resizing

	// Handle volume attachments
	if len(updatedVolume.VolumeAttachments) > 0 {
		attachment := updatedVolume.VolumeAttachments[0]
		plan.AttachedTo = types.StringValue(attachment.ComputeID)
		plan.AttachmentDevice = types.StringValue(attachment.VolumeAttachmentDeviceName)
	} else {
		plan.AttachedTo = types.StringNull()
		plan.AttachmentDevice = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volumeClient := r.volumeClient(data.SubnetID)

	// Check if volume is attached to any compute instance
	currentVolume, err := volumeClient.GetVolume(ctx, data.UUID.ValueString())
	if err != nil {
		// If volume is already gone (404), nothing to do
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume before delete: %s", err))
		return
	}

	// Detach from all compute instances before deleting
	for _, attachment := range currentVolume.VolumeAttachments {
		tflog.Info(ctx, "Detaching volume before delete", map[string]any{
			"volume_uuid": data.UUID.ValueString(),
			"compute_id":  attachment.ComputeID,
		})
		detachReq := &models.VolumeDetachRequest{
			ComputeID: attachment.ComputeID,
			VolumeID:  int(data.ID.ValueInt64()),
		}
		if err := volumeClient.DetachVolume(ctx, data.UUID.ValueString(), detachReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to detach volume before delete: %s", err))
			return
		}
	}

	// Wait for all detachments to complete
	if len(currentVolume.VolumeAttachments) > 0 {
		if err := volumeClient.WaitForVolumeDetached(ctx, data.UUID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Volume detach did not complete before delete: %s", err))
			return
		}
	}

	// Now delete the volume
	err = volumeClient.DeleteVolume(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete volume, got error: %s", err))
		return
	}
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
