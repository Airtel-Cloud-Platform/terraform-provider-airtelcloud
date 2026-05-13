package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func NewLBVirtualServerResource() resource.Resource {
	return &LBVirtualServerResource{}
}

type LBVirtualServerResource struct {
	client *client.Client
}

type LBVirtualServerResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	LBServiceID        types.String   `tfsdk:"lb_service_id"`
	Name               types.String   `tfsdk:"name"`
	VipPortID          types.Int64    `tfsdk:"vip_port_id"`
	Protocol           types.String   `tfsdk:"protocol"`
	Port               types.Int64    `tfsdk:"port"`
	RoutingAlgorithm   types.String   `tfsdk:"routing_algorithm"`
	VPCID              types.String   `tfsdk:"vpc_id"`
	Interval           types.Int64    `tfsdk:"interval"`
	Nodes              types.List     `tfsdk:"nodes"`
	PersistenceEnabled types.Bool     `tfsdk:"persistence_enabled"`
	PersistenceType    types.String   `tfsdk:"persistence_type"`
	XForwardedFor      types.Bool     `tfsdk:"x_forwarded_for"`
	RedirectHTTPS      types.Bool     `tfsdk:"redirect_https"`
	CertificateID      types.String   `tfsdk:"certificate_id"`
	MonitorProtocol    types.String   `tfsdk:"monitor_protocol"`
	Status             types.String   `tfsdk:"status"`
	VIP                types.String   `tfsdk:"vip"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

var nodeObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"compute_id": types.Int64Type,
		"compute_ip": types.StringType,
		"port":       types.Int64Type,
		"weight":     types.Int64Type,
		"max_conn":   types.Int64Type,
	},
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
				MarkdownDescription: "The ID of the parent LB service.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the virtual server.",
				Optional:            true,
			},
			"vip_port_id": schema.Int64Attribute{
				MarkdownDescription: "The VIP port ID to bind this virtual server to.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
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
				MarkdownDescription: "The VPC ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
						"compute_id": schema.Int64Attribute{
							MarkdownDescription: "The compute instance ID.",
							Required:            true,
						},
						"compute_ip": schema.StringAttribute{
							MarkdownDescription: "The compute instance IP address.",
							Required:            true,
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
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the virtual server.",
				Computed:            true,
			},
			"vip": schema.StringAttribute{
				MarkdownDescription: "The virtual IP address.",
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

func (r *LBVirtualServerResource) extractNodes(ctx context.Context, data *LBVirtualServerResourceModel) ([]models.VirtualServerNode, error) {
	var nodeObjects []types.Object
	diags := data.Nodes.ElementsAs(ctx, &nodeObjects, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract nodes")
	}

	nodes := make([]models.VirtualServerNode, len(nodeObjects))
	for i, obj := range nodeObjects {
		attrs := obj.Attributes()
		nodes[i] = models.VirtualServerNode{
			ComputeID: int(attrs["compute_id"].(types.Int64).ValueInt64()),
			ComputeIP: attrs["compute_ip"].(types.String).ValueString(),
			Port:      int(attrs["port"].(types.Int64).ValueInt64()),
		}
		if w, ok := attrs["weight"].(types.Int64); ok && !w.IsNull() {
			nodes[i].Weight = int(w.ValueInt64())
		}
		if mc, ok := attrs["max_conn"].(types.Int64); ok && !mc.IsNull() {
			nodes[i].MaxConn = int(mc.ValueInt64())
		}
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

	nodes, err := r.extractNodes(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Wait for all backend compute instances to be active before creating virtual server
	for _, node := range nodes {
		computeID := strconv.Itoa(node.ComputeID)
		tflog.Info(ctx, "Waiting for backend compute instance to be active", map[string]interface{}{
			"compute_id": computeID,
		})
		_, err := r.client.WaitForComputeReady(ctx, computeID, createTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Backend compute instance %s is not ready: %s", computeID, err))
			return
		}
	}

	params := client.BuildVirtualServerParams(
		data.Name.ValueString(),
		data.Protocol.ValueString(),
		data.VPCID.ValueString(),
		data.RoutingAlgorithm.ValueString(),
		data.MonitorProtocol.ValueString(),
		data.CertificateID.ValueString(),
		int(data.VipPortID.ValueInt64()),
		int(data.Port.ValueInt64()),
		int(data.Interval.ValueInt64()),
		data.PersistenceEnabled.ValueBool(),
		data.XForwardedFor.ValueBool(),
		data.RedirectHTTPS.ValueBool(),
		data.PersistenceType.ValueString(),
		nodes,
	)

	vs, err := r.client.CreateVirtualServer(ctx, data.LBServiceID.ValueString(), params)
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
	if readyVS.VIP != "" {
		data.VIP = types.StringValue(readyVS.VIP)
	}

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
	if vs.VIP != "" {
		data.VIP = types.StringValue(vs.VIP)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBVirtualServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LBVirtualServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("routing_algorithm", data.RoutingAlgorithm.ValueString())
	params.Set("persistence_enabled", strconv.FormatBool(data.PersistenceEnabled.ValueBool()))
	if !data.PersistenceType.IsNull() {
		params.Set("persistence_type", data.PersistenceType.ValueString())
	}
	params.Set("x_forwarded_for", strconv.FormatBool(data.XForwardedFor.ValueBool()))

	vs, err := r.client.UpdateVirtualServer(ctx, data.LBServiceID.ValueString(), data.ID.ValueString(), params)
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
