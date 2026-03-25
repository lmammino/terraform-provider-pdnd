package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *FakeServer) handleCreateAgreement(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EServiceID   string  `json:"eserviceId"`
		DescriptorID string  `json:"descriptorId"`
		DelegationID *string `json:"delegationId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	esID, ok := parseUUID(w, body.EServiceID)
	if !ok {
		return
	}
	descID, ok := parseUUID(w, body.DescriptorID)
	if !ok {
		return
	}

	var delegationID *uuid.UUID
	if body.DelegationID != nil {
		id, ok := parseUUID(w, *body.DelegationID)
		if !ok {
			return
		}
		delegationID = &id
	}

	now := time.Now().UTC()
	a := &StoredAgreement{
		ID:           uuid.New(),
		EServiceID:   esID,
		DescriptorID: descID,
		ProducerID:   s.producerID,
		ConsumerID:   s.consumerID,
		DelegationID: delegationID,
		State:        "DRAFT",
		CreatedAt:    now,
	}

	s.mu.Lock()
	s.agreements[a.ID] = a
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, agreementToJSON(a))
}

func (s *FakeServer) handleGetAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	s.mu.RLock()
	a := s.agreements[id]
	s.mu.RUnlock()

	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleListAgreements(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse filters.
	statesFilter := parseCSV(q.Get("states"))
	producerFilter := parseUUIDList(q.Get("producerIds"))
	consumerFilter := parseUUIDList(q.Get("consumerIds"))
	descriptorFilter := parseUUIDList(q.Get("descriptorIds"))
	eserviceFilter := parseUUIDList(q.Get("eserviceIds"))

	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var filtered []map[string]interface{}
	for _, a := range s.agreements {
		if len(statesFilter) > 0 && !containsString(statesFilter, a.State) {
			continue
		}
		if len(producerFilter) > 0 && !containsUUID(producerFilter, a.ProducerID) {
			continue
		}
		if len(consumerFilter) > 0 && !containsUUID(consumerFilter, a.ConsumerID) {
			continue
		}
		if len(descriptorFilter) > 0 && !containsUUID(descriptorFilter, a.DescriptorID) {
			continue
		}
		if len(eserviceFilter) > 0 && !containsUUID(eserviceFilter, a.EServiceID) {
			continue
		}
		filtered = append(filtered, agreementToJSON(a))
	}
	s.mu.RUnlock()

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

func (s *FakeServer) handleDeleteAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	switch a.State {
	case "DRAFT", "PENDING", "MISSING_CERTIFIED_ATTRIBUTES":
		delete(s.agreements, id)
		writeJSON(w, http.StatusOK, map[string]interface{}{})
	default:
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot delete agreement in %s state", a.State))
	}
}

func (s *FakeServer) handleSubmitAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	var body struct {
		ConsumerNotes *string `json:"consumerNotes,omitempty"`
	}
	// Body is optional; ignore decode errors for empty body.
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "DRAFT" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot submit agreement in %s state", a.State))
		return
	}

	if body.ConsumerNotes != nil {
		a.ConsumerNotes = *body.ConsumerNotes
	}

	now := time.Now().UTC()
	a.UpdatedAt = &now

	if s.approvalPolicy == "AUTOMATIC" {
		a.State = "ACTIVE"
	} else {
		a.State = "PENDING"
	}

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleApproveAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	// Parse optional body.
	var body struct {
		DelegationID *string `json:"delegationId,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "PENDING" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot approve agreement in %s state", a.State))
		return
	}

	a.State = "ACTIVE"
	now := time.Now().UTC()
	a.UpdatedAt = &now

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleRejectAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	if len(body.Reason) < 20 || len(body.Reason) > 1000 {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Reason must be between 20 and 1000 characters")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "PENDING" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot reject agreement in %s state", a.State))
		return
	}

	a.State = "REJECTED"
	a.RejectionReason = body.Reason
	now := time.Now().UTC()
	a.UpdatedAt = &now

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleSuspendAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	// Parse optional body.
	var body struct {
		DelegationID *string `json:"delegationId,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "ACTIVE" && a.State != "SUSPENDED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot suspend agreement in %s state", a.State))
		return
	}

	a.State = "SUSPENDED"
	a.SuspendedByConsumer = true
	now := time.Now().UTC()
	a.SuspendedAt = &now
	a.UpdatedAt = &now

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleUnsuspendAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	// Parse optional body.
	var body struct {
		DelegationID *string `json:"delegationId,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "SUSPENDED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot unsuspend agreement in %s state", a.State))
		return
	}

	a.SuspendedByConsumer = false
	now := time.Now().UTC()
	a.UpdatedAt = &now

	if !a.SuspendedByProducer && !a.SuspendedByPlatform {
		a.State = "ACTIVE"
		a.SuspendedAt = nil
	}

	writeJSON(w, http.StatusOK, agreementToJSON(a))
}

func (s *FakeServer) handleUpgradeAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "ACTIVE" && a.State != "SUSPENDED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot upgrade agreement in %s state", a.State))
		return
	}

	originalState := a.State

	// Archive the current agreement.
	a.State = "ARCHIVED"
	archiveTime := time.Now().UTC()
	a.UpdatedAt = &archiveTime

	// Create new agreement.
	now := time.Now().UTC()
	newAgreement := &StoredAgreement{
		ID:                  uuid.New(),
		EServiceID:          a.EServiceID,
		DescriptorID:        uuid.New(), // new descriptor version
		ProducerID:          a.ProducerID,
		ConsumerID:          a.ConsumerID,
		DelegationID:        a.DelegationID,
		State:               originalState,
		SuspendedByConsumer: a.SuspendedByConsumer,
		SuspendedByProducer: a.SuspendedByProducer,
		SuspendedByPlatform: a.SuspendedByPlatform,
		CreatedAt:           now,
	}
	if originalState == "SUSPENDED" {
		newAgreement.SuspendedAt = a.SuspendedAt
	}

	s.agreements[newAgreement.ID] = newAgreement

	writeJSON(w, http.StatusOK, agreementToJSON(newAgreement))
}

func (s *FakeServer) handleCloneAgreement(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a := s.agreements[id]
	if a == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	if a.State != "REJECTED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot clone agreement in %s state", a.State))
		return
	}

	now := time.Now().UTC()
	newAgreement := &StoredAgreement{
		ID:           uuid.New(),
		EServiceID:   a.EServiceID,
		DescriptorID: a.DescriptorID,
		ProducerID:   a.ProducerID,
		ConsumerID:   a.ConsumerID,
		State:        "DRAFT",
		CreatedAt:    now,
	}

	s.agreements[newAgreement.ID] = newAgreement

	writeJSON(w, http.StatusOK, agreementToJSON(newAgreement))
}

// Helper functions for query parsing.

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func parseUUIDList(s string) []uuid.UUID {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []uuid.UUID
	for _, p := range parts {
		id, err := uuid.Parse(strings.TrimSpace(p))
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func containsUUID(haystack []uuid.UUID, needle uuid.UUID) bool {
	for _, id := range haystack {
		if id == needle {
			return true
		}
	}
	return false
}
