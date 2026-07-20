package provider

import (
	"context"
	"fmt"
	"strings"
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
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &ComputeSnapshotResource{}
var _ resource.ResourceWithImportState = &ComputeSnapshotResource{}
var _ resource.ResourceWithValidateConfig = &ComputeSnapshotResource{}

func NewComputeSnapshotResource() resource.Resource {
	return &ComputeSnapshotResource{}
}

type ComputeSnapshotResource struct {
	client *client.Client
}

type ComputeSnapshotResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	ComputeID    types.String   `tfsdk:"compute_id"`
	ComputeName  types.String   `tfsdk:"compute_name"`
	SnapshotName types.String   `tfsdk:"snapshot_name"`
	Name         types.String   `tfsdk:"name"`
	Status       types.String   `tfsdk:"status"`
	IsActive     types.Bool     `tfsdk:"is_active"`
	IsImage      types.Bool     `tfsdk:"is_image"`
	ImageID      types.String   `tfsdk:"image_id"`
	Created      types.String   `tfsdk:"created"`
	Updated      types.String   `tfsdk:"updated"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
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
				MarkdownDescription: "The ID of the compute instance to snapshot. Either compute_id or compute_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_name": schema.StringAttribute{
				MarkdownDescription: "The name of the compute instance to snapshot. If set, it is resolved to compute_id. Either compute_id or compute_name must be specified.",
				Optional:            true,
			},
			"snapshot_name": schema.StringAttribute{
				MarkdownDescription: "The name to give the snapshot.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot as reported by the API.",
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
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the image created from the snapshot.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the snapshot.",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "The last update timestamp of the snapshot.",
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

// ValidateConfig enforces that exactly one of compute_id or compute_name is set.
func (r *ComputeSnapshotResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ComputeSnapshotResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ComputeID.IsNull() && !data.ComputeName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of compute_id or compute_name may be specified, not both.")
	}
	if data.ComputeID.IsNull() && data.ComputeName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of compute_id or compute_name must be specified.")
	}
}

// applySnapshot copies API-owned fields into state. It deliberately leaves
// snapshot_name alone: during Create that attribute is config-owned, and
// overwriting it from the API response would trip Terraform's "inconsistent
// result after apply" check if the backend ever normalises the value. Read
// seeds snapshot_name separately (see applySnapshotForRead), which is what makes
// drift detection and composite-ID import work.
func applySnapshot(data *ComputeSnapshotResourceModel, snapshot *models.ComputeSnapshot) {
	data.ID = types.StringValue(snapshot.UUID)
	data.Name = types.StringValue(snapshot.SnapshotName)
	data.Status = types.StringValue(snapshot.Status)
	data.IsActive = types.BoolValue(snapshot.IsActive)
	data.IsImage = types.BoolValue(snapshot.IsImage)
	data.ImageID = types.StringValue(snapshot.ImageIDString())
	data.Created = types.StringValue(snapshot.Created)
	data.Updated = types.StringValue(snapshot.Updated)
}

// applySnapshotForRead is applySnapshot plus the config-owned snapshot_name,
// which Read must populate so that drift detection and imports (where
// snapshot_name starts null) produce a clean plan.
func applySnapshotForRead(data *ComputeSnapshotResourceModel, snapshot *models.ComputeSnapshot) {
	applySnapshot(data, snapshot)
	data.SnapshotName = types.StringValue(snapshot.SnapshotName)
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
	deadline := time.Now().Add(createTimeout)

	computeID := data.ComputeID.ValueString()
	if computeID == "" && !data.ComputeName.IsNull() {
		resolved, err := r.client.ResolveComputeID(ctx, data.ComputeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve compute name %q: %s", data.ComputeName.ValueString(), err))
			return
		}
		computeID = resolved
	}
	// Persist the resolved (or user-supplied) id so the Computed compute_id attribute
	// is known in state regardless of whether the user set compute_id or compute_name.
	data.ComputeID = types.StringValue(computeID)

	snapshotName := data.SnapshotName.ValueString()

	// Scope the client once: the snapshot detail endpoints need the subnet-id
	// header, and reusing this client keeps the poll loop from re-fetching the
	// compute on every iteration.
	snapClient, err := r.client.SnapshotClientForCompute(ctx, computeID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve compute %s for snapshot: %s", computeID, err))
		return
	}

	// Record the snapshots that already exist, so that if the create response
	// carries no UUID we can identify the new one by elimination rather than by
	// name alone.
	existing := map[string]struct{}{}
	if snapshots, err := snapClient.ListComputeSnapshotsForCompute(ctx, computeID); err != nil {
		tflog.Warn(ctx, "unable to list existing snapshots before create", map[string]interface{}{"error": err.Error()})
	} else {
		for _, snapshot := range snapshots {
			existing[snapshot.UUID] = struct{}{}
		}
	}

	snapshot, err := snapClient.CreateComputeSnapshot(ctx, computeID, &models.CreateComputeSnapshotRequest{
		SnapshotName: snapshotName,
		ComputeID:    computeID,
		BillingUnit:  client.BillingUnitHourly,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create compute snapshot, got error: %s", err))
		return
	}

	snapshotUUID := ""
	if snapshot != nil {
		snapshotUUID = snapshot.UUID
	}
	if snapshotUUID == "" {
		tflog.Warn(ctx, "compute snapshot create response carried no UUID; locating the snapshot by name")

		snapshotUUID, err = snapClient.ResolveNewSnapshotUUID(ctx, computeID, snapshotName, existing, time.Until(deadline))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Snapshot was created but could not be located: %s", err))
			return
		}
	}

	ready, err := snapClient.WaitForSnapshotReady(ctx, snapshotUUID, time.Until(deadline))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for snapshot to be ready: %s", err))
		return
	}

	applySnapshot(&data, ready)

	tflog.Trace(ctx, "created compute snapshot resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComputeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeSnapshotResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshotUUID := data.ID.ValueString()
	computeID := data.ComputeID.ValueString()

	// Imported by bare UUID: the owning compute is unknown, so we cannot build
	// the subnet-id header the detail endpoint needs. The project-wide list
	// needs no header and embeds the compute, so recover compute_id from there.
	if computeID == "" {
		snapshot, err := r.client.FindComputeSnapshotByUUID(ctx, snapshotUUID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read compute snapshot, got error: %s", err))
			return
		}
		if snapshot == nil {
			resp.State.RemoveResource(ctx)
			return
		}

		computeID = snapshot.ComputeIDString()
		if computeID == "" {
			resp.Diagnostics.AddError(
				"Unable to determine compute_id",
				fmt.Sprintf("Snapshot %s does not report its compute instance. Import it as \"<compute_id>:<snapshot_uuid>\" instead.", snapshotUUID),
			)
			return
		}

		data.ComputeID = types.StringValue(computeID)
		applySnapshotForRead(&data, snapshot)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	snapClient, err := r.client.SnapshotClientForCompute(ctx, computeID)
	if err != nil {
		if client.IsNotFoundError(err) {
			// The VM is gone, but a snapshot can outlive it. Fall back to the
			// header-free list rather than assuming the snapshot is gone too.
			r.readFromList(ctx, &data, resp)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve compute %s for snapshot: %s", computeID, err))
		return
	}

	snapshot, err := snapClient.GetComputeSnapshot(ctx, snapshotUUID)
	if err != nil {
		if client.IsProviderLookupError(err) {
			// A 404 "Failed to get provider" means the header did not resolve,
			// not that the snapshot is gone. Confirm against the list.
			r.readFromList(ctx, &data, resp)
			return
		}
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read compute snapshot, got error: %s", err))
		return
	}

	applySnapshotForRead(&data, snapshot)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readFromList refreshes state from the header-free project-wide list, used
// when the compute-scoped detail GET is unavailable.
func (r *ComputeSnapshotResource) readFromList(ctx context.Context, data *ComputeSnapshotResourceModel, resp *resource.ReadResponse) {
	snapshot, err := r.client.FindComputeSnapshotByUUID(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read compute snapshot, got error: %s", err))
		return
	}
	if snapshot == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applySnapshotForRead(data, snapshot)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
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

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapClient, err := r.client.SnapshotClientForCompute(ctx, data.ComputeID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			// The compute is already gone; nothing left to scope a delete with.
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve compute for snapshot deletion: %s", err))
		return
	}

	if err := snapClient.DeleteComputeSnapshot(ctx, data.ID.ValueString(), deleteTimeout); err != nil {
		if client.IsNotFoundError(err) && !client.IsProviderLookupError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete compute snapshot, got error: %s", err))
		return
	}
}

// ImportState accepts either "<snapshot_uuid>", in which case Read recovers the
// compute from the project-wide list, or "<compute_id>:<snapshot_uuid>", which
// skips that scan.
func (r *ComputeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	switch len(parts) {
	case 1:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[0])...)
	case 2:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("compute_id"), parts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	default:
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected \"<snapshot_uuid>\" or \"<compute_id>:<snapshot_uuid>\", got: %q", req.ID),
		)
	}
}
