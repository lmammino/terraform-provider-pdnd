package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// EServiceResourceModel is the Terraform state model for pdnd_eservice resource.
type EServiceResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Technology              types.String `tfsdk:"technology"`
	Mode                    types.String `tfsdk:"mode"`
	IsSignalHubEnabled      types.Bool   `tfsdk:"is_signal_hub_enabled"`
	IsConsumerDelegable     types.Bool   `tfsdk:"is_consumer_delegable"`
	IsClientAccessDelegable types.Bool   `tfsdk:"is_client_access_delegable"`
	PersonalData            types.Bool   `tfsdk:"personal_data"`
	ProducerID              types.String `tfsdk:"producer_id"`
	TemplateID              types.String `tfsdk:"template_id"`
	// Initial descriptor (set at creation, read-only after)
	InitialDescriptorID                      types.String `tfsdk:"initial_descriptor_id"`
	InitialDescriptorAgreementApprovalPolicy types.String `tfsdk:"initial_descriptor_agreement_approval_policy"`
	InitialDescriptorAudience                types.List   `tfsdk:"initial_descriptor_audience"`
	InitialDescriptorDailyCallsPerConsumer   types.Int64  `tfsdk:"initial_descriptor_daily_calls_per_consumer"`
	InitialDescriptorDailyCallsTotal         types.Int64  `tfsdk:"initial_descriptor_daily_calls_total"`
	InitialDescriptorVoucherLifespan         types.Int64  `tfsdk:"initial_descriptor_voucher_lifespan"`
	InitialDescriptorDescription             types.String `tfsdk:"initial_descriptor_description"`
}

// EServiceDescriptorResourceModel is the Terraform state model for pdnd_eservice_descriptor resource.
type EServiceDescriptorResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	EServiceID              types.String `tfsdk:"eservice_id"`
	Version                 types.String `tfsdk:"version"`
	DesiredState            types.String `tfsdk:"desired_state"`
	State                   types.String `tfsdk:"state"`
	AgreementApprovalPolicy types.String `tfsdk:"agreement_approval_policy"`
	Audience                types.List   `tfsdk:"audience"`
	DailyCallsPerConsumer   types.Int64  `tfsdk:"daily_calls_per_consumer"`
	DailyCallsTotal         types.Int64  `tfsdk:"daily_calls_total"`
	VoucherLifespan         types.Int64  `tfsdk:"voucher_lifespan"`
	ServerUrls              types.List   `tfsdk:"server_urls"`
	Description             types.String `tfsdk:"description"`
	AllowWaitingForApproval types.Bool   `tfsdk:"allow_waiting_for_approval"`
	PublishedAt             types.String `tfsdk:"published_at"`
	SuspendedAt             types.String `tfsdk:"suspended_at"`
	DeprecatedAt            types.String `tfsdk:"deprecated_at"`
	ArchivedAt              types.String `tfsdk:"archived_at"`
}

// EServiceDataSourceModel is the Terraform state model for pdnd_eservice data source.
type EServiceDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Technology              types.String `tfsdk:"technology"`
	Mode                    types.String `tfsdk:"mode"`
	IsSignalHubEnabled      types.Bool   `tfsdk:"is_signal_hub_enabled"`
	IsConsumerDelegable     types.Bool   `tfsdk:"is_consumer_delegable"`
	IsClientAccessDelegable types.Bool   `tfsdk:"is_client_access_delegable"`
	PersonalData            types.Bool   `tfsdk:"personal_data"`
	ProducerID              types.String `tfsdk:"producer_id"`
	TemplateID              types.String `tfsdk:"template_id"`
}
