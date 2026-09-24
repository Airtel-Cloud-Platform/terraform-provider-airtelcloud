package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const asgDefaultTimeout = 15 * time.Minute

var asgVPCNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

var _ resource.Resource = &ASGResource{}
var _ resource.ResourceWithImportState = &ASGResource{}
var _ resource.ResourceWithValidateConfig = &ASGResource{}
var _ resource.ResourceWithModifyPlan = &ASGResource{}

func NewASGResource() resource.Resource { return &ASGResource{} }

type ASGResource struct{ client *client.Client }

type ASGResourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	VPCName             types.String   `tfsdk:"vpc_name"`
	VPCID               types.String   `tfsdk:"vpc_id"`
	AvailabilityZone    types.String   `tfsdk:"availability_zone"`
	Subnet              types.String   `tfsdk:"subnet"`
	NetworkID           types.String   `tfsdk:"network_id"`
	Name                types.String   `tfsdk:"name"`
	MinimumSize         types.Int64    `tfsdk:"minimum_size"`
	MaximumSize         types.Int64    `tfsdk:"maximum_size"`
	DesiredCount        types.Int64    `tfsdk:"desired_count"`
	ScaleUpStepSize     types.Int64    `tfsdk:"scale_up_step_size"`
	ScaleDownStepSize   types.Int64    `tfsdk:"scale_down_step_size"`
	ScalingInterval     types.Int64    `tfsdk:"scaling_interval"`
	CooloffPeriod       types.Int64    `tfsdk:"cooloff_period"`
	TerminationPolicy   types.String   `tfsdk:"termination_policy"`
	DrainPeriod         types.Int64    `tfsdk:"drain_period"`
	ScaleUpRules        types.List     `tfsdk:"scaleup_rules"`
	ScaleDownRules      types.List     `tfsdk:"scaledown_rules"`
	FlavorID            types.Int64    `tfsdk:"flavor_id"`
	FlavorName          types.String   `tfsdk:"flavor_name"`
	ImageID             types.Int64    `tfsdk:"image_id"`
	ImageName           types.String   `tfsdk:"image_name"`
	SnapshotName        types.String   `tfsdk:"snapshot_name"`
	SecurityGroupIDs    types.List     `tfsdk:"security_group_ids"`
	SecurityGroupNames  types.List     `tfsdk:"security_group_names"`
	KeypairID           types.String   `tfsdk:"keypair_id"`
	KeypairName         types.String   `tfsdk:"keypair_name"`
	OSType              types.String   `tfsdk:"os_type"`
	DiskSize            types.Int64    `tfsdk:"disk_size"`
	Labels              types.List     `tfsdk:"labels"`
	LoadBalancerName    types.String   `tfsdk:"load_balancer_name"`
	HostName            types.String   `tfsdk:"host_name"`
	VIP                 types.String   `tfsdk:"vip"`
	Protocol            types.String   `tfsdk:"protocol"`
	Port                types.Int64    `tfsdk:"port"`
	RoutingAlgorithm    types.String   `tfsdk:"routing_algorithm"`
	EnablePersistence   types.Bool     `tfsdk:"enable_persistance"`
	PoolName            types.String   `tfsdk:"pool_name"`
	PoolPort            types.Int64    `tfsdk:"pool_port"`
	MaxConnections      types.Int64    `tfsdk:"max_connections"`
	HealthCheckInterval types.Int64    `tfsdk:"health_check_interval"`
	HealthCheckTimeout  types.Int64    `tfsdk:"health_check_timeout"`
	PoolMonitorProtocol types.String   `tfsdk:"pool_monitor_protocol"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

type ASGRuleModel struct {
	MetricType      types.String `tfsdk:"metric_type"`
	AggregationType types.String `tfsdk:"aggregation_type"`
	TargetValue     types.Int64  `tfsdk:"target_value"`
}

func (r *ASGResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asg"
}

func (r *ASGResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceString := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceInt := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	replaceList := []planmodifier.List{listplanmodifier.RequiresReplace()}
	rule := schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
		"metric_type":      schema.StringAttribute{Required: true},
		"aggregation_type": schema.StringAttribute{Required: true},
		"target_value":     schema.Int64Attribute{Required: true},
	}}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud autoscaling group. Configuration changes replace the group.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"vpc_name":             schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.RegexMatches(asgVPCNamePattern, "must contain only lowercase letters, digits, and hyphens")}, PlanModifiers: replaceString},
			"vpc_id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"availability_zone":    schema.StringAttribute{Required: true, PlanModifiers: replaceString},
			"subnet":               schema.StringAttribute{Required: true, PlanModifiers: replaceString},
			"network_id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":                 schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(3)}, PlanModifiers: replaceString},
			"minimum_size":         schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: replaceInt},
			"maximum_size":         schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: replaceInt},
			"desired_count":        schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: replaceInt},
			"scale_up_step_size":   schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 3)}, PlanModifiers: replaceInt},
			"scale_down_step_size": schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 3)}, PlanModifiers: replaceInt},
			"scaling_interval":     schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(2, 30)}, PlanModifiers: replaceInt},
			"cooloff_period":       schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(0, 300)}, PlanModifiers: replaceInt},
			"termination_policy":   schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("oldest", "newest", "random")}, PlanModifiers: replaceString},
			"drain_period":         schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(0, 300)}, PlanModifiers: replaceInt},
			"scaleup_rules":        schema.ListNestedAttribute{Required: true, NestedObject: rule, Validators: []validator.List{listvalidator.SizeAtLeast(1)}, PlanModifiers: replaceList},
			"scaledown_rules":      schema.ListNestedAttribute{Required: true, NestedObject: rule, Validators: []validator.List{listvalidator.SizeAtLeast(1)}, PlanModifiers: replaceList},
			"flavor_id":            schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplaceIfConfigured(), int64planmodifier.UseStateForUnknown()}},
			"flavor_name":          schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured(), stringplanmodifier.UseStateForUnknown()}},
			"image_id":             schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplaceIfConfigured(), int64planmodifier.UseStateForUnknown()}},
			"image_name":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured(), stringplanmodifier.UseStateForUnknown()}},
			"snapshot_name":        schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"security_group_ids":   schema.ListAttribute{ElementType: types.Int64Type, Optional: true, Computed: true, Validators: []validator.List{listvalidator.SizeAtMost(1)}, PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplaceIfConfigured(), listplanmodifier.UseStateForUnknown()}},
			"security_group_names": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true, Validators: []validator.List{listvalidator.SizeAtMost(1)}, PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplaceIfConfigured(), listplanmodifier.UseStateForUnknown()}},
			"keypair_id":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured(), stringplanmodifier.UseStateForUnknown()}},
			"keypair_name":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured(), stringplanmodifier.UseStateForUnknown()}},
			"os_type":              schema.StringAttribute{Computed: true},
			"disk_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot volume size in GB. Defaults to 100, or 200 for Windows images.",
				Validators:          []validator.Int64{int64validator.AtLeast(20)},
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace(), int64planmodifier.UseStateForUnknown()},
			},
			"labels":                schema.ListAttribute{ElementType: types.StringType, Optional: true, PlanModifiers: replaceList},
			"load_balancer_name":    schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"host_name":             schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"vip":                   schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"protocol":              schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"port":                  schema.Int64Attribute{Optional: true, PlanModifiers: replaceInt},
			"routing_algorithm":     schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"enable_persistance":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()}},
			"pool_name":             schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"pool_port":             schema.Int64Attribute{Optional: true, PlanModifiers: replaceInt},
			"max_connections":       schema.Int64Attribute{Optional: true, PlanModifiers: replaceInt},
			"health_check_interval": schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: replaceInt},
			"health_check_timeout":  schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, PlanModifiers: replaceInt},
			"pool_monitor_protocol": schema.StringAttribute{Optional: true, PlanModifiers: replaceString},
		},
		Blocks: map[string]schema.Block{"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Delete: true})},
	}
}

func (r *ASGResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ASGResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if known(data.MinimumSize) && known(data.MaximumSize) && data.MinimumSize.ValueInt64() >= data.MaximumSize.ValueInt64() {
		resp.Diagnostics.AddError("Invalid Configuration", "minimum_size must be less than maximum_size.")
	}
	if known(data.MinimumSize) && known(data.DesiredCount) && data.MinimumSize.ValueInt64() != data.DesiredCount.ValueInt64() {
		resp.Diagnostics.AddError("Invalid Configuration", "desired_count must equal minimum_size.")
	}
	validateOneOfIntString(&resp.Diagnostics, "flavor_id", data.FlavorID, "flavor_name", data.FlavorName, true)
	imageUnknown := data.ImageID.IsUnknown() || data.ImageName.IsUnknown() || data.SnapshotName.IsUnknown()
	if !imageUnknown {
		imageSet := intSet(data.ImageID) || stringSet(data.ImageName)
		snapshotSet := stringSet(data.SnapshotName)
		if imageSet == snapshotSet {
			resp.Diagnostics.AddError("Invalid Configuration", "Exactly one image source (image_id or image_name) or snapshot_name must be configured.")
		}
		if intSet(data.ImageID) && stringSet(data.ImageName) {
			resp.Diagnostics.AddError("Invalid Configuration", "image_id and image_name are mutually exclusive.")
		}
		if snapshotSet && !data.KeypairID.IsUnknown() && !data.KeypairName.IsUnknown() && !stringSet(data.KeypairID) && !stringSet(data.KeypairName) {
			resp.Diagnostics.AddError("Invalid Configuration", "keypair_id or keypair_name is required with snapshot_name.")
		}
	}
	validateListXOR(&resp.Diagnostics, "security_group_ids", data.SecurityGroupIDs, "security_group_names", data.SecurityGroupNames, true)
	validateOneOfString(&resp.Diagnostics, "keypair_id", data.KeypairID, "keypair_name", data.KeypairName, false)
	validateLBConfig(&resp.Diagnostics, &data)
	if !data.ScaleUpRules.IsUnknown() && !data.ScaleDownRules.IsUnknown() {
		up, d := rulesFromList(ctx, data.ScaleUpRules)
		resp.Diagnostics.Append(d...)
		down, d := rulesFromList(ctx, data.ScaleDownRules)
		resp.Diagnostics.Append(d...)
		if !resp.Diagnostics.HasError() {
			validateRulePairs(&resp.Diagnostics, up, down)
		}
	}
}

func (r *ASGResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ASGResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan ASGResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client != nil {
		resp.Diagnostics.Append(r.validateASGSubnetZone(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if intSet(plan.DiskSize) {
		return
	}
	if r.client == nil {
		plan.DiskSize = types.Int64Unknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}
	osType, err := r.resolveASGOSType(ctx, &plan)
	if err != nil {
		plan.DiskSize = types.Int64Unknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}
	plan.DiskSize = types.Int64Value(defaultASGDiskSize(osType))
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *ASGResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ASGResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, d := data.Timeouts.Create(ctx, asgDefaultTimeout)
	resp.Diagnostics.Append(d...)
	createReq, scoped, diags := r.buildCreateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := scoped.CreateAutoScalingGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create autoscaling group: %s", err))
		return
	}
	data.ID = types.StringValue(group.ID)
	settleASGComputed(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ready, err := scoped.WaitForAutoScalingGroupReady(ctx, group.ID, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for autoscaling group readiness: %s", err))
		return
	}
	resp.Diagnostics.Append(applyASGToState(ctx, &data, ready)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ASGResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ASGResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.GetAutoScalingGroup(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read autoscaling group: %s", err))
		return
	}
	resp.Diagnostics.Append(applyASGToState(ctx, &data, group)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ASGResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Changes to airtelcloud_asg force replacement.")
}

func (r *ASGResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ASGResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, d := data.Timeouts.Delete(ctx, asgDefaultTimeout)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAutoScalingGroup(ctx, data.ID.ValueString()); err != nil && !client.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete autoscaling group: %s", err))
		return
	}
	if err := r.client.WaitForAutoScalingGroupDeleted(ctx, data.ID.ValueString(), timeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for autoscaling group deletion: %s", err))
	}
}

func (r *ASGResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildCreateRequest resolves every name reference and returns the create payload
// alongside a client scoped to the configured availability zone and resolved
// subnet. The create call must use that scoped client so the API receives the
// ce-availability-zone and subnet-id headers; without them the platform places
// the group in its own default zone regardless of the availability_zone field.
func (r *ASGResource) buildCreateRequest(ctx context.Context, data *ASGResourceModel) (*models.CreateASGRequest, *client.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	if _, err := r.client.ResolvePostgresAvailabilityZone(ctx, data.AvailabilityZone.ValueString()); err != nil {
		diags.AddError("Availability Zone Resolution Error", err.Error())
		return nil, nil, diags
	}
	vpcID, err := r.client.ResolveVPCID(ctx, data.VPCName.ValueString())
	if err != nil {
		diags.AddError("VPC Resolution Error", err.Error())
		return nil, nil, diags
	}
	subnet, err := r.client.ResolveSubnet(ctx, vpcID, data.Subnet.ValueString())
	if err != nil {
		diags.AddError("Subnet Resolution Error", err.Error())
		return nil, nil, diags
	}
	if msg := subnetAvailabilityZoneMismatch(data.Subnet.ValueString(), subnet.AvailabilityZone, data.AvailabilityZone.ValueString()); msg != "" {
		diags.AddError("Invalid Configuration", msg)
		return nil, nil, diags
	}
	networkID := subnet.SubnetID
	scoped := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString()).WithSubnetID(networkID)
	flavorID := data.FlavorID.ValueInt64()
	if !intSet(data.FlavorID) {
		raw, err := scoped.ResolveFlavorID(ctx, data.FlavorName.ValueString())
		if err != nil {
			diags.AddError("Flavor Resolution Error", err.Error())
			return nil, nil, diags
		}
		flavorID, _ = strconv.ParseInt(raw, 10, 64)
	}
	var imageID, snapshotID int64
	osType := ""
	if stringSet(data.SnapshotName) {
		snapshotID, err = scoped.ResolveSnapshotID(ctx, data.SnapshotName.ValueString())
	} else {
		var img *models.Image
		img, err = scoped.ResolveImage(ctx, data.ImageID.ValueInt64(), data.ImageName.ValueString())
		if err == nil {
			imageID = int64(img.ID)
			osType = img.OSFamily()
			data.OSType = types.StringValue(osType)
			if img.Name != "" {
				data.ImageName = types.StringValue(img.Name)
			}
		}
	}
	if err != nil {
		diags.AddError("Image/Snapshot Resolution Error", err.Error())
		return nil, nil, diags
	}
	diskSize := defaultASGDiskSize(osType)
	if intSet(data.DiskSize) {
		diskSize = data.DiskSize.ValueInt64()
	}
	data.DiskSize = types.Int64Value(diskSize)
	sgID, sgDiags := resolveASGSecurityGroup(ctx, scoped, data)
	diags.Append(sgDiags...)
	keypairID, kpDiags := resolveASGKeypair(ctx, scoped, data)
	diags.Append(kpDiags...)
	up, d := rulesFromList(ctx, data.ScaleUpRules)
	diags.Append(d...)
	down, d := rulesFromList(ctx, data.ScaleDownRules)
	diags.Append(d...)
	if !diags.HasError() {
		if err := scoped.ValidateASGMetricRules(ctx, append(append([]models.ASGScalingRule{}, up...), down...)); err != nil {
			diags.AddError("Metric Validation Error", err.Error())
		}
	}
	labels, d := stringSliceFromList(ctx, data.Labels)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}
	out := &models.CreateASGRequest{
		Name: data.Name.ValueString(), ImageID: imageID, SnapshotID: snapshotID, FlavorID: flavorID,
		VPCID: vpcID, NetworkID: networkID, SecurityGroupID: sgID, AvailabilityZone: data.AvailabilityZone.ValueString(),
		KeypairID: keypairID, ScalingGroupDesiredSize: data.MinimumSize.ValueInt64(), ScalingGroupMaxSize: data.MaximumSize.ValueInt64(),
		DesiredCount: data.DesiredCount.ValueInt64(), ScaleUpStepSize: data.ScaleUpStepSize.ValueInt64(),
		ScaleDownStepSize: data.ScaleDownStepSize.ValueInt64(), ScalingInterval: data.ScalingInterval.ValueInt64(),
		CooloffPeriod: data.CooloffPeriod.ValueInt64(), TerminationPolicy: data.TerminationPolicy.ValueString(),
		DrainPeriod: data.DrainPeriod.ValueInt64(), BootVolumeSize: diskSize,
		ScaleUpRules: up, ScaleDownRules: down, Labels: labels,
	}
	if stringSet(data.LoadBalancerName) {
		lbID, resolveErr := scoped.ResolveLBServiceID(ctx, data.LoadBalancerName.ValueString())
		if resolveErr != nil {
			diags.AddError("Load Balancer Resolution Error", resolveErr.Error())
			return nil, nil, diags
		}
		lbService, resolveErr := r.client.GetLBService(ctx, lbID)
		if resolveErr != nil {
			diags.AddError("Load Balancer Resolution Error", fmt.Sprintf("unable to read resolved load balancer: %s", resolveErr))
			return nil, nil, diags
		}
		vipClient := r.client.WithSubnetID(lbService.NetworkID)
		vipID, resolveErr := vipClient.ResolveVipPortID(ctx, lbID, data.VIP.ValueString())
		if resolveErr != nil {
			diags.AddError("VIP Resolution Error", resolveErr.Error())
			return nil, nil, diags
		}
		out.VSConfig = &models.ASGVSConfig{
			LBServiceID: lbID, Name: data.HostName.ValueString(), VIPPortID: vipID,
			Protocol: data.Protocol.ValueString(), Port: data.Port.ValueInt64(), RoutingAlgorithm: data.RoutingAlgorithm.ValueString(),
			PoolName: data.PoolName.ValueString(), PoolPort: data.PoolPort.ValueInt64(), Interval: data.HealthCheckInterval.ValueInt64(),
			Timeout: data.HealthCheckTimeout.ValueInt64(), XForwardedFor: true, PersistenceEnabled: data.EnablePersistence.ValueBool(),
			MonitorProtocol: data.PoolMonitorProtocol.ValueString(), MaxConn: data.MaxConnections.ValueInt64(),
		}
	}
	data.VPCID = types.StringValue(vpcID)
	data.NetworkID = types.StringValue(networkID)
	data.FlavorID = types.Int64Value(flavorID)
	if imageID != 0 {
		data.ImageID = types.Int64Value(imageID)
	}
	data.SecurityGroupIDs, _ = types.ListValueFrom(ctx, types.Int64Type, []int64{sgID})
	if keypairID != "" {
		data.KeypairID = types.StringValue(keypairID)
	}
	return out, scoped, diags
}

func applyASGToState(ctx context.Context, data *ASGResourceModel, group *models.AutoScalingGroup) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.VPCID = types.StringValue(group.VPCID)
	data.NetworkID = types.StringValue(group.NetworkID)
	diags.Append(applyASGAvailabilityZone(data, group.AZName)...)
	data.MinimumSize = types.Int64Value(group.ScalingGroupDesiredSize)
	data.DesiredCount = types.Int64Value(group.ScalingGroupDesiredSize)
	data.MaximumSize = types.Int64Value(group.ScalingGroupMaxSize)
	data.ScaleUpStepSize = types.Int64Value(group.ScaleUpStepSize)
	data.ScaleDownStepSize = types.Int64Value(group.ScaleDownStepSize)
	data.ScalingInterval = types.Int64Value(group.ScalingInterval)
	data.CooloffPeriod = types.Int64Value(group.CooloffPeriod)
	data.TerminationPolicy = types.StringValue(group.TerminationPolicy)
	data.DrainPeriod = types.Int64Value(group.DrainPeriod)
	data.FlavorID = types.Int64Value(group.FlavorID)
	if group.Flavor.Name != "" {
		data.FlavorName = types.StringValue(group.Flavor.Name)
	}
	if !stringSet(data.SnapshotName) {
		if group.ImageID != 0 {
			data.ImageID = types.Int64Value(group.ImageID)
		}
		if group.Image.Name != "" {
			data.ImageName = types.StringValue(group.Image.Name)
			data.OSType = types.StringValue(group.Image.OS)
		}
	}
	data.SecurityGroupIDs, _ = types.ListValueFrom(ctx, types.Int64Type, []int64{group.SecurityGroupID})
	if group.KeypairID != "" {
		data.KeypairID = types.StringValue(group.KeypairID)
	}
	if group.Labels.Labels != nil {
		data.Labels, _ = types.ListValueFrom(ctx, types.StringType, group.Labels.Labels)
	}
	up, err := group.ScaleUpRules()
	if err != nil {
		diags.AddError("Response Parsing Error", fmt.Sprintf("Unable to parse scaleup_rules: %s", err))
	} else {
		data.ScaleUpRules, _ = types.ListValueFrom(ctx, ruleObjectType(), apiRulesToTF(up))
	}
	down, err := group.ScaleDownRules()
	if err != nil {
		diags.AddError("Response Parsing Error", fmt.Sprintf("Unable to parse scaledown_rules: %s", err))
	} else {
		data.ScaleDownRules, _ = types.ListValueFrom(ctx, ruleObjectType(), apiRulesToTF(down))
	}
	if group.VirtualServer.Name != "" {
		data.HostName = types.StringValue(group.VirtualServer.Name)
		data.LoadBalancerName = types.StringValue(group.VirtualServer.LBService.Name)
		if len(group.VirtualServer.ASGPool.Members) > 0 {
			data.PoolPort = types.Int64Value(group.VirtualServer.ASGPool.Members[0].Port)
			data.MaxConnections = types.Int64Value(group.VirtualServer.ASGPool.Members[0].MaxConn)
		}
	}
	settleASGComputed(data)
	return diags
}

// applyASGAvailabilityZone keeps the configured zone in state and only adopts
// the API value on import, where state has none. The platform places the group
// in the zone of the chosen subnet, so it can report a zone other than the one
// requested; overwriting the configured value would make Terraform reject the
// apply, and on refresh it would plan an endless replacement back to a zone the
// platform never honors. The mismatch is surfaced as a warning instead.
func applyASGAvailabilityZone(data *ASGResourceModel, azName string) diag.Diagnostics {
	var diags diag.Diagnostics
	if azName == "" {
		return diags
	}
	if data.AvailabilityZone.IsNull() || data.AvailabilityZone.IsUnknown() {
		data.AvailabilityZone = types.StringValue(azName)
		return diags
	}
	if configured := data.AvailabilityZone.ValueString(); configured != azName {
		diags.AddAttributeWarning(
			path.Root("availability_zone"),
			"Availability Zone Mismatch",
			fmt.Sprintf("Requested availability zone %q but the group was placed in %q. "+
				"Autoscaling groups follow the zone of the configured subnet; set availability_zone to %q "+
				"or choose a subnet in %q.", configured, azName, azName, configured),
		)
	}
	return diags
}

func (r *ASGResource) validateASGSubnetZone(ctx context.Context, data *ASGResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if r.client == nil || !stringSet(data.VPCName) || !stringSet(data.Subnet) || !stringSet(data.AvailabilityZone) {
		return diags
	}
	vpcID, err := r.client.ResolveVPCID(ctx, data.VPCName.ValueString())
	if err != nil {
		return diags
	}
	subnet, err := r.client.ResolveSubnet(ctx, vpcID, data.Subnet.ValueString())
	if err != nil {
		return diags
	}
	if msg := subnetAvailabilityZoneMismatch(data.Subnet.ValueString(), subnet.AvailabilityZone, data.AvailabilityZone.ValueString()); msg != "" {
		diags.AddError("Invalid Configuration", msg)
	}
	return diags
}

func subnetAvailabilityZoneMismatch(subnetName, subnetAZ, requestedAZ string) string {
	if subnetAZ == "" || requestedAZ == "" || subnetAZ == requestedAZ {
		return ""
	}
	return fmt.Sprintf("subnet %q is in availability zone %q, but availability_zone is %q. Choose a subnet in %q or set availability_zone to %q.",
		subnetName, subnetAZ, requestedAZ, requestedAZ, subnetAZ)
}

// settleASGComputed nulls out computed attributes that are still unknown. The
// GET API returns no keypair name, and snapshot-based groups report no image or
// OS, so those values can never be filled in from the API response.
func settleASGComputed(data *ASGResourceModel) {
	data.VPCID = settledString(data.VPCID)
	data.NetworkID = settledString(data.NetworkID)
	data.FlavorID = settledInt64(data.FlavorID)
	data.FlavorName = settledString(data.FlavorName)
	data.ImageID = settledInt64(data.ImageID)
	data.ImageName = settledString(data.ImageName)
	data.KeypairID = settledString(data.KeypairID)
	data.KeypairName = settledString(data.KeypairName)
	data.OSType = settledString(data.OSType)
	data.DiskSize = settledInt64(data.DiskSize)
	data.SecurityGroupIDs = settledList(data.SecurityGroupIDs, types.Int64Type)
	data.SecurityGroupNames = settledList(data.SecurityGroupNames, types.StringType)
}

func settledString(v types.String) types.String {
	if v.IsUnknown() {
		return types.StringNull()
	}
	return v
}

func settledInt64(v types.Int64) types.Int64 {
	if v.IsUnknown() {
		return types.Int64Null()
	}
	return v
}

func settledList(v types.List, elem attr.Type) types.List {
	if v.IsUnknown() {
		return types.ListNull(elem)
	}
	return v
}

func rulesFromList(ctx context.Context, list types.List) ([]models.ASGScalingRule, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var tf []ASGRuleModel
	diags := list.ElementsAs(ctx, &tf, false)
	out := make([]models.ASGScalingRule, 0, len(tf))
	for _, rule := range tf {
		out = append(out, models.ASGScalingRule{MetricType: rule.MetricType.ValueString(), AggregationType: rule.AggregationType.ValueString(), TargetValue: rule.TargetValue.ValueInt64()})
	}
	return out, diags
}

func ruleObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"metric_type": types.StringType, "aggregation_type": types.StringType, "target_value": types.Int64Type,
	}}
}

func apiRulesToTF(rules []models.ASGScalingRule) []ASGRuleModel {
	out := make([]ASGRuleModel, 0, len(rules))
	for _, rule := range rules {
		out = append(out, ASGRuleModel{MetricType: types.StringValue(rule.MetricType), AggregationType: types.StringValue(rule.AggregationType), TargetValue: types.Int64Value(rule.TargetValue)})
	}
	return out
}

func resolveASGSecurityGroup(ctx context.Context, c *client.Client, data *ASGResourceModel) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	if listSet(data.SecurityGroupIDs) {
		var ids []int64
		diags.Append(data.SecurityGroupIDs.ElementsAs(ctx, &ids, false)...)
		if len(ids) == 1 {
			return ids[0], diags
		}
	}
	var names []string
	diags.Append(data.SecurityGroupNames.ElementsAs(ctx, &names, false)...)
	if diags.HasError() || len(names) != 1 {
		diags.AddError("Security Group Resolution Error", "Exactly one security group is required.")
		return 0, diags
	}
	id, err := c.ResolveSecurityGroupID(ctx, names[0])
	if err != nil {
		diags.AddError("Security Group Resolution Error", err.Error())
	}
	return int64(id), diags
}

func resolveASGKeypair(ctx context.Context, c *client.Client, data *ASGResourceModel) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if stringSet(data.KeypairID) {
		return data.KeypairID.ValueString(), diags
	}
	if stringSet(data.KeypairName) {
		id, err := c.ResolveKeypairID(ctx, data.KeypairName.ValueString())
		if err != nil {
			diags.AddError("Keypair Resolution Error", err.Error())
		}
		return id, diags
	}
	return "", diags
}

func validateRulePairs(diags *diag.Diagnostics, up, down []models.ASGScalingRule) {
	upByMetric := map[string]models.ASGScalingRule{}
	downByMetric := map[string]models.ASGScalingRule{}
	for _, rule := range up {
		upByMetric[rule.MetricType] = rule
	}
	for _, rule := range down {
		downByMetric[rule.MetricType] = rule
	}
	for metric, upRule := range upByMetric {
		downRule, ok := downByMetric[metric]
		if !ok {
			diags.AddError("Invalid Configuration", fmt.Sprintf("metric_type %q requires both scale-up and scale-down rules.", metric))
			continue
		}
		if upRule.AggregationType != downRule.AggregationType {
			diags.AddError("Invalid Configuration", fmt.Sprintf("metric_type %q must use the same aggregation_type in both rules.", metric))
		}
		if upRule.TargetValue <= downRule.TargetValue {
			diags.AddError("Invalid Configuration", fmt.Sprintf("scale-up target for %q must be greater than scale-down target.", metric))
		}
	}
	for metric := range downByMetric {
		if _, ok := upByMetric[metric]; !ok {
			diags.AddError("Invalid Configuration", fmt.Sprintf("metric_type %q requires both scale-up and scale-down rules.", metric))
		}
	}
}

func validateLBConfig(diags *diag.Diagnostics, data *ASGResourceModel) {
	if data.LoadBalancerName.IsUnknown() {
		return
	}
	fields := map[string]attr.Value{
		"host_name": data.HostName, "vip": data.VIP, "protocol": data.Protocol,
		"port": data.Port, "routing_algorithm": data.RoutingAlgorithm, "pool_name": data.PoolName,
		"pool_port": data.PoolPort, "max_connections": data.MaxConnections,
		"health_check_interval": data.HealthCheckInterval, "health_check_timeout": data.HealthCheckTimeout,
		"pool_monitor_protocol": data.PoolMonitorProtocol,
	}
	if !stringSet(data.LoadBalancerName) {
		for name, value := range fields {
			if value.IsUnknown() {
				continue
			}
			if !value.IsNull() {
				diags.AddError("Invalid Configuration", fmt.Sprintf("%s requires load_balancer_name.", name))
			}
		}
		return
	}
	missing := []string{}
	for name, value := range fields {
		if value.IsUnknown() {
			return
		}
		if value.IsNull() {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		diags.AddError("Invalid Configuration", "load_balancer_name also requires: "+strings.Join(missing, ", "))
	}
}

func validateOneOfIntString(diags *diag.Diagnostics, intName string, i types.Int64, stringName string, s types.String, required bool) {
	if i.IsUnknown() || s.IsUnknown() {
		return
	}
	count := 0
	if intSet(i) {
		count++
	}
	if stringSet(s) {
		count++
	}
	if count > 1 || required && count != 1 {
		diags.AddError("Invalid Configuration", fmt.Sprintf("Exactly one of %s or %s must be configured.", intName, stringName))
	}
}

func validateOneOfString(diags *diag.Diagnostics, aName string, a types.String, bName string, b types.String, required bool) {
	if a.IsUnknown() || b.IsUnknown() {
		return
	}
	count := 0
	if stringSet(a) {
		count++
	}
	if stringSet(b) {
		count++
	}
	if count > 1 || required && count != 1 {
		diags.AddError("Invalid Configuration", fmt.Sprintf("Exactly one of %s or %s must be configured.", aName, bName))
	}
}

func validateListXOR(diags *diag.Diagnostics, aName string, a types.List, bName string, b types.List, required bool) {
	if a.IsUnknown() || b.IsUnknown() {
		return
	}
	count := 0
	if listSet(a) {
		count++
	}
	if listSet(b) {
		count++
	}
	if count > 1 || required && count != 1 {
		diags.AddError("Invalid Configuration", fmt.Sprintf("Exactly one of %s or %s must be configured.", aName, bName))
	}
}

func stringSet(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown() && strings.TrimSpace(v.ValueString()) != ""
}
func intSet(v types.Int64) bool { return !v.IsNull() && !v.IsUnknown() }
func listSet(v types.List) bool { return !v.IsNull() && !v.IsUnknown() && len(v.Elements()) > 0 }
func known(v types.Int64) bool  { return !v.IsNull() && !v.IsUnknown() }

func defaultASGDiskSize(osType string) int64 {
	if strings.Contains(strings.ToLower(osType), "windows") {
		return 200
	}
	return 100
}

func (r *ASGResource) resolveASGOSType(ctx context.Context, data *ASGResourceModel) (string, error) {
	if stringSet(data.SnapshotName) {
		return "", nil
	}
	id := int64(0)
	if intSet(data.ImageID) {
		id = data.ImageID.ValueInt64()
	}
	name := ""
	if stringSet(data.ImageName) {
		name = data.ImageName.ValueString()
	}
	img, err := r.client.ResolveImage(ctx, id, name)
	if err != nil {
		return "", err
	}
	return img.OSFamily(), nil
}
