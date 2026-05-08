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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ObjectStorageAccessKeyResource{}
var _ resource.ResourceWithImportState = &ObjectStorageAccessKeyResource{}

func NewObjectStorageAccessKeyResource() resource.Resource {
	return &ObjectStorageAccessKeyResource{}
}

// ObjectStorageAccessKeyResource defines the resource implementation.
type ObjectStorageAccessKeyResource struct {
	client *client.Client
}

// ObjectStorageAccessKeyResourceModel describes the resource data model.
type ObjectStorageAccessKeyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Expiry          types.Int64  `tfsdk:"expiry"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretKey       types.String `tfsdk:"secret_key"`
	URL             types.String `tfsdk:"url"`
	CreateTime      types.Int64  `tfsdk:"create_time"`
	ExpiryTimestamp types.Int64  `tfsdk:"expiry_timestamp"`
}

func (r *ObjectStorageAccessKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_access_key"
}

func (r *ObjectStorageAccessKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud object storage access key for S3-compatible programmatic access.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the access key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expiry": schema.Int64Attribute{
				MarkdownDescription: "Duration in seconds for key validity (e.g., 2592000 for 30 days). Converted to a Unix timestamp at creation time.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "The access key identifier used for management operations.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "The S3 access key.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "The S3 secret key. Only fully available after creation.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The S3 endpoint URL.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_time": schema.Int64Attribute{
				MarkdownDescription: "Unix timestamp of when the access key was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expiry_timestamp": schema.Int64Attribute{
				MarkdownDescription: "The Unix timestamp when the access key expires.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ObjectStorageAccessKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ObjectStorageAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ObjectStorageAccessKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	expiryDuration := data.Expiry.ValueInt64()
	expiryTimestamp := time.Now().Unix() + expiryDuration

	createReq := &models.CreateAccessKeyRequest{
		Expiry: expiryTimestamp,
	}

	createResp, err := r.client.CreateAccessKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create access key, got error: %s", err))
		return
	}

	// The create response returns access_key but not access_key_id.
	// We need to list keys to find the one we just created.
	listResp, err := r.client.ListAccessKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list access keys after creation, got error: %s", err))
		return
	}

	// Find the key matching the access_key from create response
	var foundKey *models.AccessKey
	for i := range listResp.Items {
		if listResp.Items[i].AccessKey == createResp.AccessKey {
			foundKey = &listResp.Items[i]
			break
		}
	}

	if foundKey == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to find newly created access key in list response")
		return
	}

	data.ID = types.StringValue(foundKey.AccessKeyID)
	data.AccessKeyID = types.StringValue(foundKey.AccessKeyID)
	data.AccessKey = types.StringValue(createResp.AccessKey)
	data.SecretKey = types.StringValue(createResp.SecretKey)
	data.URL = types.StringValue(createResp.URL)
	data.CreateTime = types.Int64Value(int64(foundKey.CreateTime))
	data.ExpiryTimestamp = types.Int64Value(int64(foundKey.Expiry))

	tflog.Trace(ctx, "created an object storage access key resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectStorageAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ObjectStorageAccessKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := r.client.ListAccessKeys(ctx)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list access keys, got error: %s", err))
		return
	}

	// Find the key matching our access_key_id
	accessKeyID := data.AccessKeyID.ValueString()
	var foundKey *models.AccessKey
	for i := range listResp.Items {
		if listResp.Items[i].AccessKeyID == accessKeyID {
			foundKey = &listResp.Items[i]
			break
		}
	}

	if foundKey == nil {
		// Key no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(foundKey.AccessKeyID)
	data.AccessKeyID = types.StringValue(foundKey.AccessKeyID)
	data.AccessKey = types.StringValue(foundKey.AccessKey)
	// Preserve secret_key from state since the list API may not return it fully
	data.CreateTime = types.Int64Value(int64(foundKey.CreateTime))
	data.ExpiryTimestamp = types.Int64Value(int64(foundKey.Expiry))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectStorageAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update supported — expiry has RequiresReplace, so this should never be called.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Access keys cannot be updated. Changes to expiry require replacement.",
	)
}

func (r *ObjectStorageAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ObjectStorageAccessKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAccessKey(ctx, data.AccessKeyID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			// Already deleted
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete access key, got error: %s", err))
		return
	}
}

func (r *ObjectStorageAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("access_key_id"), req, resp)
}
