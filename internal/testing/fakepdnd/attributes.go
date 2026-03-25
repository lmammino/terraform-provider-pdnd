package fakepdnd

import (
	"net/http"
)

// --- Certified Attributes ---

func (s *FakeServer) handleGetCertifiedAttribute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	s.mu.RLock()
	attr := s.certifiedAttributes[id]
	s.mu.RUnlock()

	if attr == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Certified attribute not found")
		return
	}

	writeJSON(w, http.StatusOK, certifiedAttributeToJSON(attr))
}

func (s *FakeServer) handleListCertifiedAttributes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var all []map[string]interface{}
	for _, attr := range s.certifiedAttributes {
		all = append(all, certifiedAttributeToJSON(attr))
	}
	s.mu.RUnlock()

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

// --- Declared Attributes ---

func (s *FakeServer) handleGetDeclaredAttribute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	s.mu.RLock()
	attr := s.declaredAttributes[id]
	s.mu.RUnlock()

	if attr == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Declared attribute not found")
		return
	}

	writeJSON(w, http.StatusOK, declaredAttributeToJSON(attr))
}

func (s *FakeServer) handleListDeclaredAttributes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var all []map[string]interface{}
	for _, attr := range s.declaredAttributes {
		all = append(all, declaredAttributeToJSON(attr))
	}
	s.mu.RUnlock()

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

// --- Verified Attributes ---

func (s *FakeServer) handleGetVerifiedAttribute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("attributeId"))
	if !ok {
		return
	}

	s.mu.RLock()
	attr := s.verifiedAttributes[id]
	s.mu.RUnlock()

	if attr == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Verified attribute not found")
		return
	}

	writeJSON(w, http.StatusOK, verifiedAttributeToJSON(attr))
}

func (s *FakeServer) handleListVerifiedAttributes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	s.mu.RLock()
	var all []map[string]interface{}
	for _, attr := range s.verifiedAttributes {
		all = append(all, verifiedAttributeToJSON(attr))
	}
	s.mu.RUnlock()

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
