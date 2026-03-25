package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (s *FakeServer) handleCreateDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	var body struct {
		Audience                []string `json:"audience"`
		VoucherLifespan         int32    `json:"voucherLifespan"`
		DailyCallsPerConsumer   int32    `json:"dailyCallsPerConsumer"`
		DailyCallsTotal         int32    `json:"dailyCallsTotal"`
		AgreementApprovalPolicy string   `json:"agreementApprovalPolicy"`
		Description             string   `json:"description"`
		ServerUrls              []string `json:"serverUrls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.eservices[esID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "EService not found")
		return
	}

	// Check no existing DRAFT descriptor.
	for _, desc := range s.descriptors[esID] {
		if desc.State == "DRAFT" {
			writeProblem(w, http.StatusConflict, "Conflict", "A DRAFT descriptor already exists")
			return
		}
	}

	s.descriptorCount[esID]++
	version := strconv.Itoa(s.descriptorCount[esID])

	now := time.Now().UTC()
	desc := &StoredDescriptor{
		ID:                      uuid.New(),
		EServiceID:              esID,
		Version:                 version,
		State:                   "DRAFT",
		AgreementApprovalPolicy: body.AgreementApprovalPolicy,
		Audience:                body.Audience,
		DailyCallsPerConsumer:   body.DailyCallsPerConsumer,
		DailyCallsTotal:         body.DailyCallsTotal,
		VoucherLifespan:         body.VoucherLifespan,
		ServerUrls:              body.ServerUrls,
		Description:             body.Description,
		CreatedAt:               now,
	}

	s.descriptors[esID][desc.ID] = desc

	writeJSON(w, http.StatusCreated, descriptorToJSON(desc))
}

func (s *FakeServer) handleGetDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.RLock()
	descs := s.descriptors[esID]
	var desc *StoredDescriptor
	if descs != nil {
		desc = descs[descID]
	}
	s.mu.RUnlock()

	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handleListDescriptors(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}

	q := r.URL.Query()
	stateFilter := q.Get("state")
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var filtered []map[string]interface{}
	for _, desc := range s.descriptors[esID] {
		if stateFilter != "" && desc.State != stateFilter {
			continue
		}
		filtered = append(filtered, descriptorToJSON(desc))
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

func (s *FakeServer) handleDeleteDraftDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "DRAFT" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot delete descriptor in %s state", desc.State))
		return
	}

	// Check not last descriptor.
	if len(descs) <= 1 {
		writeProblem(w, http.StatusConflict, "Conflict", "Cannot delete the last descriptor")
		return
	}

	delete(descs, descID)

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (s *FakeServer) handleUpdateDraftDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
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

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "DRAFT" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot update descriptor in %s state", desc.State))
		return
	}

	// Merge-patch: update only provided fields.
	if v, ok := body["description"].(string); ok {
		desc.Description = v
	}
	if v, ok := body["agreementApprovalPolicy"].(string); ok {
		desc.AgreementApprovalPolicy = v
	}
	if v, ok := body["audience"]; ok {
		if arr, ok := v.([]interface{}); ok {
			strs := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			desc.Audience = strs
		}
	}
	if v, ok := body["dailyCallsPerConsumer"].(float64); ok {
		desc.DailyCallsPerConsumer = int32(v)
	}
	if v, ok := body["dailyCallsTotal"].(float64); ok {
		desc.DailyCallsTotal = int32(v)
	}
	if v, ok := body["voucherLifespan"].(float64); ok {
		desc.VoucherLifespan = int32(v)
	}
	if v, ok := body["serverUrls"]; ok {
		if arr, ok := v.([]interface{}); ok {
			strs := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			desc.ServerUrls = strs
		}
	}

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handleUpdatePublishedDescriptorQuotas(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	var body struct {
		DailyCallsPerConsumer *int32 `json:"dailyCallsPerConsumer"`
		DailyCallsTotal       *int32 `json:"dailyCallsTotal"`
		VoucherLifespan       *int32 `json:"voucherLifespan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "PUBLISHED" && desc.State != "SUSPENDED" && desc.State != "DEPRECATED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot update quotas for descriptor in %s state", desc.State))
		return
	}

	if body.DailyCallsPerConsumer != nil {
		desc.DailyCallsPerConsumer = *body.DailyCallsPerConsumer
	}
	if body.DailyCallsTotal != nil {
		desc.DailyCallsTotal = *body.DailyCallsTotal
	}
	if body.VoucherLifespan != nil {
		desc.VoucherLifespan = *body.VoucherLifespan
	}

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handlePublishDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "DRAFT" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot publish descriptor in %s state", desc.State))
		return
	}

	now := time.Now().UTC()

	// Auto-deprecate any previously PUBLISHED descriptor for this eservice.
	for _, d := range descs {
		if d.ID != descID && d.State == "PUBLISHED" {
			d.State = "DEPRECATED"
			d.DeprecatedAt = &now
		}
	}

	desc.State = "PUBLISHED"
	desc.PublishedAt = &now

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handleSuspendDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "PUBLISHED" && desc.State != "DEPRECATED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot suspend descriptor in %s state", desc.State))
		return
	}

	now := time.Now().UTC()
	desc.State = "SUSPENDED"
	desc.SuspendedAt = &now

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handleUnsuspendDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "SUSPENDED" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot unsuspend descriptor in %s state", desc.State))
		return
	}

	desc.State = "PUBLISHED"
	desc.SuspendedAt = nil

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}

func (s *FakeServer) handleApproveDelegatedDescriptor(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	descs := s.descriptors[esID]
	if descs == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}
	desc := descs[descID]
	if desc == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if desc.State != "WAITING_FOR_APPROVAL" {
		writeProblem(w, http.StatusConflict, "Conflict",
			fmt.Sprintf("Cannot approve descriptor in %s state", desc.State))
		return
	}

	now := time.Now().UTC()
	desc.State = "PUBLISHED"
	desc.PublishedAt = &now

	writeJSON(w, http.StatusOK, descriptorToJSON(desc))
}
