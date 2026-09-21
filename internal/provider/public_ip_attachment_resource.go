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

var _ resource.Resource = &PublicIPAttachmentResource{}
var _ resource.ResourceWithImportState = &PublicIPAttachmentResource{}
var _ resource.ResourceWithValidateConfig = &PublicIPAttachmentResource{}

const (
	defaultPublicIPAttachTimeout = 20 * time.Minute
	defaultPublicIPDetachTimeout = 20 * time.Minute
)

func NewPublicIPAttachmentResource() resource.Resource {
	return &PublicIPAttachmentResource{}
}

type PublicIPAttachmentResource struct {
	client *client.Client
}

type PublicIPAttachmentResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	PublicIPName     types.String   `tfsdk:"public_ip_name"`
	ResourceType     types.String   `tfsdk:"resource_type"`
	ResourceName     types.String   `tfsdk:"resource_name"`
	TargetVIP        types.String   `tfsdk:"target_vip"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	PublicIP         types.String   `tfsdk:"public_ip"`
	Status           types.String   `tfsdk:"status"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func (r *PublicIPAttachmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip_attachment"
}

func (r *PublicIPAttachmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a reserved Airtel Cloud Public IP to a virtual machine, load balancer, or baremetal server. Terraform takes names only: the availability zone is read from the public IP, and the target private IP is looked up from `resource_type` + `resource_name`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the attached public IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip_name": schema.StringAttribute{
				MarkdownDescription: "The `object_name` of the reserved public IP.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Type of resource to attach to: `vm`, `lb`, or `baremetal`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "Name of the VM, load balancer, or baremetal server to attach to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_vip": schema.StringAttribute{
				MarkdownDescription: "The private IP (VIP) looked up from the named VM, load balancer, or baremetal server.",
				Computed:            true,
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone, read from the public IP.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The allocated public IP address.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the public IP after attach.",
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

func (r *PublicIPAttachmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PublicIPAttachmentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data PublicIPAttachmentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ResourceType.IsNull() || data.ResourceType.IsUnknown() {
		return
	}

	if _, err := client.NormalizePublicIPResourceType(data.ResourceType.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("resource_type"), "Invalid Configuration", err.Error())
	}
}

func (r *PublicIPAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PublicIPAttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, defaultPublicIPAttachTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicIP, err := r.client.GetPublicIPByName(ctx, data.PublicIPName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find public IP named %s: %s", data.PublicIPName.ValueString(), err))
		return
	}

	az := getPublicIPAZName(publicIP)
	if az == "" {
		resp.Diagnostics.AddError(
			"Missing Availability Zone",
			fmt.Sprintf("Public IP %q does not report an availability zone, so the attach request cannot be scoped.", data.PublicIPName.ValueString()),
		)
		return
	}

	portID, resolvedVIP, err := r.client.FindPortForResource(ctx, data.ResourceType.ValueString(), data.ResourceName.ValueString(), "", az)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve attach port for %s %q: %s", data.ResourceType.ValueString(), data.ResourceName.ValueString(), err))
		return
	}
	if resolvedVIP == "" {
		resp.Diagnostics.AddError(
			"Missing Target VIP",
			fmt.Sprintf("Resource %s %q has no private IP/VIP to attach against.", data.ResourceType.ValueString(), data.ResourceName.ValueString()),
		)
		return
	}

	tflog.Debug(ctx, "Attaching public IP", map[string]interface{}{
		"public_ip_name":    data.PublicIPName.ValueString(),
		"public_ip_uuid":    publicIP.UUID,
		"resource_type":     data.ResourceType.ValueString(),
		"resource_name":     data.ResourceName.ValueString(),
		"target_vip":        resolvedVIP,
		"port_id":           portID,
		"availability_zone": az,
	})

	if err := r.client.AttachPublicIP(ctx, publicIP.UUID, portID, az); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach public IP, got error: %s", err))
		return
	}

	readyIP, err := r.client.WaitForPublicIPAttached(ctx, publicIP.UUID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for public IP to attach: %s", err))
		return
	}

	if readyIP.TargetVIP != "" {
		resolvedVIP = readyIP.TargetVIP
	}

	data.ID = types.StringValue(readyIP.UUID)
	data.Status = types.StringValue(readyIP.Status)
	data.PublicIP = types.StringValue(getPublicIPAddr(readyIP))
	data.TargetVIP = types.StringValue(resolvedVIP)
	data.AvailabilityZone = types.StringValue(az)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PublicIPAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicIP, err := r.client.GetPublicIP(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read public IP attachment, got error: %s", err))
		return
	}

	if strings.EqualFold(strings.TrimSpace(publicIP.Status), "reserved") || publicIP.TargetVIP == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(publicIP.UUID)
	if name := models.PublicIPDisplayName(*publicIP); name != "" {
		data.PublicIPName = types.StringValue(name)
	}
	data.Status = types.StringValue(publicIP.Status)
	data.PublicIP = types.StringValue(getPublicIPAddr(publicIP))
	data.TargetVIP = types.StringValue(publicIP.TargetVIP)
	if az := getPublicIPAZName(publicIP); az != "" {
		data.AvailabilityZone = types.StringValue(az)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PublicIPAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Public IP attachments cannot be updated in place. All changes require replacement.")
}

func (r *PublicIPAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PublicIPAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultPublicIPDetachTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	az := data.AvailabilityZone.ValueString()
	uuid := data.ID.ValueString()

	portID := 0
	if current, err := r.client.GetPublicIP(ctx, uuid); err == nil && current != nil && current.PortID != 0 {
		portID = current.PortID
	} else {
		resolved, _, resolveErr := r.client.FindPortForResource(ctx, data.ResourceType.ValueString(), data.ResourceName.ValueString(), "", az)
		if resolveErr != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resolve port for detach: %s", resolveErr))
			return
		}
		portID = resolved
	}

	if err := r.client.DetachPublicIP(ctx, uuid, portID, az); err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to detach public IP, got error: %s", err))
		return
	}

	deadline := time.Now().Add(deleteTimeout)
	for time.Now().Before(deadline) {
		ip, err := r.client.GetPublicIP(ctx, uuid)
		if err != nil {
			if client.IsNotFoundError(err) {
				return
			}
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for public IP detach: %s", err))
			return
		}
		status := strings.ToLower(strings.TrimSpace(ip.Status))
		if status == "reserved" || ip.TargetVIP == "" {
			return
		}
		time.Sleep(5 * time.Second)
	}

	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Public IP did not return to reserved within %s", deleteTimeout))
}

func (r *PublicIPAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
