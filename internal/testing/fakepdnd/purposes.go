package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *FakeServer) handleGetAgreementPurposes(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("agreementId"))
	if !ok {
		return
	}

	s.mu.RLock()
	a := s.agreements[id]
	if a == nil {
		s.mu.RUnlock()
		writeProblem(w, http.StatusNotFound, "Not Found", "Agreement not found")
		return
	}

	purposes := s.purposes[id]
	s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	totalCount := len(purposes)

	// Convert to JSON.
	var results []map[string]interface{}
	for i := range purposes {
		results = append(results, purposeToJSON(&purposes[i]))
	}

	// Apply pagination.
	if offset > len(results) {
		offset = len(results)
	}
	results = results[offset:]
	if limit < len(results) {
		results = results[:limit]
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"pagination": map[string]interface{}{
			"offset":     offset,
			"limit":      limit,
			"totalCount": totalCount,
		},
	})
}

func (s *FakeServer) handleCreatePurpose(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EServiceID         string  `json:"eserviceId"`
		Title              string  `json:"title"`
		Description        string  `json:"description"`
		DailyCalls         int32   `json:"dailyCalls"`
		IsFreeOfCharge     bool    `json:"isFreeOfCharge"`
		FreeOfChargeReason *string `json:"freeOfChargeReason"`
		DelegationID       *string `json:"delegationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	purposeID := uuid.New()
	versionID := uuid.New()

	p := &StoredPurpose{
		ID:                  purposeID,
		EServiceID:          uuid.MustParse(body.EServiceID),
		ConsumerID:          s.consumerID,
		Title:               body.Title,
		Description:         body.Description,
		IsFreeOfCharge:      body.IsFreeOfCharge,
		IsRiskAnalysisValid: true,
		CreatedAt:           now,
		Versions: []StoredPurposeVersion{
			{
				ID:        versionID,
				State:     "DRAFT",
				DailyCalls: body.DailyCalls,
				CreatedAt: now,
			},
		},
	}

	if body.FreeOfChargeReason != nil {
		p.FreeOfChargeReason = *body.FreeOfChargeReason
	}
	if body.DelegationID != nil {
		id := uuid.MustParse(*body.DelegationID)
		p.DelegationID = &id
	}

	s.standalonePurposes[purposeID] = p

	writeJSON(w, http.StatusCreated, fullPurposeToJSON(p))
}

func (s *FakeServer) handleGetPurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.RLock()
	p := s.standalonePurposes[id]
	s.mu.RUnlock()

	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleDeletePurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	state := deriveStoredPurposeState(p)
	if state != "DRAFT" && state != "WAITING_FOR_APPROVAL" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot delete purpose in %s state", state))
		return
	}

	delete(s.standalonePurposes, id)
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (s *FakeServer) handleUpdateDraftPurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	if deriveStoredPurposeState(p) != "DRAFT" {
		writeProblem(w, http.StatusConflict, "Conflict", "Cannot update non-draft purpose")
		return
	}

	if v, ok := body["title"].(string); ok {
		p.Title = v
	}
	if v, ok := body["description"].(string); ok {
		p.Description = v
	}
	if v, ok := body["dailyCalls"].(float64); ok {
		if len(p.Versions) > 0 {
			p.Versions[len(p.Versions)-1].DailyCalls = int32(v)
		}
	}
	if v, ok := body["isFreeOfCharge"].(bool); ok {
		p.IsFreeOfCharge = v
	}
	if v, exists := body["freeOfChargeReason"]; exists {
		if v == nil {
			p.FreeOfChargeReason = ""
		} else if s, ok := v.(string); ok {
			p.FreeOfChargeReason = s
		}
	}

	now := time.Now().UTC()
	p.UpdatedAt = &now

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleActivatePurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	// Find the draft version
	draftIdx := -1
	for i := range p.Versions {
		if p.Versions[i].State == "DRAFT" {
			draftIdx = i
			break
		}
	}
	if draftIdx == -1 {
		writeProblem(w, http.StatusConflict, "Conflict", "No draft version to activate")
		return
	}

	now := time.Now().UTC()
	if p.Versions[draftIdx].DailyCalls > s.purposeApprovalThreshold {
		p.Versions[draftIdx].State = "WAITING_FOR_APPROVAL"
	} else {
		p.Versions[draftIdx].State = "ACTIVE"
		p.Versions[draftIdx].FirstActivationAt = &now
	}
	p.Versions[draftIdx].UpdatedAt = &now
	p.UpdatedAt = &now

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleApprovePurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	wfaIdx := -1
	for i := range p.Versions {
		if p.Versions[i].State == "WAITING_FOR_APPROVAL" {
			wfaIdx = i
			break
		}
	}
	if wfaIdx == -1 {
		writeProblem(w, http.StatusConflict, "Conflict", "No version waiting for approval")
		return
	}

	now := time.Now().UTC()
	p.Versions[wfaIdx].State = "ACTIVE"
	p.Versions[wfaIdx].FirstActivationAt = &now
	p.Versions[wfaIdx].UpdatedAt = &now
	p.UpdatedAt = &now

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleSuspendPurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	activeIdx := -1
	for i := range p.Versions {
		if p.Versions[i].State == "ACTIVE" || p.Versions[i].State == "SUSPENDED" {
			activeIdx = i
			break
		}
	}
	if activeIdx == -1 {
		writeProblem(w, http.StatusConflict, "Conflict", "No active version to suspend")
		return
	}

	now := time.Now().UTC()
	p.Versions[activeIdx].State = "SUSPENDED"
	p.Versions[activeIdx].SuspendedAt = &now
	p.Versions[activeIdx].UpdatedAt = &now
	p.SuspendedByConsumer = true
	p.UpdatedAt = &now

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleUnsuspendPurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	suspIdx := -1
	for i := range p.Versions {
		if p.Versions[i].State == "SUSPENDED" {
			suspIdx = i
			break
		}
	}
	if suspIdx == -1 {
		writeProblem(w, http.StatusConflict, "Conflict", "No suspended version to unsuspend")
		return
	}

	now := time.Now().UTC()
	p.Versions[suspIdx].State = "ACTIVE"
	p.Versions[suspIdx].SuspendedAt = nil
	p.Versions[suspIdx].UpdatedAt = &now
	p.SuspendedByConsumer = false
	p.UpdatedAt = &now

	writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
}

func (s *FakeServer) handleArchivePurpose(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	for i := range p.Versions {
		if p.Versions[i].State == "ACTIVE" || p.Versions[i].State == "SUSPENDED" {
			now := time.Now().UTC()
			p.Versions[i].State = "ARCHIVED"
			p.Versions[i].UpdatedAt = &now
			p.UpdatedAt = &now
			writeJSON(w, http.StatusOK, fullPurposeToJSON(p))
			return
		}
	}

	writeProblem(w, http.StatusConflict, "Conflict", "No active or suspended version to archive")
}

func (s *FakeServer) handleCreatePurposeVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	var body struct {
		DailyCalls int32 `json:"dailyCalls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.standalonePurposes[id]
	if p == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Purpose not found")
		return
	}

	now := time.Now().UTC()
	versionID := uuid.New()

	newState := "ACTIVE"
	var firstActivation *time.Time
	if body.DailyCalls > s.purposeApprovalThreshold {
		newState = "WAITING_FOR_APPROVAL"
	} else {
		firstActivation = &now
	}

	version := StoredPurposeVersion{
		ID:                versionID,
		State:             newState,
		DailyCalls:        body.DailyCalls,
		CreatedAt:         now,
		FirstActivationAt: firstActivation,
	}

	p.Versions = append(p.Versions, version)
	p.UpdatedAt = &now

	writeJSON(w, http.StatusCreated, purposeVersionToJSON(&version))
}

// deriveStoredPurposeState derives the effective state from stored purpose versions.
func deriveStoredPurposeState(p *StoredPurpose) string {
	for i := len(p.Versions) - 1; i >= 0; i-- {
		switch p.Versions[i].State {
		case "ACTIVE", "SUSPENDED", "ARCHIVED":
			return p.Versions[i].State
		case "WAITING_FOR_APPROVAL":
			return "WAITING_FOR_APPROVAL"
		}
	}
	return "DRAFT"
}
