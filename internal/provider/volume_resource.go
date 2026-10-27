package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	VPCName          types.String `tfsdk:"vpc_name"`
	SubnetName       types.String `tfsdk:"subnet_name"`
	ComputeName      types.String `tfsdk:"compute_name"`
	IsEncrypted      types.Bool   `tfsdk:"is_encrypted"`
	Bootable         types.Bool   `tfsdk:"bootable"`
	EnableBackup     types.Bool   `tfsdk:"enable_backup"`
	AttachedTo       types.String `tfsdk:"attached_to"`
	AttachmentDevice types.String `tfsdk:"attachment_device"`
	Labels           types.List   `tfsdk:"labels"`
}

// attachmentPlanModifier marks attachment fields as unknown when compute_name changes.
type attachmentPlanModifier struct{}

func (m attachmentPlanModifier) Description(_ context.Context) string {
	return "Marks value as unknown when compute_name changes."
}

func (m attachmentPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m attachmentPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var stateName, configName types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("compute_name"), &stateName)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("compute_name"), &configName)...)
	if !stateName.Equal(configName) {
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
				MarkdownDescription: "The size of the volume in GB. API requires 50 to 4095.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(50, 4095),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Volume tier. Accepts the API type name or the UI label (for example `Standard Performance`). Looked up from GET .../volume_types?group=BLOCK_STORAGE with the subnet-id header.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the volume.",
				Computed:            true,
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone (UI value, for example `S1` or `South-AZ1`).",
				Required:            true,
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "VPC name. Resolved to a VPC ID for the create API. Forces replacement if changed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_name": schema.StringAttribute{
				MarkdownDescription: "Subnet name within the VPC. Resolved to a subnet ID (sent as the subnet-id header). Forces replacement if changed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_name": schema.StringAttribute{
				MarkdownDescription: "VM instance name to attach to. Resolved to a compute ID for the API. Omit to leave the volume unattached. Changing this value detaches and re-attaches in place.",
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
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Optional labels. Applied after create via PATCH .../volumes/labels using the numeric volume id.",
				Optional:            true,
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

// matchVolumeType matches the UI tier or API name against active BLOCK_STORAGE types.
func matchVolumeType(volumeTypes []models.VolumeType, want string) (models.VolumeType, []string, bool) {
	want = strings.TrimSpace(want)
	var options []string
	for _, vt := range volumeTypes {
		if !vt.IsActive {
			continue
		}
		label := vt.Name
		if vt.Label != "" && !strings.EqualFold(vt.Label, vt.Name) {
			label = fmt.Sprintf("%s (%s)", vt.Label, vt.Name)
		}
		options = append(options, label)
		if strings.EqualFold(vt.Name, want) || (vt.Label != "" && strings.EqualFold(vt.Label, want)) {
			return vt, options, true
		}
	}
	return models.VolumeType{}, options, false
}

func bothKnownSet(a, b types.String) bool {
	return !a.IsNull() && !a.IsUnknown() && !b.IsNull() && !b.IsUnknown()
}

func (r *VolumeResource) resolveNetwork(ctx context.Context, vpcName, subnetName types.String) (vpcID, subnetID string, err error) {
	vpcID, err = r.client.ResolveVPCID(ctx, vpcName.ValueString())
	if err != nil {
		return "", "", err
	}
	subnetID, err = r.client.ResolveSubnetID(ctx, vpcID, subnetName.ValueString())
	if err != nil {
		return "", "", err
	}
	return vpcID, subnetID, nil
}

func (r *VolumeResource) volumeClientFor(ctx context.Context, vpcName, subnetName types.String) (*client.Client, string, string, error) {
	vpcID, subnetID, err := r.resolveNetwork(ctx, vpcName, subnetName)
	if err != nil {
		return nil, "", "", err
	}
	return r.client.WithSubnetID(subnetID), vpcID, subnetID, nil
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volumeClient, vpcID, subnetID, err := r.volumeClientFor(ctx, data.VPCName, data.SubnetName)
	if err != nil {
		resp.Diagnostics.AddError("Network Resolution Error", err.Error())
		return
	}

	computeID := ""
	if !data.ComputeName.IsNull() && data.ComputeName.ValueString() != "" {
		resolved, err := volumeClient.ResolveComputeID(ctx, data.ComputeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Compute Resolution Error", fmt.Sprintf("Unable to resolve compute name %q: %s", data.ComputeName.ValueString(), err))
			return
		}
		computeID = resolved
	}

	// UI loads tiers from GET .../volume_types?group=BLOCK_STORAGE with subnet-id.
	volumeTypes, err := volumeClient.GetVolumeTypes(ctx, "BLOCK_STORAGE")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch volume types for validation: %s", err))
		return
	}
	matchedType, validNames, found := matchVolumeType(volumeTypes, data.Type.ValueString())
	if !found {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"Invalid Volume Type",
			fmt.Sprintf("Volume type %q is not valid. Available types: %v", data.Type.ValueString(), validNames),
		)
		return
	}
	volumeTypeName := matchedType.Name
	resolvedVolumeTypeID := fmt.Sprintf("%d", matchedType.ID)

	var labels []string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
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
	if len(labels) > 0 {
		if err := volumeClient.PatchVolumeLabels(ctx, volume.ID, labels); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set volume labels, got error: %s", err))
			return
		}
	}

	// Update the model with the created volume data
	data.ID = types.Int64Value(int64(volume.ID))
	data.UUID = types.StringValue(volume.UUID)
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
	data.Status = types.StringValue(volume.Status)
	if data.AvailabilityZone.IsNull() || data.AvailabilityZone.ValueString() == "" {
		az := volume.AZName
		if az == "" {
			az = volume.AvailabilityZone
		}
		data.AvailabilityZone = types.StringValue(az)
	}

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

	volumeClient, _, _, err := r.volumeClientFor(ctx, data.VPCName, data.SubnetName)
	if err != nil {
		resp.Diagnostics.AddError("Network Resolution Error", err.Error())
		return
	}

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
	data.ID = types.Int64Value(int64(volume.ID))
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
	data.Name = types.StringValue(volume.VolumeName)
	data.Size = types.Int64Value(int64(volume.VolumeSize))
	data.Status = types.StringValue(volume.Status)
	if data.AvailabilityZone.IsNull() || data.AvailabilityZone.ValueString() == "" {
		az := volume.AZName
		if az == "" {
			az = volume.AvailabilityZone
		}
		data.AvailabilityZone = types.StringValue(az)
	}
	data.Bootable = types.BoolValue(volume.Bootable)
	data.EnableBackup = types.BoolValue(volume.EnableBackup)

	if len(volume.VolumeAttachments) > 0 {
		attachment := volume.VolumeAttachments[0]
		data.AttachedTo = types.StringValue(attachment.ComputeID)
		data.AttachmentDevice = types.StringValue(attachment.VolumeAttachmentDeviceName)
	} else {
		data.AttachedTo = types.StringNull()
		data.AttachmentDevice = types.StringNull()
	}

	if volume.VolumeType.Name != "" {
		current := data.Type.ValueString()
		if !strings.EqualFold(current, volume.VolumeType.Name) &&
			(volume.VolumeType.Label == "" || !strings.EqualFold(current, volume.VolumeType.Label)) {
			data.Type = types.StringValue(volume.VolumeType.Name)
		}
	}
	if apiLabels := parseVolumeLabels(volume.Labels); len(apiLabels) > 0 {
		list, diags := types.ListValueFrom(ctx, types.StringType, apiLabels)
		resp.Diagnostics.Append(diags...)
		data.Labels = list
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

	volumeClient, _, _, err := r.volumeClientFor(ctx, plan.VPCName, plan.SubnetName)
	if err != nil {
		resp.Diagnostics.AddError("Network Resolution Error", err.Error())
		return
	}

	currentVolume, err := volumeClient.GetVolume(ctx, state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume for update: %s", err))
		return
	}

	newComputeID := ""
	if !plan.ComputeName.IsNull() && plan.ComputeName.ValueString() != "" {
		resolved, err := volumeClient.ResolveComputeID(ctx, plan.ComputeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Compute Resolution Error", fmt.Sprintf("Unable to resolve compute name %q: %s", plan.ComputeName.ValueString(), err))
			return
		}
		newComputeID = resolved
	}
	if !state.ComputeName.Equal(plan.ComputeName) {
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

	if !plan.Labels.Equal(state.Labels) {
		var labels []string
		if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
			resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if err := volumeClient.PatchVolumeLabels(ctx, int(state.ID.ValueInt64()), labels); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set volume labels, got error: %s", err))
			return
		}
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

	volumeClient, _, _, err := r.volumeClientFor(ctx, data.VPCName, data.SubnetName)
	if err != nil {
		resp.Diagnostics.AddError("Network Resolution Error", err.Error())
		return
	}

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

func parseVolumeLabels(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	var obj struct {
		Labels []string `json:"labels"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj.Labels
	}
	return nil
}
