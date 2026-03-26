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

// purposeVersionFromGenerated converts a generated PurposeVersion to a domain PurposeVersion.
func purposeVersionFromGenerated(g *generated.PurposeVersion) *PurposeVersion {
	if g == nil {
		return nil
	}
	return &PurposeVersion{
		ID:                uuid.UUID(g.Id),
		State:             string(g.State),
		DailyCalls:        g.DailyCalls,
		CreatedAt:         g.CreatedAt,
		UpdatedAt:         g.UpdatedAt,
		FirstActivationAt: g.FirstActivationAt,
		SuspendedAt:       g.SuspendedAt,
		RejectionReason:   g.RejectionReason,
	}
}

// purposeFromGenerated converts a generated Purpose to a domain Purpose.
func purposeFromGenerated(g *generated.Purpose) Purpose {
	p := Purpose{
		ID:                        uuid.UUID(g.Id),
		EServiceID:                uuid.UUID(g.EserviceId),
		ConsumerID:                uuid.UUID(g.ConsumerId),
		SuspendedByConsumer:       g.SuspendedByConsumer,
		SuspendedByProducer:       g.SuspendedByProducer,
		Title:                     g.Title,
		Description:               g.Description,
		CreatedAt:                 g.CreatedAt,
		UpdatedAt:                 g.UpdatedAt,
		IsRiskAnalysisValid:       g.IsRiskAnalysisValid,
		IsFreeOfCharge:            g.IsFreeOfCharge,
		FreeOfChargeReason:        g.FreeOfChargeReason,
		CurrentVersion:            purposeVersionFromGenerated(g.CurrentVersion),
		WaitingForApprovalVersion: purposeVersionFromGenerated(g.WaitingForApprovalVersion),
		RejectedVersion:           purposeVersionFromGenerated(g.RejectedVersion),
	}

	if g.DelegationId != nil {
		id := uuid.UUID(*g.DelegationId)
		p.DelegationID = &id
	}
	if g.PurposeTemplateId != nil {
		id := uuid.UUID(*g.PurposeTemplateId)
		p.PurposeTemplateID = &id
	}

	return p
}

// purposeSeedToGenerated converts a domain PurposeSeed to a generated PurposeSeed.
func purposeSeedToGenerated(s PurposeSeed) generated.PurposeSeed {
	gs := generated.PurposeSeed{
		EserviceId:         openapi_types.UUID(s.EServiceID),
		Title:              s.Title,
		Description:        s.Description,
		DailyCalls:         s.DailyCalls,
		IsFreeOfCharge:     s.IsFreeOfCharge,
		FreeOfChargeReason: s.FreeOfChargeReason,
	}
	if s.DelegationID != nil {
		id := openapi_types.UUID(*s.DelegationID)
		gs.DelegationId = &id
	}
	return gs
}

// purposeVersionSeedToGenerated converts a domain PurposeVersionSeed to a generated PurposeVersionSeed.
func purposeVersionSeedToGenerated(s PurposeVersionSeed) generated.PurposeVersionSeed {
	return generated.PurposeVersionSeed{
		DailyCalls: s.DailyCalls,
	}
}

// paginationFromGenerated converts a generated Pagination to a domain Pagination.
func paginationFromGenerated(g generated.Pagination) Pagination {
	return Pagination{
		Offset:     g.Offset,
		Limit:      g.Limit,
		TotalCount: g.TotalCount,
	}
}

// eserviceFromGenerated converts a generated EService to a domain EService.
func eserviceFromGenerated(g *generated.EService) *EService {
	if g == nil {
		return nil
	}

	e := &EService{
		ID:                      uuid.UUID(g.Id),
		ProducerID:              uuid.UUID(g.ProducerId),
		Name:                    g.Name,
		Description:             g.Description,
		Technology:              string(g.Technology),
		Mode:                    string(g.Mode),
		IsSignalHubEnabled:      g.IsSignalHubEnabled,
		IsConsumerDelegable:     g.IsConsumerDelegable,
		IsClientAccessDelegable: g.IsClientAccessDelegable,
		PersonalData:            g.PersonalData,
	}

	if g.TemplateId != nil {
		id := uuid.UUID(*g.TemplateId)
		e.TemplateID = &id
	}

	return e
}

// eserviceSeedToGenerated converts a domain EServiceSeed to a generated EServiceSeed.
func eserviceSeedToGenerated(s EServiceSeed) generated.EServiceSeed {
	return generated.EServiceSeed{
		Name:                    s.Name,
		Description:             s.Description,
		Technology:              generated.EServiceTechnology(s.Technology),
		Mode:                    generated.EServiceMode(s.Mode),
		IsSignalHubEnabled:      s.IsSignalHubEnabled,
		IsConsumerDelegable:     s.IsConsumerDelegable,
		IsClientAccessDelegable: s.IsClientAccessDelegable,
		PersonalData:            s.PersonalData,
		Descriptor: generated.DescriptorSeedForEServiceCreation{
			AgreementApprovalPolicy: generated.AgreementApprovalPolicy(s.Descriptor.AgreementApprovalPolicy),
			Audience:                s.Descriptor.Audience,
			DailyCallsPerConsumer:   s.Descriptor.DailyCallsPerConsumer,
			DailyCallsTotal:         s.Descriptor.DailyCallsTotal,
			VoucherLifespan:         s.Descriptor.VoucherLifespan,
			Description:             s.Descriptor.Description,
		},
	}
}

