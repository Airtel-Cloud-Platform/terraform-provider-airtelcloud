package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	UUID                  types.String `tfsdk:"uuid"`
	Flavor                types.String `tfsdk:"flavor"`
	OSImage               types.String `tfsdk:"os_image"`
	CloudInit             types.String `tfsdk:"cloud_init"`
	SubnetName            types.String `tfsdk:"subnet_name"`
	NetworkName           types.String `tfsdk:"network_name"`
	IsReserved            types.Bool   `tfsdk:"is_reserved"`
	SystemID              types.String `tfsdk:"system_id"`
	Keypair               types.String `tfsdk:"keypair"`
	KeypairID             types.String `tfsdk:"keypair_id"`
	PublicKey             types.String `tfsdk:"public_key"`
	AdditionalSubnetNames types.List   `tfsdk:"additional_subnet_names"`
	Tags                  types.List   `tfsdk:"tags"`
	Storage               types.List   `tfsdk:"storage"`
	BackupConfig          types.Object `tfsdk:"backup_config"`
	PolicyEnabled         types.Bool   `tfsdk:"policy_enabled"`
	DeleteDisks           types.Bool   `tfsdk:"delete_disks"`
	SecureErase           types.Bool   `tfsdk:"secure_erase"`
	State                 types.String `tfsdk:"state"`
	Hostname              types.String `tfsdk:"hostname"`
	AvailabilityZone      types.String `tfsdk:"availability_zone"`
	Power                 types.String `tfsdk:"power"`
	IPAddresses           types.List   `tfsdk:"ip_addresses"`
	BackendPortID         types.Int64  `tfsdk:"backend_port_id"`
}

type baremetalStorageModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.String `tfsdk:"size"`
	Path        types.String `tfsdk:"path"`
	Type        types.String `tfsdk:"type"`
	FileSystem  types.String `tfsdk:"file_system"`
	ForceFormat types.Bool   `tfsdk:"force_format"`
}

