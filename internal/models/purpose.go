package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// PurposeResourceModel is the Terraform state model for pdnd_purpose resource.
type PurposeResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	EServiceID              types.String `tfsdk:"eservice_id"`
	Title                   types.String `tfsdk:"title"`
	Description             types.String `tfsdk:"description"`
	DailyCalls              types.Int64  `tfsdk:"daily_calls"`
	IsFreeOfCharge          types.Bool   `tfsdk:"is_free_of_charge"`
	FreeOfChargeReason      types.String `tfsdk:"free_of_charge_reason"`
	DelegationID            types.String `tfsdk:"delegation_id"`
	DesiredState            types.String `tfsdk:"desired_state"`
	AllowWaitingForApproval types.Bool   `tfsdk:"allow_waiting_for_approval"`
	// Computed
	State               types.String `tfsdk:"state"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	IsRiskAnalysisValid types.Bool   `tfsdk:"is_risk_analysis_valid"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	VersionID           types.String `tfsdk:"version_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	FirstActivationAt   types.String `tfsdk:"first_activation_at"`
	SuspendedAt         types.String `tfsdk:"suspended_at"`
	RejectionReason     types.String `tfsdk:"rejection_reason"`
}
