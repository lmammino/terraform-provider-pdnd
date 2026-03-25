package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// pdndProviderModel maps provider schema data to a Go type.
type pdndProviderModel struct {
	BaseURL         types.String `tfsdk:"base_url"`
	AccessToken     types.String `tfsdk:"access_token"`
	ClientID        types.String `tfsdk:"client_id"`
	PurposeID       types.String `tfsdk:"purpose_id"`
	TokenEndpoint   types.String `tfsdk:"token_endpoint"`
	DPoPPrivateKey  types.String `tfsdk:"dpop_private_key"`
	DPoPKeyID       types.String `tfsdk:"dpop_key_id"`
	RequestTimeoutS types.Int64  `tfsdk:"request_timeout_s"`
}
