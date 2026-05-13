package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
)

var _ resource.Resource = &ComputeSnapshotResource{}
var _ resource.ResourceWithImportState = &ComputeSnapshotResource{}

func NewComputeSnapshotResource() resource.Resource {
	return &ComputeSnapshotResource{}
}

type ComputeSnapshotResource struct {
	client *client.Client
}

type ComputeSnapshotResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	ComputeID types.String   `tfsdk:"compute_id"`
	Name      types.String   `tfsdk:"name"`
	Status    types.String   `tfsdk:"status"`
	IsActive  types.Bool     `tfsdk:"is_active"`
	IsImage   types.Bool     `tfsdk:"is_image"`
	Created   types.String   `tfsdk:"created"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func (r *ComputeSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_snapshot"
}

func (r *ComputeSnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Compute Snapshot. Snapshots are immutable point-in-time copies of a VM.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compute_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the compute instance to snapshot.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the snapshot.",
				Computed:            true,
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether the snapshot is active.",
				Computed:            true,
			},
			"is_image": schema.BoolAttribute{
				MarkdownDescription: "Whether the snapshot has been converted to an image.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the snapshot.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *ComputeSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComputeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ComputeSnapshotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot, err := r.client.CreateComputeSnapshot(ctx, data.ComputeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create compute snapshot, got error: %s", err))
		return
	}

	// Wait for snapshot to be ready
	readySnapshot, err := r.client.WaitForSnapshotReady(ctx, snapshot.UUID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for snapshot to be ready: %s", err))
		return
	}

	data.ID = types.StringValue(readySnapshot.UUID)
	data.Name = types.StringValue(readySnapshot.SnapshotName)
	data.Status = types.StringValue(readySnapshot.Status)
	data.IsActive = types.BoolValue(readySnapshot.IsActive)
	data.IsImage = types.BoolValue(readySnapshot.IsImage)
	data.Created = types.StringValue(readySnapshot.Created)

	tflog.Trace(ctx, "created compute snapshot resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComputeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeSnapshotResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot, err := r.client.GetComputeSnapshot(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read compute snapshot, got error: %s", err))
		return
	}

	data.Name = types.StringValue(snapshot.SnapshotName)
	data.Status = types.StringValue(snapshot.Status)
	data.IsActive = types.BoolValue(snapshot.IsActive)
	data.IsImage = types.BoolValue(snapshot.IsImage)
	data.Created = types.StringValue(snapshot.Created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComputeSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Compute snapshots are immutable and cannot be updated.")
}

func (r *ComputeSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeSnapshotResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeSnapshot(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete compute snapshot, got error: %s", err))
		return
	}
}

func (r *ComputeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
