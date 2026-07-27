package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &LBVirtualServerResource{}
var _ resource.ResourceWithImportState = &LBVirtualServerResource{}
var _ resource.ResourceWithValidateConfig = &LBVirtualServerResource{}

func NewLBVirtualServerResource() resource.Resource {
	return &LBVirtualServerResource{}
}

type LBVirtualServerResource struct {
	client *client.Client
}

type LBVirtualServerResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	LBServiceID        types.String   `tfsdk:"lb_service_id"`
	LBServiceName      types.String   `tfsdk:"lb_service_name"`
	Name               types.String   `tfsdk:"name"`
	VipPortID          types.Int64    `tfsdk:"vip_port_id"`
	Protocol           types.String   `tfsdk:"protocol"`
	Port               types.Int64    `tfsdk:"port"`
	RoutingAlgorithm   types.String   `tfsdk:"routing_algorithm"`
	VPCID              types.String   `tfsdk:"vpc_id"`
	VPCName            types.String   `tfsdk:"vpc_name"`
	Interval           types.Int64    `tfsdk:"interval"`
	Nodes              types.List     `tfsdk:"nodes"`
	PersistenceEnabled types.Bool     `tfsdk:"persistence_enabled"`
	PersistenceType    types.String   `tfsdk:"persistence_type"`
	XForwardedFor      types.Bool     `tfsdk:"x_forwarded_for"`
	RedirectHTTPS      types.Bool     `tfsdk:"redirect_https"`
	CertificateID      types.String   `tfsdk:"certificate_id"`
	MonitorProtocol    types.String   `tfsdk:"monitor_protocol"`
	MonitorPort        types.Int64    `tfsdk:"monitor_port"`
	PoolName           types.String   `tfsdk:"pool_name"`
	MonitorName        types.String   `tfsdk:"monitor_name"`
	Timeout            types.Int64    `tfsdk:"timeout"`
	Status             types.String   `tfsdk:"status"`
	VIP                types.String   `tfsdk:"vip"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func (r *LBVirtualServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_virtual_server"
}

func (r *LBVirtualServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Virtual Server on an Airtel Cloud Load Balancer Service. A virtual server defines L4/L7 load balancing rules with backend nodes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the virtual server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lb_service_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the parent LB service. Either lb_service_id or lb_service_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lb_service_name": schema.StringAttribute{
				MarkdownDescription: "The name of the parent LB service. If set, it is resolved to lb_service_id. Either lb_service_id or lb_service_name must be specified.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the virtual server.",
				Optional:            true,
			},
			"vip": schema.StringAttribute{
				MarkdownDescription: "The VIP fixed IP address to bind this virtual server to (e.g. airtelcloud_lb_vip.example.fixed_ips). The provider resolves it to vip_port_id.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vip_port_id": schema.Int64Attribute{
				MarkdownDescription: "The VIP port ID this virtual server is bound to. Computed by resolving the vip fixed IP.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The listener protocol (HTTP, HTTPS, TCP, UDP).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The listener port number.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"routing_algorithm": schema.StringAttribute{
				MarkdownDescription: "The load balancing algorithm (e.g., ROUND_ROBIN, LEAST_CONNECTIONS).",
				Required:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The VPC ID. Either vpc_id or vpc_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPC. If set, it is resolved to vpc_id. Either vpc_id or vpc_name must be specified.",
				Optional:            true,
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "The health check interval in seconds.",
				Required:            true,
			},
			"nodes": schema.ListNestedAttribute{
				MarkdownDescription: "The backend nodes for the virtual server.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"compute_id": schema.StringAttribute{
							MarkdownDescription: "The backend instance ID (UUID). Either compute_id or compute_name must be specified for each node.",
							Optional:            true,
						},
						"compute_name": schema.StringAttribute{
							MarkdownDescription: "The backend instance name. If set, it is resolved to its ID. Either compute_id or compute_name must be specified for each node.",
							Optional:            true,
						},
						"compute_ip": schema.StringAttribute{
							MarkdownDescription: "The backend instance IP address (sent as resource_ip).",
							Required:            true,
						},
						"resource_type": schema.StringAttribute{
							MarkdownDescription: "The backend member kind: \"compute\" (a VM, the default) or \"bm\" (a baremetal server).",
							Optional:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "The backend port number.",
							Required:            true,
						},
						"weight": schema.Int64Attribute{
							MarkdownDescription: "The weight for weighted load balancing.",
							Optional:            true,
						},
						"max_conn": schema.Int64Attribute{
							MarkdownDescription: "The maximum connections for this node.",
							Optional:            true,
						},
					},
				},
			},
			"persistence_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether session persistence is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"persistence_type": schema.StringAttribute{
				MarkdownDescription: "The session persistence type (cookie or source_ip).",
				Optional:            true,
			},
			"x_forwarded_for": schema.BoolAttribute{
				MarkdownDescription: "Whether to add X-Forwarded-For header.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"redirect_https": schema.BoolAttribute{
				MarkdownDescription: "Whether to redirect HTTP to HTTPS.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"certificate_id": schema.StringAttribute{
				MarkdownDescription: "The SSL certificate ID (required for HTTPS protocol).",
				Optional:            true,
			},
			"monitor_protocol": schema.StringAttribute{
				MarkdownDescription: "The health check monitor protocol (HTTP, HTTPS, TCP, UDP).",
				Optional:            true,
			},
			"monitor_port": schema.Int64Attribute{
				MarkdownDescription: "The health check monitor port.",
				Optional:            true,
			},
			"pool_name": schema.StringAttribute{
				MarkdownDescription: "The name of the backend pool.",
				Optional:            true,
			},
			"monitor_name": schema.StringAttribute{
				MarkdownDescription: "The name of the health check monitor.",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "The health check timeout in seconds.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the virtual server.",
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

func (r *LBVirtualServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces that exactly one of lb_service_id/lb_service_name and
// exactly one of vpc_id/vpc_name is set.
func (r *LBVirtualServerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data LBVirtualServerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.LBServiceID.IsNull() && !data.LBServiceName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of lb_service_id or lb_service_name may be specified, not both.")
	}
	if data.LBServiceID.IsNull() && data.LBServiceName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of lb_service_id or lb_service_name must be specified.")
	}

	if !data.VPCID.IsNull() && !data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of vpc_id or vpc_name may be specified, not both.")
	}
	if data.VPCID.IsNull() && data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of vpc_id or vpc_name must be specified.")
	}

	// The API requires monitor_name whenever redirect_https is not enabled.
	// Surface this at plan time instead of as a 422 at apply time.
	redirectHTTPS := !data.RedirectHTTPS.IsNull() && data.RedirectHTTPS.ValueBool()
	monitorNameSet := !data.MonitorName.IsNull() && data.MonitorName.ValueString() != ""
	if !redirectHTTPS && !monitorNameSet {
		resp.Diagnostics.AddAttributeError(path.Root("monitor_name"),
			"Invalid Configuration",
			"monitor_name is required when redirect_https is not enabled.")
	}
}

func (r *LBVirtualServerResource) extractNodes(ctx context.Context, data *LBVirtualServerResourceModel) ([]models.VirtualServerNode, error) {
	var nodeObjects []types.Object
	diags := data.Nodes.ElementsAs(ctx, &nodeObjects, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract nodes")
	}

	nodes := make([]models.VirtualServerNode, len(nodeObjects))
	for i, obj := range nodeObjects {
		attrs := obj.Attributes()

		computeIDAttr := attrs["compute_id"].(types.String)
		computeNameAttr := attrs["compute_name"].(types.String)
		computeIP := attrs["compute_ip"].(types.String).ValueString()

		// Exactly one of compute_id or compute_name must be set per node.
		idSet := !computeIDAttr.IsNull() && computeIDAttr.ValueString() != ""
		nameSet := !computeNameAttr.IsNull() && computeNameAttr.ValueString() != ""
		if idSet && nameSet {
			return nil, fmt.Errorf("node %d: only one of compute_id or compute_name may be specified, not both", i)
		}
		if !idSet && !nameSet {
			return nil, fmt.Errorf("node %d: one of compute_id or compute_name must be specified", i)
		}

		// resource_type selects the backend member kind: "compute" (a VM, the
		// default) or "bm" (a baremetal server).
		resourceType := "compute"
		if rt, ok := attrs["resource_type"].(types.String); ok && !rt.IsNull() && rt.ValueString() != "" {
			resourceType = rt.ValueString()
		}

		node := models.VirtualServerNode{
			ResourceIP: computeIP,
			Port:       int(attrs["port"].(types.Int64).ValueInt64()),
		}

		switch resourceType {
		case "compute", "vm", "":
			compute, err := r.client.ResolveComputeNode(ctx, computeIDAttr.ValueString(), computeNameAttr.ValueString())
			if err != nil {
				return nil, fmt.Errorf("node %d: unable to resolve compute: %w", i, err)
			}
			portID, err := client.BackendPortIDForIP(compute.Ports, computeIP)
			if err != nil {
				return nil, fmt.Errorf("node %d: unable to determine backend port for compute %q (ip %s): %w", i, compute.InstanceName, computeIP, err)
			}
			node.SourceType = "vm"
			node.ResourceType = "compute"
			node.ResourceID = compute.ID
			node.InstanceName = compute.InstanceName
			node.ResourceName = compute.InstanceName
			node.BackendPortID = portID
		case "bm":
			bm, err := r.client.ResolveBaremetalNode(ctx, computeIDAttr.ValueString(), computeNameAttr.ValueString())
			if err != nil {
				return nil, fmt.Errorf("node %d: unable to resolve baremetal server: %w", i, err)
			}
			node.SourceType = "bm"
			node.ResourceType = "baremetal"
			node.ResourceID = bm.UUID
			node.InstanceName = bm.Name
			node.ResourceName = bm.Name
			node.BackendPortID = bm.PortID
		default:
			return nil, fmt.Errorf("node %d: invalid resource_type %q, must be \"compute\" or \"bm\"", i, resourceType)
		}

		if w, ok := attrs["weight"].(types.Int64); ok && !w.IsNull() {
			node.Weight = int(w.ValueInt64())
		}
		if mc, ok := attrs["max_conn"].(types.Int64); ok && !mc.IsNull() {
			node.MaxConn = int(mc.ValueInt64())
		}
		nodes[i] = node
	}
	return nodes, nil
}

func (r *LBVirtualServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LBVirtualServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve lb_service_name -> lb_service_id and vpc_name -> vpc_id, persisting
	// the resolved values into their Computed attributes.
	lbServiceID := data.LBServiceID.ValueString()
	if lbServiceID == "" && !data.LBServiceName.IsNull() {
		resolved, err := r.client.ResolveLBServiceID(ctx, data.LBServiceName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("LB Service Resolution Error", err.Error())
			return
		}
		lbServiceID = resolved
	}
	data.LBServiceID = types.StringValue(lbServiceID)

	vpcID := data.VPCID.ValueString()
	if vpcID == "" && !data.VPCName.IsNull() {
		resolved, err := r.client.ResolveVPCID(ctx, data.VPCName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VPC Resolution Error", err.Error())
			return
		}
		vpcID = resolved
	}
	data.VPCID = types.StringValue(vpcID)

	// Resolve the vip fixed IP to its VIP port id. ListLBVips requires the LB
	// service's network scope (subnet-id header), so scope the client first.
	lbService, err := r.client.GetLBService(ctx, data.LBServiceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB service: %s", err))
		return
	}
	scopedClient := r.client.WithSubnetID(lbService.NetworkID)
	vipPortID, err := scopedClient.ResolveVipPortID(ctx, data.LBServiceID.ValueString(), data.VIP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("VIP Resolution Error", err.Error())
		return
	}
	data.VipPortID = types.Int64Value(int64(vipPortID))

	nodes, err := r.extractNodes(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Wait for all backend VM instances to be active before creating the virtual
	// server. Baremetal members were already validated as "Ready" during resolution.
	for _, node := range nodes {
		if node.SourceType != "vm" {
			continue
		}
		tflog.Info(ctx, "Waiting for backend compute instance to be active", map[string]interface{}{
			"resource_id": node.ResourceID,
		})
		_, err := r.client.WaitForComputeReady(ctx, node.ResourceID, createTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Backend compute instance %s is not ready: %s", node.ResourceID, err))
			return
		}
	}

	formData := client.BuildVirtualServerFormData(client.VirtualServerCreateParams{
		Name:               data.Name.ValueString(),
		Protocol:           data.Protocol.ValueString(),
		VPCID:              data.VPCID.ValueString(),
		RoutingAlgorithm:   data.RoutingAlgorithm.ValueString(),
		MonitorProtocol:    data.MonitorProtocol.ValueString(),
		CertificateID:      data.CertificateID.ValueString(),
		PoolName:           data.PoolName.ValueString(),
		MonitorName:        data.MonitorName.ValueString(),
		PersistenceType:    data.PersistenceType.ValueString(),
		VipPortID:          vipPortID,
		Port:               int(data.Port.ValueInt64()),
		Interval:           int(data.Interval.ValueInt64()),
		MonitorPort:        int(data.MonitorPort.ValueInt64()),
		Timeout:            int(data.Timeout.ValueInt64()),
		PersistenceEnabled: data.PersistenceEnabled.ValueBool(),
		XForwardedFor:      data.XForwardedFor.ValueBool(),
		RedirectHTTPS:      data.RedirectHTTPS.ValueBool(),
		Nodes:              nodes,
	})

	vs, err := r.client.CreateVirtualServer(ctx, data.LBServiceID.ValueString(), formData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create virtual server, got error: %s", err))
		return
	}

	// Wait for virtual server to become ready
	readyVS, err := r.client.WaitForVirtualServerReady(ctx, data.LBServiceID.ValueString(), vs.ID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for virtual server to be ready: %s", err))
		return
	}

	data.ID = types.StringValue(readyVS.ID)
	data.Status = types.StringValue(readyVS.Status)
	// data.VIP is a user-supplied input (the fixed IP); leave it as configured.

	tflog.Trace(ctx, "created LB virtual server resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVirtualServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LBVirtualServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.client.GetVirtualServer(ctx, data.LBServiceID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read virtual server, got error: %s", err))
		return
	}

	if vs.Name != "" {
		data.Name = types.StringValue(vs.Name)
	}
	data.Protocol = types.StringValue(vs.Protocol)
	data.Port = types.Int64Value(int64(vs.Port))
	data.RoutingAlgorithm = types.StringValue(vs.RoutingAlgorithm)
	data.PersistenceEnabled = types.BoolValue(vs.PersistenceEnabled)
	if vs.PersistenceType != "" {
		data.PersistenceType = types.StringValue(vs.PersistenceType)
	}
	data.XForwardedFor = types.BoolValue(vs.XForwardedFor)
	data.RedirectHTTPS = types.BoolValue(vs.RedirectHTTPS)
	data.Status = types.StringValue(vs.Status)
	// data.VIP is a user-supplied input (the fixed IP); do not overwrite it.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVirtualServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LBVirtualServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formData := map[string]interface{}{
		"routing_algorithm":   data.RoutingAlgorithm.ValueString(),
		"persistence_enabled": strconv.FormatBool(data.PersistenceEnabled.ValueBool()),
		"x_forwarded_for":     strconv.FormatBool(data.XForwardedFor.ValueBool()),
	}
	if !data.PersistenceType.IsNull() {
		formData["persistence_type"] = data.PersistenceType.ValueString()
	}

	vs, err := r.client.UpdateVirtualServer(ctx, data.LBServiceID.ValueString(), data.ID.ValueString(), formData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update virtual server, got error: %s", err))
		return
	}

	data.RoutingAlgorithm = types.StringValue(vs.RoutingAlgorithm)
	data.PersistenceEnabled = types.BoolValue(vs.PersistenceEnabled)
	if vs.PersistenceType != "" {
		data.PersistenceType = types.StringValue(vs.PersistenceType)
	}
	data.XForwardedFor = types.BoolValue(vs.XForwardedFor)
	data.Status = types.StringValue(vs.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVirtualServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LBVirtualServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVirtualServer(ctx, data.LBServiceID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete virtual server, got error: %s", err))
		return
	}

	err = r.client.WaitForVirtualServerDeleted(ctx, data.LBServiceID.ValueString(), data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for virtual server deletion: %s", err))
		return
	}
}

func (r *LBVirtualServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: lb_service_id/virtual_server_id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: lb_service_id/virtual_server_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lb_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
