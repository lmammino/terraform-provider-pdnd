package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// DescriptorDocumentResourceModel is the Terraform state for pdnd_eservice_descriptor_document.
type DescriptorDocumentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	EServiceID   types.String `tfsdk:"eservice_id"`
	DescriptorID types.String `tfsdk:"descriptor_id"`
	PrettyName   types.String `tfsdk:"pretty_name"`
	FilePath     types.String `tfsdk:"file_path"`
	ContentType  types.String `tfsdk:"content_type"`
	DocumentID   types.String `tfsdk:"document_id"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"created_at"`
	FileHash     types.String `tfsdk:"file_hash"`
}

// DescriptorInterfaceResourceModel is the Terraform state for pdnd_eservice_descriptor_interface.
type DescriptorInterfaceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	EServiceID   types.String `tfsdk:"eservice_id"`
	DescriptorID types.String `tfsdk:"descriptor_id"`
	PrettyName   types.String `tfsdk:"pretty_name"`
	FilePath     types.String `tfsdk:"file_path"`
	ContentType  types.String `tfsdk:"content_type"`
	DocumentID   types.String `tfsdk:"document_id"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"created_at"`
	FileHash     types.String `tfsdk:"file_hash"`
}
