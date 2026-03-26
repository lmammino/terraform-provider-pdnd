package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// DelegationResourceModel is the Terraform state model for delegation resources.
type DelegationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	EServiceID      types.String `tfsdk:"eservice_id"`
	DelegateID      types.String `tfsdk:"delegate_id"`
	// Computed
	DelegatorID     types.String `tfsdk:"delegator_id"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
	SubmittedAt     types.String `tfsdk:"submitted_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	ActivatedAt     types.String `tfsdk:"activated_at"`
	RejectedAt      types.String `tfsdk:"rejected_at"`
	RevokedAt       types.String `tfsdk:"revoked_at"`
	RejectionReason types.String `tfsdk:"rejection_reason"`
}

// DelegationDataSourceModel is the Terraform state model for singular delegation data sources.
type DelegationDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	EServiceID      types.String `tfsdk:"eservice_id"`
	DelegateID      types.String `tfsdk:"delegate_id"`
	DelegatorID     types.String `tfsdk:"delegator_id"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
	SubmittedAt     types.String `tfsdk:"submitted_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	ActivatedAt     types.String `tfsdk:"activated_at"`
	RejectedAt      types.String `tfsdk:"rejected_at"`
	RevokedAt       types.String `tfsdk:"revoked_at"`
	RejectionReason types.String `tfsdk:"rejection_reason"`
}
