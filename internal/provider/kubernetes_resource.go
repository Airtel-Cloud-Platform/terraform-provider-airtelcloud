package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const kubernetesDefaultTimeout = 45 * time.Minute

var kubernetesClusterNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

var _ resource.Resource = &KubernetesResource{}
var _ resource.ResourceWithImportState = &KubernetesResource{}
var _ resource.ResourceWithValidateConfig = &KubernetesResource{}

func NewKubernetesResource() resource.Resource {
	return &KubernetesResource{}
}

// KubernetesResource defines the airtelcloud_kubernetes resource.
type KubernetesResource struct {
	client *client.Client
}

type KubernetesResourceModel struct {
	ID                   types.String              `tfsdk:"id"`
	Name                 types.String              `tfsdk:"name"`
	Description          types.String              `tfsdk:"description"`
	K8sName              types.String              `tfsdk:"k8s_name"`
	K8sVersion           types.String              `tfsdk:"k8s_version"`
	CNIName              types.String              `tfsdk:"cni_name"`
	CNIVersion           types.String              `tfsdk:"cni_version"`
	MasterNodes          types.Int64               `tfsdk:"master_nodes"`
	WorkerNodes          types.Int64               `tfsdk:"worker_nodes"`
	AvailabilityZone     types.String              `tfsdk:"availability_zone"`
	ProviderType         types.String              `tfsdk:"provider_type"`
	ControlPlaneProvider types.String              `tfsdk:"control_plane_provider"`
	MasterHostGroup      types.String              `tfsdk:"master_host_group"`
	VPCName              types.String              `tfsdk:"vpc_name"`
	NodePools            []KubernetesNodePoolModel `tfsdk:"node_pools"`
	State                types.String              `tfsdk:"state"`
	CreatedBy            types.String              `tfsdk:"created_by"`
	Timeouts             timeouts.Value            `tfsdk:"timeouts"`
}

