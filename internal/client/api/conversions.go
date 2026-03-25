package api

import (
	"github.com/google/uuid"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// agreementFromGenerated converts a generated Agreement to a domain Agreement.
func agreementFromGenerated(g *generated.Agreement) *Agreement {
	if g == nil {
		return nil
	}

	a := &Agreement{
		ID:                  uuid.UUID(g.Id),
		EServiceID:          uuid.UUID(g.EserviceId),
		DescriptorID:        uuid.UUID(g.DescriptorId),
		ProducerID:          uuid.UUID(g.ProducerId),
		ConsumerID:          uuid.UUID(g.ConsumerId),
		State:               string(g.State),
		SuspendedByConsumer: g.SuspendedByConsumer,
		SuspendedByProducer: g.SuspendedByProducer,
		SuspendedByPlatform: g.SuspendedByPlatform,
		ConsumerNotes:       g.ConsumerNotes,
		RejectionReason:     g.RejectionReason,
		CreatedAt:           g.CreatedAt,
		UpdatedAt:           g.UpdatedAt,
		SuspendedAt:         g.SuspendedAt,
	}

	if g.DelegationId != nil {
		id := uuid.UUID(*g.DelegationId)
		a.DelegationID = &id
	}

	return a
}

// agreementSeedToGenerated converts a domain AgreementSeed to a generated AgreementSeed.
func agreementSeedToGenerated(s AgreementSeed) generated.AgreementSeed {
	gs := generated.AgreementSeed{
		EserviceId:   openapi_types.UUID(s.EServiceID),
		DescriptorId: openapi_types.UUID(s.DescriptorID),
	}
	if s.DelegationID != nil {
		id := openapi_types.UUID(*s.DelegationID)
		gs.DelegationId = &id
	}
	return gs
}

// agreementSubmissionToGenerated converts a domain AgreementSubmission to a generated AgreementSubmission.
func agreementSubmissionToGenerated(s AgreementSubmission) generated.AgreementSubmission {
	return generated.AgreementSubmission{
		ConsumerNotes: s.ConsumerNotes,
	}
}

// agreementRejectionToGenerated converts a domain AgreementRejection to a generated AgreementRejection.
func agreementRejectionToGenerated(r AgreementRejection) generated.AgreementRejection {
	return generated.AgreementRejection{
		Reason: r.Reason,
	}
}

// delegationRefToGenerated converts a domain DelegationRef to a generated DelegationRef.
// If ref is nil, returns a zero-value DelegationRef (with nil DelegationId).
func delegationRefToGenerated(ref *DelegationRef) generated.DelegationRef {
	if ref == nil {
		return generated.DelegationRef{}
	}
	id := openapi_types.UUID(ref.DelegationID)
	return generated.DelegationRef{
		DelegationId: &id,
	}
}

// purposeFromGenerated converts a generated Purpose to a domain Purpose.
func purposeFromGenerated(g *generated.Purpose) Purpose {
	p := Purpose{
		ID:                  uuid.UUID(g.Id),
		EServiceID:          uuid.UUID(g.EserviceId),
		ConsumerID:          uuid.UUID(g.ConsumerId),
		SuspendedByConsumer: g.SuspendedByConsumer,
		SuspendedByProducer: g.SuspendedByProducer,
		Title:               g.Title,
		Description:         g.Description,
		CreatedAt:           g.CreatedAt,
		UpdatedAt:           g.UpdatedAt,
		IsRiskAnalysisValid: g.IsRiskAnalysisValid,
		IsFreeOfCharge:      g.IsFreeOfCharge,
		FreeOfChargeReason:  g.FreeOfChargeReason,
	}

	if g.DelegationId != nil {
		id := uuid.UUID(*g.DelegationId)
		p.DelegationID = &id
	}

	return p
}

// paginationFromGenerated converts a generated Pagination to a domain Pagination.
func paginationFromGenerated(g generated.Pagination) Pagination {
	return Pagination{
		Offset:     g.Offset,
		Limit:      g.Limit,
		TotalCount: g.TotalCount,
	}
}

// uuidsToOpenAPI converts a slice of uuid.UUID to a slice of openapi_types.UUID.
func uuidsToOpenAPI(ids []uuid.UUID) []openapi_types.UUID {
	if ids == nil {
		return nil
	}
	result := make([]openapi_types.UUID, len(ids))
	for i, id := range ids {
		result[i] = openapi_types.UUID(id)
	}
	return result
}
