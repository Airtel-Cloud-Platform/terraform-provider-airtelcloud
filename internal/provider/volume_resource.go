package provider

import (
	"context"
	"fmt"

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

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}

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
	SubnetID         types.String `tfsdk:"subnet_id"`
	ComputeID        types.String `tfsdk:"compute_id"`
	IsEncrypted      types.Bool   `tfsdk:"is_encrypted"`
	Bootable         types.Bool   `tfsdk:"bootable"`
	EnableBackup     types.Bool   `tfsdk:"enable_backup"`
	AttachedTo       types.String `tfsdk:"attached_to"`
	AttachmentDevice types.String `tfsdk:"attachment_device"`
}

// computeIDPlanModifier ensures that removing compute_id from config
// produces a null plan value instead of preserving the prior state.
type computeIDPlanModifier struct{}

func (m computeIDPlanModifier) Description(_ context.Context) string {
	return "Sets plan value to null when compute_id is removed from configuration."
}

func (m computeIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m computeIDPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringNull()
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
				MarkdownDescription: "The VPC network ID for the volume.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "The subnet ID for the volume.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_id": schema.StringAttribute{
				MarkdownDescription: "The compute instance ID to attach the volume to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					computeIDPlanModifier{},
				},
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

	volumeClient := r.volumeClient(data.SubnetID)

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
		VPCID:            data.VPCID.ValueString(),
		Network:          data.VPCID.ValueString(),
		SubnetID:         data.SubnetID.ValueString(),
		ComputeID:        data.ComputeID.ValueString(),
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

	// Handle attach/detach if compute_id changed
	oldComputeID := state.ComputeID.ValueString()
	newComputeID := plan.ComputeID.ValueString()
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
