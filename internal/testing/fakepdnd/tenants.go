package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// tenantToJSON converts a StoredTenant to a JSON-serializable map.
func tenantToJSON(t *StoredTenant) map[string]interface{} {
	m := map[string]interface{}{
		"id":        t.ID.String(),
		"name":      t.Name,
		"createdAt": t.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"externalId": map[string]interface{}{
			"origin": t.ExternalOrigin,
			"value":  t.ExternalValue,
		},
	}
	if t.Kind != "" {
		m["kind"] = t.Kind
	}
	if t.UpdatedAt != nil {
		m["updatedAt"] = t.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if t.OnboardedAt != nil {
		m["onboardedAt"] = t.OnboardedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}

// tenantCertifiedAttrToJSON converts a StoredTenantCertifiedAttr to a JSON-serializable map.
func tenantCertifiedAttrToJSON(a *StoredTenantCertifiedAttr) map[string]interface{} {
	m := map[string]interface{}{
		"id":         a.ID.String(),
		"assignedAt": a.AssignedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if a.RevokedAt != nil {
		m["revokedAt"] = a.RevokedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m
}

// tenantDeclaredAttrToJSON converts a StoredTenantDeclaredAttr to a JSON-serializable map.
func tenantDeclaredAttrToJSON(a *StoredTenantDeclaredAttr) map[string]interface{} {
	m := map[string]interface{}{
		"id":         a.ID.String(),
		"assignedAt": a.AssignedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if a.RevokedAt != nil {
		m["revokedAt"] = a.RevokedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if a.DelegationID != nil {
		m["delegationId"] = a.DelegationID.String()
	}
	return m
}

// tenantVerifiedAttrToJSON converts a StoredTenantVerifiedAttr to a JSON-serializable map.
func tenantVerifiedAttrToJSON(a *StoredTenantVerifiedAttr) map[string]interface{} {
	return map[string]interface{}{
		"id":         a.ID.String(),
		"assignedAt": a.AssignedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// handleListTenants filters tenants by IPACode/taxCode and paginates.
func (s *FakeServer) handleListTenants(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)
	ipaCode := q.Get("IPACode")
	taxCode := q.Get("taxCode")

	var filtered []map[string]interface{}
	for _, t := range s.tenants {
		if ipaCode != "" && !strings.EqualFold(t.ExternalValue, ipaCode) {
			continue
		}
		if taxCode != "" && !strings.EqualFold(t.ExternalValue, taxCode) {
			continue
		}
		filtered = append(filtered, tenantToJSON(t))
	}

	totalCount := len(filtered)

	// Apply pagination.
	if offset > len(filtered) {
		offset = len(filtered)
	}
	filtered = filtered[offset:]
	if limit < len(filtered) {
		filtered = filtered[:limit]
	}
	if filtered == nil {
		filtered = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": filtered,
		"pagination": map[string]interface{}{
			"offset":     offset,
			"limit":      limit,
			"totalCount": totalCount,
		},
	})
}

// handleGetTenant returns a tenant by ID.
func (s *FakeServer) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	s.mu.RLock()
	t := s.tenants[id]
	s.mu.RUnlock()

	if t == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Tenant %s not found", id))
		return
	}

	writeJSON(w, http.StatusOK, tenantToJSON(t))
}

// handleListTenantCertifiedAttrs lists certified attributes for a tenant with pagination.
func (s *FakeServer) handleListTenantCertifiedAttrs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	attrs := s.tenantCertifiedAttrs[tenantID]
	var all []map[string]interface{}
	for i := range attrs {
		all = append(all, tenantCertifiedAttrToJSON(&attrs[i]))
	}

	totalCount := len(all)

	if offset > len(all) {
		offset = len(all)
	}
	all = all[offset:]
	if limit < len(all) {
		all = all[:limit]
	}
	if all == nil {
		all = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": all,
		"pagination": map[string]interface{}{
			"offset":     offset,
			"limit":      limit,
			"totalCount": totalCount,
		},
	})
}

// handleAssignTenantCertifiedAttr assigns a certified attribute to a tenant.
func (s *FakeServer) handleAssignTenantCertifiedAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	attrID, ok := parseUUID(w, body.ID)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attr := StoredTenantCertifiedAttr{
		ID:         attrID,
		AssignedAt: time.Now().UTC(),
	}
	s.tenantCertifiedAttrs[tenantID] = append(s.tenantCertifiedAttrs[tenantID], attr)

	writeJSON(w, http.StatusOK, tenantCertifiedAttrToJSON(&attr))
}

// handleRevokeTenantCertifiedAttr revokes a certified attribute from a tenant.
func (s *FakeServer) handleRevokeTenantCertifiedAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}
	attrID, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := s.tenantCertifiedAttrs[tenantID]
	for i := range attrs {
		if attrs[i].ID == attrID {
			now := time.Now().UTC()
			attrs[i].RevokedAt = &now
			s.tenantCertifiedAttrs[tenantID] = attrs
			writeJSON(w, http.StatusOK, tenantCertifiedAttrToJSON(&attrs[i]))
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Certified attribute %s not found on tenant %s", attrID, tenantID))
}

// handleListTenantDeclaredAttrs lists declared attributes for a tenant with pagination.
func (s *FakeServer) handleListTenantDeclaredAttrs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	attrs := s.tenantDeclaredAttrs[tenantID]
	var all []map[string]interface{}
	for i := range attrs {
		all = append(all, tenantDeclaredAttrToJSON(&attrs[i]))
	}

	totalCount := len(all)

	if offset > len(all) {
		offset = len(all)
	}
	all = all[offset:]
	if limit < len(all) {
		all = all[:limit]
	}
	if all == nil {
		all = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": all,
		"pagination": map[string]interface{}{
			"offset":     offset,
			"limit":      limit,
			"totalCount": totalCount,
		},
	})
}

