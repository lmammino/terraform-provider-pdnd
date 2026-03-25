package fakepdnd

import (
	"net/http"
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
