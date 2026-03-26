package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// ClientKeyResourceModel is the Terraform state for pdnd_client_key resource.
type ClientKeyResourceModel struct {
	ID       types.String `tfsdk:"id"`        // Composite: client_id/kid
	ClientID types.String `tfsdk:"client_id"`
	Key      types.String `tfsdk:"key"`  // Base64 PEM public key
	Use      types.String `tfsdk:"use"`  // SIG or ENC
	Alg      types.String `tfsdk:"alg"`
	Name     types.String `tfsdk:"name"` // 5-60 chars
	// Computed from response
	Kid types.String `tfsdk:"kid"` // Server-assigned key ID
	Kty types.String `tfsdk:"kty"` // Key type (RSA, EC, etc.)
}

// ClientPurposeResourceModel is the Terraform state for pdnd_client_purpose resource.
type ClientPurposeResourceModel struct {
	ID        types.String `tfsdk:"id"`         // Composite: client_id/purpose_id
	ClientID  types.String `tfsdk:"client_id"`
	PurposeID types.String `tfsdk:"purpose_id"`
}

// ClientDataSourceModel is the Terraform state for pdnd_client data source.
type ClientDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ConsumerID  types.String `tfsdk:"consumer_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// ClientKeyDataSourceModel is for the nested key items in pdnd_client_keys.
type ClientKeyDataSourceModel struct {
	Kid types.String `tfsdk:"kid"`
	Kty types.String `tfsdk:"kty"`
	Alg types.String `tfsdk:"alg"`
	Use types.String `tfsdk:"use"`
}
