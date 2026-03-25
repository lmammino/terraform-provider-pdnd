package resources

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
)

// populateModelFromAgreement populates the Terraform resource model from an API Agreement.
// It sets all computed fields but preserves desired_state and allow_pending from the plan/state.
func populateModelFromAgreement(model *models.AgreementResourceModel, a *api.Agreement) {
	model.ID = types.StringValue(a.ID.String())
	model.EServiceID = types.StringValue(a.EServiceID.String())
	model.DescriptorID = types.StringValue(a.DescriptorID.String())
	model.State = types.StringValue(a.State)
	model.ProducerID = types.StringValue(a.ProducerID.String())
	model.ConsumerID = types.StringValue(a.ConsumerID.String())
	model.CreatedAt = types.StringValue(a.CreatedAt.Format(time.RFC3339))

	if a.DelegationID != nil {
		model.DelegationID = types.StringValue(a.DelegationID.String())
	} else {
		model.DelegationID = types.StringNull()
	}

	if a.SuspendedByConsumer != nil {
		model.SuspendedByConsumer = types.BoolValue(*a.SuspendedByConsumer)
	} else {
		model.SuspendedByConsumer = types.BoolNull()
	}

	if a.SuspendedByProducer != nil {
		model.SuspendedByProducer = types.BoolValue(*a.SuspendedByProducer)
	} else {
		model.SuspendedByProducer = types.BoolNull()
	}

	if a.SuspendedByPlatform != nil {
		model.SuspendedByPlatform = types.BoolValue(*a.SuspendedByPlatform)
	} else {
		model.SuspendedByPlatform = types.BoolNull()
	}

	if a.ConsumerNotes != nil {
		model.ConsumerNotes = types.StringValue(*a.ConsumerNotes)
	} else {
		model.ConsumerNotes = types.StringNull()
	}

	if a.RejectionReason != nil {
		model.RejectionReason = types.StringValue(*a.RejectionReason)
	} else {
		model.RejectionReason = types.StringNull()
	}

	if a.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(a.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if a.SuspendedAt != nil {
		model.SuspendedAt = types.StringValue(a.SuspendedAt.Format(time.RFC3339))
	} else {
		model.SuspendedAt = types.StringNull()
	}
}

// canDelete returns true if an agreement in the given state can be deleted.
func canDelete(state string) bool {
	switch state {
	case "DRAFT", "PENDING", "MISSING_CERTIFIED_ATTRIBUTES":
		return true
	default:
		return false
	}
}

// inferDesiredState maps an observed API state to a desired_state value for import.
func inferDesiredState(observedState string) (string, error) {
	switch observedState {
	case "ACTIVE":
		return "ACTIVE", nil
	case "SUSPENDED":
		return "SUSPENDED", nil
	case "DRAFT":
		return "DRAFT", nil
	default:
		return "", fmt.Errorf("cannot import agreement in state %s: only ACTIVE, SUSPENDED, and DRAFT agreements can be imported", observedState)
	}
}
