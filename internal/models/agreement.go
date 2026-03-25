package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AgreementResourceModel is the Terraform state model for pdnd_agreement resource.
type AgreementResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	DescriptorID        types.String `tfsdk:"descriptor_id"`
	DelegationID        types.String `tfsdk:"delegation_id"`
	DesiredState        types.String `tfsdk:"desired_state"`
	ConsumerNotes       types.String `tfsdk:"consumer_notes"`
	AllowPending        types.Bool   `tfsdk:"allow_pending"`
	State               types.String `tfsdk:"state"`
	ProducerID          types.String `tfsdk:"producer_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	SuspendedByPlatform types.Bool   `tfsdk:"suspended_by_platform"`
	RejectionReason     types.String `tfsdk:"rejection_reason"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	SuspendedAt         types.String `tfsdk:"suspended_at"`
}

// AgreementDataSourceModel is the Terraform state model for pdnd_agreement data source.
type AgreementDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	EServiceID          types.String `tfsdk:"eservice_id"`
	DescriptorID        types.String `tfsdk:"descriptor_id"`
	ProducerID          types.String `tfsdk:"producer_id"`
	ConsumerID          types.String `tfsdk:"consumer_id"`
	DelegationID        types.String `tfsdk:"delegation_id"`
	State               types.String `tfsdk:"state"`
	SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
	SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
	SuspendedByPlatform types.Bool   `tfsdk:"suspended_by_platform"`
	ConsumerNotes       types.String `tfsdk:"consumer_notes"`
	RejectionReason     types.String `tfsdk:"rejection_reason"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	SuspendedAt         types.String `tfsdk:"suspended_at"`
}