// descriptorFromGenerated converts a generated EServiceDescriptor to a domain Descriptor.
func descriptorFromGenerated(g *generated.EServiceDescriptor) *Descriptor {
	if g == nil {
		return nil
	}

	return &Descriptor{
		ID:                      uuid.UUID(g.Id),
		Version:                 g.Version,
		State:                   string(g.State),
		AgreementApprovalPolicy: string(g.AgreementApprovalPolicy),
		Audience:                g.Audience,
		DailyCallsPerConsumer:   g.DailyCallsPerConsumer,
		DailyCallsTotal:         g.DailyCallsTotal,
		VoucherLifespan:         g.VoucherLifespan,
		ServerUrls:              g.ServerUrls,
		Description:             g.Description,
		PublishedAt:             g.PublishedAt,
		SuspendedAt:             g.SuspendedAt,
		DeprecatedAt:            g.DeprecatedAt,
		ArchivedAt:              g.ArchivedAt,
	}
}

// descriptorSeedToGenerated converts a domain DescriptorSeed to a generated EServiceDescriptorSeed.
func descriptorSeedToGenerated(s DescriptorSeed) generated.EServiceDescriptorSeed {
	return generated.EServiceDescriptorSeed{
		AgreementApprovalPolicy: generated.AgreementApprovalPolicy(s.AgreementApprovalPolicy),
		Audience:                s.Audience,
		DailyCallsPerConsumer:   s.DailyCallsPerConsumer,
		DailyCallsTotal:         s.DailyCallsTotal,
		VoucherLifespan:         s.VoucherLifespan,
		Description:             s.Description,
	}
}

// eserviceDraftUpdateToGenerated converts a domain EServiceDraftUpdate to a generated EServiceDraftUpdateSeed.
func eserviceDraftUpdateToGenerated(s EServiceDraftUpdate) generated.EServiceDraftUpdateSeed {
	gs := generated.EServiceDraftUpdateSeed{
		Name:                    s.Name,
		Description:             s.Description,
		IsSignalHubEnabled:      s.IsSignalHubEnabled,
		IsConsumerDelegable:     s.IsConsumerDelegable,
		IsClientAccessDelegable: s.IsClientAccessDelegable,
		PersonalData:            s.PersonalData,
	}
	if s.Technology != nil {
		t := generated.EServiceTechnology(*s.Technology)
		gs.Technology = &t
	}
	if s.Mode != nil {
		m := generated.EServiceMode(*s.Mode)
		gs.Mode = &m
	}
	return gs
}

// descriptorDraftUpdateToGenerated converts a domain DescriptorDraftUpdate to a generated EServiceDescriptorDraftUpdateSeed.
func descriptorDraftUpdateToGenerated(s DescriptorDraftUpdate) generated.EServiceDescriptorDraftUpdateSeed {
	gs := generated.EServiceDescriptorDraftUpdateSeed{
		DailyCallsPerConsumer: s.DailyCallsPerConsumer,
		DailyCallsTotal:       s.DailyCallsTotal,
		VoucherLifespan:       s.VoucherLifespan,
		Description:           s.Description,
	}
	if s.AgreementApprovalPolicy != nil {
		p := generated.AgreementApprovalPolicy(*s.AgreementApprovalPolicy)
		gs.AgreementApprovalPolicy = &p
	}
	if s.Audience != nil {
		gs.Audience = &s.Audience
	}
	return gs
}

// descriptorQuotasUpdateToGenerated converts a domain DescriptorQuotasUpdate to a generated EServiceDescriptorQuotasUpdateSeed.
func descriptorQuotasUpdateToGenerated(s DescriptorQuotasUpdate) generated.EServiceDescriptorQuotasUpdateSeed {
	return generated.EServiceDescriptorQuotasUpdateSeed{
		DailyCallsPerConsumer: s.DailyCallsPerConsumer,
		DailyCallsTotal:       s.DailyCallsTotal,
		VoucherLifespan:       s.VoucherLifespan,
	}
}

// certifiedAttributeFromGenerated converts a generated CertifiedAttribute to a domain CertifiedAttribute.
func certifiedAttributeFromGenerated(g *generated.CertifiedAttribute) *CertifiedAttribute {
	if g == nil {
		return nil
	}
	return &CertifiedAttribute{
		ID:          uuid.UUID(g.Id),
		Name:        g.Name,
		Description: g.Description,
		Code:        g.Code,
		Origin:      g.Origin,
		CreatedAt:   g.CreatedAt,
	}
}

// declaredAttributeFromGenerated converts a generated DeclaredAttribute to a domain DeclaredAttribute.
func declaredAttributeFromGenerated(g *generated.DeclaredAttribute) *DeclaredAttribute {
	if g == nil {
		return nil
	}
	return &DeclaredAttribute{
		ID:          uuid.UUID(g.Id),
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
	}
}

// verifiedAttributeFromGenerated converts a generated VerifiedAttribute to a domain VerifiedAttribute.
func verifiedAttributeFromGenerated(g *generated.VerifiedAttribute) *VerifiedAttribute {
	if g == nil {
		return nil
	}
	return &VerifiedAttribute{
		ID:          uuid.UUID(g.Id),
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
	}
}

// documentFromGenerated converts a generated Document to a domain DescriptorDocument.
func documentFromGenerated(g *generated.Document) *DescriptorDocument {
	if g == nil {
		return nil
	}
	return &DescriptorDocument{
		ID:          uuid.UUID(g.Id),
		Name:        g.Name,
		PrettyName:  g.PrettyName,
		ContentType: g.ContentType,
		CreatedAt:   g.CreatedAt,
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
