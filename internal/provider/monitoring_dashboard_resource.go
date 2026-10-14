package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

var _ resource.Resource = &MonitoringDashboardResource{}
var _ resource.ResourceWithImportState = &MonitoringDashboardResource{}

func NewMonitoringDashboardResource() resource.Resource {
	return &MonitoringDashboardResource{}
}

type MonitoringDashboardResource struct {
	client *client.Client
}

type MonitoringDashboardResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ResourceType types.String `tfsdk:"resource_type"`
	Active       types.Bool   `tfsdk:"active"`
	Description  types.String `tfsdk:"description"`
	Panels       types.List   `tfsdk:"panels"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

type monitoringDashboardPanelModel struct {
	ID            types.String  `tfsdk:"id"`
	Title         types.String  `tfsdk:"title"`
	ChartType     types.String  `tfsdk:"chart_type"`
	Description   types.String  `tfsdk:"description"`
	X             types.Int64   `tfsdk:"x"`
	Y             types.Int64   `tfsdk:"y"`
	W             types.Int64   `tfsdk:"w"`
	H             types.Int64   `tfsdk:"h"`
	MinW          types.Int64   `tfsdk:"min_w"`
	MinH          types.Int64   `tfsdk:"min_h"`
	DisplayOrder  types.Int64   `tfsdk:"display_order"`
	DisplayConfig types.Dynamic `tfsdk:"display_config"`
	Queries       types.List    `tfsdk:"queries"`
}

type monitoringDashboardQueryModel struct {
	ID                types.String  `tfsdk:"id"`
	QueryTemplate     types.String  `tfsdk:"query_template"`
	QueryType         types.String  `tfsdk:"query_type"`
	RefID             types.String  `tfsdk:"ref_id"`
	DisplayOrder      types.Int64   `tfsdk:"display_order"`
	PlaceholderConfig types.Dynamic `tfsdk:"placeholder_config"`
	SeriesConfig      types.Dynamic `tfsdk:"series_config"`
}

func monitoringDashboardPanelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"title":          types.StringType,
		"chart_type":     types.StringType,
		"description":    types.StringType,
		"x":              types.Int64Type,
		"y":              types.Int64Type,
		"w":              types.Int64Type,
		"h":              types.Int64Type,
		"min_w":          types.Int64Type,
		"min_h":          types.Int64Type,
		"display_order":  types.Int64Type,
		"display_config": types.DynamicType,
		"queries":        types.ListType{ElemType: types.ObjectType{AttrTypes: monitoringDashboardQueryAttrTypes()}},
	}
}

func monitoringDashboardQueryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.StringType,
		"query_template":     types.StringType,
		"query_type":         types.StringType,
		"ref_id":             types.StringType,
		"display_order":      types.Int64Type,
		"placeholder_config": types.DynamicType,
		"series_config":      types.DynamicType,
	}
}

func (r *MonitoringDashboardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_dashboard"
}

func (r *MonitoringDashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud monitoring dashboard definition.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dashboard ID returned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Dashboard name.",
			},
			"resource_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Resource type this dashboard applies to (for example `compute`, `mssql`).",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the dashboard is active. Defaults to `true`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional dashboard description.",
			},
			"panels": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Dashboard panels.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Panel ID returned by the API.",
						},
						"title": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Panel title.",
						},
						"chart_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Chart type. One of `line`, `area`, `gauge`, `stats`, `table`.",
							Validators: []validator.String{
								stringvalidator.OneOf("line", "area", "gauge", "stats", "table"),
							},
						},
						"description": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Panel description.",
						},
						"x": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							MarkdownDescription: "Grid x position. Defaults to `0`.",
						},
						"y": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							MarkdownDescription: "Grid y position. Defaults to `0`.",
						},
						"w": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(12),
							MarkdownDescription: "Grid width (1-12). Defaults to `12`.",
							Validators: []validator.Int64{
								int64validator.Between(1, 12),
							},
						},
						"h": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(4),
							MarkdownDescription: "Grid height. Defaults to `4`.",
						},
						"min_w": schema.Int64Attribute{
							Optional:            true,
							MarkdownDescription: "Minimum panel width.",
						},
						"min_h": schema.Int64Attribute{
							Optional:            true,
							MarkdownDescription: "Minimum panel height.",
						},
						"display_order": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
							MarkdownDescription: "Panel display order. Defaults to `0`.",
						},
						"display_config": schema.DynamicAttribute{
							Optional:            true,
							MarkdownDescription: "Arbitrary display configuration object.",
						},
						"queries": schema.ListNestedAttribute{
							Optional:            true,
							MarkdownDescription: "Queries for this panel.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Query ID returned by the API.",
									},
									"query_template": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Query template with placeholders.",
									},
									"query_type": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Query backend type (for example `prometheus`, `zabbix`, `elasticsearch`).",
									},
									"ref_id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("A"),
										MarkdownDescription: "Reference ID for multi-query panels. Defaults to `A`.",
									},
									"display_order": schema.Int64Attribute{
										Optional:            true,
										MarkdownDescription: "Order of the query within the panel.",
									},
									"placeholder_config": schema.DynamicAttribute{
										Optional:            true,
										MarkdownDescription: "Arbitrary placeholder configuration object.",
									},
									"series_config": schema.DynamicAttribute{
										Optional:            true,
										MarkdownDescription: "Arbitrary series configuration object.",
									},
								},
							},
						},
					},
				},
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

func (r *MonitoringDashboardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitoringDashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitoringDashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildMonitoringDashboardCreateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.CreateMonitoringDashboard(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create monitoring dashboard: %s", err))
		return
	}

	resp.Diagnostics.Append(applyMonitoringDashboardToState(ctx, &data, dashboard)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created monitoring dashboard")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringDashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitoringDashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.GetMonitoringDashboard(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read monitoring dashboard: %s", err))
		return
	}

	resp.Diagnostics.Append(applyMonitoringDashboardToState(ctx, &data, dashboard)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringDashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitoringDashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildMonitoringDashboardUpdateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.UpdateMonitoringDashboard(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update monitoring dashboard: %s", err))
		return
	}

	resp.Diagnostics.Append(applyMonitoringDashboardToState(ctx, &data, dashboard)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitoringDashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitoringDashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMonitoringDashboard(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete monitoring dashboard: %s", err))
		return
	}
}

func (r *MonitoringDashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildMonitoringDashboardCreateRequest(ctx context.Context, data *MonitoringDashboardResourceModel) (*models.CreateMonitoringDashboardRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	active := data.Active.ValueBool()
	req := &models.CreateMonitoringDashboardRequest{
		Name:         data.Name.ValueString(),
		ResourceType: data.ResourceType.ValueString(),
		Active:       &active,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}
	panels, panelDiags := expandMonitoringDashboardPanels(ctx, data.Panels)
	diags.Append(panelDiags...)
	req.Panels = panels
	return req, diags
}

func buildMonitoringDashboardUpdateRequest(ctx context.Context, data *MonitoringDashboardResourceModel) (*models.UpdateMonitoringDashboardRequest, diag.Diagnostics) {
	createReq, diags := buildMonitoringDashboardCreateRequest(ctx, data)
	if diags.HasError() {
		return nil, diags
	}
	return &models.UpdateMonitoringDashboardRequest{
		Name:         createReq.Name,
		ResourceType: createReq.ResourceType,
		Active:       createReq.Active,
		Description:  createReq.Description,
		Panels:       createReq.Panels,
	}, diags
}

func expandMonitoringDashboardPanels(ctx context.Context, list types.List) ([]models.MonitoringDashboardPanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var panels []monitoringDashboardPanelModel
	diags.Append(list.ElementsAs(ctx, &panels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]models.MonitoringDashboardPanel, 0, len(panels))
	for _, p := range panels {
		panel := models.MonitoringDashboardPanel{
			Title:        p.Title.ValueString(),
			ChartType:    p.ChartType.ValueString(),
			X:            p.X.ValueInt64(),
			Y:            p.Y.ValueInt64(),
			W:            p.W.ValueInt64(),
			H:            p.H.ValueInt64(),
			DisplayOrder: p.DisplayOrder.ValueInt64(),
		}
		if !p.ID.IsNull() && !p.ID.IsUnknown() && p.ID.ValueString() != "" {
			id := p.ID.ValueString()
			panel.ID = &id
		}
		if !p.Description.IsNull() && !p.Description.IsUnknown() {
			desc := p.Description.ValueString()
			panel.Description = &desc
		}
		if !p.MinW.IsNull() && !p.MinW.IsUnknown() {
			v := p.MinW.ValueInt64()
			panel.MinW = &v
		}
		if !p.MinH.IsNull() && !p.MinH.IsUnknown() {
			v := p.MinH.ValueInt64()
			panel.MinH = &v
		}
		cfg, cfgDiags := monitoringDynamicToMap(ctx, p.DisplayConfig)
		diags.Append(cfgDiags...)
		panel.DisplayConfig = cfg

		queries, queryDiags := expandMonitoringDashboardQueries(ctx, p.Queries)
		diags.Append(queryDiags...)
		panel.Queries = queries
		out = append(out, panel)
	}
	return out, diags
}

func expandMonitoringDashboardQueries(ctx context.Context, list types.List) ([]models.MonitoringDashboardQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	var queries []monitoringDashboardQueryModel
	diags.Append(list.ElementsAs(ctx, &queries, false)...)
	if diags.HasError() {
		return nil, diags
	}

	out := make([]models.MonitoringDashboardQuery, 0, len(queries))
	for _, q := range queries {
		query := models.MonitoringDashboardQuery{
			QueryTemplate: q.QueryTemplate.ValueString(),
			QueryType:     q.QueryType.ValueString(),
			RefID:         q.RefID.ValueString(),
		}
		if !q.ID.IsNull() && !q.ID.IsUnknown() && q.ID.ValueString() != "" {
			id := q.ID.ValueString()
			query.ID = &id
		}
		if !q.DisplayOrder.IsNull() && !q.DisplayOrder.IsUnknown() {
			v := q.DisplayOrder.ValueInt64()
			query.DisplayOrder = &v
		}
		ph, phDiags := monitoringDynamicToMap(ctx, q.PlaceholderConfig)
		diags.Append(phDiags...)
		query.PlaceholderConfig = ph
		sc, scDiags := monitoringDynamicToMap(ctx, q.SeriesConfig)
		diags.Append(scDiags...)
		query.SeriesConfig = sc
		out = append(out, query)
	}
	return out, diags
}

func applyMonitoringDashboardToState(ctx context.Context, data *MonitoringDashboardResourceModel, dashboard *models.MonitoringDashboard) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(dashboard.ID)
	data.Name = types.StringValue(dashboard.Name)
	data.ResourceType = types.StringValue(dashboard.ResourceType)
	data.Active = types.BoolValue(dashboard.Active)
	if dashboard.Description != nil {
		data.Description = types.StringValue(*dashboard.Description)
	} else {
		data.Description = types.StringNull()
	}
	if dashboard.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*dashboard.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if dashboard.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*dashboard.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	panels, panelDiags := flattenMonitoringDashboardPanels(ctx, dashboard.Panels)
	diags.Append(panelDiags...)
	data.Panels = panels
	return diags
}

func flattenMonitoringDashboardPanels(ctx context.Context, panels []models.MonitoringDashboardPanel) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if panels == nil {
		return types.ListNull(types.ObjectType{AttrTypes: monitoringDashboardPanelAttrTypes()}), diags
	}

	elems := make([]attr.Value, 0, len(panels))
	for _, p := range panels {
		queries, queryDiags := flattenMonitoringDashboardQueries(ctx, p.Queries)
		diags.Append(queryDiags...)

		displayConfig, cfgDiags := monitoringMapToDynamic(ctx, p.DisplayConfig)
		diags.Append(cfgDiags...)

		attrs := map[string]attr.Value{
			"id":             optionalStringPtr(p.ID),
			"title":          types.StringValue(p.Title),
			"chart_type":     types.StringValue(p.ChartType),
			"description":    optionalStringPtr(p.Description),
			"x":              types.Int64Value(p.X),
			"y":              types.Int64Value(p.Y),
			"w":              types.Int64Value(p.W),
			"h":              types.Int64Value(p.H),
			"min_w":          optionalInt64Ptr(p.MinW),
			"min_h":          optionalInt64Ptr(p.MinH),
			"display_order":  types.Int64Value(p.DisplayOrder),
			"display_config": displayConfig,
			"queries":        queries,
		}
		obj, objDiags := types.ObjectValue(monitoringDashboardPanelAttrTypes(), attrs)
		diags.Append(objDiags...)
		elems = append(elems, obj)
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: monitoringDashboardPanelAttrTypes()}, elems)
	diags.Append(listDiags...)
	return list, diags
}

func flattenMonitoringDashboardQueries(ctx context.Context, queries []models.MonitoringDashboardQuery) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if queries == nil {
		return types.ListNull(types.ObjectType{AttrTypes: monitoringDashboardQueryAttrTypes()}), diags
	}

	elems := make([]attr.Value, 0, len(queries))
	for _, q := range queries {
		ph, phDiags := monitoringMapToDynamic(ctx, q.PlaceholderConfig)
		diags.Append(phDiags...)
		sc, scDiags := monitoringMapToDynamic(ctx, q.SeriesConfig)
		diags.Append(scDiags...)

		refID := q.RefID
		if refID == "" {
			refID = "A"
		}
		attrs := map[string]attr.Value{
			"id":                 optionalStringPtr(q.ID),
			"query_template":     types.StringValue(q.QueryTemplate),
			"query_type":         types.StringValue(q.QueryType),
			"ref_id":             types.StringValue(refID),
			"display_order":      optionalInt64Ptr(q.DisplayOrder),
			"placeholder_config": ph,
			"series_config":      sc,
		}
		obj, objDiags := types.ObjectValue(monitoringDashboardQueryAttrTypes(), attrs)
		diags.Append(objDiags...)
		elems = append(elems, obj)
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: monitoringDashboardQueryAttrTypes()}, elems)
	diags.Append(listDiags...)
	return list, diags
}

func optionalStringPtr(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func optionalInt64Ptr(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func monitoringDynamicToMap(ctx context.Context, d types.Dynamic) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if d.IsNull() || d.IsUnknown() {
		return nil, diags
	}
	raw, err := attrValueToInterface(d.UnderlyingValue())
	if err != nil {
		diags.AddError("Invalid Configuration", fmt.Sprintf("Unable to convert dynamic value: %s", err))
		return nil, diags
	}
	if raw == nil {
		return nil, diags
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		diags.AddError("Invalid Configuration", "Dynamic map attributes must be objects.")
		return nil, diags
	}
	_ = ctx
	return m, diags
}

func monitoringMapToDynamic(ctx context.Context, m map[string]interface{}) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m == nil {
		return types.DynamicNull(), diags
	}
	val, err := interfaceToAttrValue(m)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to convert map to dynamic value: %s", err))
		return types.DynamicNull(), diags
	}
	_ = ctx
	return types.DynamicValue(val), diags
}

func attrValueToInterface(val attr.Value) (interface{}, error) {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil, nil
	}
	switch v := val.(type) {
	case types.String:
		return v.ValueString(), nil
	case types.Bool:
		return v.ValueBool(), nil
	case types.Int64:
		return v.ValueInt64(), nil
	case types.Float64:
		return v.ValueFloat64(), nil
	case types.Number:
		f, _ := v.ValueBigFloat().Float64()
		return f, nil
	case types.List:
		elems := v.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			converted, err := attrValueToInterface(e)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	case types.Tuple:
		elems := v.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			converted, err := attrValueToInterface(e)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	case types.Set:
		elems := v.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			converted, err := attrValueToInterface(e)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	case types.Map:
		elems := v.Elements()
		out := make(map[string]interface{}, len(elems))
		for k, e := range elems {
			converted, err := attrValueToInterface(e)
			if err != nil {
				return nil, err
			}
			out[k] = converted
		}
		return out, nil
	case types.Object:
		attrs := v.Attributes()
		out := make(map[string]interface{}, len(attrs))
		for k, e := range attrs {
			converted, err := attrValueToInterface(e)
			if err != nil {
				return nil, err
			}
			out[k] = converted
		}
		return out, nil
	case types.Dynamic:
		return attrValueToInterface(v.UnderlyingValue())
	default:
		// Fallback via JSON for uncommon attr types.
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("unsupported attr type %T", val)
		}
		var raw interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			return nil, err
		}
		return raw, nil
	}
}

func interfaceToAttrValue(v interface{}) (attr.Value, error) {
	switch val := v.(type) {
	case nil:
		return types.DynamicNull(), nil
	case string:
		return types.StringValue(val), nil
	case bool:
		return types.BoolValue(val), nil
	case int:
		return types.Int64Value(int64(val)), nil
	case int64:
		return types.Int64Value(val), nil
	case float64:
		if val == float64(int64(val)) {
			return types.Int64Value(int64(val)), nil
		}
		return types.Float64Value(val), nil
	case json.Number:
		i, err := val.Int64()
		if err == nil {
			return types.Int64Value(i), nil
		}
		f, err := val.Float64()
		if err != nil {
			return nil, err
		}
		return types.Float64Value(f), nil
	case []interface{}:
		elems := make([]attr.Value, 0, len(val))
		typesList := make([]attr.Type, 0, len(val))
		for _, item := range val {
			converted, err := interfaceToAttrValue(item)
			if err != nil {
				return nil, err
			}
			elems = append(elems, converted)
			typesList = append(typesList, converted.Type(context.Background()))
		}
		return types.TupleValueMust(typesList, elems), nil
	case map[string]interface{}:
		attrs := make(map[string]attr.Value, len(val))
		attrTypes := make(map[string]attr.Type, len(val))
		for k, item := range val {
			converted, err := interfaceToAttrValue(item)
			if err != nil {
				return nil, err
			}
			attrs[k] = converted
			attrTypes[k] = converted.Type(context.Background())
		}
		return types.ObjectValueMust(attrTypes, attrs), nil
	case *big.Float:
		f, _ := val.Float64()
		return types.Float64Value(f), nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", v)
	}
}
