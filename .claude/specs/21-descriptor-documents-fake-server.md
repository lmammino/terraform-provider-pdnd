# Spec 21: Fake Server for Descriptor Documents/Interface

## Objective

Extend the fake PDND server with document and interface endpoints for descriptors. When done, the fake server compiles and all existing tests pass. New handlers support upload, list, download, and delete for both documents and interfaces.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/testing/fakepdnd/server.go` — FakeServer struct (line 13), NewFakeServer (line 34), setupRoutes (line 228+), existing Seed/Get methods
2. `internal/testing/fakepdnd/state.go` — storage types (StoredDescriptor, StoredDescriptorAttributeGroup, StoredPurposeVersion, etc.)
3. `internal/testing/fakepdnd/descriptor_attributes.go` — handler pattern for descriptor sub-resources (multipart-type handlers, storage by eservice→descriptor)
4. `internal/testing/fakepdnd/helpers.go` — `writeJSON`, `writeProblem`, `parseUUID`, `descriptorToJSON`, `documentToJSON` (does NOT exist yet)
5. `internal/testing/fakepdnd/agreements.go` — `parseIntDefault` helper (line 472)
6. `internal/testing/fakepdnd/descriptors.go` — descriptor handlers (for reference on validating eservice/descriptor exists)

## Files to Create/Modify

### 1. Extend `internal/testing/fakepdnd/state.go` — Add StoredDocument

Add after `StoredDescriptorAttributeGroup`:

```go
// StoredDocument represents a document or interface file in the fake server's state.
type StoredDocument struct {
    ID          uuid.UUID
    Name        string  // filename
    PrettyName  string
    ContentType string
    Content     []byte  // actual file bytes
    CreatedAt   time.Time
}
```

### 2. Extend `internal/testing/fakepdnd/server.go`

**Add to FakeServer struct** (after `descVerifiedAttrGroups`):

```go
// descriptor documents: eserviceID -> descriptorID -> []StoredDocument
descDocuments  map[uuid.UUID]map[uuid.UUID][]StoredDocument
// descriptor interface: eserviceID -> descriptorID -> *StoredDocument (nil = no interface)
descInterfaces map[uuid.UUID]map[uuid.UUID]*StoredDocument
```

**Add to NewFakeServer** initialization:

```go
descDocuments:  make(map[uuid.UUID]map[uuid.UUID][]StoredDocument),
descInterfaces: make(map[uuid.UUID]map[uuid.UUID]*StoredDocument),
```

**Add seed/getter methods** (before `setupRoutes`):

```go
// SeedDocument pre-populates a document for a descriptor.
func (s *FakeServer) SeedDocument(eserviceID, descriptorID uuid.UUID, doc StoredDocument) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.descDocuments[eserviceID] == nil {
        s.descDocuments[eserviceID] = make(map[uuid.UUID][]StoredDocument)
    }
    s.descDocuments[eserviceID][descriptorID] = append(s.descDocuments[eserviceID][descriptorID], doc)
}

// GetDocuments returns stored documents for test assertions.
func (s *FakeServer) GetDocuments(eserviceID, descriptorID uuid.UUID) []StoredDocument {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.descDocuments[eserviceID] == nil {
        return nil
    }
    return s.descDocuments[eserviceID][descriptorID]
}

// SeedInterface pre-populates the interface for a descriptor.
func (s *FakeServer) SeedInterface(eserviceID, descriptorID uuid.UUID, doc StoredDocument) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.descInterfaces[eserviceID] == nil {
        s.descInterfaces[eserviceID] = make(map[uuid.UUID]*StoredDocument)
    }
    s.descInterfaces[eserviceID][descriptorID] = &doc
}

// GetInterface returns the stored interface for test assertions.
func (s *FakeServer) GetInterface(eserviceID, descriptorID uuid.UUID) *StoredDocument {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.descInterfaces[eserviceID] == nil {
        return nil
    }
    return s.descInterfaces[eserviceID][descriptorID]
}
```

**Add routes in setupRoutes** (before the `// Descriptor attribute group routes` comment):

