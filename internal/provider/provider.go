package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Airtel-Cloud-Platform/terraform-provider-airtelcloud/internal/client"
)

// Ensure AirtelCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &AirtelCloudProvider{}

// AirtelCloudProvider defines the provider implementation.
type AirtelCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AirtelCloudProviderModel describes the provider data model.
type AirtelCloudProviderModel struct {
	APIEndpoint  types.String `tfsdk:"api_endpoint"`
	APIKey       types.String `tfsdk:"api_key"`
	APISecret    types.String `tfsdk:"api_secret"`
	Region       types.String `tfsdk:"region"`
	Organization types.String `tfsdk:"organization"`
	ProjectName  types.String `tfsdk:"project_name"`
	SubnetID     types.String `tfsdk:"subnet_id"`
}

func (p *AirtelCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "airtelcloud"
	resp.Version = p.version
}

func (p *AirtelCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "Airtel Cloud API endpoint URL",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Airtel Cloud API key for authentication. Can also be set via the `AIRTEL_API_KEY` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"api_secret": schema.StringAttribute{
				MarkdownDescription: "Airtel Cloud API secret for HMAC authentication. Can also be set via the `AIRTEL_API_SECRET` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Airtel Cloud region",
				Optional:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization name or domain for Airtel Cloud",
				Optional:            true,
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name for Airtel Cloud API calls",
				Optional:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "Default subnet ID used for volume API provider lookup",
				Optional:            true,
			},
		},
	}
}

func (p *AirtelCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AirtelCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// Each value falls back to an environment variable if not set in config.
	apiEndpoint := stringValueOrEnv(data.APIEndpoint, "AIRTEL_API_ENDPOINT", "https://api.south.cloud.airtel.in")
	region := stringValueOrEnv(data.Region, "AIRTEL_REGION", "south")
	organization := stringValueOrEnv(data.Organization, "AIRTEL_ORGANIZATION", "")
	projectName := stringValueOrEnv(data.ProjectName, "AIRTEL_PROJECT_NAME", "")
	subnetID := stringValueOrEnv(data.SubnetID, "AIRTEL_SUBNET_ID", "")

	apiKey := stringValueOrEnv(data.APIKey, "AIRTEL_API_KEY", "")
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Airtel Cloud API Key",
			"The provider requires an API key. Set it in the provider configuration or via the AIRTEL_API_KEY environment variable.",
		)
	}

	apiSecret := stringValueOrEnv(data.APISecret, "AIRTEL_API_SECRET", "")
	if apiSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Airtel Cloud API Secret",
			"The provider requires an API secret. Set it in the provider configuration or via the AIRTEL_API_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	c, err := client.NewClient(apiEndpoint, apiKey, apiSecret, region, organization, projectName, subnetID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Airtel Cloud API Client",
			"An unexpected error occurred when creating the Airtel Cloud API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Airtel Cloud Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c

	tflog.Info(ctx, "Configured Airtel Cloud client", map[string]any{
		"api_endpoint": apiEndpoint,
		"region":       region,
	})
}

func (p *AirtelCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
		NewVolumeResource,
		NewVPCResource,
		NewSubnetResource,
		NewObjectStorageBucketResource,
		NewObjectStorageAccessKeyResource,
		NewFileStorageResource,
		NewFileStorageExportPathResource,
		NewDNSZoneResource,
		NewDNSRecordResource,
		NewSecurityGroupResource,
		NewSecurityGroupRuleResource,
		NewVPCPeeringResource,
		NewLBServiceResource,
		NewLBVipResource,
		NewLBCertificateResource,
		NewLBVirtualServerResource,
		NewComputeSnapshotResource,
		NewProtectionResource,
		NewProtectionPlanResource,
		NewPublicIPResource,
		NewPublicIPPolicyRuleResource,
	}
}

func (p *AirtelCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be implemented later
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AirtelCloudProvider{
			version: version,
		}
	}
}

// stringValueOrEnv returns the Terraform config value if set, otherwise falls back
// to the named environment variable, and finally to the provided default.
func stringValueOrEnv(val types.String, envVar, defaultVal string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return defaultVal
}
