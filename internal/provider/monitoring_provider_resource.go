package provider

import (
	"context"
	"fmt"

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

var _ resource.Resource = &MonitoringProviderResource{}
var _ resource.ResourceWithImportState = &MonitoringProviderResource{}

func NewMonitoringProviderResource() resource.Resource {
	return &MonitoringProviderResource{}
}

type MonitoringProviderResource struct {
	client *client.Client
}

type MonitoringProviderResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ResourceType     types.String `tfsdk:"resource_type"`
	ProviderType     types.String `tfsdk:"provider_type"`
	DashboardAuthURL types.String `tfsdk:"dashboard_auth_url"`
	APIKeyName       types.String `tfsdk:"api_key_name"`
	APIToken         types.String `tfsdk:"api_token"`
	DataSourceUID    types.String `tfsdk:"data_source_uid"`
	DataSourceID     types.String `tfsdk:"data_source_id"`
	AZName           types.String `tfsdk:"az_name"`
	Active           types.Bool   `tfsdk:"active"`
	Description      types.String `tfsdk:"description"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *MonitoringProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_provider"
}

func (r *MonitoringProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud monitoring metrics provider (Prometheus, Zabbix, or Elasticsearch).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Provider ID returned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Provider name.",
			},
			"resource_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Resource type this provider serves (for example `compute`, `mssql`).",
			},
			"provider_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Provider backend type. One of `prometheus`, `zabbix`, `elasticsearch`.",
				Validators: []validator.String{
					stringvalidator.OneOf("prometheus", "zabbix", "elasticsearch"),
				},
			},
			"dashboard_auth_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Dashboard authentication URL.",
			},
			"api_key_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "API key name.",
			},
			"api_token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "API token. Required on create. On update, omit or leave unchanged to keep the existing token. Never returned by the API.",
			},
			"data_source_uid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Data source UID.",
			},
			"data_source_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Data source ID.",
			},
			"az_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional availability zone name.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the provider is active. Defaults to `true`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional provider description.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp from the API.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp from the API.",
			},
		},
	}
}

func (r *MonitoringProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitoringProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitoringProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildMonitoringProviderCreateRequest(&data)
	provider, err := r.client.CreateMonitoringProvider(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create monitoring provider: %s", err))
		return
	}

	// Preserve api_token from plan — API never returns it.
	token := data.APIToken
	applyMonitoringProviderToState(&data, provider)
	data.APIToken = token

	tflog.Trace(ctx, "created monitoring provider")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitoringProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve sensitive token from state; API never returns it.
	token := data.APIToken

	provider, err := r.client.GetMonitoringProvider(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read monitoring provider: %s", err))
		return
	}

	applyMonitoringProviderToState(&data, provider)
	data.APIToken = token
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitoringProviderResourceModel
	var state MonitoringProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildMonitoringProviderUpdateRequest(&plan, &state)
	provider, err := r.client.UpdateMonitoringProvider(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update monitoring provider: %s", err))
		return
	}

	token := plan.APIToken
	applyMonitoringProviderToState(&plan, provider)
	data := plan
	// Never overwrite token with empty API response.
	if token.IsNull() || token.IsUnknown() || token.ValueString() == "" {
		data.APIToken = state.APIToken
	} else {
		data.APIToken = token
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitoringProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMonitoringProvider(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete monitoring provider: %s", err))
		return
	}
}

func (r *MonitoringProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildMonitoringProviderCreateRequest(data *MonitoringProviderResourceModel) *models.CreateMonitoringProviderRequest {
	active := data.Active.ValueBool()
	req := &models.CreateMonitoringProviderRequest{
		Name:             data.Name.ValueString(),
		ResourceType:     data.ResourceType.ValueString(),
		DashboardAuthURL: data.DashboardAuthURL.ValueString(),
		APIKeyName:       data.APIKeyName.ValueString(),
		APIToken:         data.APIToken.ValueString(),
		DataSourceUID:    data.DataSourceUID.ValueString(),
		DataSourceID:     data.DataSourceID.ValueString(),
		Active:           &active,
	}
	if !data.ProviderType.IsNull() && !data.ProviderType.IsUnknown() && data.ProviderType.ValueString() != "" {
		pt := data.ProviderType.ValueString()
		req.ProviderType = &pt
	}
	if !data.AZName.IsNull() && !data.AZName.IsUnknown() {
		az := data.AZName.ValueString()
		req.AZName = &az
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}
	return req
}

func buildMonitoringProviderUpdateRequest(plan, state *MonitoringProviderResourceModel) *models.UpdateMonitoringProviderRequest {
	active := plan.Active.ValueBool()
	req := &models.UpdateMonitoringProviderRequest{
		Name:             plan.Name.ValueString(),
		ResourceType:     plan.ResourceType.ValueString(),
		DashboardAuthURL: plan.DashboardAuthURL.ValueString(),
		APIKeyName:       plan.APIKeyName.ValueString(),
		DataSourceUID:    plan.DataSourceUID.ValueString(),
		DataSourceID:     plan.DataSourceID.ValueString(),
		Active:           &active,
	}
	if !plan.ProviderType.IsNull() && !plan.ProviderType.IsUnknown() && plan.ProviderType.ValueString() != "" {
		pt := plan.ProviderType.ValueString()
		req.ProviderType = &pt
	}
	if !plan.AZName.IsNull() && !plan.AZName.IsUnknown() {
		az := plan.AZName.ValueString()
		req.AZName = &az
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc := plan.Description.ValueString()
		req.Description = &desc
	}

	// Omit api_token when unchanged or empty so the API keeps the existing token.
	token := plan.APIToken.ValueString()
	if token != "" && (state.APIToken.IsNull() || state.APIToken.IsUnknown() || token != state.APIToken.ValueString()) {
		req.APIToken = &token
	}
	return req
}

func applyMonitoringProviderToState(data *MonitoringProviderResourceModel, provider *models.MonitoringProvider) {
	data.ID = types.StringValue(provider.ID)
	data.Name = types.StringValue(provider.Name)
	data.ResourceType = types.StringValue(provider.ResourceType)
	if provider.ProviderType != "" {
		data.ProviderType = types.StringValue(provider.ProviderType)
	}
	data.DashboardAuthURL = types.StringValue(provider.DashboardAuthURL)
	data.APIKeyName = types.StringValue(provider.APIKeyName)
	data.DataSourceUID = types.StringValue(provider.DataSourceUID)
	data.DataSourceID = types.StringValue(provider.DataSourceID)
	data.Active = types.BoolValue(provider.Active)
	if provider.AZName != nil {
		data.AZName = types.StringValue(*provider.AZName)
	} else if data.AZName.IsUnknown() {
		data.AZName = types.StringNull()
	}
	if provider.Description != nil {
		data.Description = types.StringValue(*provider.Description)
	} else if data.Description.IsUnknown() {
		data.Description = types.StringNull()
	}
	if provider.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*provider.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if provider.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*provider.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}
}
