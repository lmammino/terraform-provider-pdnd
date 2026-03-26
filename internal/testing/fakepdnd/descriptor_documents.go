package fakepdnd

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// detectContentType infers MIME type from file extension, falling back to the provided default.
func detectContentType(filename, fallback string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return "application/yaml"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	default:
		return fallback
	}
}

// documentToJSON converts a StoredDocument to a JSON-serializable map.
func documentToJSON(d *StoredDocument) map[string]interface{} {
	return map[string]interface{}{
		"id":          d.ID.String(),
		"name":        d.Name,
		"prettyName":  d.PrettyName,
		"contentType": d.ContentType,
		"createdAt":   d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func (s *FakeServer) handleListDocuments(w http.ResponseWriter, r *http.Request) {
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
	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		s.mu.RUnlock()
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	docs := s.descDocuments[esID][descID]
	var all []map[string]interface{}
	for i := range docs {
		all = append(all, documentToJSON(&docs[i]))
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

func (s *FakeServer) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Missing file field")
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to read file")
		return
	}

	prettyName := r.FormValue("prettyName")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	ct := detectContentType(header.Filename, header.Header.Get("Content-Type"))

	doc := StoredDocument{
		ID:          uuid.New(),
		Name:        header.Filename,
		PrettyName:  prettyName,
		ContentType: ct,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
	}

	if s.descDocuments[esID] == nil {
		s.descDocuments[esID] = make(map[uuid.UUID][]StoredDocument)
	}
	s.descDocuments[esID][descID] = append(s.descDocuments[esID][descID], doc)

	writeJSON(w, http.StatusCreated, documentToJSON(&doc))
}

func (s *FakeServer) handleDownloadDocument(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}
	docID, ok := parseUUID(w, r.PathValue("documentId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	docs := s.descDocuments[esID][descID]
	for i := range docs {
		if docs[i].ID == docID {
			_, _ = w.Write(docs[i].Content)
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", "Document not found")
}

func (s *FakeServer) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}
	docID, ok := parseUUID(w, r.PathValue("documentId"))
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	docs := s.descDocuments[esID][descID]
	for i := range docs {
		if docs[i].ID == docID {
			s.descDocuments[esID][descID] = append(docs[:i], docs[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]interface{}{})
			return
		}
	}

	writeProblem(w, http.StatusNotFound, "Not Found", "Document not found")
}

func (s *FakeServer) handleUploadInterface(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Missing file field")
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to read file")
		return
	}

	prettyName := r.FormValue("prettyName")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	ct2 := detectContentType(header.Filename, header.Header.Get("Content-Type"))

	doc := StoredDocument{
		ID:          uuid.New(),
		Name:        header.Filename,
		PrettyName:  prettyName,
		ContentType: ct2,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
	}

	if s.descInterfaces[esID] == nil {
		s.descInterfaces[esID] = make(map[uuid.UUID]*StoredDocument)
	}
	s.descInterfaces[esID][descID] = &doc

	writeJSON(w, http.StatusCreated, documentToJSON(&doc))
}

func (s *FakeServer) handleDownloadInterface(w http.ResponseWriter, r *http.Request) {
	esID, ok := parseUUID(w, r.PathValue("eserviceId"))
	if !ok {
		return
	}
	descID, ok := parseUUID(w, r.PathValue("descriptorId"))
	if !ok {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	iface := s.descInterfaces[esID][descID]
	if iface == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Interface not found")
		return
	}

	_, _ = w.Write(iface.Content)
}

func (s *FakeServer) handleDeleteInterface(w http.ResponseWriter, r *http.Request) {
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

	// Validate descriptor exists.
	descs := s.descriptors[esID]
	if descs == nil || descs[descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Descriptor not found")
		return
	}

	if s.descInterfaces[esID] == nil || s.descInterfaces[esID][descID] == nil {
		writeProblem(w, http.StatusNotFound, "Not Found", "Interface not found")
		return
	}

	s.descInterfaces[esID][descID] = nil

	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
