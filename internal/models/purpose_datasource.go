package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// PurposeDataSourceModel is the Terraform state model for pdnd_purpose data source.
type PurposeDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	DailyCalls          types.Int64  `tfsdk:"daily_calls"`
	State               types.String `tfsdk:"state"`
	IsFreeOfCharge      types.Bool   `tfsdk:"is_free_of_charge"`
	FreeOfChargeReason  types.String `tfsdk:"free_of_charge_reason"`
	IsRiskAnalysisValid types.Bool   `tfsdk:"is_risk_analysis_valid"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	DelegationID        types.String `tfsdk:"delegation_id"`
	VersionID           types.String `tfsdk:"version_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}
