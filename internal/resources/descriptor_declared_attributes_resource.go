package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

var _ resource.Resource = &descriptorDeclaredAttributesResource{}
var _ resource.ResourceWithImportState = &descriptorDeclaredAttributesResource{}

type descriptorDeclaredAttributesResource struct {
	client api.DescriptorAttributesAPI
}

func NewDescriptorDeclaredAttributesResource() resource.Resource {
	return &descriptorDeclaredAttributesResource{}
}

func (r *descriptorDeclaredAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eservice_descriptor_declared_attributes"
}

func (r *descriptorDeclaredAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = descriptorAttributesSchema("declared")
}

func (r *descriptorDeclaredAttributesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = descriptorAttributesConfigure(req, resp)
}

func (r *descriptorDeclaredAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	descriptorAttributesCreate(ctx, req, resp, r.client, "declared")
}

func (r *descriptorDeclaredAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	descriptorAttributesRead(ctx, req, resp, r.client, "declared")
}

func (r *descriptorDeclaredAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	descriptorAttributesUpdate(ctx, req, resp, r.client, "declared")
}

func (r *descriptorDeclaredAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	descriptorAttributesDelete(ctx, req, resp, r.client, "declared")
}

func (r *descriptorDeclaredAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	descriptorAttributesImportState(ctx, req, resp)
}
