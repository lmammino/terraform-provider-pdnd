package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

var _ resource.Resource = &producerDelegationResource{}
var _ resource.ResourceWithImportState = &producerDelegationResource{}

type producerDelegationResource struct {
	client api.DelegationsAPI
}

func NewProducerDelegationResource() resource.Resource {
	return &producerDelegationResource{}
}

func (r *producerDelegationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_producer_delegation"
}

func (r *producerDelegationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = delegationSchema("producer")
}

func (r *producerDelegationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = delegationConfigure(req, resp)
}

func (r *producerDelegationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	delegationCreate(ctx, req, resp, r.client, "producer")
}

func (r *producerDelegationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	delegationRead(ctx, req, resp, r.client, "producer")
}

func (r *producerDelegationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Producer delegations do not support in-place updates. All user-configurable fields require replacement.",
	)
}

func (r *producerDelegationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	delegationDelete(ctx, req, resp, r.client, "producer")
}

func (r *producerDelegationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	delegationImportState(ctx, req, resp)
}
