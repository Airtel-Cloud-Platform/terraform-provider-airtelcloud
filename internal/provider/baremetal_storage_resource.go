package provider

import (
	"context"
	"fmt"
	"time"

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

var _ resource.Resource = &BaremetalStorageResource{}
var _ resource.ResourceWithImportState = &BaremetalStorageResource{}

func NewBaremetalStorageResource() resource.Resource {
	return &BaremetalStorageResource{}
}

type BaremetalStorageResource struct {
	client *client.Client
}

type BaremetalStorageResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Size             types.Int64  `tfsdk:"size"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	State            types.String `tfsdk:"state"`
	FailedStateError types.String `tfsdk:"failed_state_error"`
	CreatedAt        types.String `tfsdk:"created_at"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UUID             types.String `tfsdk:"uuid"`
	ProviderVolumeID types.String `tfsdk:"provider_volume_id"`
}

func (r *BaremetalStorageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_baremetal_storage"
}

func (r *BaremetalStorageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud baremetal block storage volume (`POST /api/storage-plugin/v1/.../block-storage/volume`).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the volume (same as name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the baremetal storage volume.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the volume.",
				Optional:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the volume in GB.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone where the volume will be created (e.g. `S1`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current state of the volume.",
			},
			"failed_state_error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Error message in case of failed state.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the volume was created.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user who created the volume.",
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the volume.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_volume_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The provider-specific volume identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *BaremetalStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BaremetalStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BaremetalStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateBlockStorageVolumeRequest{
		Name:             data.Name.ValueString(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
		Size:             data.Size.ValueInt64(),
		Description:      data.Description.ValueString(),
	}

	err := r.client.CreateBlockStorageVolume(ctx, createReq)
	if err != nil {
		if _, getErr := r.client.GetBlockStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString()); getErr != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create baremetal storage volume, got error: %s", err))
			return
		}
	}

	err = r.client.WaitForBlockStorageVolumeReady(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 10*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Baremetal storage volume failed to become ready: %s", err))
		return
	}

	volume, err := r.client.GetBlockStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	r.mapBlockStorageToModel(volume, &data)
	tflog.Trace(ctx, "created a baremetal storage volume")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BaremetalStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := r.client.GetBlockStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	r.mapBlockStorageToModel(volume, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BaremetalStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &models.UpdateBlockStorageVolumeRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		Size:             data.Size.ValueInt64(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	err := r.client.UpdateBlockStorageVolume(ctx, data.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update baremetal storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	err = r.client.WaitForBlockStorageVolumeReady(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 10*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Baremetal storage volume failed to become ready after update: %s", err))
		return
	}

	volume, err := r.client.GetBlockStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal storage volume %s after update, got error: %s", data.Name.ValueString(), err))
		return
	}

	r.mapBlockStorageToModel(volume, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BaremetalStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBlockStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), data.UUID.ValueString())
	if err != nil && !client.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete baremetal storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	tflog.Trace(ctx, "deleted a baremetal storage volume")
}

func (r *BaremetalStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *BaremetalStorageResource) mapBlockStorageToModel(volume *models.BlockStorageVolume, data *BaremetalStorageResourceModel) {
	data.ID = types.StringValue(volume.Name)
	data.Name = types.StringValue(volume.Name)
	if volume.Description != "" {
		data.Description = types.StringValue(volume.Description)
	}
	if volume.Size != 0 {
		data.Size = types.Int64Value(int64(volume.Size))
	}
	if volume.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(volume.AvailabilityZone)
	}
	data.State = types.StringValue(string(volume.State))
	data.FailedStateError = types.StringValue(volume.FailedStateError)
	data.CreatedAt = types.StringValue(volume.CreatedAt)
	data.CreatedBy = types.StringValue(volume.CreatedBy)
	data.UUID = types.StringValue(volume.UUID)
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
}