type KubernetesNodePoolModel struct {
	Name             types.String `tfsdk:"name"`
	HostGroup        types.String `tfsdk:"host_group"`
	GroupName        types.String `tfsdk:"group_name"`
	SubnetName       types.String `tfsdk:"subnet_name"`
	OSDistribution   types.String `tfsdk:"os_distribution"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Count            types.Int64  `tfsdk:"count"`
}

func (r *KubernetesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

func (r *KubernetesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud Kubernetes (CKP) cluster. v1 supports BYOH create, read, import, and delete. Changes to configuration force a new cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster name. Same as `name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Cluster name sent as `cluster` in the create API.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(kubernetesClusterNamePattern, "name must be lowercase alphanumeric and hyphens, and cannot start or end with a hyphen."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Cluster description (`desc`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"k8s_name": schema.StringAttribute{
				MarkdownDescription: "Kubernetes distribution. Defaults to `CKP`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(models.KubernetesDistributionCKP),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"k8s_version": schema.StringAttribute{
				MarkdownDescription: "Kubernetes version (for example `v1.33.7`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cni_name": schema.StringAttribute{
				MarkdownDescription: "CNI plugin name (for example `calico`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cni_version": schema.StringAttribute{
				MarkdownDescription: "CNI plugin version (for example `v3.30.6`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"master_nodes": schema.Int64Attribute{
				MarkdownDescription: "Number of control-plane nodes.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"worker_nodes": schema.Int64Attribute{
				MarkdownDescription: "Number of worker nodes. Defaults to the sum of `node_pools` counts.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone code (for example `S1`). Sent as `ce-availability-zone` and `requiredSchedulingTags.availabilityZone`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "CKP provider. v1 supports `BringYourOwnHost` only.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(models.KubernetesProviderBYOH),
				Validators: []validator.String{
					stringvalidator.OneOf(models.KubernetesProviderBYOH),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"control_plane_provider": schema.StringAttribute{
				MarkdownDescription: "CAPI control-plane provider. One of `Kamaji` or `Kubeadm`. Defaults to `Kamaji`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(models.KubernetesControlPlaneKamaji),
				Validators: []validator.String{
					stringvalidator.OneOf(models.KubernetesControlPlaneKamaji, models.KubernetesControlPlaneKubeadm),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"master_host_group": schema.StringAttribute{
				MarkdownDescription: "Optional BYOH master host group. The console often sends an empty string when using Kamaji.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "VPC name used to resolve `node_pools.subnet_name` to a subnet UUID. Optional if the subnet name is unique in the project.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"node_pools": schema.ListNestedAttribute{
				MarkdownDescription: "BYOH worker node pools. Map keys in the API are `node_pools[].name`.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Node pool key (for example `md0`).",
							Required:            true,
						},
						"host_group": schema.StringAttribute{
							MarkdownDescription: "Host group identifier (for example `ccd.xLarge`).",
							Required:            true,
						},
						"group_name": schema.StringAttribute{
							MarkdownDescription: "Display name of the host group (for example `Compute dense`). Sent as `groupName`.",
							Optional:            true,
						},
						"subnet_name": schema.StringAttribute{
							MarkdownDescription: "Subnet name. Resolved to the subnet UUID the cluster API expects.",
							Required:            true,
						},
						"os_distribution": schema.StringAttribute{
							MarkdownDescription: "OS distribution. One of `Ubuntu` or `Rhel`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("Ubuntu", "Rhel"),
							},
						},
						"availability_zone": schema.StringAttribute{
							MarkdownDescription: "Pool AZ. Defaults to the cluster `availability_zone`.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"count": schema.Int64Attribute{
							MarkdownDescription: "Number of machines in the pool.",
							Required:            true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster lifecycle state (`Unknown`, `Creating`, `Deleting`, `Ready`).",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User that created the cluster, when returned by the API.",
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

func (r *KubernetesResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data KubernetesResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := map[string]struct{}{}
	for i, pool := range data.NodePools {
		if pool.Name.IsUnknown() || pool.Name.IsNull() {
			continue
		}
		name := pool.Name.ValueString()
		if _, ok := seen[name]; ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("node_pools").AtListIndex(i).AtName("name"),
				"Duplicate node pool name",
				fmt.Sprintf("node pool name %q is used more than once.", name),
			)
		}
		seen[name] = struct{}{}
	}
}

func (r *KubernetesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, kubernetesDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnetIDs, resolveDiags := r.resolveKubernetesSubnetIDs(ctx, &data)
	resp.Diagnostics.Append(resolveDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildKubernetesCreateRequest(&data, subnetIDs)
	azClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
	if err := azClient.CreateKubernetesCluster(ctx, createReq); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create kubernetes cluster, got error: %s", err))
		return
	}

	ready, err := azClient.WaitForKubernetesReady(ctx, data.Name.ValueString(), kubernetesHostGroups(&data), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for kubernetes cluster to be ready: %s", err))
		return
	}

	applyKubernetesClusterToState(&data, ready)
	tflog.Trace(ctx, "created kubernetes resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.ID.ValueString()
	if name == "" {
		name = data.Name.ValueString()
	}

	az := data.AvailabilityZone.ValueString()
	c := r.client
	if az != "" {
		c = c.WithAvailabilityZone(az)
	}

	cluster, err := c.GetKubernetesCluster(ctx, name, kubernetesHostGroups(&data))
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read kubernetes cluster, got error: %s", err))
		return
	}

	if cluster.IsDeleted || cluster.State == models.KubernetesStateDeleting {
		resp.State.RemoveResource(ctx)
		return
	}

	applyKubernetesClusterToState(&data, cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"airtelcloud_kubernetes does not support in-place updates. Changing configuration forces a new cluster.",
	)
}

func (r *KubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, kubernetesDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.ID.ValueString()
	if name == "" {
		name = data.Name.ValueString()
	}

	az := data.AvailabilityZone.ValueString()
	c := r.client
	if az != "" {
		c = c.WithAvailabilityZone(az)
	}

	if err := c.DeleteKubernetesCluster(ctx, name); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete kubernetes cluster, got error: %s", err))
		return
	}

	if err := c.WaitForKubernetesDeleted(ctx, name, kubernetesHostGroups(&data), deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for kubernetes cluster deletion: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted kubernetes resource")
}

func (r *KubernetesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "imported kubernetes resource")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *KubernetesResource) resolveKubernetesSubnetIDs(ctx context.Context, data *KubernetesResourceModel) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	ids := make(map[string]string, len(data.NodePools))
	vpcName := ""
	if !data.VPCName.IsNull() && !data.VPCName.IsUnknown() {
		vpcName = data.VPCName.ValueString()
	}
	for i, pool := range data.NodePools {
		name := pool.SubnetName.ValueString()
		id, err := r.client.ResolveSubnetIDByName(ctx, name, vpcName)
		if err != nil {
			diags.AddAttributeError(
				path.Root("node_pools").AtListIndex(i).AtName("subnet_name"),
				"Subnet Resolution Error",
				err.Error(),
			)
			continue
		}
		ids[pool.Name.ValueString()] = id
	}
	return ids, diags
}

func buildKubernetesCreateRequest(data *KubernetesResourceModel, subnetIDs map[string]string) *models.CreateKubernetesClusterRequest {
	az := data.AvailabilityZone.ValueString()
	pools := make(map[string]models.KubernetesNodePool, len(data.NodePools))
	workerSum := 0
	for _, pool := range data.NodePools {
		poolAZ := az
		if !pool.AvailabilityZone.IsNull() && !pool.AvailabilityZone.IsUnknown() && pool.AvailabilityZone.ValueString() != "" {
			poolAZ = pool.AvailabilityZone.ValueString()
		}
		if pool.AvailabilityZone.IsUnknown() || pool.AvailabilityZone.IsNull() || pool.AvailabilityZone.ValueString() == "" {
			pool.AvailabilityZone = types.StringValue(az)
		}
		count := int(pool.Count.ValueInt64())
		workerSum += count
		item := models.KubernetesNodePool{
			HostGroup:      pool.HostGroup.ValueString(),
			Count:          count,
			Subnet:         subnetIDs[pool.Name.ValueString()],
			AZ:             poolAZ,
			OSDistribution: pool.OSDistribution.ValueString(),
		}
		if !pool.GroupName.IsNull() && !pool.GroupName.IsUnknown() {
			item.GroupName = pool.GroupName.ValueString()
		}
		pools[pool.Name.ValueString()] = item
	}

	workerNodes := workerSum
	if !data.WorkerNodes.IsNull() && !data.WorkerNodes.IsUnknown() {
		workerNodes = int(data.WorkerNodes.ValueInt64())
	} else {
		data.WorkerNodes = types.Int64Value(int64(workerSum))
	}

	masterHostGroup := ""
	if !data.MasterHostGroup.IsNull() && !data.MasterHostGroup.IsUnknown() {
		masterHostGroup = data.MasterHostGroup.ValueString()
	}

	req := &models.CreateKubernetesClusterRequest{
		Cluster: data.Name.ValueString(),
		Desc:    strings.TrimSpace(data.Description.ValueString()),
		K8sInfo: models.KubernetesKubeInfo{
			K8sName:     data.K8sName.ValueString(),
			K8sVersion:  data.K8sVersion.ValueString(),
			CNIName:     data.CNIName.ValueString(),
			CNIVersion:  data.CNIVersion.ValueString(),
			MasterNodes: int(data.MasterNodes.ValueInt64()),
			WorkerNodes: workerNodes,
		},
		CKPProvider: models.KubernetesCKPProvider{
			Provider: data.ProviderType.ValueString(),
			BYOH: &models.KubernetesBYOHProvider{
				MasterHostGroup:      masterHostGroup,
				ControlPlaneProvider: data.ControlPlaneProvider.ValueString(),
				NodePools:            pools,
			},
		},
		RequiredSchedulingTags: map[string]string{
			"availabilityZone": az,
		},
	}
	return req
}

func applyKubernetesClusterToState(data *KubernetesResourceModel, cluster *models.KubernetesCluster) {
	name := cluster.ClusterName()
	if name == "" {
		name = data.Name.ValueString()
	}
	data.ID = types.StringValue(name)
	if data.Name.IsNull() || data.Name.IsUnknown() || data.Name.ValueString() == "" {
		data.Name = types.StringValue(name)
	}
	if cluster.State != "" {
		data.State = types.StringValue(cluster.State)
	} else if data.State.IsNull() || data.State.IsUnknown() {
		data.State = types.StringValue(models.KubernetesStateReady)
	}
	if cluster.CreatedBy != "" {
		data.CreatedBy = types.StringValue(cluster.CreatedBy)
	} else if data.CreatedBy.IsUnknown() {
		data.CreatedBy = types.StringNull()
	}
	if cluster.Desc != "" && (data.Description.IsNull() || data.Description.IsUnknown()) {
		data.Description = types.StringValue(cluster.Desc)
	}
	if cluster.K8sInfo != nil {
		if cluster.K8sInfo.K8sVersion != "" {
			data.K8sVersion = types.StringValue(cluster.K8sInfo.K8sVersion)
		}
		if cluster.K8sInfo.K8sName != "" {
			data.K8sName = types.StringValue(cluster.K8sInfo.K8sName)
		}
	} else if cluster.K8sVersion != "" && (data.K8sVersion.IsNull() || data.K8sVersion.IsUnknown()) {
		data.K8sVersion = types.StringValue(cluster.K8sVersion)
	}

	az := data.AvailabilityZone.ValueString()
	for i := range data.NodePools {
		if data.NodePools[i].AvailabilityZone.IsNull() || data.NodePools[i].AvailabilityZone.IsUnknown() || data.NodePools[i].AvailabilityZone.ValueString() == "" {
			data.NodePools[i].AvailabilityZone = types.StringValue(az)
		}
	}

	if data.WorkerNodes.IsNull() || data.WorkerNodes.IsUnknown() {
		sum := int64(0)
		for _, pool := range data.NodePools {
			sum += pool.Count.ValueInt64()
		}
		data.WorkerNodes = types.Int64Value(sum)
	}
}

func kubernetesHostGroups(data *KubernetesResourceModel) []string {
	seen := make(map[string]struct{}, len(data.NodePools))
	out := make([]string, 0, len(data.NodePools))
	for _, pool := range data.NodePools {
		if pool.HostGroup.IsNull() || pool.HostGroup.IsUnknown() {
			continue
		}
		name := pool.HostGroup.ValueString()
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
