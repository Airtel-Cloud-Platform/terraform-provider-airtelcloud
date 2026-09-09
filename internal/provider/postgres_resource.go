package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const postgresDefaultTimeout = 30 * time.Minute

var (
	postgresClusterNamePattern  = regexp.MustCompile(`^[a-z0-9-]+$`)
	postgresIdentifierPattern   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	postgresScheduleTimePattern = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
)

var _ resource.Resource = &PostgresResource{}
var _ resource.ResourceWithImportState = &PostgresResource{}
var _ resource.ResourceWithValidateConfig = &PostgresResource{}

func NewPostgresResource() resource.Resource {
	return &PostgresResource{}
}

// PostgresResource defines the airtelcloud_postgres resource.
type PostgresResource struct {
	client *client.Client
}

// PostgresResourceModel is the Terraform state model. It is separate from API structs.
type PostgresResourceModel struct {
	ID               types.String                `tfsdk:"id"`
	ClusterName      types.String                `tfsdk:"cluster_name"`
	Description      types.String                `tfsdk:"description"`
	Version          types.String                `tfsdk:"version"`
	HighAvailability types.Bool                  `tfsdk:"high_availability"`
	NumReplicas      types.Int64                 `tfsdk:"num_replicas"`
	DatabaseName     types.String                `tfsdk:"database_name"`
	PostgresUsername types.String                `tfsdk:"postgres_username"`
	Password         types.String                `tfsdk:"password"`
	IsSuperuser      types.Bool                  `tfsdk:"is_superuser"`
	NetworkType      types.String                `tfsdk:"network_type"`
	ComputeSize      types.String                `tfsdk:"compute_size"`
	StorageSize      types.Int64                 `tfsdk:"storage_size"`
	StorageType      types.String                `tfsdk:"storage_type"`
	AvailabilityZone types.String                `tfsdk:"availability_zone"`
	PGExtensions     types.List                  `tfsdk:"pg_extensions"`
	Labels           types.List                  `tfsdk:"labels"`
	Topology         types.String                `tfsdk:"topology"`
	ConnectionString types.String                `tfsdk:"connection_string"`
	CreatedAt        types.String                `tfsdk:"created_at"`
	Status           types.String                `tfsdk:"status"`
	Backup           *PostgresBackupModel        `tfsdk:"backup"`
	SecurityGroup    *PostgresSecurityGroupModel `tfsdk:"security_group"`
	Timeouts         timeouts.Value              `tfsdk:"timeouts"`
}

// PostgresBackupModel is the nested backup block.
type PostgresBackupModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	ProtectionPlan   types.String `tfsdk:"protection_plan"`
	CompressionLevel types.Int64  `tfsdk:"compression_level"`
	Retention        types.Int64  `tfsdk:"retention"`
	ScheduleTime     types.String `tfsdk:"schedule_time"`
	ScheduleDay      types.String `tfsdk:"schedule_day"`
}

// PostgresSecurityGroupModel is the nested security_group block.
type PostgresSecurityGroupModel struct {
	AllowedIPs types.List `tfsdk:"allowed_ips"`
}

