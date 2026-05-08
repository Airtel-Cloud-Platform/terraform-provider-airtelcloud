package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FileStorageExportPathResource{}
var _ resource.ResourceWithImportState = &FileStorageExportPathResource{}

func NewFileStorageExportPathResource() resource.Resource {
	return &FileStorageExportPathResource{}
}

// FileStorageExportPathResource defines the resource implementation.
type FileStorageExportPathResource struct {
	client *client.Client
}

// FileStorageExportPathResourceModel describes the resource data model.
type FileStorageExportPathResourceModel struct {
	ID                types.String `tfsdk:"id"`
	PathID            types.String `tfsdk:"path_id"`
	Volume            types.String `tfsdk:"volume"`
	Description       types.String `tfsdk:"description"`
	Protocol          types.String `tfsdk:"protocol"`
	AvailabilityZone  types.String `tfsdk:"availability_zone"`
	NFSExportPath     types.String `tfsdk:"nfs_export_path"`
	DefaultAccessType types.String `tfsdk:"default_access_type"`
	DefaultUserSquash types.String `tfsdk:"default_user_squash"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedBy         types.String `tfsdk:"created_by"`
	ProviderExpPathID types.String `tfsdk:"provider_export_path_id"`
}

func (r *FileStorageExportPathResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_export_path"
}

func (r *FileStorageExportPathResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud NFS export path for file storage volumes. This resource allows you to create NFS mount points for your file storage volumes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the export path (same as path_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique path ID assigned by the system.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"volume": schema.StringAttribute{
				MarkdownDescription: "The name of the file storage volume to create the export path for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the NFS export path.",
				Optional:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The NFS protocol version. Valid values: 'NFSv3', 'NFSv4'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NFSv4"),
				Validators: []validator.String{
					stringvalidator.OneOf("NFSv3", "NFSv4"),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone for the export path.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nfs_export_path": schema.StringAttribute{
				MarkdownDescription: "The NFS export directory path name.",
				Optional:            true,
				Computed:            true,
			},
			"default_access_type": schema.StringAttribute{
				MarkdownDescription: "Default access type for the export. Valid values: 'NoAccess', 'ReadOnly', 'ReadWrite'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ReadWrite"),
				Validators: []validator.String{
					stringvalidator.OneOf("NoAccess", "ReadOnly", "ReadWrite"),
				},
			},
			"default_user_squash": schema.StringAttribute{
				MarkdownDescription: "Default user squash setting. Valid values: 'NoSquash', 'RootIdSquash', 'RootSquash', 'AllSquash'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NoSquash"),
				Validators: []validator.String{
					stringvalidator.OneOf("NoSquash", "RootIdSquash", "RootSquash", "AllSquash"),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the export path was created.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user who created the export path.",
			},
			"provider_export_path_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The provider-specific export path identifier.",
			},
		},
	}
}

func (r *FileStorageExportPathResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FileStorageExportPathResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FileStorageExportPathResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the NFS info
	nfsInfo := &models.NFSExportInfo{
		DefaultAccessType: models.NFSAccessType(data.DefaultAccessType.ValueString()),
		DefaultUserSquash: models.NFSSquashType(data.DefaultUserSquash.ValueString()),
	}

	if !data.NFSExportPath.IsNull() && !data.NFSExportPath.IsUnknown() {
		nfsInfo.NFSExportPath = data.NFSExportPath.ValueString()
	}

	// Create the export path request
	createReq := &models.CreateFileStorageExportPathRequest{
		Volume:           data.Volume.ValueString(),
		Description:      data.Description.ValueString(),
		Protocol:         models.NFSProtocolType(data.Protocol.ValueString()),
		NFSInfo:          nfsInfo,
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	result, err := r.client.CreateFileStorageExportPath(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create file storage export path, got error: %s", err))
		return
	}

	// Set the path ID from the response
	data.PathID = types.StringValue(result.PathID)
	data.ID = types.StringValue(result.PathID)

	// Refresh the export path data
	exportPath, err := r.client.GetFileStorageExportPath(ctx, result.PathID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage export path after creation, got error: %s", err))
		return
	}

	// Map the response to the model
	r.mapExportPathToModel(ctx, exportPath, &data, &resp.Diagnostics)

	tflog.Trace(ctx, "created a file storage export path")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageExportPathResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FileStorageExportPathResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	exportPath, err := r.client.GetFileStorageExportPath(ctx, data.PathID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage export path, got error: %s", err))
		return
	}

	// Map the response to the model
	r.mapExportPathToModel(ctx, exportPath, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageExportPathResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FileStorageExportPathResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the NFS info
	nfsInfo := &models.NFSExportInfo{
		DefaultAccessType: models.NFSAccessType(data.DefaultAccessType.ValueString()),
		DefaultUserSquash: models.NFSSquashType(data.DefaultUserSquash.ValueString()),
	}

	if !data.NFSExportPath.IsNull() && !data.NFSExportPath.IsUnknown() {
		nfsInfo.NFSExportPath = data.NFSExportPath.ValueString()
	}

	// Update the export path request
	updateReq := &models.UpdateFileStorageExportPathRequest{
		PathID:           data.PathID.ValueString(),
		Volume:           data.Volume.ValueString(),
		Description:      data.Description.ValueString(),
		Protocol:         models.NFSProtocolType(data.Protocol.ValueString()),
		NFSInfo:          nfsInfo,
		AvailabilityZone: data.AvailabilityZone.ValueString(),
	}

	err := r.client.UpdateFileStorageExportPath(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update file storage export path, got error: %s", err))
		return
	}

	// Refresh the export path data
	exportPath, err := r.client.GetFileStorageExportPath(ctx, data.PathID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read file storage export path after update, got error: %s", err))
		return
	}

	// Map the response to the model
	r.mapExportPathToModel(ctx, exportPath, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FileStorageExportPathResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FileStorageExportPathResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFileStorageExportPath(ctx, data.PathID.ValueString(), data.AvailabilityZone.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete file storage export path, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a file storage export path")
}

func (r *FileStorageExportPathResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: path_id
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// mapExportPathToModel maps the API response to the Terraform model
func (r *FileStorageExportPathResource) mapExportPathToModel(ctx context.Context, exportPath *models.FileStorageExportPath, data *FileStorageExportPathResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(exportPath.PathID)
	data.PathID = types.StringValue(exportPath.PathID)
	data.Volume = types.StringValue(exportPath.Volume)
	if exportPath.Description != "" {
		data.Description = types.StringValue(exportPath.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Protocol = types.StringValue(string(exportPath.Protocol))
	data.AvailabilityZone = types.StringValue(exportPath.AvailabilityZone)
	data.CreatedAt = types.StringValue(exportPath.CreatedAt)
	data.CreatedBy = types.StringValue(exportPath.CreatedBy)
	data.ProviderExpPathID = types.StringValue(exportPath.ProviderExpPathID)

	if exportPath.NFSInfo != nil {
		data.NFSExportPath = types.StringValue(exportPath.NFSInfo.NFSExportPath)
		data.DefaultAccessType = types.StringValue(string(exportPath.NFSInfo.DefaultAccessType))
		data.DefaultUserSquash = types.StringValue(string(exportPath.NFSInfo.DefaultUserSquash))
	}
}