// handleAssignTenantDeclaredAttr assigns a declared attribute to a tenant.
func (s *FakeServer) handleAssignTenantDeclaredAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	var body struct {
		ID           string  `json:"id"`
		DelegationID *string `json:"delegationId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	attrID, ok := parseUUID(w, body.ID)
	if !ok {
		return
	}

	var delegationID *uuid.UUID
	if body.DelegationID != nil {
		did, ok := parseUUID(w, *body.DelegationID)
		if !ok {
			return
		}
		delegationID = &did
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attr := StoredTenantDeclaredAttr{
		ID:           attrID,
		AssignedAt:   time.Now().UTC(),
		DelegationID: delegationID,
	}
	s.tenantDeclaredAttrs[tenantID] = append(s.tenantDeclaredAttrs[tenantID], attr)

	writeJSON(w, http.StatusOK, tenantDeclaredAttrToJSON(&attr))
}

// handleRevokeTenantDeclaredAttr revokes a declared attribute from a tenant.
func (s *FakeServer) handleRevokeTenantDeclaredAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}
	attrID, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := s.tenantDeclaredAttrs[tenantID]
	for i := range attrs {
		if attrs[i].ID == attrID {
			now := time.Now().UTC()
			attrs[i].RevokedAt = &now
			s.tenantDeclaredAttrs[tenantID] = attrs
			writeJSON(w, http.StatusOK, tenantDeclaredAttrToJSON(&attrs[i]))
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Declared attribute %s not found on tenant %s", attrID, tenantID))
}

// handleListTenantVerifiedAttrs lists verified attributes for a tenant with pagination.
func (s *FakeServer) handleListTenantVerifiedAttrs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	attrs := s.tenantVerifiedAttrs[tenantID]
	var all []map[string]interface{}
	for i := range attrs {
		all = append(all, tenantVerifiedAttrToJSON(&attrs[i]))
	}

	totalCount := len(all)

	if offset > len(all) {
		offset = len(all)
	}
	all = all[offset:]
	if limit < len(all) {
		all = all[:limit]
	}
	if all == nil {
		all = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": all,
		"pagination": map[string]interface{}{
			"offset":     offset,
			"limit":      limit,
			"totalCount": totalCount,
		},
	})
}

// handleAssignTenantVerifiedAttr assigns a verified attribute to a tenant.
func (s *FakeServer) handleAssignTenantVerifiedAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}

	var body struct {
		ID             string  `json:"id"`
		AgreementID    string  `json:"agreementId"`
		ExpirationDate *string `json:"expirationDate,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	attrID, ok := parseUUID(w, body.ID)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attr := StoredTenantVerifiedAttr{
		ID:         attrID,
		AssignedAt: time.Now().UTC(),
	}
	s.tenantVerifiedAttrs[tenantID] = append(s.tenantVerifiedAttrs[tenantID], attr)

	writeJSON(w, http.StatusOK, tenantVerifiedAttrToJSON(&attr))
}

// handleRevokeTenantVerifiedAttr revokes a verified attribute from a tenant.
// The agreementId is read from the query parameter, not a path parameter.
func (s *FakeServer) handleRevokeTenantVerifiedAttr(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseUUID(w, r.PathValue("tenantId"))
	if !ok {
		return
	}
	attrID, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	agreementIDStr := r.URL.Query().Get("agreementId")
	if agreementIDStr == "" {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Missing required query parameter: agreementId")
		return
	}
	if _, ok := parseUUID(w, agreementIDStr); !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	attrs := s.tenantVerifiedAttrs[tenantID]
	for i := range attrs {
		if attrs[i].ID == attrID {
			s.tenantVerifiedAttrs[tenantID] = append(attrs[:i], attrs[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]interface{}{})
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Verified attribute %s not found on tenant %s", attrID, tenantID))
}
