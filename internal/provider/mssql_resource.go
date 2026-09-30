package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const mssqlDefaultTimeout = 15 * time.Minute

var (
	mssqlNamePattern       = regexp.MustCompile(`^[a-z0-9-]+$`)
	mssqlIdentifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
)

var _ resource.Resource = &MSSQLResource{}
var _ resource.ResourceWithImportState = &MSSQLResource{}
var _ resource.ResourceWithValidateConfig = &MSSQLResource{}

func NewMSSQLResource() resource.Resource {
	return &MSSQLResource{}
}

// MSSQLResource defines the airtelcloud_mssql resource.
type MSSQLResource struct {
	client *client.Client
}

// MSSQLResourceModel is the Terraform state model. It is separate from API structs.
type MSSQLResourceModel struct {
	ID                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	Description             types.String   `tfsdk:"description"`
	DatabaseEdition         types.String   `tfsdk:"database_edition"`
	Collation               types.String   `tfsdk:"collation"`
	Features                types.List     `tfsdk:"features"`
	ComputeSize             types.String   `tfsdk:"compute_size"`
	StorageSize             types.Int64    `tfsdk:"storage_size"`
	StorageType             types.String   `tfsdk:"storage_type"`
	AvailabilityZone        types.String   `tfsdk:"availability_zone"`
	VPC                     types.String   `tfsdk:"vpc"`
	Subnet                  types.String   `tfsdk:"subnet"`
	DBName                  types.String   `tfsdk:"db_name"`
	MSSQLUsername           types.String   `tfsdk:"mssql_username"`
	Password                types.String   `tfsdk:"password"`
	Labels                  types.List     `tfsdk:"labels"`
	EnableTDE               types.Bool     `tfsdk:"enable_tde"`
	AdvancedSettingsEnabled types.Bool     `tfsdk:"advanced_settings_enabled"`
	Version                 types.String   `tfsdk:"version"`
	ConnectionString        types.String   `tfsdk:"connection_string"`
	CreatedAt               types.String   `tfsdk:"created_at"`
	Port                    types.Int64    `tfsdk:"port"`
	Status                  types.String   `tfsdk:"status"`
	Backup                  types.Object   `tfsdk:"backup"`
	SecurityGroup           types.Object   `tfsdk:"security_group"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}

// MSSQLBackupModel is the nested backup attribute. Enabled is derived from protection_plan.
type MSSQLBackupModel struct {
	ProtectionPlan types.String `tfsdk:"protection_plan"`
}

// MSSQLSecurityGroupModel is the nested security_group attribute.
type MSSQLSecurityGroupModel struct {
	AllowedIPs types.List `tfsdk:"allowed_ips"`
}

func (r *MSSQLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mssql"
}

func (r *MSSQLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud MSSQL instance. Create, read, import, and delete are supported. Changes force a new instance. `vpc` and `subnet` are stored for later use and are not sent to the API.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the MSSQL instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Instance name. Lowercase letters, digits, and hyphens only.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(mssqlNamePattern, "name must contain only lowercase letters, digits, and hyphens."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the instance.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_edition": schema.StringAttribute{
				MarkdownDescription: "Database edition, such as `web`, `standard`, or `enterprise`. Version is taken from that edition's catalog entry.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collation": schema.StringAttribute{
				MarkdownDescription: "Collation name from the collations catalog for the selected edition.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"features": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Feature values for the collation, such as `SQLENGINE`. Each value must be offered for the selected edition.",
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"compute_size": schema.StringAttribute{
				MarkdownDescription: "Flavor name from the flavors catalog for the selected edition.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_size": schema.Int64Attribute{
				MarkdownDescription: "Storage size in GiB. Must be from 200 to 2048 in steps of 10 starting at 200.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(models.MSSQLStorageMinGiB, models.MSSQLStorageMaxGiB),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"storage_type": schema.StringAttribute{
				MarkdownDescription: "Volume type label, such as `High Performance`. Defaults to `High Performance`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(models.MSSQLDefaultStorageType),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone code, such as `S1`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc": schema.StringAttribute{
				MarkdownDescription: "VPC name collected for later use. Not sent on create.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "Subnet name collected for later use. Not sent on create.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_name": schema.StringAttribute{
				MarkdownDescription: "Initial database name. Must start with a letter and contain only letters, digits, and underscore.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(mssqlIdentifierPattern, "db_name must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mssql_username": schema.StringAttribute{
				MarkdownDescription: "Admin username. Must start with a letter and contain only letters, digits, and underscore.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(mssqlIdentifierPattern, "mssql_username must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Admin password. The API does not return this value. The same value is sent as confirm_password.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Labels sent as add_labels.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"enable_tde": schema.BoolAttribute{
				MarkdownDescription: "Whether transparent data encryption is enabled. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"advanced_settings_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether advanced settings are enabled. Defaults to false. Maintenance windows are not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Database version resolved from the selected edition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_string": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Connection string returned by the API.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp returned by the API.",
			},
			"port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Database port returned by the API.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Instance status. Create waits until Active. Requested, Initiated, and Resizing are still in progress. Failed fails the operation.",
			},
			"backup": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Backup settings. When protection_plan is set, backup is enabled and the plan is resolved from the netbackup catalog. When omitted, backup is disabled.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"protection_plan": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Protection plan name from the netbackup catalog.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"security_group": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Client addresses allowed to connect.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"allowed_ips": schema.ListAttribute{
						ElementType:         types.StringType,
						Required:            true,
						MarkdownDescription: "CIDR blocks allowed to connect, for example `192.168.1.0/24`.",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
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

func (r *MSSQLResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data MSSQLResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.StorageSize.IsNull() && !data.StorageSize.IsUnknown() && !models.MSSQLStorageSizeAllowed(data.StorageSize.ValueInt64()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("storage_size"),
			"Invalid Configuration",
			fmt.Sprintf("storage_size must be from %d to %d GiB in steps of %d starting at %d.", models.MSSQLStorageMinGiB, models.MSSQLStorageMaxGiB, models.MSSQLStorageStepGiB, models.MSSQLStorageMinGiB),
		)
	}
}

func (r *MSSQLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MSSQLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MSSQLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, mssqlDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, version, diags := r.buildMSSQLCreateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateMSSQL(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create mssql instance, got error: %s", err))
		return
	}
	if created.UUID == "" {
		resp.Diagnostics.AddError("Client Error", "Create mssql instance returned an empty uuid.")
		return
	}

	ready, err := r.client.WaitForMSSQLReady(ctx, created.UUID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for mssql instance to be ready: %s", err))
		return
	}

	data.Version = types.StringValue(version)
	resp.Diagnostics.Append(applyMSSQLToState(ctx, &data, ready)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created mssql resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MSSQLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MSSQLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.GetMSSQL(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read mssql instance, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(applyMSSQLToState(ctx, &data, instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MSSQLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"airtelcloud_mssql does not support in-place updates. Changing configuration forces a new instance.",
	)
}

func (r *MSSQLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MSSQLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, mssqlDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMSSQL(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete mssql instance, got error: %s", err))
		return
	}

	if err := r.client.WaitForMSSQLDeleted(ctx, data.ID.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for mssql instance deletion: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted mssql resource")
}

func (r *MSSQLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MSSQLResource) buildMSSQLCreateRequest(ctx context.Context, data *MSSQLResourceModel) (*models.CreateMSSQLRequest, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	az := data.AvailabilityZone.ValueString()

	zone, err := r.client.ResolvePostgresAvailabilityZone(ctx, az)
	if err != nil {
		diags.AddError("Availability Zone Resolution Error", err.Error())
		return nil, "", diags
	}

	edition, version, err := r.client.ResolveMSSQLEdition(ctx, data.DatabaseEdition.ValueString())
	if err != nil {
		diags.AddError("Edition Resolution Error", err.Error())
		return nil, "", diags
	}

	features, featureDiags := stringSliceFromList(ctx, data.Features)
	diags.Append(featureDiags...)
	labels, labelDiags := stringSliceFromList(ctx, data.Labels)
	diags.Append(labelDiags...)
	if diags.HasError() {
		return nil, "", diags
	}
	if labels == nil {
		labels = []string{}
	}

	if err := r.client.ResolveMSSQLCollation(ctx, edition.Edition, data.Collation.ValueString(), features); err != nil {
		diags.AddError("Collation Resolution Error", err.Error())
		return nil, "", diags
	}

	flavor, ramGB, err := r.client.ResolveMSSQLFlavor(ctx, zone.AZCode, edition.Edition, data.ComputeSize.ValueString())
	if err != nil {
		diags.AddError("Flavor Resolution Error", err.Error())
		return nil, "", diags
	}

	storage, err := r.client.ResolveMSSQLStorage(ctx, zone.AZCode, data.StorageType.ValueString())
	if err != nil {
		diags.AddError("Storage Resolution Error", err.Error())
		return nil, "", diags
	}

	size := strconv.FormatInt(data.StorageSize.ValueInt64(), 10)
	password := data.Password.ValueString()
	createReq := &models.CreateMSSQLRequest{
		Name:        data.Name.ValueString(),
		User:        models.MSSQLFixedUser,
		Description: data.Description.ValueString(),
		AZName:      zone.AZCode,
		AZIDs:       []string{zone.AZCode},
		CollationFeatures: models.MSSQLCollationFeatures{
			Collation: data.Collation.ValueString(),
			Features:  features,
		},
		DBConfig: models.MSSQLDBConfig{
			Version: version,
			Edition: edition.Edition,
		},
		ComputeStorageConfig: models.MSSQLComputeStorageConfig{
			Flavor: models.MSSQLFlavorRef{
				Name:  flavor.Name,
				VCPUs: flavor.VCPUs,
				RAM:   ramGB,
			},
			FlavorIDs: []models.MSSQLFlavorAZID{{
				AZID: zone.AZCode,
				ID:   flavor.ID,
			}},
			TPS:                models.MSSQLFixedTPS,
			PerTransactionSize: models.MSSQLFixedPerTransactionSize,
			DatabaseSize:       size,
			StorageType:        storage.Name,
		},
		Backup: models.MSSQLBackup{Enabled: false},
		AdvancedSettings: models.MSSQLAdvancedSettings{
			Enabled: data.AdvancedSettingsEnabled.ValueBool(),
		},
		VolumeSize:           size,
		VolumeTypeProviderID: storage.ProviderVTID,
		VolumeTypeID:         storage.ID,
		DBName:               data.DBName.ValueString(),
		MSSQLUsername:        data.MSSQLUsername.ValueString(),
		Password:             password,
		ConfirmPassword:      password,
		AddLabels:            labels,
		EnableTDE:            data.EnableTDE.ValueBool(),
	}

	protectionPlan, hasBackup, backupDiags := mssqlProtectionPlan(ctx, data.Backup)
	diags.Append(backupDiags...)
	if diags.HasError() {
		return nil, "", diags
	}
	if hasBackup {
		plan, err := r.client.ResolveMSSQLProtectionPlan(ctx, zone.AZCode, protectionPlan)
		if err != nil {
			diags.AddError("Protection Plan Resolution Error", err.Error())
			return nil, "", diags
		}
		createReq.Backup = models.MSSQLBackup{
			Enabled:     true,
			NetbackupID: plan.ID,
			NetbackupName: []models.MSSQLNetbackupName{{
				Name:        plan.Attributes.Name,
				Description: plan.Attributes.Description,
			}},
		}
	}

	allowedIPs, hasSecurityGroup, sgDiags := mssqlAllowedIPs(ctx, data.SecurityGroup)
	diags.Append(sgDiags...)
	if diags.HasError() {
		return nil, "", diags
	}
	if hasSecurityGroup {
		createReq.SecurityGroup = &models.MSSQLSecurityGroup{AllowedIPs: allowedIPs}
	}

	return createReq, version, diags
}

func applyMSSQLToState(ctx context.Context, data *MSSQLResourceModel, instance *models.MSSQLInstance) diag.Diagnostics {
	var diags diag.Diagnostics

	if instance.UUID != "" {
		data.ID = types.StringValue(instance.UUID)
	}
	data.Status = types.StringValue(instance.Status)
	if instance.ConnectionString != nil && *instance.ConnectionString != "" {
		data.ConnectionString = types.StringValue(*instance.ConnectionString)
	}
	if instance.CreatedAt != "" {
		data.CreatedAt = types.StringValue(instance.CreatedAt)
	}
	if instance.Port != 0 {
		data.Port = types.Int64Value(int64(instance.Port))
	}
	if instance.DBConfig.Version != "" {
		data.Version = types.StringValue(instance.DBConfig.Version)
	}
	if instance.Name != "" {
		data.Name = types.StringValue(instance.Name)
	}
	if instance.DBName != "" {
		data.DBName = types.StringValue(instance.DBName)
	}
	if instance.MSSQLUsername != "" {
		data.MSSQLUsername = types.StringValue(instance.MSSQLUsername)
	}
	if instance.Description != "" {
		data.Description = types.StringValue(instance.Description)
	}
	if instance.AZName != "" {
		data.AvailabilityZone = types.StringValue(instance.AZName)
	}
	if instance.DBConfig.Edition != "" {
		data.DatabaseEdition = types.StringValue(instance.DBConfig.Edition)
	}
	if instance.CollationFeatures.Collation != "" {
		data.Collation = types.StringValue(instance.CollationFeatures.Collation)
	}
	if len(instance.CollationFeatures.Features) > 0 {
		features, featureDiags := types.ListValueFrom(ctx, types.StringType, instance.CollationFeatures.Features)
		diags.Append(featureDiags...)
		data.Features = features
	}
	if instance.ComputeStorageConfig.Flavor.Name != "" {
		data.ComputeSize = types.StringValue(instance.ComputeStorageConfig.Flavor.Name)
	}
	if instance.VolumeSize != 0 {
		data.StorageSize = types.Int64Value(int64(instance.VolumeSize))
	}
	data.EnableTDE = types.BoolValue(instance.EnableTDE)
	data.AdvancedSettingsEnabled = types.BoolValue(instance.AdvancedSettings.Enabled)
	if len(instance.AddLabels) > 0 {
		labels, labelDiags := types.ListValueFrom(ctx, types.StringType, instance.AddLabels)
		diags.Append(labelDiags...)
		data.Labels = labels
	}
	if !data.SecurityGroup.IsNull() && !data.SecurityGroup.IsUnknown() && instance.AllowedIPs != nil {
		ips, ipDiags := types.ListValueFrom(ctx, types.StringType, instance.AllowedIPs)
		diags.Append(ipDiags...)
		if !ipDiags.HasError() {
			obj, objDiags := types.ObjectValue(mssqlSecurityGroupAttrTypes(), map[string]attr.Value{
				"allowed_ips": ips,
			})
			diags.Append(objDiags...)
			data.SecurityGroup = obj
		}
	}

	return diags
}

func mssqlProtectionPlan(ctx context.Context, obj types.Object) (string, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() {
		return "", false, diags
	}
	if obj.IsUnknown() {
		diags.AddError("Invalid Configuration", "backup is unknown.")
		return "", false, diags
	}

	var model MSSQLBackupModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return "", false, diags
	}
	if model.ProtectionPlan.IsNull() || model.ProtectionPlan.IsUnknown() || model.ProtectionPlan.ValueString() == "" {
		return "", false, diags
	}
	return model.ProtectionPlan.ValueString(), true, diags
}

func mssqlAllowedIPs(ctx context.Context, obj types.Object) ([]string, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() {
		return nil, false, diags
	}
	if obj.IsUnknown() {
		diags.AddError("Invalid Configuration", "security_group is unknown.")
		return nil, false, diags
	}

	var model MSSQLSecurityGroupModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, false, diags
	}
	ips, ipDiags := stringSliceFromList(ctx, model.AllowedIPs)
	diags.Append(ipDiags...)
	return ips, true, diags
}

func mssqlSecurityGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_ips": types.ListType{ElemType: types.StringType},
	}
}