func (r *PostgresResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

func (r *PostgresResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud PostgreSQL cluster. v1 supports create, read, import, and delete. Changes to configuration force a new cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the PostgreSQL cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the cluster. The API may suffix a UUID onto the stored name; Terraform keeps this configured value.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(postgresClusterNamePattern, "cluster_name must contain only lowercase letters, digits, and hyphens."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "PostgreSQL major version (for example `17` or `18`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"high_availability": schema.BoolAttribute{
				MarkdownDescription: "When true, the cluster is created with primary-standby topology. When false, standalone. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"num_replicas": schema.Int64Attribute{
				MarkdownDescription: "Number of standby replicas. Required and must be between 1 and 10 when high_availability is true. Defaults to 0.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"database_name": schema.StringAttribute{
				MarkdownDescription: "Name of the initial database.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(postgresIdentifierPattern, "database_name must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"postgres_username": schema.StringAttribute{
				MarkdownDescription: "Admin username for the cluster.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(postgresIdentifierPattern, "postgres_username must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Admin password. The API does not return this value; it is stored only in Terraform state.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_superuser": schema.BoolAttribute{
				MarkdownDescription: "Whether the admin user is a superuser. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"network_type": schema.StringAttribute{
				MarkdownDescription: "Network type for the cluster. Defaults to `private`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("private"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_size": schema.StringAttribute{
				MarkdownDescription: "Flavor name from the postgres flavors catalog (for example `db.postgres.uhper.ccs.xlarge`). Resolved to flavor ID and RAM internally.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_size": schema.Int64Attribute{
				MarkdownDescription: "Data volume size in GB. Must be at least 200.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(200),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"storage_type": schema.StringAttribute{
				MarkdownDescription: "Storage class label from the volume-types catalog. Defaults to `High Performance`. Resolved to backend type, name, and volume type ID internally.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("High Performance"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "Availability zone code (for example `S1`). Must match `azCode` from the zones API.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pg_extensions": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "PostgreSQL extensions to enable (for example `pgvector`).",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Labels to assign to the cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"topology": schema.StringAttribute{
				MarkdownDescription: "Resolved topology (`standalone` or `primary-standby`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_string": schema.StringAttribute{
				MarkdownDescription: "PostgreSQL connection string. Null until the cluster is Active. The password is masked by the API.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the cluster was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current cluster status (for example `Creating`, `Active`, `Failed`).",
				Computed:            true,
			},
			"backup": schema.SingleNestedAttribute{
				MarkdownDescription: "Backup configuration. Changes force a new cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether backup is enabled. Defaults to false.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"protection_plan": schema.StringAttribute{
						MarkdownDescription: "Protection plan value from the protection-plans catalog (for example `weekly-full-daily-incr`).",
						Optional:            true,
					},
					"compression_level": schema.Int64Attribute{
						MarkdownDescription: "Backup compression level. Defaults to 6.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(6),
					},
					"retention": schema.Int64Attribute{
						MarkdownDescription: "Backup retention in days. Defaults to 15.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(15),
					},
					"schedule_time": schema.StringAttribute{
						MarkdownDescription: "Backup schedule time as returned by the API.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(postgresScheduleTimePattern, "must be in HH:MM format"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"schedule_day": schema.StringAttribute{
						MarkdownDescription: "Backup schedule day as returned by the API.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"security_group": schema.SingleNestedAttribute{
				MarkdownDescription: "Allowed client CIDRs. Changes force a new cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"allowed_ips": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "CIDR blocks allowed to connect (for example `192.168.1.0/24`).",
						Optional:            true,
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

func (r *PostgresResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data PostgresResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if msg := validatePostgresReplicas(data.HighAvailability, data.NumReplicas); msg != "" {
		resp.Diagnostics.AddError("Invalid Configuration", msg)
	}

	if msg := validatePostgresClusterName(data.ClusterName); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("cluster_name"), "Invalid cluster_name", msg)
	}

	if msg := validatePostgresIdentifier(data.DatabaseName, "database_name"); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("database_name"), "Invalid database_name", msg)
	}

	if msg := validatePostgresIdentifier(data.PostgresUsername, "postgres_username"); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("postgres_username"), "Invalid postgres_username", msg)
	}

	if msg := validatePostgresPassword(data.Password, data.PostgresUsername, data.ClusterName); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Invalid password", msg)
	}

	if data.Backup != nil {
		if msg := validatePostgresBackupScheduleTimeRequired(data.Backup); msg != "" {
			resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_time"), "Invalid schedule_time", msg)
		}
		if msg := validatePostgresBackupScheduleDayRequired(data.Backup); msg != "" {
			resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_day"), "Invalid schedule_day", msg)
		}

		if msg := validatePostgresScheduleTime(data.Backup.ScheduleTime); msg != "" {
			resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_time"), "Invalid schedule_time", msg)
		}
		if msg := validatePostgresScheduleDay(data.Backup.ScheduleDay); msg != "" {
			resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_day"), "Invalid schedule_day", msg)
		}
	}
}

func validatePostgresReplicas(ha types.Bool, replicas types.Int64) string {
	if ha.IsUnknown() {
		return ""
	}

	highAvailability := !ha.IsNull() && ha.ValueBool()
	if !highAvailability {
		return ""
	}

	if replicas.IsUnknown() {
		return ""
	}

	n := int64(0)
	if !replicas.IsNull() {
		n = replicas.ValueInt64()
	}
	if n < 1 || n > 10 {
		return "num_replicas is required when high_availability is true and must be between 1 and 10."
	}
	return ""
}

func validatePostgresClusterName(cluster types.String) string {
	if cluster.IsNull() || cluster.IsUnknown() {
		return ""
	}
	if !postgresClusterNamePattern.MatchString(cluster.ValueString()) {
		return "cluster_name must contain only lowercase letters, digits, and hyphens."
	}
	return ""
}

