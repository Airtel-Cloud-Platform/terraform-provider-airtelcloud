package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}
var _ resource.ResourceWithValidateConfig = &VMResource{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

// VMResource defines the resource implementation.
type VMResource struct {
	client *client.Client
}

// VMResourceModel describes the resource data model.
type VMResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	ProviderInstanceID types.String   `tfsdk:"provider_instance_id"`
	InstanceName       types.String   `tfsdk:"instance_name"`
	FlavorID           types.String   `tfsdk:"flavor_id"`
	FlavorName         types.String   `tfsdk:"flavor_name"`
	ImageID            types.String   `tfsdk:"image_id"`
	ImageName          types.String   `tfsdk:"image_name"`
	SnapshotName       types.String   `tfsdk:"snapshot_name"`
	VPCID              types.String   `tfsdk:"vpc_id"`
	VPCName            types.String   `tfsdk:"vpc_name"`
	SubnetID           types.String   `tfsdk:"subnet_id"`
	SubnetName         types.String   `tfsdk:"subnet_name"`
	SecurityGroupIDs   types.List     `tfsdk:"security_group_ids"`
	SecurityGroupNames types.List     `tfsdk:"security_group_names"`
	KeypairID          types.String   `tfsdk:"keypair_id"`
	KeypairName        types.String   `tfsdk:"keypair_name"`
	AdminUsername      types.String   `tfsdk:"admin_username"`
	AdminPassword      types.String   `tfsdk:"admin_password"`
	PublicIP           types.String   `tfsdk:"public_ip"`
	PrivateIP          types.String   `tfsdk:"private_ip"`
	Status             types.String   `tfsdk:"status"`
	UserData           types.String   `tfsdk:"user_data"`
	AvailabilityZone   types.String   `tfsdk:"availability_zone"`
	OSType             types.String   `tfsdk:"os_type"`
	Region             types.String   `tfsdk:"region"`
	DiskSize           types.Int64    `tfsdk:"disk_size"`
	BootFromVolume     types.Bool     `tfsdk:"boot_from_volume"`
	VolumeTypeID       types.String   `tfsdk:"volume_type_id"`
	Description        types.String   `tfsdk:"description"`
	EnableBackup       types.Bool     `tfsdk:"enable_backup"`
	ProtectionPlan     types.String   `tfsdk:"protection_plan"`
	StartDate          types.String   `tfsdk:"start_date"`
	Weekday            types.String   `tfsdk:"weekday"`
	StartTime          types.String   `tfsdk:"start_time"`
	VMCount            types.Int64    `tfsdk:"vm_count"`
	Labels             types.List     `tfsdk:"labels"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Airtel Cloud compute instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the compute instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_instance_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The provider-specific instance ID.",
			},
			"instance_name": schema.StringAttribute{
				MarkdownDescription: "The name of the compute instance.",
				Required:            true,
			},
			"flavor_id": schema.StringAttribute{
				MarkdownDescription: "The flavor ID for the compute instance. Either flavor_id or flavor_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flavor_name": schema.StringAttribute{
				MarkdownDescription: "The flavor name for the compute instance. Either flavor_id or flavor_name must be specified.",
				Optional:            true,
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the image to use for the instance. Exactly one of image_id, image_name, or snapshot_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_name": schema.StringAttribute{
				MarkdownDescription: "The name of the image to use for the instance. Exactly one of image_id, image_name, or snapshot_name must be specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_name": schema.StringAttribute{
				MarkdownDescription: "The name of the compute snapshot to use. The provider resolves this through the snapshot list API and uses the snapshot image_id. Exactly one of image_id, image_name, or snapshot_name must be specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the VPC. Either vpc_id or vpc_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPC. Either vpc_id or vpc_name must be specified.",
				Optional:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subnet. Either subnet_id or subnet_name must be specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_name": schema.StringAttribute{
				MarkdownDescription: "The name of the subnet. Either subnet_id or subnet_name must be specified.",
				Optional:            true,
			},
			"security_group_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group IDs. Mutually exclusive with security_group_names.",
				Optional:            true,
			},
			"security_group_names": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of security group names. Each name is resolved to an ID. Mutually exclusive with security_group_ids.",
				Optional:            true,
			},
			"keypair_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the key pair for SSH access. Mutually exclusive with keypair_name.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"keypair_name": schema.StringAttribute{
				MarkdownDescription: "The name of the key pair for SSH access. Mutually exclusive with keypair_id.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_username": schema.StringAttribute{
				MarkdownDescription: "Login username to create on the instance. Supported for os_type values: `linux`, `ubuntu`, `rhel`, `suse`, `centos`. " +
					"Must be set together with admin_password. Mutually exclusive with keypair_id and keypair_name.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_password": schema.StringAttribute{
				MarkdownDescription: "Login password for admin_username. Supported for os_type values: `linux`, `ubuntu`, `rhel`, `suse`, `centos`. " +
					"Must be set together with admin_username. Mutually exclusive with keypair_id and keypair_name. " +
					"Minimum 8 characters, must contain at least one uppercase letter, one lowercase letter, and one special character. " +
					"Stored in plaintext in Terraform state.",
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "The public IP address of the instance.",
				Computed:            true,
			},
			"private_ip": schema.StringAttribute{
				MarkdownDescription: "The private IP address of the instance.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the instance.",
				Computed:            true,
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "User data / cloud-init script to run on instance initialization.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "The availability zone where the instance is placed.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
				"os_type": schema.StringAttribute{
				MarkdownDescription: "The OS type of the instance. Accepted values: `linux`, `ubuntu`, `rhel`, `suse`, `centos`, `windows`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region for the instance. Defaults to the provider region.",
				Optional:            true,
				Computed:            true,
			},
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: "The disk size in GB.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(20),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"boot_from_volume": schema.BoolAttribute{
				MarkdownDescription: "Whether to boot from volume.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"volume_type_id": schema.StringAttribute{
				MarkdownDescription: "The volume type ID.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the compute instance.",
				Optional:            true,
			},
			"enable_backup": schema.BoolAttribute{
				MarkdownDescription: "Whether backup is enabled for the instance.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"protection_plan": schema.StringAttribute{
				MarkdownDescription: "Protection plan UUID or name for the instance. The provider accepts either the UUID (e.g. `4cb5b1b6-f62f-4fea-b348-d17aa407d64d`) or the plan name (e.g. `daily-backup`). Names are resolved to UUIDs at apply time.",
				Optional:            true,
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date for backup scheduling. Mutually exclusive with weekday. Accepts YYYY-MM-DD and converts to the API's MM/DD/YYYY format.",
				Optional:            true,
			},
			"weekday": schema.StringAttribute{
				MarkdownDescription: "Optional weekday convenience input for VM backup scheduling (`monday`..`sunday` or `mon`..`sun`). Mutually exclusive with start_date. Converted internally to the next matching start_date in IST (`Asia/Kolkata`).",
				Optional:            true,
			},
			"start_time": schema.StringAttribute{
				MarkdownDescription: "The start time for backup scheduling. Accepts 24-hour HH:MM and converts to the API's 12-hour AM/PM format.",
				Optional:            true,
			},
			"vm_count": schema.Int64Attribute{
				MarkdownDescription: "Number of VM instances to create. Must be between 1 and 10.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"labels": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of labels to assign to the instance. Up to 5 labels are supported, and each label must be between 3 and 15 characters long.",
				Optional:            true,
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

func (r *VMResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data VMResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate flavor: exactly one of flavor_id or flavor_name must be set
	if !data.FlavorID.IsNull() && !data.FlavorName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of flavor_id or flavor_name may be specified, not both.")
	}
	if data.FlavorID.IsNull() && data.FlavorName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of flavor_id or flavor_name must be specified.")
	}

	imageIDSet := !data.ImageID.IsNull() && (data.ImageID.IsUnknown() || strings.TrimSpace(data.ImageID.ValueString()) != "")
	imageNameSet := !data.ImageName.IsNull() && (data.ImageName.IsUnknown() || strings.TrimSpace(data.ImageName.ValueString()) != "")
	snapshotNameSet := !data.SnapshotName.IsNull() && (data.SnapshotName.IsUnknown() || strings.TrimSpace(data.SnapshotName.ValueString()) != "")

	// Validate image source: exactly one of image_id, image_name, or snapshot_name must be set
	configuredImageSources := 0
	if imageIDSet {
		configuredImageSources++
	}
	if imageNameSet {
		configuredImageSources++
	}
	if snapshotNameSet {
		configuredImageSources++
	}

	if configuredImageSources > 1 {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of image_id, image_name, or snapshot_name may be specified.")
	}
	if configuredImageSources == 0 {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of image_id, image_name, or snapshot_name must be specified.")
	}

	// Validate VPC: exactly one of vpc_id or vpc_name must be set
	if !data.VPCID.IsNull() && !data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of vpc_id or vpc_name may be specified, not both.")
	}
	if data.VPCID.IsNull() && data.VPCName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of vpc_id or vpc_name must be specified.")
	}

	// Validate subnet: exactly one of subnet_id or subnet_name must be set
	if !data.SubnetID.IsNull() && !data.SubnetName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of subnet_id or subnet_name may be specified, not both.")
	}
	if data.SubnetID.IsNull() && data.SubnetName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"One of subnet_id or subnet_name must be specified.")
	}

	// Validate security group list input: only one input style allowed.
	if !data.SecurityGroupIDs.IsNull() && !data.SecurityGroupNames.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of security_group_ids or security_group_names may be specified, not both.")
	}

	if !data.SecurityGroupIDs.IsNull() && !data.SecurityGroupIDs.IsUnknown() {
		var securityGroupIDs []string
		resp.Diagnostics.Append(data.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, sgID := range securityGroupIDs {
			if strings.TrimSpace(sgID) == "" {
				resp.Diagnostics.AddError("Invalid Configuration",
					"security_group_ids must not contain empty values.")
				break
			}
		}
	}

	if !data.SecurityGroupNames.IsNull() && !data.SecurityGroupNames.IsUnknown() {
		var securityGroupNames []string
		resp.Diagnostics.Append(data.SecurityGroupNames.ElementsAs(ctx, &securityGroupNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, sgName := range securityGroupNames {
			if strings.TrimSpace(sgName) == "" {
				resp.Diagnostics.AddError("Invalid Configuration",
					"security_group_names must not contain empty values.")
				break
			}
		}
	}

	// Validate keypair: mutual exclusion only (both are optional)
	if !data.KeypairID.IsNull() && !data.KeypairName.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of keypair_id or keypair_name may be specified, not both.")
	}

	// Validate backup schedule inputs: start_date and weekday are mutually exclusive.
	hasStartDate := !data.StartDate.IsNull() && strings.TrimSpace(data.StartDate.ValueString()) != ""
	hasWeekday := !data.Weekday.IsNull() && strings.TrimSpace(data.Weekday.ValueString()) != ""
	if hasStartDate && hasWeekday {
		resp.Diagnostics.AddError("Invalid Configuration",
			"Only one of start_date or weekday may be specified, not both.")
	}

	// Validate VM credentials (admin_username / admin_password).
	//
	// Presence is tested with IsNull() alone: a value sourced from a variable or
	// another resource is Unknown rather than Null, so IsNull() still answers
	// "is this set in config?". Reading a value is guarded by IsUnknown().
	usernameSet := !data.AdminUsername.IsNull()
	passwordSet := !data.AdminPassword.IsNull()
	keypairSet := !data.KeypairID.IsNull() || !data.KeypairName.IsNull()
	credentialsSet := usernameSet || passwordSet

	if usernameSet != passwordSet {
		resp.Diagnostics.AddError("Invalid Configuration",
			"admin_username and admin_password must be specified together.")
	}

	if credentialsSet {
		if keypairSet {
			resp.Diagnostics.AddError("Invalid Configuration",
				"admin_username/admin_password may not be combined with keypair_id or keypair_name. "+
					"Use either password credentials or an SSH keypair, not both.")
		}

		if !data.OSType.IsUnknown() && !isPasswordSupportedOS(data.OSType.ValueString()) {
			resp.Diagnostics.AddError("Invalid Configuration",
				"admin_username/admin_password are only supported for os_type: linux, ubuntu, rhel, suse, centos.")
		}

		// The form encoder drops empty strings, so an explicit "" would silently
		// send a half-populated request.
		if usernameSet && !data.AdminUsername.IsUnknown() && data.AdminUsername.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid Configuration",
				"admin_username may not be an empty string.")
		}
		if passwordSet && !data.AdminPassword.IsUnknown() && data.AdminPassword.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid Configuration",
				"admin_password may not be an empty string.")
		}
		if passwordSet && !data.AdminPassword.IsUnknown() && data.AdminPassword.ValueString() != "" {
			if err := validateVMPassword(data.AdminPassword.ValueString()); err != nil {
				resp.Diagnostics.AddError("Invalid Configuration", err.Error())
			}
		}
	}

	// non-Windows instances need exactly one authentication method.
	if !data.OSType.IsUnknown() && isPasswordSupportedOS(data.OSType.ValueString()) &&
		!credentialsSet && !keypairSet {
		resp.Diagnostics.AddError("Invalid Configuration",
			fmt.Sprintf("A %s instance requires an authentication method: set either keypair_id/keypair_name "+
				"or admin_username and admin_password.", data.OSType.ValueString()))
	}

	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		var labels []string
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if err := validateVMLabelValues(labels); err != nil {
			resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		}
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func resolveVMBackupStartDateInput(startDate, weekday string, now time.Time) (string, error) {
	if strings.TrimSpace(startDate) != "" {
		return startDate, nil
	}
	return resolveWeekdayStartDate(weekday, now)
}

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var labels []string
	if !data.Labels.IsNull() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create a scoped client with availability-zone and subnet-id headers.
	// The API requires the subnet-id header to resolve the correct provider
	// for compute endpoints (flavors, images, keypairs, etc.).
	computeClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
	if !data.SubnetID.IsNull() && data.SubnetID.ValueString() != "" {
		computeClient = computeClient.WithSubnetID(data.SubnetID.ValueString())
	}

	// Resolve VPC first (needed for subnet resolution if using subnet_name)
	vpcID := data.VPCID.ValueString()
	if vpcID == "" && !data.VPCName.IsNull() {
		resolved, err := computeClient.ResolveVPCID(ctx, data.VPCName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("VPC Resolution Error", err.Error())
			return
		}
		vpcID = resolved
	}

	// Resolve subnet (needs resolved VPC ID)
	subnetID := data.SubnetID.ValueString()
	if subnetID == "" && !data.SubnetName.IsNull() {
		resolved, err := computeClient.ResolveSubnetID(ctx, vpcID, data.SubnetName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Subnet Resolution Error", err.Error())
			return
		}
		subnetID = resolved
		// Now set the resolved subnet-id header for subsequent API calls
		computeClient = computeClient.WithSubnetID(subnetID)
	}

	// Resolve flavor (requires subnet-id header)
	flavorID := data.FlavorID.ValueString()
	if flavorID == "" && !data.FlavorName.IsNull() {
		resolved, err := computeClient.ResolveFlavorID(ctx, data.FlavorName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Flavor Resolution Error", err.Error())
			return
		}
		flavorID = resolved
	}

	// Resolve image (requires subnet-id header)
	imageID := data.ImageID.ValueString()
	if imageID == "" && !data.ImageName.IsNull() {
		resolved, err := computeClient.ResolveImageID(ctx, data.ImageName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Image Resolution Error", err.Error())
			return
		}
		imageID = resolved
	} else if imageID == "" && !data.SnapshotName.IsNull() {
		resolved, err := computeClient.ResolveSnapshotImageID(ctx, data.SnapshotName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Snapshot Resolution Error", err.Error())
			return
		}
		imageID = resolved
	}

	// Resolve security groups
	var secGroupIDs []string
	if !data.SecurityGroupNames.IsNull() {
		var securityGroupNames []string
		resp.Diagnostics.Append(data.SecurityGroupNames.ElementsAs(ctx, &securityGroupNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, sgName := range securityGroupNames {
			resolved, err := computeClient.ResolveSecurityGroupID(ctx, strings.TrimSpace(sgName))
			if err != nil {
				resp.Diagnostics.AddError("Security Group Resolution Error", err.Error())
				return
			}
			secGroupIDs = append(secGroupIDs, strconv.Itoa(resolved))
		}
	} else if !data.SecurityGroupIDs.IsNull() {
		resp.Diagnostics.Append(data.SecurityGroupIDs.ElementsAs(ctx, &secGroupIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for i, sgID := range secGroupIDs {
			secGroupIDs[i] = strings.TrimSpace(sgID)
		}
	}

	// Resolve keypair
	keypairID := data.KeypairID.ValueString()
	if keypairID == "" && !data.KeypairName.IsNull() {
		resolved, err := computeClient.ResolveKeypairID(ctx, data.KeypairName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Keypair Resolution Error", err.Error())
			return
		}
		keypairID = resolved
	}

	startDateInput, err := resolveVMBackupStartDateInput(data.StartDate.ValueString(), data.Weekday.ValueString(), time.Now())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startDate, err := formatProtectionStartDate(startDateInput)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}
	startTime, err := formatProtectionStartTime(data.StartTime.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// VM credentials. Mutual exclusion with the keypair is enforced in ValidateConfig.
	// The API's confirm field is populated from admin_password: double entry guards
	// against typos in an interactive form, but here the config is the source of truth.
	adminPassword := data.AdminPassword.ValueString()

	// Use provider region as default if not set
	region := data.Region.ValueString()
	if region == "" {
		region = r.client.Region
	}

	protectionPlan := strings.TrimSpace(data.ProtectionPlan.ValueString())
	if protectionPlan != "" {
		resolved, err := computeClient.ResolveProtectionPlanID(ctx, protectionPlan, subnetID)
		if err != nil {
			resp.Diagnostics.AddError("Protection Plan Resolution Error", err.Error())
			return
		}
		protectionPlan = resolved
	}

	// Description defaults
	description := data.Description.ValueString()
	if description == "" {
		description = "Created by Terraform"
	}

	// Convert volume_type_id to int if provided
	var volumeTypeID int
	if !data.VolumeTypeID.IsNull() && data.VolumeTypeID.ValueString() != "" {
		if vtID, err := strconv.Atoi(data.VolumeTypeID.ValueString()); err == nil {
			volumeTypeID = vtID
		}
	}

	createReq := &models.CreateComputeRequest{
		InstanceName:         data.InstanceName.ValueString(),
		Description:          description,
		ImageID:              imageID,
		FlavorID:             flavorID,
		VPCID:                vpcID,
		SubnetID:             subnetID,
		NetworkID:            subnetID,
		SecurityGroupIDs:     secGroupIDs,
		KeypairID:            keypairID,
		VMUsername:           data.AdminUsername.ValueString(),
		VMPassword:           adminPassword,
		VMConfirmPassword:    adminPassword,
		UserCloudInitScripts: data.UserData.ValueString(),
		AZName:               data.AvailabilityZone.ValueString(),
		Region:               region,
		OSType:               data.OSType.ValueString(),
		VolumeSize:           int(data.DiskSize.ValueInt64()),
		BootFromVolume:       data.BootFromVolume.ValueBool(),
		VolumeTypeID:         volumeTypeID,
		EnableBackup:         data.EnableBackup.ValueBool(),
		ProtectionPlan:       protectionPlan,
		StartDate:            startDate,
		StartTime:            startTime,
		VMCount:              int(data.VMCount.ValueInt64()),
	}

	compute, err := computeClient.CreateCompute(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create compute instance, got error: %s", err))
		return
	}

	// Wait for compute instance to become active and report private IP.
	readyCompute, err := r.client.WaitForComputeReadyWithPrivateIP(ctx, compute.ID, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error waiting for compute instance to be ready: %s", err))
		return
	}

	if len(labels) > 0 {
		patchedCompute, err := computeClient.PatchComputeLabels(ctx, readyCompute.ID, labels)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply labels to compute instance, got error: %s", err))
			return
		}
		readyCompute = patchedCompute
	}

	// Update the model with the ready compute data
	data.ID = types.StringValue(readyCompute.ID)
	data.ProviderInstanceID = types.StringValue(readyCompute.ProviderInstanceID)
	data.FlavorID = types.StringValue(flavorID)
	data.ImageID = types.StringValue(imageID)
	data.VPCID = types.StringValue(vpcID)
	data.SubnetID = types.StringValue(subnetID)
	data.Status = types.StringValue(readyCompute.Status)
	data.PublicIP = types.StringValue(models.FlexString(readyCompute.PublicIPs))
	data.PrivateIP = types.StringValue(readyCompute.PrivateIP())
	if readyCompute.AZName != "" {
		data.AvailabilityZone = types.StringValue(readyCompute.AZName)
	} else if readyCompute.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(readyCompute.AvailabilityZone)
	}
	if region != "" {
		data.Region = types.StringValue(region)
	}

	tflog.Trace(ctx, "created a compute instance resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	compute, err := r.client.GetCompute(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read compute instance, got error: %s", err))
		return
	}

	// If the API returned a deleted status, remove from state
	if compute.Status == "Deleted" || compute.Status == "deleted" || compute.Status == "soft-deleted" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with the current compute data
	data.ProviderInstanceID = types.StringValue(compute.ProviderInstanceID)
	data.InstanceName = types.StringValue(compute.InstanceName)
	data.FlavorID = types.StringValue(models.FlexString(compute.FlavorID))
	data.ImageID = types.StringValue(models.FlexString(compute.ImageID))
	if compute.VPCID != "" {
		data.VPCID = types.StringValue(compute.VPCID)
	}
	if compute.SubnetID != "" {
		data.SubnetID = types.StringValue(compute.SubnetID)
	}
	// security_group_ids/security_group_names are configuration-owned. The compute API
	// does not return a full SG list in a shape that can be mapped back without drift,
	// so these fields are left unchanged during read.
	// The compute API only ever reports the keypair by name; it returns no keypair ID.
	// So keypair_id is left untouched here — there is nothing to refresh it from, and
	// both keypair attributes are Optional (not Computed) with RequiresReplace, meaning
	// any value written back that the practitioner didn't configure reads as drift and
	// forces the VM to be replaced. Refresh keypair_name only when the keypair was
	// configured by name; if it was configured by ID, keypair_name must stay null.
	if compute.KeypairName != "" && data.KeypairID.IsNull() {
		data.KeypairName = types.StringValue(compute.KeypairName)
	}
	data.Status = types.StringValue(compute.Status)
	data.PublicIP = types.StringValue(models.FlexString(compute.PublicIPs))
	data.PrivateIP = types.StringValue(compute.PrivateIP())
	if compute.UserData != "" {
		data.UserData = types.StringValue(compute.UserData)
	}
	if compute.AZName != "" {
		data.AvailabilityZone = types.StringValue(compute.AZName)
	} else if compute.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(compute.AvailabilityZone)
	}
	if compute.Region != "" {
		data.Region = types.StringValue(compute.Region)
	}
	if compute.OSType != "" {
		data.OSType = types.StringValue(compute.OSType)
	}
	if compute.VolumeSize != 0 {
		data.DiskSize = types.Int64Value(int64(compute.VolumeSize))
	}
	data.BootFromVolume = types.BoolValue(compute.BootFromVolume)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMResourceModel
	var state VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a scoped client with AZ and subnet-id headers for resolution calls
	computeClient := r.client.WithAvailabilityZone(data.AvailabilityZone.ValueString())
	if !data.SubnetID.IsNull() && data.SubnetID.ValueString() != "" {
		computeClient = computeClient.WithSubnetID(data.SubnetID.ValueString())
	}

	// Handle flavor resize if flavor_id changed
	newFlavorID := data.FlavorID.ValueString()
	if !data.FlavorName.IsNull() && data.FlavorName.ValueString() != "" && newFlavorID == "" {
		resolved, err := computeClient.ResolveFlavorID(ctx, data.FlavorName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Flavor Resolution Error", err.Error())
			return
		}
		newFlavorID = resolved
	}

	oldFlavorID := state.FlavorID.ValueString()
	if newFlavorID != "" && newFlavorID != oldFlavorID {
		tflog.Info(ctx, "Resizing compute instance", map[string]interface{}{
			"id":            data.ID.ValueString(),
			"old_flavor_id": oldFlavorID,
			"new_flavor_id": newFlavorID,
		})

		err := computeClient.ResizeCompute(ctx, data.ID.ValueString(), newFlavorID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to resize compute instance, got error: %s", err))
			return
		}

		// Poll until the compute instance returns to ACTIVE status
		deadline := time.Now().Add(10 * time.Minute)
		for time.Now().Before(deadline) {
			compute, err := r.client.GetCompute(ctx, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error polling compute status after resize: %s", err))
				return
			}
			if compute.Status == "ACTIVE" || compute.Status == "active" {
				data.FlavorID = types.StringValue(models.FlexString(compute.FlavorID))
				break
			}
			if compute.Status == "ERROR" || compute.Status == "error" {
				resp.Diagnostics.AddError("Resize Error", "Compute instance entered error state during resize")
				return
			}
			time.Sleep(10 * time.Second)
		}
	}

	var labels []string
	if !data.Labels.IsNull() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Resolve security groups for update.
	var secGroupIDs []string
	if !data.SecurityGroupNames.IsNull() {
		var securityGroupNames []string
		resp.Diagnostics.Append(data.SecurityGroupNames.ElementsAs(ctx, &securityGroupNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, sgName := range securityGroupNames {
			resolved, err := computeClient.ResolveSecurityGroupID(ctx, strings.TrimSpace(sgName))
			if err != nil {
				resp.Diagnostics.AddError("Security Group Resolution Error", err.Error())
				return
			}
			secGroupIDs = append(secGroupIDs, strconv.Itoa(resolved))
		}
	} else if !data.SecurityGroupIDs.IsNull() {
		resp.Diagnostics.Append(data.SecurityGroupIDs.ElementsAs(ctx, &secGroupIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for i, sgID := range secGroupIDs {
			secGroupIDs[i] = strings.TrimSpace(sgID)
		}
	}

	description := data.Description.ValueString()
	if description == "" {
		description = "Updated by Terraform"
	}

	updateReq := &models.UpdateComputeRequest{
		InstanceName:     data.InstanceName.ValueString(),
		Description:      description,
		SecurityGroupIDs: secGroupIDs,
	}

	compute, err := r.client.UpdateCompute(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update compute instance, got error: %s", err))
		return
	}

	if !data.Labels.IsNull() {
		patchedCompute, err := computeClient.PatchComputeLabels(ctx, data.ID.ValueString(), labels)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply labels to compute instance, got error: %s", err))
			return
		}
		compute = patchedCompute
	}

	// Update the model with the updated compute data
	data.Status = types.StringValue(compute.Status)
	data.PublicIP = types.StringValue(models.FlexString(compute.PublicIPs))
	data.PrivateIP = types.StringValue(compute.PrivateIP())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCompute(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete compute instance, got error: %s", err))
		return
	}
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// isPasswordSupportedOS reports whether os_type supports admin_username/admin_password auth.
func isPasswordSupportedOS(osType string) bool {
	switch strings.ToLower(osType) {
	case "linux", "ubuntu", "rhel", "suse", "centos":
		return true
	}
	return false
}

// validateVMPassword enforces: min 8 chars, ≥1 uppercase, ≥1 lowercase, ≥1 special character.
func validateVMPassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("admin_password must be at least 8 characters long")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("admin_password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("admin_password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		return fmt.Errorf("admin_password must contain at least one special character")
	}
	return nil
}

func validateVMLabelValues(labels []string) error {
	if len(labels) > 5 {
		return fmt.Errorf("labels supports at most 5 labels, got %d", len(labels))
	}

	for _, label := range labels {
		trimmedLabel := strings.TrimSpace(label)
		if len(trimmedLabel) < 3 {
			return fmt.Errorf("each label must be at least 3 characters long, got %q", label)
		}
		if len(trimmedLabel) > 15 {
			return fmt.Errorf("each label must be at most 15 characters long, got %q", label)
		}
	}

	return nil
}
