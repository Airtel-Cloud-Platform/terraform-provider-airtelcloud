package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &BaremetalResource{}
var _ resource.ResourceWithImportState = &BaremetalResource{}

func NewBaremetalResource() resource.Resource {
	return &BaremetalResource{}
}

type BaremetalResource struct {
	client *client.Client
}

type BaremetalResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	UUID             types.String `tfsdk:"uuid"`
	Flavor           types.String `tfsdk:"flavor"`
	OSImage          types.String `tfsdk:"os_image"`
	CloudInit        types.String `tfsdk:"cloud_init"`
	SubnetID         types.String `tfsdk:"subnet_id"`
	NetworkName      types.String `tfsdk:"network_name"`
	IsReserved       types.Bool   `tfsdk:"is_reserved"`
	SystemID         types.String `tfsdk:"system_id"`
	Keypair          types.String `tfsdk:"keypair"`
	KeypairID        types.String `tfsdk:"keypair_id"`
	PublicKey        types.String `tfsdk:"public_key"`
	Tags             types.List   `tfsdk:"tags"`
	PolicyEnabled    types.Bool   `tfsdk:"policy_enabled"`
	DeleteDisks      types.Bool   `tfsdk:"delete_disks"`
	SecureErase      types.Bool   `tfsdk:"secure_erase"`
	State            types.String `tfsdk:"state"`
	Hostname         types.String `tfsdk:"hostname"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Power            types.String `tfsdk:"power"`
	IPAddresses      types.List   `tfsdk:"ip_addresses"`
	BackendPortID    types.Int64  `tfsdk:"backend_port_id"`
}

func (r *BaremetalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_baremetal"
}

func (r *BaremetalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allocates and manages an Airtel Cloud baremetal server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource id. This provider uses the baremetal name as the id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Baremetal server name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Baremetal server UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flavor": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Flavor profile to allocate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"os_image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OS image name to install.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_init": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cloud-init script for first boot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Subnet id for primary network interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Availability zone used by baremetal APIs (for example `N1`, `N2`, `S1`, `S2`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional network interface name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_reserved": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to allocate from reserved capacity.",
			},
			"system_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional system id used by the backend for reservation/release.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"keypair": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional SSH keypair name to inject.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"keypair_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional keypair UUID, sent as `keypairId` in baremetal allocate API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional SSH public key, sent as `publicKey` in baremetal allocate API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Optional list of tags.",
			},
			"policy_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Mutable backup policy toggle sent to update API.",
			},
			"delete_disks": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to delete disks on resource destroy.",
			},
			"secure_erase": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to perform secure erase on resource destroy.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current server state.",
			},
			"hostname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resolved hostname from API.",
			},
			"power": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current power state.",
			},
			"ip_addresses": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Assigned IP addresses.",
			},
			"backend_port_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Backend port id from networkInfo.portId.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *BaremetalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BaremetalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BaremetalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allocateReq := &models.AllocateBaremetalRequest{
		Name:       data.Name.ValueString(),
		Flavor:     data.Flavor.ValueString(),
		OSImage:    data.OSImage.ValueString(),
		CloudInit:  data.CloudInit.ValueString(),
		KeypairID:  data.KeypairID.ValueString(),
		PublicKey:  data.PublicKey.ValueString(),
		IsReserved: data.IsReserved.ValueBool(),
		SystemID:   data.SystemID.ValueString(),
	}

	if !data.NetworkName.IsNull() || !data.SubnetID.IsNull() {
		subnetID := data.SubnetID.ValueString()
		allocateReq.NetworkInterface = &models.BaremetalNetworkInterface{
			Name:     data.NetworkName.ValueString(),
			SubnetID: subnetID,
			Subnets: []models.BaremetalSubnetConfig{
				{
					SubnetID:  subnetID,
					IsPrimary: true,
				},
			},
		}
	}

	if !data.Keypair.IsNull() && data.Keypair.ValueString() != "" {
		allocateReq.Metadata = &models.BaremetalMetadata{Keypair: data.Keypair.ValueString()}
	}

	if !data.Tags.IsNull() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		allocateReq.Tags = tags
	}

	tflog.Debug(ctx, "Baremetal create request prepared", map[string]interface{}{
		"name":              allocateReq.Name,
		"flavor":            allocateReq.Flavor,
		"os_image":          allocateReq.OSImage,
		"availability_zone": data.AvailabilityZone.ValueString(),
		"subnet_id":         data.SubnetID.ValueString(),
		"network_name":      data.NetworkName.ValueString(),
		"is_reserved":       allocateReq.IsReserved,
		"has_system_id":     strings.TrimSpace(allocateReq.SystemID) != "",
		"has_keypair":       allocateReq.Metadata != nil && strings.TrimSpace(allocateReq.Metadata.Keypair) != "",
		"has_keypair_id":    strings.TrimSpace(allocateReq.KeypairID) != "",
		"has_public_key":    strings.TrimSpace(allocateReq.PublicKey) != "",
		"tags_count":        len(allocateReq.Tags),
		"has_cloud_init":    strings.TrimSpace(allocateReq.CloudInit) != "",
	})

	bmClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
	if err := bmClient.AllocateBaremetal(ctx, allocateReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to allocate baremetal server, got error: %s", err))
		return
	}
	tflog.Debug(ctx, "Baremetal allocate request accepted by API", map[string]interface{}{
		"name":              data.Name.ValueString(),
		"availability_zone": data.AvailabilityZone.ValueString(),
	})

	data.ID = types.StringValue(data.Name.ValueString())

	// The allocate API is asynchronous and often returns before the server has a
	// meaningful lifecycle state. Wait for a stable non-placeholder state.
	if err := r.waitForBaremetalState(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 5*time.Minute); err != nil {
		resp.Diagnostics.AddError("Baremetal Provisioning Error", err.Error())
		return
	}

	if err := r.readIntoState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal server after create, got error: %s", err))
		return
	}

	if strings.EqualFold(strings.TrimSpace(data.State.ValueString()), "NoResource") {
		tflog.Warn(ctx, "Baremetal create finished in placeholder state", map[string]interface{}{
			"name":              data.Name.ValueString(),
			"availability_zone": data.AvailabilityZone.ValueString(),
			"state":             data.State.ValueString(),
			"power":             data.Power.ValueString(),
		})
		resp.Diagnostics.AddError(
			"Baremetal Provisioning Error",
			fmt.Sprintf("Baremetal server %q was created but remains in placeholder state %q. This indicates backend provisioning did not complete properly.", data.Name.ValueString(), data.State.ValueString()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BaremetalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() || data.Name.ValueString() == "" {
		data.Name = types.StringValue(data.ID.ValueString())
	}

	if err := r.readIntoState(ctx, &data); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal server, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BaremetalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.PolicyEnabled.IsNull() {
		enabled := data.PolicyEnabled.ValueBool()
		updateReq := &models.UpdateBaremetalRequest{PolicyEnabled: &enabled}
		bmClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
		if err := bmClient.UpdateBaremetal(ctx, data.Name.ValueString(), updateReq); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update baremetal server, got error: %s", err))
			return
		}
	}

	if err := r.readIntoState(ctx, &data); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal server after update, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BaremetalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BaremetalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	if name == "" {
		name = data.ID.ValueString()
	}

	opts := &models.ReleaseBaremetalOptions{}
	if !data.SystemID.IsNull() && data.SystemID.ValueString() != "" {
		opts.SystemID = data.SystemID.ValueString()
	}
	if !data.DeleteDisks.IsNull() {
		v := data.DeleteDisks.ValueBool()
		opts.DeleteDisks = &v
	}
	if !data.SecureErase.IsNull() {
		v := data.SecureErase.ValueBool()
		opts.SecureErase = &v
	}

	bmClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
	if err := bmClient.ReleaseBaremetal(ctx, name, opts); err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to release baremetal server, got error: %s", err))
		return
	}
}

func (r *BaremetalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *BaremetalResource) readIntoState(ctx context.Context, data *BaremetalResourceModel) error {
	name := data.Name.ValueString()
	if name == "" {
		name = data.ID.ValueString()
	}

	az := data.AvailabilityZone.ValueString()
	if az == "" {
		resolvedAZ, err := r.resolveBaremetalAZ(ctx, name, 30*time.Second)
		if err != nil {
			return err
		}
		az = resolvedAZ
	}

	bm, err := r.client.GetBaremetal(ctx, name, az)
	if err != nil {
		return err
	}

	// The detail endpoint may return placeholder state (for example "NoResource")
	// and can omit fields like name/ips in some stages. Merge with list data.
	if summary, err := r.findBaremetalSummary(ctx, name, bm.UUID); err == nil && summary != nil {
		if bm.Name == "" {
			bm.Name = summary.Name
		}
		if bm.UUID == "" {
			bm.UUID = summary.UUID
		}
		if bm.Flavor == "" {
			bm.Flavor = summary.Flavor
		}
		if bm.Hostname == "" {
			bm.Hostname = summary.Hostname
		}
		if bm.AvailabilityZone == "" {
			bm.AvailabilityZone = summary.AvailabilityZone
		}
		if bm.PowerState == "" {
			bm.PowerState = summary.PowerState
		}
		if len(bm.IPAddr) == 0 {
			bm.IPAddr = summary.IPAddr
		}
		if bm.State == "" || strings.EqualFold(bm.State, "NoResource") {
			if summary.State != "" {
				bm.State = summary.State
			}
		}
	}

	data.ID = types.StringValue(name)
	if bm.Name == "" {
		data.Name = types.StringValue(name)
	} else {
		data.Name = types.StringValue(bm.Name)
	}
	if bm.UUID != "" {
		data.UUID = types.StringValue(bm.UUID)
	} else {
		data.UUID = types.StringNull()
	}
	if bm.Flavor != "" {
		data.Flavor = types.StringValue(bm.Flavor)
	}
	state := strings.TrimSpace(bm.State)
	if state != "" {
		data.State = types.StringValue(state)
	} else {
		data.State = types.StringNull()
	}
	if bm.Hostname != "" {
		data.Hostname = types.StringValue(bm.Hostname)
	} else {
		data.Hostname = types.StringNull()
	}
	if bm.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(bm.AvailabilityZone)
	}
	power := strings.TrimSpace(bm.PowerState)
	if power != "" {
		data.Power = types.StringValue(power)
	} else {
		data.Power = types.StringNull()
	}

	portID := int64(bm.NetworkInfo.PortID)
	if portID == 0 {
		data.BackendPortID = types.Int64Null()
	} else {
		data.BackendPortID = types.Int64Value(portID)
	}

	ips, diags := types.ListValueFrom(ctx, types.StringType, bm.IPAddr)
	if diags.HasError() {
		return fmt.Errorf("failed to map ip_addresses")
	}
	data.IPAddresses = ips

	return nil
}

func (r *BaremetalResource) findBaremetalSummary(ctx context.Context, name, uuid string) (*models.Baremetal, error) {
	items, err := r.client.ListBaremetals(ctx)
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].Name == name {
			return &items[i], nil
		}
	}
	if uuid != "" {
		for i := range items {
			if items[i].UUID == uuid {
				return &items[i], nil
			}
		}
	}

	return nil, nil
}

func (r *BaremetalResource) waitForBaremetalState(ctx context.Context, name, az string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	start := time.Now()
	lastState := ""
	lastPower := ""
	stableProvisionedReads := 0
	attempt := 0
	for {
		attempt++
		if time.Now().After(deadline) {
			tflog.Warn(ctx, "Baremetal provisioning timeout", map[string]interface{}{
				"name":              name,
				"availability_zone": az,
				"attempt":           attempt,
				"elapsed_seconds":   int(time.Since(start).Seconds()),
				"stable_reads":      stableProvisionedReads,
				"last_state":        lastState,
				"last_power":        lastPower,
				"timeout_seconds":   int(timeout.Seconds()),
			})
			return fmt.Errorf("baremetal server %q did not reach a valid provisioning state within %s (last state=%q, last power=%q)", name, timeout.String(), lastState, lastPower)
		}

		detailState := ""
		detailPower := ""
		detailErr := ""
		detail, err := r.client.GetBaremetal(ctx, name, az)
		if err == nil {
			state := strings.TrimSpace(detail.State)
			power := strings.TrimSpace(detail.PowerState)
			detailState = state
			detailPower = power
			if state != "" {
				lastState = state
			}
			if power != "" {
				lastPower = power
			}
			if isProvisionedBaremetalState(state) {
				stableProvisionedReads++
				if stableProvisionedReads >= 2 {
					return nil
				}
			} else {
				stableProvisionedReads = 0
			}
		} else {
			detailErr = err.Error()
		}

		summaryState := ""
		summaryPower := ""
		summaryErr := ""
		summary, err := r.findBaremetalSummary(ctx, name, "")
		if err == nil && summary != nil {
			state := strings.TrimSpace(summary.State)
			power := strings.TrimSpace(summary.PowerState)
			summaryState = state
			summaryPower = power
			if state != "" {
				lastState = state
			}
			if power != "" {
				lastPower = power
			}
			if isProvisionedBaremetalState(state) {
				stableProvisionedReads++
				if stableProvisionedReads >= 2 {
					return nil
				}
			} else {
				stableProvisionedReads = 0
			}
		} else if err != nil {
			summaryErr = err.Error()
		}

		tflog.Debug(ctx, "Baremetal provisioning poll", map[string]interface{}{
			"name":              name,
			"availability_zone": az,
			"attempt":           attempt,
			"elapsed_seconds":   int(time.Since(start).Seconds()),
			"stable_reads":      stableProvisionedReads,
			"detail_state":      detailState,
			"detail_power":      detailPower,
			"detail_error":      detailErr,
			"summary_state":     summaryState,
			"summary_power":     summaryPower,
			"summary_error":     summaryErr,
			"last_state":        lastState,
			"last_power":        lastPower,
		})

		select {
		case <-ctx.Done():
			tflog.Warn(ctx, "Baremetal provisioning wait interrupted", map[string]interface{}{
				"name":              name,
				"availability_zone": az,
				"attempt":           attempt,
				"last_state":        lastState,
				"last_power":        lastPower,
				"error":             ctx.Err().Error(),
			})
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

func isProvisionedBaremetalState(state string) bool {
	trimmed := strings.TrimSpace(state)
	if trimmed == "" {
		return false
	}
	if strings.EqualFold(trimmed, "NoResource") {
		return false
	}
	return true
}

func (r *BaremetalResource) resolveBaremetalAZ(ctx context.Context, name string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for {
		servers, err := r.client.ListBaremetals(ctx)
		if err != nil {
			return "", err
		}

		for _, s := range servers {
			if s.Name == name {
				if s.AvailabilityZone == "" {
					return "", fmt.Errorf("baremetal server %q was found but availability zone is empty", name)
				}
				return s.AvailabilityZone, nil
			}
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("baremetal server %q not found", name)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
