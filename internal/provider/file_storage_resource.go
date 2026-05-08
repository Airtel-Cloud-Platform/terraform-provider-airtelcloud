package provider

import (
	"context"
	"fmt"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FileStorageResource{}
var _ resource.ResourceWithImportState = &FileStorageResource{}

func NewFileStorageResource() resource.Resource {
	return &FileStorageResource{}
}

// FileStorageResource defines the resource implementation.
type FileStorageResource struct {
	client *client.Client
}

// FileStorageResourceModel describes the resource data model.
type FileStorageResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Size             types.String `tfsdk:"size"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	State            types.String `tfsdk:"state"`
	FailedStateError types.String `tfsdk:"failed_state_error"`
	CreatedAt        types.String `tfsdk:"created_at"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UUID             types.String `tfsdk:"uuid"`
	ProviderVolumeID types.String `tfsdk:"provider_volume_id"`
}

func (r *FileStorageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage"
}

func (r *FileStorageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud file storage volume for shared file storage (NFS).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the file storage volume (same as name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the file storage volume.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the file storage volume.",
				Optional:            true,
			},
			"size": schema.StringAttribute{
				MarkdownDescription: "The size of the file storage volume in GB.",
				Required:            true,
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone where the file storage volume will be created.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current state of the file storage volume.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"failed_state_error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Error message in case of failed state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the file storage volume was created.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user who created the file storage volume.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the file storage volume.",
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

func (r *FileStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FileStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FileStorageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the file storage volume
	createReq := &models.CreateFileStorageVolumeRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		Size:             data.Size.ValueString(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	err := r.client.CreateFileStorageVolume(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create file storage volume, got error: %s", err))
		return
	}

	// Wait for the file storage volume to become ready
	err = r.client.WaitForFileStorageVolumeReady(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 10*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("File storage volume failed to become ready: %s", err))
		return
	}

	// Refresh the file storage volume data
	volume, err := r.client.GetFileStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	// Save data into Terraform state
	r.mapFileStorageToModel(volume, data)

	tflog.Trace(ctx, "created a file storage volume")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FileStorageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := r.client.GetFileStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	// Save updated data into Terraform state
	r.mapFileStorageToModel(volume, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FileStorageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the file storage volume
	updateReq := &models.UpdateFileStorageVolumeRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		Size:             data.Size.ValueString(),
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	err := r.client.UpdateFileStorageVolume(ctx, data.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update file storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	// Wait for the file storage volume to become ready
	err = r.client.WaitForFileStorageVolumeReady(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 10*time.Minute)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("File storage volume failed to become ready after update: %s", err))
		return
	}

	// Refresh the file storage volume data
	volume, err := r.client.GetFileStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage volume %s after update, got error: %s", data.Name.ValueString(), err))
		return
	}

	// Save updated data into Terraform state
	r.mapFileStorageToModel(volume, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FileStorageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorageVolume(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete file storage volume %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	tflog.Trace(ctx, "deleted a file storage volume")
}

func (r *FileStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapFileStorageToModel maps a file storage volume model to the resource model
func (r *FileStorageResource) mapFileStorageToModel(volume *models.FileStorageVolume, data *FileStorageResourceModel) {
	data.ID = types.StringValue(volume.Name) // Using name as ID since that's the identifier
	data.Name = types.StringValue(volume.Name)
	if volume.Description != "" {
		data.Description = types.StringValue(volume.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Size = types.StringValue(volume.Size)
	data.AvailabilityZone = types.StringValue(volume.AvailabilityZone)
	data.State = types.StringValue(string(volume.State))
	data.FailedStateError = types.StringValue(volume.FailedStateError)
	data.CreatedAt = types.StringValue(volume.CreatedAt)
	data.CreatedBy = types.StringValue(volume.CreatedBy)
	data.UUID = types.StringValue(volume.UUID)
	data.ProviderVolumeID = types.StringValue(volume.ProviderVolumeID)
}
