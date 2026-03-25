package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// DescriptorAttributeGroupModel represents a single attribute group.
type DescriptorAttributeGroupModel struct {
	AttributeIDs types.List `tfsdk:"attribute_ids"` // List of string UUIDs
}

// DescriptorAttributesResourceModel is the Terraform state model for descriptor attribute resources.
type DescriptorAttributesResourceModel struct {
	ID           types.String `tfsdk:"id"`
	EServiceID   types.String `tfsdk:"eservice_id"`
	DescriptorID types.String `tfsdk:"descriptor_id"`
	Groups       types.List   `tfsdk:"group"` // List of DescriptorAttributeGroupModel
}
