package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

var _ resource.Resource = &descriptorVerifiedAttributesResource{}
var _ resource.ResourceWithImportState = &descriptorVerifiedAttributesResource{}

type descriptorVerifiedAttributesResource struct {
	client api.DescriptorAttributesAPI
}

func NewDescriptorVerifiedAttributesResource() resource.Resource {
	return &descriptorVerifiedAttributesResource{}
}

func (r *descriptorVerifiedAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor_verified_attributes"
}

func (r *descriptorVerifiedAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = descriptorAttributesSchema("verified")
}

func (r *descriptorVerifiedAttributesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = descriptorAttributesConfigure(req, resp)
}

func (r *descriptorVerifiedAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	descriptorAttributesCreate(ctx, req, resp, r.client, "verified")
}

func (r *descriptorVerifiedAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	descriptorAttributesRead(ctx, req, resp, r.client, "verified")
}

func (r *descriptorVerifiedAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	descriptorAttributesUpdate(ctx, req, resp, r.client, "verified")
}

func (r *descriptorVerifiedAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	descriptorAttributesDelete(ctx, req, resp, r.client, "verified")
}

func (r *descriptorVerifiedAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	descriptorAttributesImportState(ctx, req, resp)
}
