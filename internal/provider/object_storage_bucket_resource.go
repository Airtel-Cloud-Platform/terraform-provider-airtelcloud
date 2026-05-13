package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ObjectStorageBucketResource{}
var _ resource.ResourceWithImportState = &ObjectStorageBucketResource{}
var _ resource.ResourceWithValidateConfig = &ObjectStorageBucketResource{}

func NewObjectStorageBucketResource() resource.Resource {
	return &ObjectStorageBucketResource{}
}

// ObjectStorageBucketResource defines the resource implementation.
type ObjectStorageBucketResource struct {
	client *client.Client
}

// ObjectStorageBucketResourceModel describes the resource data model.
type ObjectStorageBucketResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ReplicationType  types.String `tfsdk:"replication_type"`
	ReplicationTag   types.String `tfsdk:"replication_tag"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Versioning       types.Bool   `tfsdk:"versioning"`
	ObjectLocking    types.Bool   `tfsdk:"object_locking"`
	Tags             types.Map    `tfsdk:"tags"`
	// Computed fields from response
	S3Endpoint     types.String `tfsdk:"s3_endpoint"`
	PublicEndpoint types.String `tfsdk:"public_endpoint"`
	State          types.String `tfsdk:"state"`
}

func (r *ObjectStorageBucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket"
}

func (r *ObjectStorageBucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud object storage bucket.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the bucket. Must be globally unique.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replication_type": schema.StringAttribute{
				MarkdownDescription: "The replication type for the bucket. Valid values: `Local`, `Replicated within region`, `Replicated across region`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Local", "Replicated within region", "Replicated across region"),
				},
			},
			"replication_tag": schema.StringAttribute{
				MarkdownDescription: "The replication tag for the bucket. Valid values: `north_N1`, `north_N2`, `north_N1_N2`, `north_N2_N1`, `north_south_N1_S1`, `north_south_N2_S2`, `south_S1`, `south_S2`, `south_S1_S2`, `south_S2_S1`, `south_north_S1_N1`, `south_north_S2_N2`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"north_N1",
						"north_N2",
						"north_N1_N2",
						"north_N2_N1",
						"north_south_N1_S1",
						"north_south_N2_S2",
						"south_S1",
						"south_S2",
						"south_S1_S2",
						"south_S2_S1",
						"south_north_S1_N1",
						"south_north_S2_N2",
					),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone where the object storage bucket will be created (e.g., 'S1').",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"versioning": schema.BoolAttribute{
				MarkdownDescription: "Whether versioning is enabled for the bucket.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"object_locking": schema.BoolAttribute{
				MarkdownDescription: "Whether object locking is enabled for the bucket.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of tags to assign to the bucket.",
				Optional:            true,
			},
			// Computed fields from API response
			"s3_endpoint": schema.StringAttribute{
				MarkdownDescription: "The S3 endpoint for the bucket.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_endpoint": schema.StringAttribute{
				MarkdownDescription: "The public endpoint for the bucket.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the bucket.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ObjectStorageBucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ObjectStorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ObjectStorageBucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags map to map[string]string
	var tags map[string]string
	if !data.Tags.IsNull() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build replication config
	replication := &models.BucketReplicationConfig{
		ReplicationType: data.ReplicationType.ValueString(),
		AZ:              data.AvailabilityZone.ValueString(),
		Tag:             data.ReplicationTag.ValueString(),
	}

	// Build config
	config := &models.BucketCreateConfig{
		Versioning:  data.Versioning.ValueBool(),
		ObjLocking:  data.ObjectLocking.ValueBool(),
		Replication: replication,
	}

	createReq := &models.CreateObjectStorageBucketRequest{
		Bucket: data.Name.ValueString(),
		Config: config,
		Tags:   tags,
	}

	_, err := r.client.CreateObjectStorageBucket(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create bucket, got error: %s", err))
		return
	}

	// Always do a fresh GET to retrieve complete bucket state
	bucket, err := r.client.GetObjectStorageBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read bucket after creation, got error: %s", err))
		return
	}

	// Update computed fields from the API response.
	// Keep versioning/object_locking from the plan — the API may not return
	// these fields reliably right after creation.
	data.ID = types.StringValue(bucket.Name)
	data.State = types.StringValue(bucket.State)
	data.S3Endpoint = types.StringValue(bucket.S3Endpoint)
	data.PublicEndpoint = types.StringValue(bucket.PublicEndpoint)

	tflog.Trace(ctx, "created an object storage bucket resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectStorageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ObjectStorageBucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetObjectStorageBucket(ctx, data.Name.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read bucket, got error: %s", err))
		return
	}

	// Update the model with the current bucket data
	data.ID = types.StringValue(bucket.Name) // Use name as ID
	data.Name = types.StringValue(bucket.Name)
	data.State = types.StringValue(bucket.State)
	data.S3Endpoint = types.StringValue(bucket.S3Endpoint)
	data.PublicEndpoint = types.StringValue(bucket.PublicEndpoint)

	// Only update versioning/object_locking if the API actually returned them
	if bucket.Versioning != nil {
		data.Versioning = types.BoolValue(bool(*bucket.Versioning))
	}
	if bucket.ObjLocking != nil {
		data.ObjectLocking = types.BoolValue(bool(*bucket.ObjLocking))
	}

	// Map replication config to flat attributes
	if bucket.ReplicationConfig != nil {
		data.ReplicationType = types.StringValue(bucket.ReplicationConfig.ReplicationType)
		data.ReplicationTag = types.StringValue(bucket.ReplicationConfig.Tag)
		data.AvailabilityZone = types.StringValue(bucket.ReplicationConfig.AZ)
	}

	// Convert tags to Terraform map
	if len(bucket.Tags) > 0 {
		tagsValue, diags := types.MapValueFrom(ctx, types.StringType, bucket.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Tags = tagsValue
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectStorageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ObjectStorageBucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags map to map[string]string
	var tags map[string]string
	if !data.Tags.IsNull() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	versioning := data.Versioning.ValueBool()
	objLocking := data.ObjectLocking.ValueBool()

	updateReq := &models.UpdateObjectStorageBucketRequest{
		Versioning: &versioning,
		ObjLocking: &objLocking,
		Tags:       tags,
	}

	_, err := r.client.UpdateObjectStorageBucket(ctx, data.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bucket, got error: %s", err))
		return
	}

	// Always do a fresh GET to retrieve complete bucket state
	bucket, err := r.client.GetObjectStorageBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read bucket after update, got error: %s", err))
		return
	}

	// Update computed fields from the API response.
	// Keep versioning/object_locking from the plan — the API may not return
	// these fields reliably right after an update.
	data.ID = types.StringValue(bucket.Name)
	data.State = types.StringValue(bucket.State)
	data.S3Endpoint = types.StringValue(bucket.S3Endpoint)
	data.PublicEndpoint = types.StringValue(bucket.PublicEndpoint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectStorageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ObjectStorageBucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteObjectStorageBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete bucket, got error: %s", err))
		return
	}
}

func (r *ObjectStorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ValidateConfig validates the resource configuration.
func (r *ObjectStorageBucketResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ObjectStorageBucketResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if values are unknown (e.g., during plan with variables)
	if data.ReplicationType.IsUnknown() || data.ReplicationTag.IsUnknown() {
		return
	}

	replicationType := data.ReplicationType.ValueString()
	replicationTag := data.ReplicationTag.ValueString()

	// Define valid replication_tag values for each replication_type
	validTagsByType := map[string][]string{
		"Local": {"north_N1", "north_N2", "south_S1", "south_S2"},
		"Replicated within region": {
			"north_N1_N2", "north_N2_N1",
			"south_S1_S2", "south_S2_S1",
		},
		"Replicated across region": {
			"north_south_N1_S1", "north_south_N2_S2",
			"south_north_S1_N1", "south_north_S2_N2",
		},
	}

	validTags, ok := validTagsByType[replicationType]
	if !ok {
		// replication_type validation is handled by the schema validator
		return
	}

	// Check if replicationTag is in the valid list
	isValid := false
	for _, tag := range validTags {
		if replicationTag == tag {
			isValid = true
			break
		}
	}

	if !isValid {
		resp.Diagnostics.AddAttributeError(
			path.Root("replication_tag"),
			"Invalid Replication Tag",
			fmt.Sprintf("When replication_type is %q, replication_tag must be one of: %v. Got: %q",
				replicationType, validTags, replicationTag),
		)
	}
}
