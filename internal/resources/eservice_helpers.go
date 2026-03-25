package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
)

// populateEServiceModel populates the Terraform resource model from an API EService and optional initial descriptor.
func populateEServiceModel(model *models.EServiceResourceModel, es *api.EService, initialDescriptor *api.Descriptor) {
	model.ID = types.StringValue(es.ID.String())
	model.Name = types.StringValue(es.Name)
	model.Description = types.StringValue(es.Description)
	model.Technology = types.StringValue(es.Technology)
	model.Mode = types.StringValue(es.Mode)
	model.ProducerID = types.StringValue(es.ProducerID.String())

	if es.IsSignalHubEnabled != nil {
		model.IsSignalHubEnabled = types.BoolValue(*es.IsSignalHubEnabled)
	} else {
		model.IsSignalHubEnabled = types.BoolValue(false)
	}
	if es.IsConsumerDelegable != nil {
		model.IsConsumerDelegable = types.BoolValue(*es.IsConsumerDelegable)
	} else {
		model.IsConsumerDelegable = types.BoolValue(false)
	}
	if es.IsClientAccessDelegable != nil {
		model.IsClientAccessDelegable = types.BoolValue(*es.IsClientAccessDelegable)
	} else {
		model.IsClientAccessDelegable = types.BoolValue(false)
	}
	if es.PersonalData != nil {
		model.PersonalData = types.BoolValue(*es.PersonalData)
	} else {
		model.PersonalData = types.BoolValue(false)
	}

	if es.TemplateID != nil {
		model.TemplateID = types.StringValue(es.TemplateID.String())
	} else {
		model.TemplateID = types.StringNull()
	}

	if initialDescriptor != nil {
		model.InitialDescriptorID = types.StringValue(initialDescriptor.ID.String())
	}
}

// populateInitialDescriptorFromAPI populates the initial_descriptor_* fields from an actual API descriptor.
// Used during import when these fields are not yet in state.
func populateInitialDescriptorFromAPI(model *models.EServiceResourceModel, d *api.Descriptor) {
	model.InitialDescriptorAgreementApprovalPolicy = types.StringValue(d.AgreementApprovalPolicy)

	audienceValues := make([]types.String, len(d.Audience))
	for i, a := range d.Audience {
		audienceValues[i] = types.StringValue(a)
	}
	model.InitialDescriptorAudience, _ = types.ListValueFrom(context.Background(), types.StringType, d.Audience)

	model.InitialDescriptorDailyCallsPerConsumer = types.Int64Value(int64(d.DailyCallsPerConsumer))
	model.InitialDescriptorDailyCallsTotal = types.Int64Value(int64(d.DailyCallsTotal))
	model.InitialDescriptorVoucherLifespan = types.Int64Value(int64(d.VoucherLifespan))

	if d.Description != nil {
		model.InitialDescriptorDescription = types.StringValue(*d.Description)
	} else {
		model.InitialDescriptorDescription = types.StringNull()
	}
}

// isEServiceDraft returns true if all descriptors are in DRAFT state (or there are no descriptors).
func isEServiceDraft(descriptors []api.Descriptor) bool {
	for _, d := range descriptors {
		if d.State != "DRAFT" {
			return false
		}
	}
	return true
}

// findInitialDescriptor finds the initial descriptor (version "1") from a list of descriptors.
// If no descriptor has version "1", returns the first descriptor. Returns nil if the list is empty.
func findInitialDescriptor(descriptors []api.Descriptor) *api.Descriptor {
	if len(descriptors) == 0 {
		return nil
	}
	for i := range descriptors {
		if descriptors[i].Version == "1" {
			return &descriptors[i]
		}
	}
	return &descriptors[0]
}
