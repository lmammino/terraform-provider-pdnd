package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

var _ resource.Resource = &descriptorCertifiedAttributesResource{}
var _ resource.ResourceWithImportState = &descriptorCertifiedAttributesResource{}

type descriptorCertifiedAttributesResource struct {
	client api.DescriptorAttributesAPI
}

func NewDescriptorCertifiedAttributesResource() resource.Resource {
	return &descriptorCertifiedAttributesResource{}
}

func (r *descriptorCertifiedAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor_certified_attributes"
}

func (r *descriptorCertifiedAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = descriptorAttributesSchema("certified")
}

func (r *descriptorCertifiedAttributesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = descriptorAttributesConfigure(req, resp)
}

func (r *descriptorCertifiedAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	descriptorAttributesCreate(ctx, req, resp, r.client, "certified")
}

func (r *descriptorCertifiedAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	descriptorAttributesRead(ctx, req, resp, r.client, "certified")
}

func (r *descriptorCertifiedAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	descriptorAttributesUpdate(ctx, req, resp, r.client, "certified")
}

func (r *descriptorCertifiedAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	descriptorAttributesDelete(ctx, req, resp, r.client, "certified")
}

func (r *descriptorCertifiedAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	descriptorAttributesImportState(ctx, req, resp)
}
