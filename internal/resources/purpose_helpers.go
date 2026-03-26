package resources

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
)

// derivePurposeState derives the effective state from purpose versions.
func derivePurposeState(p *api.Purpose) string {
	if p.CurrentVersion != nil {
		return p.CurrentVersion.State
	}
	if p.WaitingForApprovalVersion != nil {
		return "WAITING_FOR_APPROVAL"
	}
	return "DRAFT"
}

// populateModelFromPurpose populates the Terraform resource model from an API Purpose.
func populateModelFromPurpose(model *models.PurposeResourceModel, p *api.Purpose) {
	model.ID = types.StringValue(p.ID.String())
	model.EServiceID = types.StringValue(p.EServiceID.String())
	model.ConsumerID = types.StringValue(p.ConsumerID.String())
	model.Title = types.StringValue(p.Title)
	model.Description = types.StringValue(p.Description)
	model.IsFreeOfCharge = types.BoolValue(p.IsFreeOfCharge)
	model.IsRiskAnalysisValid = types.BoolValue(p.IsRiskAnalysisValid)
	model.State = types.StringValue(derivePurposeState(p))
	model.CreatedAt = types.StringValue(p.CreatedAt.Format(time.RFC3339))

	// daily_calls and version_id from the most relevant version
	if p.CurrentVersion != nil {
		model.DailyCalls = types.Int64Value(int64(p.CurrentVersion.DailyCalls))
		model.VersionID = types.StringValue(p.CurrentVersion.ID.String())
		if p.CurrentVersion.FirstActivationAt != nil {
			model.FirstActivationAt = types.StringValue(p.CurrentVersion.FirstActivationAt.Format(time.RFC3339))
		} else {
			model.FirstActivationAt = types.StringNull()
		}
		if p.CurrentVersion.SuspendedAt != nil {
			model.SuspendedAt = types.StringValue(p.CurrentVersion.SuspendedAt.Format(time.RFC3339))
		} else {
			model.SuspendedAt = types.StringNull()
		}
	} else if p.WaitingForApprovalVersion != nil {
		model.DailyCalls = types.Int64Value(int64(p.WaitingForApprovalVersion.DailyCalls))
		model.VersionID = types.StringValue(p.WaitingForApprovalVersion.ID.String())
		model.FirstActivationAt = types.StringNull()
		model.SuspendedAt = types.StringNull()
	} else {
		// Should not happen for created purposes, but handle gracefully
		model.VersionID = types.StringNull()
		model.FirstActivationAt = types.StringNull()
		model.SuspendedAt = types.StringNull()
	}

	// Optional fields
	if p.FreeOfChargeReason != nil {
		model.FreeOfChargeReason = types.StringValue(*p.FreeOfChargeReason)
	} else {
		model.FreeOfChargeReason = types.StringNull()
	}

	if p.DelegationID != nil {
		model.DelegationID = types.StringValue(p.DelegationID.String())
	} else {
		model.DelegationID = types.StringNull()
	}

	if p.SuspendedByConsumer != nil {
		model.SuspendedByConsumer = types.BoolValue(*p.SuspendedByConsumer)
	} else {
		model.SuspendedByConsumer = types.BoolNull()
	}

	if p.SuspendedByProducer != nil {
		model.SuspendedByProducer = types.BoolValue(*p.SuspendedByProducer)
	} else {
		model.SuspendedByProducer = types.BoolNull()
	}

	if p.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(p.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if p.RejectedVersion != nil && p.RejectedVersion.RejectionReason != nil {
		model.RejectionReason = types.StringValue(*p.RejectedVersion.RejectionReason)
	} else {
		model.RejectionReason = types.StringNull()
	}
}

// canDeletePurpose returns true if a purpose in the given state can be deleted.
func canDeletePurpose(state string) bool {
	switch state {
	case "DRAFT", "WAITING_FOR_APPROVAL":
		return true
	default:
		return false
	}
}

// inferPurposeDesiredState maps an observed API state to a desired_state value for import.
func inferPurposeDesiredState(observedState string) (string, error) {
	switch observedState {
	case "ACTIVE":
		return "ACTIVE", nil
	case "SUSPENDED":
		return "SUSPENDED", nil
	case "DRAFT":
		return "DRAFT", nil
	default:
		return "", fmt.Errorf("cannot import purpose in state %s: only ACTIVE, SUSPENDED, and DRAFT purposes can be imported", observedState)
	}
}