func validatePostgresIdentifier(value types.String, fieldName string) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	if !postgresIdentifierPattern.MatchString(value.ValueString()) {
		return fmt.Sprintf("%s must start with a letter and contain only letters, digits, and underscore.", fieldName)
	}
	return ""
}

func validatePostgresPassword(password, username, clusterName types.String) string {
	if password.IsNull() || password.IsUnknown() {
		return ""
	}

	p := password.ValueString()
	length := utf8.RuneCountInString(p)
	if length < 8 || length > 128 {
		return "password must be between 8 and 128 characters."
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			if r == '@' || r == '_' {
				return "password cannot contain '@' or '_'."
			}
			hasSpecial = true
		}
	}

	if !hasUpper {
		return "password must contain at least 1 uppercase letter."
	}
	if !hasLower {
		return "password must contain at least 1 lowercase letter."
	}
	if !hasDigit {
		return "password must contain at least 1 digit."
	}
	if !hasSpecial {
		return "password must contain at least 1 special character."
	}
	lowerPassword := strings.ToLower(p)
	if containsIdentifierWord(lowerPassword, username) || containsIdentifierWord(lowerPassword, clusterName) {
		return "password cannot contain words from postgres_username or cluster_name."
	}
	return ""
}

func validatePostgresScheduleTime(scheduleTime types.String) string {
	if scheduleTime.IsNull() || scheduleTime.IsUnknown() {
		return ""
	}
	if !postgresScheduleTimePattern.MatchString(scheduleTime.ValueString()) {
		return "schedule_time must be in HH:MM format."
	}
	return ""
}

func validatePostgresBackupScheduleTimeRequired(backup *PostgresBackupModel) string {
	if backup == nil || backup.Enabled.IsNull() || backup.Enabled.IsUnknown() || !backup.Enabled.ValueBool() {
		return ""
	}
	if backup.ScheduleTime.IsNull() || backup.ScheduleTime.IsUnknown() || strings.TrimSpace(backup.ScheduleTime.ValueString()) == "" {
		return "schedule_time is required when backup.enabled is true."
	}
	return ""
}

func validatePostgresBackupScheduleDayRequired(backup *PostgresBackupModel) string {
	if backup == nil || backup.Enabled.IsNull() || backup.Enabled.IsUnknown() || !backup.Enabled.ValueBool() {
		return ""
	}
	if backup.ScheduleDay.IsNull() || backup.ScheduleDay.IsUnknown() || strings.TrimSpace(backup.ScheduleDay.ValueString()) == "" {
		return "schedule_day is required when backup.enabled is true."
	}
	return ""
}

func validatePostgresScheduleDay(scheduleDay types.String) string {
	if scheduleDay.IsNull() || scheduleDay.IsUnknown() {
		return ""
	}
	validDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range validDays {
		if scheduleDay.ValueString() == day {
			return ""
		}
	}
	return "schedule_day must be one of [\"Monday\", \"Tuesday\", \"Wednesday\", \"Thursday\", \"Friday\", \"Saturday\", \"Sunday\"]."
}

func containsIdentifierWord(passwordLower string, v types.String) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}

	identifier := strings.ToLower(v.ValueString())
	if len(identifier) >= 3 && strings.Contains(passwordLower, identifier) {
		return true
	}

	parts := strings.FieldsFunc(identifier, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, part := range parts {
		if len(part) < 3 {
			continue
		}
		if strings.Contains(passwordLower, part) {
			return true
		}
	}

	return false
}