type baremetalBackupConfigModel struct {
	ScheduleType      types.String `tfsdk:"schedule_type"`
	StartTime         types.String `tfsdk:"start_time"`
	IncrDays          types.List   `tfsdk:"incr_days"`
	FullDays          types.List   `tfsdk:"full_days"`
	FullRetention     types.Int64  `tfsdk:"full_retention"`
	FullRetentionUnit types.String `tfsdk:"full_retention_unit"`
	BackupSelections  types.List   `tfsdk:"backup_selections"`
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

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
			"subnet_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Primary subnet display name. Resolved to `subnetId` before POST /server.",
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
				MarkdownDescription: "VPC name or UUID sent as `networkInterface.name`. Names are resolved to the VPC UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"additional_subnet_names": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extra subnet display names appended to `networkInterface.subnets` after the primary subnet.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"storage": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Optional extra disks sent as allocate `storage`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Storage volume name.",
						},
						"size": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Size as sent by the console (for example `10`).",
						},
						"path": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Mount path.",
						},
						"type": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Storage type (for example `BlockStorage`).",
						},
						"file_system": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Filesystem (for example `xfs`).",
						},
						"force_format": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Whether to force format the volume.",
						},
					},
				},
			},
			"backup_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Optional backup schedule sent as allocate `backupConfig`.",
				Attributes: map[string]schema.Attribute{
					"schedule_type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Schedule type (for example `weekly_full`).",
					},
					"start_time": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Backup start time (for example `21:00`).",
					},
					"incr_days": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.Int64Type,
						MarkdownDescription: "Incremental backup days.",
					},
					"full_days": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.Int64Type,
						MarkdownDescription: "Full backup days.",
					},
					"full_retention": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Full backup retention count.",
					},
					"full_retention_unit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Full backup retention unit (for example `MONTHS`).",
					},
					"backup_selections": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Paths included in backup.",
					},
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
		KeypairID:  data.KeypairID.ValueString(),
		PublicKey:  data.PublicKey.ValueString(),
		IsReserved: data.IsReserved.ValueBool(),
		SystemID:   data.SystemID.ValueString(),
	}
	if cloudInit := strings.TrimSpace(data.CloudInit.ValueString()); cloudInit != "" {
		if pub := strings.TrimSpace(allocateReq.PublicKey); pub != "" {
			// Keep user cloud-init and still inject the console public-key runcmd.
			allocateReq.CloudInit = cloudInit + "\n" + uiBaremetalCloudInit(pub)
		} else {
			allocateReq.CloudInit = cloudInit
		}
	} else if pub := strings.TrimSpace(allocateReq.PublicKey); pub != "" {
		// Console allocate always sends this runcmd block with the selected public key.
		allocateReq.CloudInit = uiBaremetalCloudInit(pub)
	}

	networkID, subnetID, extraIDs, err := r.resolveBaremetalNetwork(ctx, data, resp)
	if err != nil || resp.Diagnostics.HasError() {
		if err != nil {
			resp.Diagnostics.AddError("Client Error", err.Error())
		}
		return
	}
	subnets := []models.BaremetalSubnetConfig{
		{
			SubnetID:  subnetID,
			IsPrimary: true,
		},
	}
	for _, id := range extraIDs {
		subnets = append(subnets, models.BaremetalSubnetConfig{SubnetID: id})
	}
	allocateReq.NetworkInterface = &models.BaremetalNetworkInterface{
		Name:     networkID,
		SubnetID: subnetID,
		Subnets:  subnets,
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

	storage, err := expandBaremetalStorage(ctx, data.Storage)
	if err != nil {
		resp.Diagnostics.AddError("Invalid storage", err.Error())
		return
	}
	allocateReq.Storage = storage

	backupCfg, err := expandBaremetalBackupConfig(ctx, data.BackupConfig, data.PolicyEnabled)
	if err != nil {
		resp.Diagnostics.AddError("Invalid backup_config", err.Error())
		return
	}
	allocateReq.BackupConfig = backupCfg

	tflog.Debug(ctx, "Baremetal create request prepared", map[string]interface{}{
		"name":              allocateReq.Name,
		"flavor":            allocateReq.Flavor,
		"os_image":          allocateReq.OSImage,
		"availability_zone": data.AvailabilityZone.ValueString(),
		"subnet_name":       data.SubnetName.ValueString(),
		"network_name":      data.NetworkName.ValueString(),
		"network_id":        networkID,
		"subnet_id":         subnetID,
		"is_reserved":       allocateReq.IsReserved,
		"has_system_id":     strings.TrimSpace(allocateReq.SystemID) != "",
		"has_keypair":       allocateReq.Metadata != nil && strings.TrimSpace(allocateReq.Metadata.Keypair) != "",
		"has_keypair_id":    strings.TrimSpace(allocateReq.KeypairID) != "",
		"has_public_key":    strings.TrimSpace(allocateReq.PublicKey) != "",
		"tags_count":        len(allocateReq.Tags),
		"has_cloud_init":    strings.TrimSpace(allocateReq.CloudInit) != "",
		"storage_count":     len(allocateReq.Storage),
		"has_backup_config": allocateReq.BackupConfig != nil,
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

	// Allocate returns before install finishes. Wait until state is Ready and power is On.
	if err := r.waitForBaremetalState(ctx, data.Name.ValueString(), data.AvailabilityZone.ValueString(), 30*time.Minute); err != nil {
		// POST already created the server; keep it in Terraform state so destroy
		// can release it instead of leaving a 409 on the next apply.
		if readErr := r.readIntoState(ctx, &data); readErr != nil {
			tflog.Warn(ctx, "Unable to refresh baremetal after provisioning failure", map[string]interface{}{
				"name":  data.Name.ValueString(),
				"error": readErr.Error(),
			})
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Baremetal Provisioning Error", err.Error())
		return
	}

	if err := r.readIntoState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read baremetal server after create, got error: %s", err))
		return
	}

	if !isBaremetalReadyAndPoweredOn(data.State.ValueString(), data.Power.ValueString()) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError(
			"Baremetal Provisioning Error",
			fmt.Sprintf("Baremetal server %q is not Ready with power On yet (state=%q, power=%q). Terraform will keep the resource in state so destroy can release it.", data.Name.ValueString(), data.State.ValueString(), data.Power.ValueString()),
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
	lastErrMsg := ""
	stableReadyPolls := 0
	attempt := 0
	for {
		attempt++
		if time.Now().After(deadline) {
			tflog.Warn(ctx, "Baremetal provisioning timeout", map[string]interface{}{
				"name":               name,
				"availability_zone":  az,
				"attempt":            attempt,
				"elapsed_seconds":    int(time.Since(start).Seconds()),
				"stable_ready_polls": stableReadyPolls,
				"last_state":         lastState,
				"last_power":         lastPower,
				"last_err_msg":       lastErrMsg,
				"timeout_seconds":    int(timeout.Seconds()),
			})
			if lastErrMsg != "" {
				return fmt.Errorf("baremetal server %q did not reach Ready with power On within %s (last state=%q, last power=%q): %s", name, timeout.String(), lastState, lastPower, lastErrMsg)
			}
			return fmt.Errorf("baremetal server %q did not reach Ready with power On within %s (last state=%q, last power=%q)", name, timeout.String(), lastState, lastPower)
		}

		readySources := 0
		notReadySources := 0

		detailState := ""
		detailPower := ""
		detailErr := ""
		detail, err := r.client.GetBaremetal(ctx, name, az)
		if err == nil {
			state := strings.TrimSpace(detail.State)
			power := strings.TrimSpace(detail.PowerState)
			detailState = state
			detailPower = power
			if msg := strings.TrimSpace(detail.LastErrMsg); msg != "" {
				lastErrMsg = msg
			}
			if state != "" {
				lastState = state
			}
			if power != "" {
				lastPower = power
			}
			if err := terminalBaremetalAllocationError(name, state, lastErrMsg); err != nil {
				return err
			}
			if isBaremetalReadyAndPoweredOn(state, power) {
				readySources++
			} else {
				notReadySources++
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
			if msg := strings.TrimSpace(summary.LastErrMsg); msg != "" {
				lastErrMsg = msg
			}
			if state != "" {
				lastState = state
			}
			if power != "" {
				lastPower = power
			}
			if err := terminalBaremetalAllocationError(name, state, lastErrMsg); err != nil {
				return err
			}
			if isBaremetalReadyAndPoweredOn(state, power) {
				readySources++
			} else {
				notReadySources++
			}
		} else if err != nil {
			summaryErr = err.Error()
		}

		// Count GET detail + GET list as one poll. The previous counter treated
		// both as successes, so apply returned in ~15s while the UI was still installing.
		if readySources > 0 && notReadySources == 0 {
			stableReadyPolls++
			if stableReadyPolls >= 2 {
				return nil
			}
		} else {
			stableReadyPolls = 0
		}

		tflog.Info(ctx, "Waiting for baremetal Ready and power On", map[string]interface{}{
			"name":               name,
			"availability_zone":  az,
			"attempt":            attempt,
			"elapsed_seconds":    int(time.Since(start).Seconds()),
			"stable_ready_polls": stableReadyPolls,
			"detail_state":       detailState,
			"detail_power":       detailPower,
			"detail_error":       detailErr,
			"summary_state":      summaryState,
			"summary_power":      summaryPower,
			"summary_error":      summaryErr,
			"last_state":         lastState,
			"last_power":         lastPower,
			"last_err_msg":       lastErrMsg,
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

func isBaremetalReadyAndPoweredOn(state, power string) bool {
	if !strings.EqualFold(strings.TrimSpace(state), "Ready") {
		return false
	}
	p := strings.TrimSpace(power)
	return strings.EqualFold(p, "On") || strings.EqualFold(p, "PoweredOn")
}

func isBaremetalFailedState(state string) bool {
	s := strings.ToLower(strings.TrimSpace(state))
	switch s {
	case "failed", "error", "failure", "errorstate", "failedstate":
		return true
	}
	return false
}

func isBaremetalHardLastErr(lastErrMsg string) bool {
	m := strings.ToLower(strings.TrimSpace(lastErrMsg))
	if m == "" {
		return false
	}
	if strings.Contains(m, "volume fetch") {
		return true
	}
	return strings.Contains(m, "volume") && strings.Contains(m, "not found")
}

func terminalBaremetalAllocationError(name, state, lastErrMsg string) error {
	msg := strings.TrimSpace(lastErrMsg)
	st := strings.TrimSpace(state)
	if isBaremetalFailedState(st) {
		if msg == "" {
			return fmt.Errorf("baremetal server %q provisioning failed (state=%q)", name, st)
		}
		return fmt.Errorf("baremetal server %q provisioning failed (state=%q): %s", name, st, msg)
	}
	if msg == "" {
		return nil
	}
	if strings.EqualFold(st, "NoResource") || isBaremetalHardLastErr(msg) {
		return fmt.Errorf("baremetal server %q allocation failed (state=%q): %s", name, st, msg)
	}
	return nil
}

func looksLikeUUID(value string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(value))
}

func (r *BaremetalResource) resolveBaremetalNetwork(ctx context.Context, data BaremetalResourceModel, resp *resource.CreateResponse) (networkID, subnetID string, extra []string, err error) {
	networkRef := strings.TrimSpace(data.NetworkName.ValueString())
	if networkRef == "" {
		return "", "", nil, fmt.Errorf("network_name is required to resolve subnet_name")
	}
	if looksLikeUUID(networkRef) {
		networkID = networkRef
	} else {
		networkID, err = r.client.ResolveVPCID(ctx, networkRef)
		if err != nil {
			return "", "", nil, fmt.Errorf("unable to resolve network_name %q: %w", networkRef, err)
		}
	}

	subnetName := strings.TrimSpace(data.SubnetName.ValueString())
	subnetID, err = r.client.ResolveSubnetID(ctx, networkID, subnetName)
	if err != nil {
		return "", "", nil, fmt.Errorf("unable to resolve subnet_name %q: %w", subnetName, err)
	}

	if !data.AdditionalSubnetNames.IsNull() && !data.AdditionalSubnetNames.IsUnknown() {
		var names []string
		resp.Diagnostics.Append(data.AdditionalSubnetNames.ElementsAs(ctx, &names, false)...)
		if resp.Diagnostics.HasError() {
			return networkID, subnetID, nil, nil
		}
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			id, resolveErr := r.client.ResolveSubnetID(ctx, networkID, name)
			if resolveErr != nil {
				return "", "", nil, fmt.Errorf("unable to resolve additional subnet_name %q: %w", name, resolveErr)
			}
			extra = append(extra, id)
		}
	}

	return networkID, subnetID, extra, nil
}

func expandBaremetalStorage(ctx context.Context, list types.List) ([]models.BaremetalAllocateStorageMap, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var items []baremetalStorageModel
	diags := list.ElementsAs(ctx, &items, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse storage")
	}
	out := make([]models.BaremetalAllocateStorageMap, 0, len(items))
	for _, item := range items {
		out = append(out, models.BaremetalAllocateStorageMap{
			Name:        item.Name.ValueString(),
			Size:        item.Size.ValueString(),
			Path:        item.Path.ValueString(),
			Type:        item.Type.ValueString(),
			FileSystem:  item.FileSystem.ValueString(),
			ForceFormat: item.ForceFormat.ValueBool(),
		})
	}
	return out, nil
}

func expandBaremetalBackupConfig(ctx context.Context, obj types.Object, policyEnabled types.Bool) (*models.BaremetalBackupConfig, error) {
	if obj.IsNull() || obj.IsUnknown() {
		if policyEnabled.IsNull() || policyEnabled.IsUnknown() {
			return nil, nil
		}
		return &models.BaremetalBackupConfig{PolicyEnabled: policyEnabled.ValueBool()}, nil
	}

	var cfg baremetalBackupConfigModel
	diags := obj.As(ctx, &cfg, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse backup_config")
	}

	incrDays, err := intListFromAttr(ctx, cfg.IncrDays)
	if err != nil {
		return nil, err
	}
	fullDays, err := intListFromAttr(ctx, cfg.FullDays)
	if err != nil {
		return nil, err
	}
	selections, err := stringListFromAttr(ctx, cfg.BackupSelections)
	if err != nil {
		return nil, err
	}

	out := &models.BaremetalBackupConfig{
		ScheduleType:      cfg.ScheduleType.ValueString(),
		StartTime:         cfg.StartTime.ValueString(),
		IncrDays:          incrDays,
		FullDays:          fullDays,
		FullRetention:     int(cfg.FullRetention.ValueInt64()),
		FullRetentionUnit: cfg.FullRetentionUnit.ValueString(),
		BackupSelections:  selections,
	}
	if !policyEnabled.IsNull() && !policyEnabled.IsUnknown() {
		out.PolicyEnabled = policyEnabled.ValueBool()
	}
	return out, nil
}

func intListFromAttr(ctx context.Context, list types.List) ([]int, error) {
	if list.IsNull() || list.IsUnknown() {
		return []int{}, nil
	}
	var vals []int64
	diags := list.ElementsAs(ctx, &vals, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse integer list")
	}
	out := make([]int, len(vals))
	for i, v := range vals {
		out[i] = int(v)
	}
	return out, nil
}

func stringListFromAttr(ctx context.Context, list types.List) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var vals []string
	diags := list.ElementsAs(ctx, &vals, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse string list")
	}
	return vals, nil
}

func uiBaremetalCloudInit(publicKey string) string {
	return "runcmd:\n\n" +
		"- useradd -m cloud-user\n" +
		"- mkdir -p /home/cloud-user/.ssh\n" +
		"- echo " + strconv.Quote(publicKey) + " >> /home/cloud-user/.ssh/authorized_keys\n" +
		"- chmod 700 /home/cloud-user/.ssh\n" +
		"- chmod 600 /home/cloud-user/.ssh/authorized_keys\n" +
		"- chown -R cloud-user:cloud-user /home/cloud-user/.ssh"
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
