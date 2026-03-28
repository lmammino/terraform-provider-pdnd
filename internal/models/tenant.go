package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// TenantCertifiedAttrResourceModel for pdnd_tenant_certified_attribute.
type TenantCertifiedAttrResourceModel struct {
	ID          types.String `tfsdk:"id"`          // composite: tenant_id/attribute_id
	TenantID    types.String `tfsdk:"tenant_id"`
	AttributeID types.String `tfsdk:"attribute_id"`
	AssignedAt  types.String `tfsdk:"assigned_at"`
	RevokedAt   types.String `tfsdk:"revoked_at"`
}

// TenantDeclaredAttrResourceModel for pdnd_tenant_declared_attribute.
type TenantDeclaredAttrResourceModel struct {
	ID           types.String `tfsdk:"id"`
	TenantID     types.String `tfsdk:"tenant_id"`
	AttributeID  types.String `tfsdk:"attribute_id"`
	DelegationID types.String `tfsdk:"delegation_id"`
	AssignedAt   types.String `tfsdk:"assigned_at"`
	RevokedAt    types.String `tfsdk:"revoked_at"`
}

// TenantVerifiedAttrResourceModel for pdnd_tenant_verified_attribute.
type TenantVerifiedAttrResourceModel struct {
	ID             types.String `tfsdk:"id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	AttributeID    types.String `tfsdk:"attribute_id"`
	AgreementID    types.String `tfsdk:"agreement_id"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
	AssignedAt     types.String `tfsdk:"assigned_at"`
}

// TenantDataSourceModel for pdnd_tenant data source.
type TenantDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Kind        types.String `tfsdk:"kind"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	OnboardedAt types.String `tfsdk:"onboarded_at"`
}