func (r *PostgresResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PostgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PostgresResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, postgresDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := r.buildPostgresCreateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.CreatePostgresCluster(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create postgres cluster, got error: %s", err))
		return
	}

	ready, err := r.client.WaitForPostgresReady(ctx, cluster.UUID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for postgres cluster to be ready: %s", err))
		return
	}

	resp.Diagnostics.Append(applyPostgresClusterToState(ctx, &data, ready, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created postgres resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PostgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PostgresResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetPostgresCluster(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read postgres cluster, got error: %s", err))
		return
	}

	if cluster.Status == models.PostgresStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(applyPostgresClusterToState(ctx, &data, cluster, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PostgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"airtelcloud_postgres does not support in-place updates. Changing configuration forces a new cluster.",
	)
}

func (r *PostgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PostgresResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, postgresDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePostgresCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete postgres cluster, got error: %s", err))
		return
	}

	if err := r.client.WaitForPostgresDeleted(ctx, data.ID.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for postgres cluster deletion: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted postgres resource")
}

func (r *PostgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "imported postgres resource")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PostgresResource) buildPostgresCreateRequest(ctx context.Context, data *PostgresResourceModel) (*models.CreatePostgresClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	az := data.AvailabilityZone.ValueString()

	zone, err := r.client.ResolvePostgresAvailabilityZone(ctx, az)
	if err != nil {
		diags.AddError("Availability Zone Resolution Error", err.Error())
		return nil, diags
	}

	flavor, err := r.client.ResolvePostgresFlavor(ctx, zone.AZCode, data.ComputeSize.ValueString())
	if err != nil {
		diags.AddError("Flavor Resolution Error", err.Error())
		return nil, diags
	}

	storage, err := r.client.ResolvePostgresStorage(ctx, zone.AZCode, data.StorageType.ValueString())
	if err != nil {
		diags.AddError("Storage Resolution Error", err.Error())
		return nil, diags
	}

	ha := data.HighAvailability.ValueBool()
	topology, err := r.client.ResolvePostgresTopology(ctx, ha)
	if err != nil {
		diags.AddError("Topology Resolution Error", err.Error())
		return nil, diags
	}

	extensions, extDiags := stringSliceFromList(ctx, data.PGExtensions)
	diags.Append(extDiags...)
	labels, labelDiags := stringSliceFromList(ctx, data.Labels)
	diags.Append(labelDiags...)
	if diags.HasError() {
		return nil, diags
	}

	createReq := &models.CreatePostgresClusterRequest{
		Name:             data.ClusterName.ValueString(),
		Description:      data.Description.ValueString(),
		Version:          data.Version.ValueString(),
		Topology:         topology,
		DatabaseName:     data.DatabaseName.ValueString(),
		PostgresUsername: data.PostgresUsername.ValueString(),
		Password:         data.Password.ValueString(),
		IsSuperuser:      data.IsSuperuser.ValueBool(),
		NetworkType:      data.NetworkType.ValueString(),
		ComputeStorageConfig: models.PostgresComputeStorageConfig{
			Flavor: models.PostgresFlavorRef{
				Name: flavor.Name,
				RAM:  flavor.RAM,
			},
			FlavorIDs: []models.PostgresFlavorAZID{{
				AZID: zone.AZCode,
				ID:   flavor.ID,
			}},
			StorageSizeGB:    int(data.StorageSize.ValueInt64()),
			StorageType:      storage.Name,
			StorageName:      storage.Label,
			DataVolumeTypeID: storage.ID,
		},
		AZIDs:        []string{zone.AZCode},
		PGExtensions: extensions,
		Labels:       labels,
	}

	if ha {
		createReq.NumReplicas = int(data.NumReplicas.ValueInt64())
	}

	if data.Backup != nil {
		backup := &models.PostgresBackupConfig{
			Enabled:          data.Backup.Enabled.ValueBool(),
			CompressionLevel: int(data.Backup.CompressionLevel.ValueInt64()),
			Retention:        int(data.Backup.Retention.ValueInt64()),
		}
		if !data.Backup.ScheduleTime.IsNull() && !data.Backup.ScheduleTime.IsUnknown() {
			s := strings.TrimSpace(data.Backup.ScheduleTime.ValueString())
			if s != "" {
				backup.ScheduleTime = s
			}
		}
		if !data.Backup.ScheduleDay.IsNull() && !data.Backup.ScheduleDay.IsUnknown() {
			s := strings.TrimSpace(data.Backup.ScheduleDay.ValueString())
			if s != "" {
				backup.ScheduleDay = s
			}
		}
		if !data.Backup.ProtectionPlan.IsNull() && data.Backup.ProtectionPlan.ValueString() != "" {
			plan, err := r.client.ResolvePostgresProtectionPlan(ctx, data.Backup.ProtectionPlan.ValueString())
			if err != nil {
				diags.AddError("Protection Plan Resolution Error", err.Error())
				return nil, diags
			}
			backup.ProtectionPlan = plan.Value
		}
		createReq.Backup = backup
	}

	if data.SecurityGroup != nil {
		allowedIPs, ipDiags := stringSliceFromList(ctx, data.SecurityGroup.AllowedIPs)
		diags.Append(ipDiags...)
		if diags.HasError() {
			return nil, diags
		}
		createReq.SecurityGroup = &models.PostgresSecurityGroup{AllowedIPs: allowedIPs}
	}

	return createReq, diags
}

func applyPostgresClusterToState(ctx context.Context, data *PostgresResourceModel, cluster *models.PostgresCluster, refreshConfig bool) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(cluster.UUID)
	data.Status = types.StringValue(cluster.Status)
	data.Topology = types.StringValue(cluster.Topology)
	data.HighAvailability = types.BoolValue(cluster.Topology == models.PostgresTopologyPrimaryStandby)
	if cluster.CreatedAt != "" {
		data.CreatedAt = types.StringValue(cluster.CreatedAt)
	}
	if cluster.ConnectionString != nil && *cluster.ConnectionString != "" {
		data.ConnectionString = types.StringValue(*cluster.ConnectionString)
	}

	if data.Backup != nil && cluster.Backup != nil {
		if cluster.Backup.ScheduleTime != "" {
			data.Backup.ScheduleTime = types.StringValue(cluster.Backup.ScheduleTime)
		}
		if cluster.Backup.ScheduleDay != "" {
			data.Backup.ScheduleDay = types.StringValue(cluster.Backup.ScheduleDay)
		}
	}

	if !refreshConfig {
		return diags
	}

	if data.ClusterName.IsNull() || data.ClusterName.IsUnknown() {
		name := cluster.LevelName
		if name == "" {
			name = cluster.Name
		}
		if name != "" {
			data.ClusterName = types.StringValue(name)
		}
	}
	if cluster.Version != "" {
		data.Version = types.StringValue(cluster.Version)
	}
	if cluster.Description != "" {
		data.Description = types.StringValue(cluster.Description)
	}
	if cluster.DatabaseName != "" {
		data.DatabaseName = types.StringValue(cluster.DatabaseName)
	}
	if cluster.PostgresUsername != "" {
		data.PostgresUsername = types.StringValue(cluster.PostgresUsername)
	}
	data.IsSuperuser = types.BoolValue(cluster.IsSuperuser)
	data.NumReplicas = types.Int64Value(int64(cluster.NumReplicas))
	if cluster.NetworkConfig != nil && cluster.NetworkConfig.NetworkType != "" {
		data.NetworkType = types.StringValue(cluster.NetworkConfig.NetworkType)
	}
	if cluster.ComputeStorageConfig.Flavor.Name != "" && (data.ComputeSize.IsNull() || data.ComputeSize.IsUnknown()) {
		data.ComputeSize = types.StringValue(cluster.ComputeStorageConfig.Flavor.Name)
	}
	if cluster.ComputeStorageConfig.StorageSizeGB > 0 {
		data.StorageSize = types.Int64Value(int64(cluster.ComputeStorageConfig.StorageSizeGB))
	}
	if cluster.ComputeStorageConfig.StorageName != "" && (data.StorageType.IsNull() || data.StorageType.IsUnknown()) {
		data.StorageType = types.StringValue(cluster.ComputeStorageConfig.StorageName)
	}
	if (data.AvailabilityZone.IsNull() || data.AvailabilityZone.IsUnknown()) && len(cluster.AzNames) > 0 {
		data.AvailabilityZone = types.StringValue(cluster.AzNames[0])
	}

	if len(cluster.PGExtensions) > 0 {
		list, d := types.ListValueFrom(ctx, types.StringType, cluster.PGExtensions)
		diags.Append(d...)
		data.PGExtensions = list
	}
	if len(cluster.Labels) > 0 {
		list, d := types.ListValueFrom(ctx, types.StringType, cluster.Labels)
		diags.Append(d...)
		data.Labels = list
	}

	if cluster.Backup != nil && (data.Backup != nil || cluster.Backup.Enabled) {
		if data.Backup == nil {
			data.Backup = &PostgresBackupModel{}
		}
		data.Backup.Enabled = types.BoolValue(cluster.Backup.Enabled)
		if cluster.Backup.ProtectionPlan != "" {
			data.Backup.ProtectionPlan = types.StringValue(cluster.Backup.ProtectionPlan)
		}
		if cluster.Backup.CompressionLevel != 0 {
			data.Backup.CompressionLevel = types.Int64Value(int64(cluster.Backup.CompressionLevel))
		}
		if cluster.Backup.Retention != 0 {
			data.Backup.Retention = types.Int64Value(int64(cluster.Backup.Retention))
		}
		if cluster.Backup.ScheduleTime != "" {
			data.Backup.ScheduleTime = types.StringValue(cluster.Backup.ScheduleTime)
		}
		if cluster.Backup.ScheduleDay != "" {
			data.Backup.ScheduleDay = types.StringValue(cluster.Backup.ScheduleDay)
		}
	}

	return diags
}

func stringSliceFromList(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	return values, diags
}
