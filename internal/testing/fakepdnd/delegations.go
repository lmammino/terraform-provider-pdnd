package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// delegationToJSON converts a StoredDelegation to a JSON-serializable map.
func delegationToJSON(d *StoredDelegation) map[string]interface{} {
	m := map[string]interface{}{
		"id":          d.ID.String(),
		"delegatorId": d.DelegatorID.String(),
		"delegateId":  d.DelegateID.String(),
		"eserviceId":  d.EServiceID.String(),
		"state":       d.State,
		"createdAt":   d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"submittedAt": d.SubmittedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if d.UpdatedAt != nil {
		m["updatedAt"] = d.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.ActivatedAt != nil {
		m["activatedAt"] = d.ActivatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.RejectedAt != nil {
		m["rejectedAt"] = d.RejectedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.RevokedAt != nil {
		m["revokedAt"] = d.RevokedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if d.RejectionReason != nil {
		m["rejectionReason"] = *d.RejectionReason
	}
	return m
}

func (s *FakeServer) makeHandleCreateDelegation(delegationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			EServiceID string `json:"eserviceId"`
			DelegateID string `json:"delegateId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
			return
		}

		esID, ok := parseUUID(w, body.EServiceID)
		if !ok {
			return
		}
		delegateID, ok := parseUUID(w, body.DelegateID)
		if !ok {
			return
		}

		var delegatorID uuid.UUID
		if delegationType == "consumer" {
			delegatorID = s.consumerID
		} else {
			delegatorID = s.producerID
		}

		now := time.Now().UTC()
		d := &StoredDelegation{
			ID:          uuid.New(),
			DelegatorID: delegatorID,
			DelegateID:  delegateID,
			EServiceID:  esID,
			State:       "WAITING_FOR_APPROVAL",
			CreatedAt:   now,
			SubmittedAt: now,
		}

		s.mu.Lock()
		s.getDelegationStore(delegationType)[d.ID] = d
		s.mu.Unlock()

		writeJSON(w, http.StatusCreated, delegationToJSON(d))
	}
}

func (s *FakeServer) makeHandleGetDelegation(delegationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUUID(w, r.PathValue("delegationId"))
		if !ok {
			return
		}

		s.mu.RLock()
		d := s.getDelegationStore(delegationType)[id]
		s.mu.RUnlock()

		if d == nil {
			writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Delegation %s not found", id))
			return
		}

		writeJSON(w, http.StatusOK, delegationToJSON(d))
	}
}

func (s *FakeServer) makeHandleListDelegations(delegationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		q := r.URL.Query()
		offset := parseIntDefault(q.Get("offset"), 0)
		limit := parseIntDefault(q.Get("limit"), 50)

		states := q["states"]
		delegatorIDs := q["delegatorIds"]
		delegateIDs := q["delegateIds"]
		eserviceIDs := q["eserviceIds"]

		// Build set helpers for filtering.
		stateSet := make(map[string]bool, len(states))
		for _, st := range states {
			stateSet[st] = true
		}
		delegatorSet := make(map[string]bool, len(delegatorIDs))
		for _, id := range delegatorIDs {
			delegatorSet[id] = true
		}
		delegateSet := make(map[string]bool, len(delegateIDs))
		for _, id := range delegateIDs {
			delegateSet[id] = true
		}
		eserviceSet := make(map[string]bool, len(eserviceIDs))
		for _, id := range eserviceIDs {
			eserviceSet[id] = true
		}

		store := s.getDelegationStore(delegationType)
		var filtered []map[string]interface{}
		for _, d := range store {
			if len(stateSet) > 0 && !stateSet[d.State] {
				continue
			}
			if len(delegatorSet) > 0 && !delegatorSet[d.DelegatorID.String()] {
				continue
			}
			if len(delegateSet) > 0 && !delegateSet[d.DelegateID.String()] {
				continue
			}
			if len(eserviceSet) > 0 && !eserviceSet[d.EServiceID.String()] {
				continue
			}
			filtered = append(filtered, delegationToJSON(d))
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
}

func (s *FakeServer) makeHandleAcceptDelegation(delegationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUUID(w, r.PathValue("delegationId"))
		if !ok {
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		d := s.getDelegationStore(delegationType)[id]
		if d == nil {
			writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Delegation %s not found", id))
			return
		}

		if d.State != "WAITING_FOR_APPROVAL" {
			writeProblem(w, http.StatusConflict, "Conflict",
				fmt.Sprintf("Cannot accept delegation in %s state", d.State))
			return
		}

		now := time.Now().UTC()
		d.State = "ACTIVE"
		d.ActivatedAt = &now
		d.UpdatedAt = &now

		writeJSON(w, http.StatusOK, delegationToJSON(d))
	}
}

func (s *FakeServer) makeHandleRejectDelegation(delegationType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseUUID(w, r.PathValue("delegationId"))
		if !ok {
			return
		}

		var body struct {
			RejectionReason string `json:"rejectionReason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		d := s.getDelegationStore(delegationType)[id]
		if d == nil {
			writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Delegation %s not found", id))
			return
		}

		if d.State != "WAITING_FOR_APPROVAL" {
			writeProblem(w, http.StatusConflict, "Conflict",
				fmt.Sprintf("Cannot reject delegation in %s state", d.State))
			return
		}

		now := time.Now().UTC()
		d.State = "REJECTED"
		d.RejectedAt = &now
		d.RejectionReason = &body.RejectionReason
		d.UpdatedAt = &now

		writeJSON(w, http.StatusOK, delegationToJSON(d))
	}
}
