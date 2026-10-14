package provider

import (
	"context"
	"fmt"
	"strconv"

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

var _ resource.Resource = &AlertNotificationResource{}
var _ resource.ResourceWithImportState = &AlertNotificationResource{}

func NewAlertNotificationResource() resource.Resource {
	return &AlertNotificationResource{}
}

// AlertNotificationResource manages airtelcloud_alert_notification.
type AlertNotificationResource struct {
	client *client.Client
}

// AlertNotificationResourceModel is the Terraform state model.
type AlertNotificationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	AlertID   types.String `tfsdk:"alert_id"`
	Type      types.String `tfsdk:"type"`
	Target    types.String `tfsdk:"target"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *AlertNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_notification"
}

func (r *AlertNotificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud alert notification channel (email or SMS).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Notification id as a string (API returns an integer).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alert_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Alert id (string form of the integer Alert.id). Force-new.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Notification type. One of `email`, `sms`. Force-new.",
				Validators:          []validator.String{stringvalidator.OneOf("email", "sms")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination address or phone number.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the notification is enabled. Defaults to true.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp.",
			},
		},
	}
}

func (r *AlertNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AlertNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlertNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertID, err := strconv.Atoi(data.AlertID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert_id"),
			"Invalid Configuration",
			fmt.Sprintf("alert_id must be a numeric string, got %q: %s", data.AlertID.ValueString(), err),
		)
		return
	}

	created, err := r.client.CreateAlertNotification(ctx, &models.CreateAlertNotificationRequest{
		Type:    data.Type.ValueString(),
		Target:  data.Target.ValueString(),
		Enabled: data.Enabled.ValueBool(),
		AlertID: alertID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert notification, got error: %s", err))
		return
	}

	mapAlertNotification(&data, created)
	tflog.Trace(ctx, "created an alert notification", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlertNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	n, err := r.client.GetAlertNotification(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alert notification, got error: %s", err))
		return
	}

	mapAlertNotification(&data, n)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlertNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := data.Target.ValueString()
	enabled := data.Enabled.ValueBool()
	updated, err := r.client.UpdateAlertNotification(ctx, data.ID.ValueString(), &models.UpdateAlertNotificationRequest{
		Target:  &target,
		Enabled: &enabled,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert notification, got error: %s", err))
		return
	}

	mapAlertNotification(&data, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlertNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAlertNotification(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete alert notification, got error: %s", err))
		return
	}
}

func (r *AlertNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapAlertNotification(data *AlertNotificationResourceModel, n *models.AlertNotification) {
	data.ID = types.StringValue(strconv.Itoa(n.ID))
	data.AlertID = types.StringValue(strconv.Itoa(n.AlertID))
	data.Type = types.StringValue(n.Type)
	data.Target = types.StringValue(n.Target)
	data.Enabled = types.BoolValue(n.Enabled)
	data.CreatedAt = types.StringValue(n.CreatedAt)
}
