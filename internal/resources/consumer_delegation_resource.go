package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

var _ resource.Resource = &consumerDelegationResource{}
var _ resource.ResourceWithImportState = &consumerDelegationResource{}

type consumerDelegationResource struct {
	client api.DelegationsAPI
}

func NewConsumerDelegationResource() resource.Resource {
	return &consumerDelegationResource{}
}

func (r *consumerDelegationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consumer_delegation"
}

func (r *consumerDelegationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = delegationSchema("consumer")
}

func (r *consumerDelegationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = delegationConfigure(req, resp)
}

func (r *consumerDelegationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	delegationCreate(ctx, req, resp, r.client, "consumer")
}

func (r *consumerDelegationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	delegationRead(ctx, req, resp, r.client, "consumer")
}

func (r *consumerDelegationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Consumer delegations do not support in-place updates. All user-configurable fields require replacement.",
	)
}

func (r *consumerDelegationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	delegationDelete(ctx, req, resp, r.client, "consumer")
}

func (r *consumerDelegationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	delegationImportState(ctx, req, resp)
}
