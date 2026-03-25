package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (s *FakeServer) makeHandleListDescriptorAttributes(attrType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		esID, ok := parseUUID(w, r.PathValue("eserviceId"))
		if !ok {
			return
		}
		descID, ok := parseUUID(w, r.PathValue("descriptorId"))
		if !ok {
			return
		}

		q := r.URL.Query()
		offset := parseIntDefault(q.Get("offset"), 0)
		limit := parseIntDefault(q.Get("limit"), 50)

		s.mu.RLock()
		store := s.getAttrGroupStore(attrType)
		groups := store[esID][descID]

		// Flatten groups into entries with groupIndex
		var all []map[string]interface{}
		for groupIdx, group := range groups {
			for _, attrID := range group.Attributes {
				entry := map[string]interface{}{
					"attribute":  s.buildAttributeJSON(attrType, attrID),
					"groupIndex": groupIdx,
				}
				all = append(all, entry)
			}
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
}

func (s *FakeServer) makeHandleCreateDescriptorAttributeGroup(attrType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		esID, ok := parseUUID(w, r.PathValue("eserviceId"))
		if !ok {
			return
		}
		descID, ok := parseUUID(w, r.PathValue("descriptorId"))
		if !ok {
			return
		}

		var body struct {
			AttributeIDs []string `json:"attributeIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		// Validate descriptor exists
		descs := s.descriptors[esID]
		if descs == nil || descs[descID] == nil {
			writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
			return
		}

		attrs := make([]uuid.UUID, len(body.AttributeIDs))
		for i, idStr := range body.AttributeIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				writeProblem(w, http.StatusBadRequest, "Bad Request", fmt.Sprintf("Invalid attribute UUID: %s", idStr))
				return
			}
			attrs[i] = id
		}

		store := s.getAttrGroupStore(attrType)
		if store[esID] == nil {
			store[esID] = make(map[uuid.UUID][]StoredDescriptorAttributeGroup)
		}
		store[esID][descID] = append(store[esID][descID], StoredDescriptorAttributeGroup{
			Attributes: attrs,
		})

		// Build response: list of attribute entries in the new group
		var resultAttrs []map[string]interface{}
		groupIndex := len(store[esID][descID]) - 1
		for _, attrID := range attrs {
			resultAttrs = append(resultAttrs, map[string]interface{}{
				"attribute":  s.buildAttributeJSON(attrType, attrID),
				"groupIndex": groupIndex,
			})
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"attributes": resultAttrs,
		})
	}
}

func (s *FakeServer) makeHandleAssignDescriptorAttributesToGroup(attrType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		esID, ok := parseUUID(w, r.PathValue("eserviceId"))
		if !ok {
			return
		}
		descID, ok := parseUUID(w, r.PathValue("descriptorId"))
		if !ok {
			return
		}
		groupIndex, err := strconv.Atoi(r.PathValue("groupIndex"))
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid group index")
			return
		}

		var body struct {
			AttributeIDs []string `json:"attributeIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		store := s.getAttrGroupStore(attrType)
		groups := store[esID][descID]
		if groupIndex < 0 || groupIndex >= len(groups) {
			writeProblem(w, http.StatusNotFound, "Not Found", "Group not found")
			return
		}

		for _, idStr := range body.AttributeIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				writeProblem(w, http.StatusBadRequest, "Bad Request", fmt.Sprintf("Invalid attribute UUID: %s", idStr))
				return
			}
			groups[groupIndex].Attributes = append(groups[groupIndex].Attributes, id)
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{})
	}
}

func (s *FakeServer) makeHandleDeleteDescriptorAttributeFromGroup(attrType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		esID, ok := parseUUID(w, r.PathValue("eserviceId"))
		if !ok {
			return
		}
		descID, ok := parseUUID(w, r.PathValue("descriptorId"))
		if !ok {
			return
		}
		groupIndex, err := strconv.Atoi(r.PathValue("groupIndex"))
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid group index")
			return
		}
		attrID, ok := parseUUID(w, r.PathValue("attributeId"))
		if !ok {
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		store := s.getAttrGroupStore(attrType)
		groups := store[esID][descID]
		if groupIndex < 0 || groupIndex >= len(groups) {
			writeProblem(w, http.StatusNotFound, "Not Found", "Group not found")
			return
		}

		group := &groups[groupIndex]
		found := false
		for i, id := range group.Attributes {
			if id == attrID {
				group.Attributes = append(group.Attributes[:i], group.Attributes[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			writeProblem(w, http.StatusNotFound, "Not Found", "Attribute not found in group")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{})
	}
}

// buildAttributeJSON creates a minimal attribute JSON for responses.
func (s *FakeServer) buildAttributeJSON(attrType string, id uuid.UUID) map[string]interface{} {
	attr := map[string]interface{}{
		"id":          id.String(),
		"name":        "Attribute " + id.String()[:8],
		"description": "Test attribute",
		"createdAt":   "2024-01-01T00:00:00Z",
	}

	switch attrType {
	case "certified":
		if stored := s.certifiedAttributes[id]; stored != nil {
			return certifiedAttributeToJSON(stored)
		}
		attr["code"] = "TEST"
		attr["origin"] = "TEST"
	case "declared":
		if stored := s.declaredAttributes[id]; stored != nil {
			return declaredAttributeToJSON(stored)
		}
	case "verified":
		if stored := s.verifiedAttributes[id]; stored != nil {
			return verifiedAttributeToJSON(stored)
		}
	}

	return attr
}
