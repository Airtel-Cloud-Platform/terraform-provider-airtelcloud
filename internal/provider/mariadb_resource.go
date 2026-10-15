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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

const mariadbDefaultTimeout = 15 * time.Minute

var (
	mariadbClusterNamePattern  = regexp.MustCompile(`^[a-z0-9-]+$`)
	mariadbIdentifierPattern   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	mariadbScheduleTimePattern = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
)

var _ resource.Resource = &MariaDBResource{}
var _ resource.ResourceWithImportState = &MariaDBResource{}
var _ resource.ResourceWithValidateConfig = &MariaDBResource{}

func NewMariaDBResource() resource.Resource {
	return &MariaDBResource{}
}

// MariaDBResource defines the airtelcloud_mariadb resource.
type MariaDBResource struct {
	client *client.Client
}

// MariaDBResourceModel is the Terraform state model. It is separate from API structs.
type MariaDBResourceModel struct {
	ID               types.String               `tfsdk:"id"`
	ClusterName      types.String               `tfsdk:"cluster_name"`
	Description      types.String               `tfsdk:"description"`
	Version          types.String               `tfsdk:"version"`
	HighAvailability types.Bool                 `tfsdk:"high_availability"`
	NumReplicas      types.Int64                `tfsdk:"num_replicas"`
	DatabaseName     types.String               `tfsdk:"database_name"`
	MariaDBUsername  types.String               `tfsdk:"mariadb_username"`
	Password         types.String               `tfsdk:"password"`
	IsSuperuser      types.Bool                 `tfsdk:"is_superuser"`
	NetworkType      types.String               `tfsdk:"network_type"`
	ComputeSize      types.String               `tfsdk:"compute_size"`
	StorageSize      types.Int64                `tfsdk:"storage_size"`
	StorageType      types.String               `tfsdk:"storage_type"`
	AvailabilityZone types.String               `tfsdk:"availability_zone"`
	Labels           types.List                 `tfsdk:"labels"`
	Topology         types.String               `tfsdk:"topology"`
	ConnectionString types.String               `tfsdk:"connection_string"`
	Port             types.Int64                `tfsdk:"port"`
	CreatedAt        types.String               `tfsdk:"created_at"`
	Users            types.String   `tfsdk:"users"`
	Status           types.String   `tfsdk:"status"`
	Backup           types.Object   `tfsdk:"backup"`
	SecurityGroup    types.Object   `tfsdk:"security_group"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

// MariaDBBackupModel is the nested backup block.
type MariaDBBackupModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	ProtectionPlan   types.String `tfsdk:"protection_plan"`
	CompressionLevel types.Int64  `tfsdk:"compression_level"`
	Retention        types.Int64  `tfsdk:"retention"`
	ScheduleTime     types.String `tfsdk:"schedule_time"`
	ScheduleDay      types.String `tfsdk:"schedule_day"`
}

// MariaDBSecurityGroupModel is the nested security_group block.
type MariaDBSecurityGroupModel struct {
	AllowedIPs types.List `tfsdk:"allowed_ips"`
}

func (r *MariaDBResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mariadb"
}

func (r *MariaDBResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud MariaDB cluster. v1 supports create, read, import, and delete. Changes to configuration force a new cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier (UUID) of the MariaDB cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the cluster. The API may suffix a UUID onto the stored name; Terraform keeps this configured value.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(mariadbClusterNamePattern, "cluster_name must contain only lowercase letters, digits, and hyphens."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the cluster. When omitted or empty, the API is sent an empty string.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "MariaDB engine version (for example `10.6`, `10.11`, `11.4`, or `11.8`). Validated against the versions catalog.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"high_availability": schema.BoolAttribute{
				MarkdownDescription: "When true, the cluster is created with master-slave topology. When false, master. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"num_replicas": schema.Int64Attribute{
				MarkdownDescription: "Number of slave replicas. Required and must be between 1 and 10 when high_availability is true. Defaults to 0.",
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
					stringvalidator.RegexMatches(mariadbIdentifierPattern, "database_name must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mariadb_username": schema.StringAttribute{
				MarkdownDescription: "Admin username for the cluster.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(mariadbIdentifierPattern, "mariadb_username must start with a letter and contain only letters, digits, and underscore."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Admin password. The API does not return this value; it is stored only in Terraform state. Must be 8-128 characters; include uppercase, lowercase, digit, and special character; must not contain '@' or '_'; must not contain words from mariadb_username or cluster_name.",
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
				MarkdownDescription: "Flavor name from the mariadb flavors catalog (for example `db.maria.uhper.ccs.xlarge`). Resolved to flavor ID, vCPUs, and RAM internally.",
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
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Labels to assign to the cluster.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"topology": schema.StringAttribute{
				MarkdownDescription: "Resolved topology (`master` or `master-slave`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_string": schema.StringAttribute{
				MarkdownDescription: "MariaDB connection string. Null until the cluster is Active. The password is masked by the API.",
				Computed:            true,
				Sensitive:           true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Database port (typically 3350).",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the cluster was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"users": schema.StringAttribute{
				MarkdownDescription: "Admin username returned by the API (`mariadb_username`).",
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
						MarkdownDescription: "Backup compression level. Defaults to 1.",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(1),
					},
					"retention": schema.Int64Attribute{
						MarkdownDescription: "Backup retention in days.",
						Optional:            true,
						Computed:            true,
					},
					"schedule_time": schema.StringAttribute{
						MarkdownDescription: "Backup schedule time in HH:MM format. Required when enabled is true.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(mariadbScheduleTimePattern, "must be in HH:MM format"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"schedule_day": schema.StringAttribute{
						MarkdownDescription: "Backup schedule day. Required when enabled is true.",
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
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"allowed_ips": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "CIDR blocks allowed to connect (for example `192.168.1.0/24`). At least one CIDR is required.",
						Required:            true,
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

func (r *MariaDBResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data MariaDBResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if msg := validateMariaDBReplicas(data.HighAvailability, data.NumReplicas); msg != "" {
		resp.Diagnostics.AddError("Invalid Configuration", msg)
	}

	if msg := validateMariaDBClusterName(data.ClusterName); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("cluster_name"), "Invalid cluster_name", msg)
	}

	if msg := validateMariaDBIdentifier(data.DatabaseName, "database_name"); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("database_name"), "Invalid database_name", msg)
	}

	if msg := validateMariaDBIdentifier(data.MariaDBUsername, "mariadb_username"); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("mariadb_username"), "Invalid mariadb_username", msg)
	}

	if msg := validateMariaDBPassword(data.Password, data.MariaDBUsername, data.ClusterName); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Invalid password", msg)
	}

	if !data.Backup.IsNull() && !data.Backup.IsUnknown() {
		backup, backupDiags := expandMariaDBBackup(ctx, data.Backup)
		resp.Diagnostics.Append(backupDiags...)
		if backup != nil {
			if msg := validateMariaDBBackupScheduleTimeRequired(backup); msg != "" {
				resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_time"), "Invalid schedule_time", msg)
			}
			if msg := validateMariaDBBackupScheduleDayRequired(backup); msg != "" {
				resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_day"), "Invalid schedule_day", msg)
			}
			if msg := validateMariaDBScheduleTime(backup.ScheduleTime); msg != "" {
				resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_time"), "Invalid schedule_time", msg)
			}
			if msg := validateMariaDBScheduleDay(backup.ScheduleDay); msg != "" {
				resp.Diagnostics.AddAttributeError(path.Root("backup").AtName("schedule_day"), "Invalid schedule_day", msg)
			}
		}
	}

	// Skip when unknown (common under modules during ValidateConfig).
	if data.SecurityGroup.IsUnknown() {
		return
	}
	if data.SecurityGroup.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("security_group"), "Invalid security_group", "security_group is required.")
		return
	}

	group, groupDiags := expandMariaDBSecurityGroup(ctx, data.SecurityGroup)
	resp.Diagnostics.Append(groupDiags...)
	if group != nil {
		if msg := validateMariaDBSecurityGroup(group); msg != "" {
			resp.Diagnostics.AddAttributeError(path.Root("security_group"), "Invalid security_group", msg)
		}
	}
}

func validateMariaDBSecurityGroup(group *MariaDBSecurityGroupModel) string {
	if group == nil {
		return "security_group is required."
	}
	if group.AllowedIPs.IsUnknown() {
		return ""
	}
	if group.AllowedIPs.IsNull() {
		return "security_group.allowed_ips is required."
	}
	var ips []string
	if diags := group.AllowedIPs.ElementsAs(context.Background(), &ips, false); diags.HasError() {
		return "security_group.allowed_ips is invalid."
	}
	if len(ips) == 0 {
		return "security_group.allowed_ips must contain at least one CIDR."
	}
	for _, ip := range ips {
		if strings.TrimSpace(ip) == "" {
			return "security_group.allowed_ips cannot contain empty values."
		}
	}
	return ""
}

func validateMariaDBReplicas(ha types.Bool, replicas types.Int64) string {
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

func validateMariaDBClusterName(cluster types.String) string {
	if cluster.IsNull() || cluster.IsUnknown() {
		return ""
	}
	if !mariadbClusterNamePattern.MatchString(cluster.ValueString()) {
		return "cluster_name must contain only lowercase letters, digits, and hyphens."
	}
	return ""
}

func validateMariaDBIdentifier(value types.String, fieldName string) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	if !mariadbIdentifierPattern.MatchString(value.ValueString()) {
		return fmt.Sprintf("%s must start with a letter and contain only letters, digits, and underscore.", fieldName)
	}
	return ""
}

func validateMariaDBPassword(password, username, clusterName types.String) string {
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
		return "password cannot contain words from mariadb_username or cluster_name."
	}
	return ""
}

func validateMariaDBScheduleTime(scheduleTime types.String) string {
	if scheduleTime.IsNull() || scheduleTime.IsUnknown() {
		return ""
	}
	if !mariadbScheduleTimePattern.MatchString(scheduleTime.ValueString()) {
		return "schedule_time must be in HH:MM format."
	}
	return ""
}

func validateMariaDBBackupScheduleTimeRequired(backup *MariaDBBackupModel) string {
	if backup == nil || backup.Enabled.IsNull() || backup.Enabled.IsUnknown() || !backup.Enabled.ValueBool() {
		return ""
	}
	if backup.ScheduleTime.IsNull() || backup.ScheduleTime.IsUnknown() || strings.TrimSpace(backup.ScheduleTime.ValueString()) == "" {
		return "schedule_time is required when backup.enabled is true."
	}
	return ""
}

func validateMariaDBBackupScheduleDayRequired(backup *MariaDBBackupModel) string {
	if backup == nil || backup.Enabled.IsNull() || backup.Enabled.IsUnknown() || !backup.Enabled.ValueBool() {
		return ""
	}
	if backup.ScheduleDay.IsNull() || backup.ScheduleDay.IsUnknown() || strings.TrimSpace(backup.ScheduleDay.ValueString()) == "" {
		return "schedule_day is required when backup.enabled is true."
	}
	return ""
}

func validateMariaDBScheduleDay(scheduleDay types.String) string {
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

func (r *MariaDBResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MariaDBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MariaDBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, mariadbDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := r.buildMariaDBCreateRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.CreateMariaDBCluster(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create mariadb cluster, got error: %s", err))
		return
	}

	ready, err := r.client.WaitForMariaDBReady(ctx, cluster.UUID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for mariadb cluster to be ready: %s", err))
		return
	}

	resp.Diagnostics.Append(applyMariaDBClusterToState(ctx, &data, ready, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created mariadb resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MariaDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MariaDBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetMariaDBCluster(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read mariadb cluster, got error: %s", err))
		return
	}

	if cluster.Status == models.MariaDBStatusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(applyMariaDBClusterToState(ctx, &data, cluster, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MariaDBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"airtelcloud_mariadb does not support in-place updates. Changing configuration forces a new cluster.",
	)
}

func (r *MariaDBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MariaDBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, mariadbDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMariaDBCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete mariadb cluster, got error: %s", err))
		return
	}

	if err := r.client.WaitForMariaDBDeleted(ctx, data.ID.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for mariadb cluster deletion: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted mariadb resource")
}

func (r *MariaDBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, "imported mariadb resource")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MariaDBResource) buildMariaDBCreateRequest(ctx context.Context, data *MariaDBResourceModel) (*models.CreateMariaDBClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	az := data.AvailabilityZone.ValueString()

	zone, err := r.client.ResolveMariaDBAvailabilityZone(ctx, az)
	if err != nil {
		diags.AddError("Availability Zone Resolution Error", err.Error())
		return nil, diags
	}

	version, err := r.client.ResolveMariaDBVersion(ctx, data.Version.ValueString())
	if err != nil {
		diags.AddError("Version Resolution Error", err.Error())
		return nil, diags
	}

	flavor, err := r.client.ResolveMariaDBFlavor(ctx, zone.AZCode, data.ComputeSize.ValueString())
	if err != nil {
		diags.AddError("Flavor Resolution Error", err.Error())
		return nil, diags
	}

	storage, err := r.client.ResolveMariaDBStorage(ctx, zone.AZCode, data.StorageType.ValueString())
	if err != nil {
		diags.AddError("Storage Resolution Error", err.Error())
		return nil, diags
	}

	ha := data.HighAvailability.ValueBool()
	topology, err := r.client.ResolveMariaDBTopology(ctx, ha)
	if err != nil {
		diags.AddError("Topology Resolution Error", err.Error())
		return nil, diags
	}

	labels, labelDiags := stringSliceFromList(ctx, data.Labels)
	diags.Append(labelDiags...)
	if diags.HasError() {
		return nil, diags
	}

	numReplicas := 0
	if ha {
		numReplicas = int(data.NumReplicas.ValueInt64())
	}

	createReq := &models.CreateMariaDBClusterRequest{
		Name:            data.ClusterName.ValueString(),
		Description:     data.Description.ValueString(),
		Version:         version.Version,
		Topology:        topology,
		NumReplicas:     numReplicas,
		DatabaseName:    data.DatabaseName.ValueString(),
		MariaDBUsername: data.MariaDBUsername.ValueString(),
		Password:        data.Password.ValueString(),
		IsSuperuser:     data.IsSuperuser.ValueBool(),
		NetworkType:     data.NetworkType.ValueString(),
		ComputeStorageConfig: models.MariaDBComputeStorageConfig{
			Flavor: models.MariaDBFlavorRef{
				Name:  flavor.Name,
				VCPUs: flavor.VCPUs,
				RAM:   flavor.RAM,
			},
			FlavorIDs: []models.MariaDBFlavorAZID{{
				AZID: zone.AZCode,
				ID:   flavor.ID,
			}},
			StorageSizeGB:    int(data.StorageSize.ValueInt64()),
			StorageType:      storage.Name,
			StorageName:      storage.Label,
			DataVolumeTypeID: storage.ID,
		},
		AZIDs:  []string{zone.AZCode},
		Labels: labels,
	}

	if !data.Backup.IsNull() && !data.Backup.IsUnknown() {
		backupModel, backupDiags := expandMariaDBBackup(ctx, data.Backup)
		diags.Append(backupDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if backupModel != nil {
			backup := &models.MariaDBBackupConfig{
				Enabled: backupModel.Enabled.ValueBool(),
			}
			if !backupModel.CompressionLevel.IsNull() && !backupModel.CompressionLevel.IsUnknown() {
				backup.CompressionLevel = int(backupModel.CompressionLevel.ValueInt64())
			}
			if !backupModel.Retention.IsNull() && !backupModel.Retention.IsUnknown() {
				backup.Retention = int(backupModel.Retention.ValueInt64())
			}
			if !backupModel.ScheduleTime.IsNull() && !backupModel.ScheduleTime.IsUnknown() {
				s := strings.TrimSpace(backupModel.ScheduleTime.ValueString())
				if s != "" {
					backup.ScheduleTime = s
				}
			}
			if !backupModel.ScheduleDay.IsNull() && !backupModel.ScheduleDay.IsUnknown() {
				s := strings.TrimSpace(backupModel.ScheduleDay.ValueString())
				if s != "" {
					backup.ScheduleDay = s
				}
			}
			if !backupModel.ProtectionPlan.IsNull() && backupModel.ProtectionPlan.ValueString() != "" {
				plan, err := r.client.ResolveMariaDBProtectionPlan(ctx, backupModel.ProtectionPlan.ValueString())
				if err != nil {
					diags.AddError("Protection Plan Resolution Error", err.Error())
					return nil, diags
				}
				backup.ProtectionPlan = plan.Value
			}
			createReq.Backup = backup
		}
	}

	if !data.SecurityGroup.IsNull() && !data.SecurityGroup.IsUnknown() {
		group, groupDiags := expandMariaDBSecurityGroup(ctx, data.SecurityGroup)
		diags.Append(groupDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if group != nil {
			allowedIPs, ipDiags := stringSliceFromList(ctx, group.AllowedIPs)
			diags.Append(ipDiags...)
			if diags.HasError() {
				return nil, diags
			}
			createReq.SecurityGroup = &models.MariaDBSecurityGroup{AllowedIPs: allowedIPs}
		}
	}

	return createReq, diags
}

func applyMariaDBClusterToState(ctx context.Context, data *MariaDBResourceModel, cluster *models.MariaDBCluster, refreshConfig bool) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(cluster.UUID)
	data.Status = types.StringValue(cluster.Status)
	data.Topology = types.StringValue(cluster.Topology)
	data.HighAvailability = types.BoolValue(cluster.Topology == models.MariaDBTopologyMasterSlave)
	if cluster.CreatedAt != "" {
		data.CreatedAt = types.StringValue(cluster.CreatedAt)
	}
	if cluster.ConnectionString != nil && *cluster.ConnectionString != "" {
		data.ConnectionString = types.StringValue(*cluster.ConnectionString)
	}
	if cluster.Port != 0 {
		data.Port = types.Int64Value(int64(cluster.Port))
	}
	if cluster.MariaDBUsername != "" {
		data.Users = types.StringValue(cluster.MariaDBUsername)
	}

	if !data.Backup.IsNull() && !data.Backup.IsUnknown() && cluster.Backup != nil {
		backupModel, backupDiags := expandMariaDBBackup(ctx, data.Backup)
		diags.Append(backupDiags...)
		if backupModel != nil {
			if cluster.Backup.ScheduleTime != "" {
				backupModel.ScheduleTime = types.StringValue(cluster.Backup.ScheduleTime)
			}
			if cluster.Backup.ScheduleDay != "" {
				backupModel.ScheduleDay = types.StringValue(cluster.Backup.ScheduleDay)
			}
			obj, objDiags := flattenMariaDBBackup(ctx, backupModel)
			diags.Append(objDiags...)
			data.Backup = obj
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
	if cluster.MariaDBUsername != "" {
		data.MariaDBUsername = types.StringValue(cluster.MariaDBUsername)
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

	if len(cluster.Labels) > 0 {
		list, d := types.ListValueFrom(ctx, types.StringType, cluster.Labels)
		diags.Append(d...)
		data.Labels = list
	}

	if cluster.Backup != nil && (!data.Backup.IsNull() || cluster.Backup.Enabled) {
		backupModel := &MariaDBBackupModel{}
		if !data.Backup.IsNull() && !data.Backup.IsUnknown() {
			existing, existingDiags := expandMariaDBBackup(ctx, data.Backup)
			diags.Append(existingDiags...)
			if existing != nil {
				backupModel = existing
			}
		}
		backupModel.Enabled = types.BoolValue(cluster.Backup.Enabled)
		if cluster.Backup.ProtectionPlan != "" {
			backupModel.ProtectionPlan = types.StringValue(cluster.Backup.ProtectionPlan)
		}
		if cluster.Backup.CompressionLevel != 0 {
			backupModel.CompressionLevel = types.Int64Value(int64(cluster.Backup.CompressionLevel))
		}
		if cluster.Backup.Retention != 0 {
			backupModel.Retention = types.Int64Value(int64(cluster.Backup.Retention))
		}
		if cluster.Backup.ScheduleTime != "" {
			backupModel.ScheduleTime = types.StringValue(cluster.Backup.ScheduleTime)
		}
		if cluster.Backup.ScheduleDay != "" {
			backupModel.ScheduleDay = types.StringValue(cluster.Backup.ScheduleDay)
		}
		obj, objDiags := flattenMariaDBBackup(ctx, backupModel)
		diags.Append(objDiags...)
		data.Backup = obj
	}

	return diags
}

func expandMariaDBBackup(ctx context.Context, obj types.Object) (*MariaDBBackupModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var model MariaDBBackupModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	return &model, diags
}

func expandMariaDBSecurityGroup(ctx context.Context, obj types.Object) (*MariaDBSecurityGroupModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var model MariaDBSecurityGroupModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	return &model, diags
}

func flattenMariaDBBackup(ctx context.Context, model *MariaDBBackupModel) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if model == nil {
		return types.ObjectNull(mariadbBackupAttrTypes()), diags
	}
	obj, objDiags := types.ObjectValueFrom(ctx, mariadbBackupAttrTypes(), model)
	diags.Append(objDiags...)
	return obj, diags
}

func mariadbBackupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":           types.BoolType,
		"protection_plan":   types.StringType,
		"compression_level": types.Int64Type,
		"retention":         types.Int64Type,
		"schedule_time":     types.StringType,
		"schedule_day":      types.StringType,
	}
}

func mariadbSecurityGroupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_ips": types.ListType{ElemType: types.StringType},
	}
}