```go
// Descriptor document routes.
s.mux.HandleFunc("GET /eservices/{eserviceId}/descriptors/{descriptorId}/documents", s.handleListDocuments)
s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/documents", s.handleUploadDocument)
s.mux.HandleFunc("GET /eservices/{eserviceId}/descriptors/{descriptorId}/documents/{documentId}", s.handleDownloadDocument)
s.mux.HandleFunc("DELETE /eservices/{eserviceId}/descriptors/{descriptorId}/documents/{documentId}", s.handleDeleteDocument)

// Descriptor interface routes.
s.mux.HandleFunc("POST /eservices/{eserviceId}/descriptors/{descriptorId}/interface", s.handleUploadInterface)
s.mux.HandleFunc("GET /eservices/{eserviceId}/descriptors/{descriptorId}/interface", s.handleDownloadInterface)
s.mux.HandleFunc("DELETE /eservices/{eserviceId}/descriptors/{descriptorId}/interface", s.handleDeleteInterface)
```

### 3. Create `internal/testing/fakepdnd/descriptor_documents.go` — Handlers

7 handlers + 1 JSON helper:

**`handleListDocuments`**: Parse eserviceId/descriptorId, read-lock, get documents, paginate, respond with `{"results": [...], "pagination": {...}}`.

**`handleUploadDocument`**: Parse eserviceId/descriptorId, parse multipart form (`r.ParseMultipartForm(10 << 20)`), extract file via `r.FormFile("file")` and prettyName via `r.FormValue("prettyName")`, read file content, create StoredDocument with new UUID, store it, respond 201 with document JSON.

**`handleDownloadDocument`**: Parse all three IDs, find document in slice by documentId, write the file content bytes. If not found, 404.

**`handleDeleteDocument`**: Parse all three IDs, find and remove document from slice. If not found, 404. Respond 200 with `{}`.

**`handleUploadInterface`**: Same multipart parsing as document. Store in `descInterfaces`. Respond 201 with document JSON.

**`handleDownloadInterface`**: If interface exists, write content bytes. If not, 404.

**`handleDeleteInterface`**: Remove interface (set to nil). If not found, 404. Respond 200 with `{}`.

**`documentToJSON`** helper:
```go
func documentToJSON(d *StoredDocument) map[string]interface{} {
    return map[string]interface{}{
        "id":          d.ID.String(),
        "name":        d.Name,
        "prettyName":  d.PrettyName,
        "contentType": d.ContentType,
        "createdAt":   d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
    }
}
```

## Execution Steps

1. Read existing handler patterns thoroughly
2. Add `StoredDocument` to `state.go`
3. Add state fields, initialization, seed/get methods, routes to `server.go`
4. Create `descriptor_documents.go` with all 7 handlers + `documentToJSON`
5. Run `go build ./internal/testing/fakepdnd/...`
6. Run `go test ./internal/testing/fakepdnd/... -v`
7. Run `go build ./...`

## Verification

```shell
go build ./internal/testing/fakepdnd/...
go test ./internal/testing/fakepdnd/... -v
go build ./...
```

## Important Notes

- Do NOT break existing tests or handlers
- For multipart parsing in upload handlers: `r.ParseMultipartForm(10 << 20)` then `r.FormFile("file")` and `r.FormValue("prettyName")`
- The download response in the real API is multipart/form-data, but the generated client reads raw `Body []byte`. For the fake server, just write the file content bytes with `w.Write(doc.Content)` — don't build multipart response
- Always validate that eservice and descriptor exist before operations: check `s.descriptors[esID][descID] != nil`
- Use `s.mu.Lock()` for write operations, `s.mu.RLock()` for read operations
- `parseIntDefault` is defined in `agreements.go` (line 472) and available in the package
- `parseUUID` is defined in `helpers.go` and available in the package
