package fakepdnd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// clientToJSON converts a StoredClient to a JSON-serializable map.
func clientToJSON(c *StoredClient) map[string]interface{} {
	m := map[string]interface{}{
		"id":         c.ID.String(),
		"consumerId": c.ConsumerID.String(),
		"name":       c.Name,
		"createdAt":  c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if c.Description != "" {
		m["description"] = c.Description
	}
	return m
}

// clientKeyToJWK converts a StoredClientKey to a JSON-serializable JWK map.
func clientKeyToJWK(k *StoredClientKey) map[string]interface{} {
	return map[string]interface{}{
		"kid": k.Kid,
		"kty": k.Kty,
		"alg": k.Alg,
		"use": k.Use,
	}
}

func (s *FakeServer) handleListClients(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)
	nameFilter := q.Get("name")
	consumerIdFilter := q.Get("consumerId")

	var filtered []map[string]interface{}
	for _, c := range s.clients {
		if nameFilter != "" && !strings.Contains(strings.ToLower(c.Name), strings.ToLower(nameFilter)) {
			continue
		}
		if consumerIdFilter != "" && c.ConsumerID.String() != consumerIdFilter {
			continue
		}
		filtered = append(filtered, clientToJSON(c))
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

func (s *FakeServer) handleGetClient(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}

	s.mu.RLock()
	c := s.clients[id]
	s.mu.RUnlock()

	if c == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Client %s not found", id))
		return
	}

	writeJSON(w, http.StatusOK, clientToJSON(c))
}

func (s *FakeServer) handleListClientKeys(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query()
	offset := parseIntDefault(q.Get("offset"), 0)
	limit := parseIntDefault(q.Get("limit"), 50)

	keys := s.clientKeys[clientID]
	var all []map[string]interface{}
	for i := range keys {
		all = append(all, clientKeyToJWK(&keys[i]))
	}

	totalCount := len(all)

	// Apply pagination.
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

func (s *FakeServer) handleCreateClientKey(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}

	var body struct {
		Key  string `json:"key"`
		Use  string `json:"use"`
		Alg  string `json:"alg"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clients[clientID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Client %s not found", clientID))
		return
	}

	kid := uuid.New().String()
	key := StoredClientKey{
		Kid:  kid,
		Kty:  "RSA",
		Alg:  body.Alg,
		Use:  body.Use,
		Name: body.Name,
		Key:  body.Key,
	}
	s.clientKeys[clientID] = append(s.clientKeys[clientID], key)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"clientId": clientID.String(),
		"jwk":      clientKeyToJWK(&key),
	})
}

func (s *FakeServer) handleDeleteClientKey(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}
	keyID := r.PathValue("keyId")

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := s.clientKeys[clientID]
	for i := range keys {
		if keys[i].Kid == keyID {
			s.clientKeys[clientID] = append(keys[:i], keys[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]interface{}{})
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Key %s not found", keyID))
}

func (s *FakeServer) handleAddClientPurpose(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}

	var body struct {
		PurposeID string `json:"purposeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON body")
		return
	}

	purposeID, ok := parseUUID(w, body.PurposeID)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clients[clientID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Client %s not found", clientID))
		return
	}

	if s.clientPurposes[clientID] == nil {
		s.clientPurposes[clientID] = make(map[uuid.UUID]bool)
	}
	s.clientPurposes[clientID][purposeID] = true

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (s *FakeServer) handleRemoveClientPurpose(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseUUID(w, r.PathValue("clientId"))
	if !ok {
		return
	}
	purposeID, ok := parseUUID(w, r.PathValue("purposeId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clientPurposes[clientID], purposeID)

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
