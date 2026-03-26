package datasources

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	"github.com/lmammino/terraform-provider-pdnd/internal/models"
)

// populateDelegationDataSourceModel maps an API Delegation to a DelegationDataSourceModel.
func populateDelegationDataSourceModel(model *models.DelegationDataSourceModel, d *api.Delegation) {
	model.ID = types.StringValue(d.ID.String())
	model.EServiceID = types.StringValue(d.EServiceID.String())
	model.DelegateID = types.StringValue(d.DelegateID.String())
	model.DelegatorID = types.StringValue(d.DelegatorID.String())
	model.State = types.StringValue(d.State)
	model.CreatedAt = types.StringValue(d.CreatedAt.Format(time.RFC3339))
	model.SubmittedAt = types.StringValue(d.SubmittedAt.Format(time.RFC3339))

	if d.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(d.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if d.ActivatedAt != nil {
		model.ActivatedAt = types.StringValue(d.ActivatedAt.Format(time.RFC3339))
	} else {
		model.ActivatedAt = types.StringNull()
	}

	if d.RejectedAt != nil {
		model.RejectedAt = types.StringValue(d.RejectedAt.Format(time.RFC3339))
	} else {
		model.RejectedAt = types.StringNull()
	}

	if d.RevokedAt != nil {
		model.RevokedAt = types.StringValue(d.RevokedAt.Format(time.RFC3339))
	} else {
		model.RevokedAt = types.StringNull()
	}

	if d.RejectionReason != nil {
		model.RejectionReason = types.StringValue(*d.RejectionReason)
	} else {
		model.RejectionReason = types.StringNull()
	}
}
