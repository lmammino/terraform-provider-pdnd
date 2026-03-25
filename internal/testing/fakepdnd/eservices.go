package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *FakeServer) handleCreateEService(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Technology  string `json:"technology"`
		Mode        string `json:"mode"`
		Descriptor  *struct {
			Audience                []string `json:"audience"`
			VoucherLifespan         int32    `json:"voucherLifespan"`
			DailyCallsPerConsumer   int32    `json:"dailyCallsPerConsumer"`
			DailyCallsTotal         int32    `json:"dailyCallsTotal"`
			AgreementApprovalPolicy string   `json:"agreementApprovalPolicy"`
			Description             string   `json:"description"`
			ServerUrls              []string `json:"serverUrls"`
		} `json:"descriptor,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	now := time.Now().UTC()
	es := &StoredEService{
		ID:          uuid.New(),
		ProducerID:  s.producerID,
		Name:        body.Name,
		Description: body.Description,
		Technology:  body.Technology,
		Mode:        body.Mode,
	}

	s.mu.Lock()
	s.eservices[es.ID] = es
	s.descriptors[es.ID] = make(map[uuid.UUID]*StoredDescriptor)
	s.descriptorCount[es.ID] = 1

	// Create initial descriptor in DRAFT state.
	desc := &StoredDescriptor{
		ID:         uuid.New(),
		EServiceID: es.ID,
		Version:    "1",
		State:      "DRAFT",
		CreatedAt:  now,
	}
	if body.Descriptor != nil {
		desc.Audience = body.Descriptor.Audience
		desc.VoucherLifespan = body.Descriptor.VoucherLifespan
		desc.DailyCallsPerConsumer = body.Descriptor.DailyCallsPerConsumer
		desc.DailyCallsTotal = body.Descriptor.DailyCallsTotal
		desc.AgreementApprovalPolicy = body.Descriptor.AgreementApprovalPolicy
		desc.Description = body.Descriptor.Description
		desc.ServerUrls = body.Descriptor.ServerUrls
	}
	s.descriptors[es.ID][desc.ID] = desc
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, eserviceToJSON(es))
}

func (s *FakeServer) handleGetEService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	s.mu.RLock()
	es := s.eservices[id]
	s.mu.RUnlock()

	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

func (s *FakeServer) handleListEServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	producerFilter := parseUUIDList(q.Get("producerIds"))
	nameFilter := q.Get("name")
	technologyFilter := q.Get("technology")
	modeFilter := q.Get("mode")

	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var filtered []map[string]interface{}
	for _, es := range s.eservices {
		if len(producerFilter) > 0 && !containsUUID(producerFilter, es.ProducerID) {
			continue
		}
		if nameFilter != "" && es.Name != nameFilter {
			continue
		}
		if technologyFilter != "" && es.Technology != technologyFilter {
			continue
		}
		if modeFilter != "" && es.Mode != modeFilter {
			continue
		}
		filtered = append(filtered, eserviceToJSON(es))
	}
	s.mu.RUnlock()

	totalCount := len(filtered)

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

func (s *FakeServer) handleDeleteEService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	// Check all descriptors are DRAFT.
	for _, desc := range s.descriptors[id] {
		if desc.State != "DRAFT" {
			writeProblem(w, http.StatusConflict, "Conflict",
				fmt.Sprintf("Cannot delete eservice: descriptor %s is in %s state", desc.ID, desc.State))
			return
		}
	}

	delete(s.eservices, id)
	delete(s.descriptors, id)
	delete(s.descriptorCount, id)

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (s *FakeServer) handleUpdateDraftEService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
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

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	// Check all descriptors are DRAFT.
	for _, desc := range s.descriptors[id] {
		if desc.State != "DRAFT" {
			writeProblem(w, http.StatusConflict, "Conflict",
				"Cannot update eservice with non-DRAFT descriptors")
			return
		}
	}

	// Merge-patch: update only provided fields.
	if v, ok := body["name"].(string); ok {
		es.Name = v
	}
	if v, ok := body["description"].(string); ok {
		es.Description = v
	}
	if v, ok := body["technology"].(string); ok {
		es.Technology = v
	}
	if v, ok := body["mode"].(string); ok {
		es.Mode = v
	}

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

func (s *FakeServer) handleUpdatePublishedEServiceName(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	es.Name = body.Name

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

func (s *FakeServer) handleUpdatePublishedEServiceDescription(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	var body struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	es.Description = body.Description

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

func (s *FakeServer) handleUpdatePublishedEServiceDelegation(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	var body struct {
		IsConsumerDelegable     *bool `json:"isConsumerDelegable"`
		IsClientAccessDelegable *bool `json:"isClientAccessDelegable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	if body.IsConsumerDelegable != nil {
		es.IsConsumerDelegable = *body.IsConsumerDelegable
	}
	if body.IsClientAccessDelegable != nil {
		es.IsClientAccessDelegable = *body.IsClientAccessDelegable
	}

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

func (s *FakeServer) handleUpdatePublishedEServiceSignalHub(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	var body struct {
		IsSignalHubEnabled *bool `json:"isSignalHubEnabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	es := s.eservices[id]
	if es == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	if body.IsSignalHubEnabled != nil {
		es.IsSignalHubEnabled = *body.IsSignalHubEnabled
	}

	writeJSON(w, http.StatusOK, eserviceToJSON(es))
}

