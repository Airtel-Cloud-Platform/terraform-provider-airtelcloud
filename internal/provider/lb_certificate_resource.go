package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

var _ resource.Resource = &LBCertificateResource{}
var _ resource.ResourceWithImportState = &LBCertificateResource{}

func NewLBCertificateResource() resource.Resource {
	return &LBCertificateResource{}
}

type LBCertificateResource struct {
	client *client.Client
}

type LBCertificateResourceModel struct {
	ID            types.String `tfsdk:"id"`
	LBServiceID   types.String `tfsdk:"lb_service_id"`
	Name          types.String `tfsdk:"name"`
	SSLCert       types.String `tfsdk:"ssl_cert"`
	SSLPrivateKey types.String `tfsdk:"ssl_private_key"`
	CACert        types.String `tfsdk:"ca_cert"`
	Status        types.String `tfsdk:"status"`
}

func (r *LBCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_certificate"
}

func (r *LBCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an SSL certificate for an Airtel Cloud Load Balancer Service. Certificates are immutable; any change requires replacement.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the certificate.",
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
				MarkdownDescription: "The name of the certificate.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssl_cert": schema.StringAttribute{
				MarkdownDescription: "The SSL certificate in PEM format.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssl_private_key": schema.StringAttribute{
				MarkdownDescription: "The SSL private key in PEM format.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "The CA certificate in PEM format.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the certificate.",
				Computed:            true,
			},
		},
	}
}

func (r *LBCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LBCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LBCertificateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &models.CreateLBCertificateRequest{
		Name:      data.Name.ValueString(),
		SSLCert:   data.SSLCert.ValueString(),
		SSLPvtKey: data.SSLPrivateKey.ValueString(),
		CACert:    data.CACert.ValueString(),
	}

	cert, err := r.client.CreateLBCertificate(ctx, data.LBServiceID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create LB certificate, got error: %s", err))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(cert.ID))
	data.Name = types.StringValue(cert.Name)
	data.Status = types.StringValue(cert.Status)

	tflog.Trace(ctx, "created LB certificate resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LBCertificateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse certificate ID: %s", err))
		return
	}

	// List all certificates and find by ID (no get-by-ID endpoint)
	certs, err := r.client.ListLBCertificates(ctx, data.LBServiceID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read LB certificates, got error: %s", err))
		return
	}

	var found bool
	for _, cert := range certs {
		if cert.ID == certID {
			data.Name = types.StringValue(cert.Name)
			data.Status = types.StringValue(cert.Status)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// ssl_cert and ssl_private_key are write-only: keep plan values in state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LBCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "LB certificates are immutable and cannot be updated.")
}

func (r *LBCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LBCertificateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse certificate ID: %s", err))
		return
	}

	err = r.client.DeleteLBCertificate(ctx, data.LBServiceID.ValueString(), certID)
	if err != nil {
		if client.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete LB certificate, got error: %s", err))
		return
	}
}

func (r *LBCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: lb_service_id/certificate_id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: lb_service_id/certificate_id")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lb_service_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
