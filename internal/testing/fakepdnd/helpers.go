package fakepdnd

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeProblem writes a Problem+JSON error response.
func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"type":   "about:blank",
		"status": status,
		"title":  title,
		"detail": detail,
	})
}

// parseUUID parses a UUID string, writes 400 error if invalid.
// Returns the parsed UUID and true on success, or zero UUID and false on failure.
func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid UUID: "+s)
		return uuid.Nil, false
	}
	return id, true
}

// agreementToJSON converts a StoredAgreement to a JSON-serializable map
// matching the OpenAPI Agreement schema field names (camelCase).
func agreementToJSON(a *StoredAgreement) map[string]interface{} {
	m := map[string]interface{}{
		"id":                  a.ID.String(),
		"eserviceId":          a.EServiceID.String(),
		"descriptorId":        a.DescriptorID.String(),
		"producerId":          a.ProducerID.String(),
		"consumerId":          a.ConsumerID.String(),
		"state":               a.State,
		"createdAt":           a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"suspendedByConsumer": a.SuspendedByConsumer,
		"suspendedByProducer": a.SuspendedByProducer,
		"suspendedByPlatform": a.SuspendedByPlatform,
	}
	if a.DelegationID != nil {
		m["delegationId"] = a.DelegationID.String()
	}
	if a.ConsumerNotes != "" {
		m["consumerNotes"] = a.ConsumerNotes
	}
	if a.RejectionReason != "" {
		m["rejectionReason"] = a.RejectionReason
	}
	if a.UpdatedAt != nil {
		m["updatedAt"] = a.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if a.SuspendedAt != nil {
		m["suspendedAt"] = a.SuspendedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}

// eserviceToJSON converts a StoredEService to a JSON-serializable map
// matching the OpenAPI EService schema field names (camelCase).
func eserviceToJSON(e *StoredEService) map[string]interface{} {
	return map[string]interface{}{
		"id":                      e.ID.String(),
		"producerId":              e.ProducerID.String(),
		"name":                    e.Name,
		"description":             e.Description,
		"technology":              e.Technology,
		"mode":                    e.Mode,
		"isSignalHubEnabled":      e.IsSignalHubEnabled,
		"isConsumerDelegable":     e.IsConsumerDelegable,
		"isClientAccessDelegable": e.IsClientAccessDelegable,
		"personalData":            e.PersonalData,
	}
}

// descriptorToJSON converts a StoredDescriptor to a JSON-serializable map
// matching the OpenAPI Descriptor schema field names (camelCase).
func descriptorToJSON(d *StoredDescriptor) map[string]interface{} {
	audience := d.Audience
	if audience == nil {
		audience = []string{}
	}
	serverUrls := d.ServerUrls
	if serverUrls == nil {
		serverUrls = []string{}
	}
	m := map[string]interface{}{
		"id":                      d.ID.String(),
		"eserviceId":              d.EServiceID.String(),
		"version":                 d.Version,
		"state":                   d.State,
		"agreementApprovalPolicy": d.AgreementApprovalPolicy,
		"audience":                audience,
		"dailyCallsPerConsumer":   d.DailyCallsPerConsumer,
		"dailyCallsTotal":        d.DailyCallsTotal,
		"voucherLifespan":        d.VoucherLifespan,
		"serverUrls":             serverUrls,
		"createdAt":              d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if d.Description != "" {
		m["description"] = d.Description
	}
	if d.PublishedAt != nil {
		m["publishedAt"] = d.PublishedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.SuspendedAt != nil {
		m["suspendedAt"] = d.SuspendedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.DeprecatedAt != nil {
		m["deprecatedAt"] = d.DeprecatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.ArchivedAt != nil {
		m["archivedAt"] = d.ArchivedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}

// certifiedAttributeToJSON converts a StoredCertifiedAttribute to a JSON-serializable map
// matching the OpenAPI CertifiedAttribute schema field names (camelCase).
func certifiedAttributeToJSON(a *StoredCertifiedAttribute) map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID.String(),
		"name":        a.Name,
		"description": a.Description,
		"code":        a.Code,
		"origin":      a.Origin,
		"createdAt":   a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// declaredAttributeToJSON converts a StoredDeclaredAttribute to a JSON-serializable map
// matching the OpenAPI DeclaredAttribute schema field names (camelCase).
func declaredAttributeToJSON(a *StoredDeclaredAttribute) map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID.String(),
		"name":        a.Name,
		"description": a.Description,
		"createdAt":   a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// verifiedAttributeToJSON converts a StoredVerifiedAttribute to a JSON-serializable map
// matching the OpenAPI VerifiedAttribute schema field names (camelCase).
func verifiedAttributeToJSON(a *StoredVerifiedAttribute) map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID.String(),
		"name":        a.Name,
		"description": a.Description,
		"createdAt":   a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// purposeToJSON converts a StoredPurpose to a JSON-serializable map (legacy, for agreement purposes).
func purposeToJSON(p *StoredPurpose) map[string]interface{} {
	m := map[string]interface{}{
		"id":                  p.ID.String(),
		"eserviceId":          p.EServiceID.String(),
		"consumerId":          p.ConsumerID.String(),
		"suspendedByConsumer": p.SuspendedByConsumer,
		"suspendedByProducer": p.SuspendedByProducer,
		"title":               p.Title,
		"description":         p.Description,
		"createdAt":           p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"isRiskAnalysisValid": p.IsRiskAnalysisValid,
		"isFreeOfCharge":      p.IsFreeOfCharge,
	}
	if p.DelegationID != nil {
		m["delegationId"] = p.DelegationID.String()
	}
	if p.FreeOfChargeReason != "" {
		m["freeOfChargeReason"] = p.FreeOfChargeReason
	}
	if p.UpdatedAt != nil {
		m["updatedAt"] = p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}

// fullPurposeToJSON converts a StoredPurpose to a JSON-serializable map with version fields.
func fullPurposeToJSON(p *StoredPurpose) map[string]interface{} {
	m := purposeToJSON(p)

	// Derive version fields from the versions list
	for i := len(p.Versions) - 1; i >= 0; i-- {
		v := &p.Versions[i]
		switch v.State {
		case "ACTIVE", "SUSPENDED", "ARCHIVED":
			m["currentVersion"] = purposeVersionToJSON(v)
		case "WAITING_FOR_APPROVAL":
			m["waitingForApprovalVersion"] = purposeVersionToJSON(v)
		case "REJECTED":
			m["rejectedVersion"] = purposeVersionToJSON(v)
		case "DRAFT":
			// Draft version is exposed as currentVersion if no active/suspended version exists
			if _, hasCV := m["currentVersion"]; !hasCV {
				m["currentVersion"] = purposeVersionToJSON(v)
			}
		}
	}

	return m
}

// purposeVersionToJSON converts a StoredPurposeVersion to a JSON-serializable map.
func purposeVersionToJSON(v *StoredPurposeVersion) map[string]interface{} {
	m := map[string]interface{}{
		"id":         v.ID.String(),
		"state":      v.State,
		"dailyCalls": v.DailyCalls,
		"createdAt":  v.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if v.UpdatedAt != nil {
		m["updatedAt"] = v.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if v.FirstActivationAt != nil {
		m["firstActivationAt"] = v.FirstActivationAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if v.SuspendedAt != nil {
		m["suspendedAt"] = v.SuspendedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if v.RejectionReason != nil {
		m["rejectionReason"] = *v.RejectionReason
	}
	return m
}
