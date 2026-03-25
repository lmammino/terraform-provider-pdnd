package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
)

// populateDescriptorModel populates the Terraform resource model from an API Descriptor.
// It sets all computed fields but preserves desired_state and allow_waiting_for_approval from the plan/state.
func populateDescriptorModel(ctx context.Context, model *models.EServiceDescriptorResourceModel, d *api.Descriptor) {
	model.ID = types.StringValue(d.ID.String())
	model.Version = types.StringValue(d.Version)
	model.State = types.StringValue(d.State)
	model.AgreementApprovalPolicy = types.StringValue(d.AgreementApprovalPolicy)
	model.DailyCallsPerConsumer = types.Int64Value(int64(d.DailyCallsPerConsumer))
	model.DailyCallsTotal = types.Int64Value(int64(d.DailyCallsTotal))
	model.VoucherLifespan = types.Int64Value(int64(d.VoucherLifespan))

	// Audience
	model.Audience, _ = types.ListValueFrom(ctx, types.StringType, d.Audience)

	// ServerUrls
	if d.ServerUrls != nil {
		model.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, d.ServerUrls)
	} else {
		model.ServerUrls, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	// Description
	if d.Description != nil {
		model.Description = types.StringValue(*d.Description)
	} else {
		model.Description = types.StringNull()
	}

	// Timestamps
	if d.PublishedAt != nil {
		model.PublishedAt = types.StringValue(d.PublishedAt.Format(time.RFC3339))
	} else {
		model.PublishedAt = types.StringNull()
	}

	if d.SuspendedAt != nil {
		model.SuspendedAt = types.StringValue(d.SuspendedAt.Format(time.RFC3339))
	} else {
		model.SuspendedAt = types.StringNull()
	}

	if d.DeprecatedAt != nil {
		model.DeprecatedAt = types.StringValue(d.DeprecatedAt.Format(time.RFC3339))
	} else {
		model.DeprecatedAt = types.StringNull()
	}

	if d.ArchivedAt != nil {
		model.ArchivedAt = types.StringValue(d.ArchivedAt.Format(time.RFC3339))
	} else {
		model.ArchivedAt = types.StringNull()
	}
}

// canDeleteDescriptor returns true if a descriptor in the given state can be deleted.
func canDeleteDescriptor(state string) bool {
	return state == "DRAFT"
}

// inferDescriptorDesiredState maps an observed API state to a desired_state value for import.
func inferDescriptorDesiredState(observedState string) (string, error) {
	switch observedState {
	case "PUBLISHED":
		return "PUBLISHED", nil
	case "SUSPENDED":
		return "SUSPENDED", nil
	case "DRAFT":
		return "DRAFT", nil
	default:
		return "", fmt.Errorf("cannot import descriptor in state %s: only PUBLISHED, SUSPENDED, and DRAFT descriptors can be imported", observedState)
	}
}
