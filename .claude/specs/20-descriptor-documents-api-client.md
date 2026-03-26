# Spec 20: API Client for Descriptor Documents/Interface

## Objective

Add `DescriptorDocumentsAPI` interface and implementation to the API wrapper layer, with contract tests. When done, `go test ./internal/client/api/... -v` must pass all existing + new tests.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/descriptor_attributes.go` — closest pattern (descriptor sub-resource interface + implementation with switch dispatch)
2. `internal/client/api/types.go` — existing domain types, add `DescriptorDocument` here
3. `internal/client/api/conversions.go` — existing conversions, add `documentFromGenerated` here
4. `internal/client/api/descriptor_attributes_test.go` — contract test pattern for descriptor sub-resources
5. `internal/client/api/agreements_test.go` — `newTestClient` helper (lines 60-68)
6. `internal/client/generated/client.gen.go` — generated client. Find these:
   - `type Document struct` (line 1231) — 5 fields: Id, Name, PrettyName, ContentType, CreatedAt
   - `type Documents struct` (line 1249) — Results []Document + Pagination
   - `GetEServiceDescriptorDocumentsParams` (line 3001) — Offset, Limit int32
   - `UploadEServiceDescriptorDocumentWithBodyWithResponse` (line 26114) — takes contentType string + body io.Reader
   - `DeleteEServiceDescriptorDocumentWithResponse` (line 26123) — takes documentId openapi_types.UUID
   - `DownloadEServiceDescriptorDocumentWithResponse` (line 26132) — returns raw Body
   - `UploadEServiceDescriptorInterfaceWithBodyWithResponse` (line 26159) — takes contentType + body io.Reader
   - `DeleteEServiceDescriptorInterfaceWithResponse` (line 26141)
   - `DownloadEServiceDescriptorInterfaceWithResponse` (line 26150)
   - Response types: Upload returns `JSON201 *Document`, List returns `JSON200 *Documents`, Delete returns `JSON200 *VoidObject`, Download returns raw `Body []byte`
7. `internal/client/errors.go` — `CheckResponse`, `IsNotFound`

## Files to Create/Modify

### 1. Extend `internal/client/api/types.go` — Add Document domain type

Add after the `DescriptorAttributeEntry` type:

```go
// DescriptorDocument represents a file (document or interface) on a descriptor.
type DescriptorDocument struct {
    ID          uuid.UUID
    Name        string
    PrettyName  string
    ContentType string
    CreatedAt   time.Time
}
```

### 2. Extend `internal/client/api/conversions.go` — Add conversion function

Add before the `uuidsToOpenAPI` function:

```go
// documentFromGenerated converts a generated Document to a domain DescriptorDocument.
func documentFromGenerated(g *generated.Document) *DescriptorDocument {
    if g == nil {
        return nil
    }
    return &DescriptorDocument{
        ID:          uuid.UUID(g.Id),
        Name:        g.Name,
        PrettyName:  g.PrettyName,
        ContentType: g.ContentType,
        CreatedAt:   g.CreatedAt,
    }
}
```

### 3. Create `internal/client/api/descriptor_documents.go` — Interface + Implementation

```go
// DescriptorDocumentsAPI defines operations on descriptor documents and interfaces.
type DescriptorDocumentsAPI interface {
    // Documents (multiple per descriptor)
    ListDocuments(ctx context.Context, eserviceID, descriptorID uuid.UUID, offset, limit int32) ([]DescriptorDocument, Pagination, error)
    UploadDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error)
    DeleteDocument(ctx context.Context, eserviceID, descriptorID uuid.UUID, documentID uuid.UUID) error
    GetDocumentByID(ctx context.Context, eserviceID, descriptorID, documentID uuid.UUID) (*DescriptorDocument, error)

    // Interface (singular, one per descriptor)
    UploadInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID, fileName string, fileContent []byte, prettyName string) (*DescriptorDocument, error)
    DeleteInterface(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
    CheckInterfaceExists(ctx context.Context, eserviceID, descriptorID uuid.UUID) (bool, error)
}
```

Implementation struct: `descriptorDocumentsClient` with `*generated.ClientWithResponses`.

Constructor: `NewDescriptorDocumentsClient(c *generated.ClientWithResponses) DescriptorDocumentsAPI`

**Key implementation details:**

- **`ListDocuments`**: Call `GetEServiceDescriptorDocumentsWithResponse`, convert results using `documentFromGenerated`
- **`UploadDocument`**: Build multipart body using helper, call `UploadEServiceDescriptorDocumentWithBodyWithResponse` with the multipart content type and body reader
- **`DeleteDocument`**: Call `DeleteEServiceDescriptorDocumentWithResponse`
- **`GetDocumentByID`**: Paginate through `ListDocuments` looking for a document with matching ID. Return `nil` + not-found-style error if not found. Use the `client.IsNotFound` error pattern (or return a wrapped error that `client.IsNotFound` can match).
- **`UploadInterface`**: Build multipart body, call `UploadEServiceDescriptorInterfaceWithBodyWithResponse`
- **`DeleteInterface`**: Call `DeleteEServiceDescriptorInterfaceWithResponse`
- **`CheckInterfaceExists`**: Call `DownloadEServiceDescriptorInterfaceWithResponse`. If HTTP status is 200, return true. If 404, return false. For other errors, return the error.

**Multipart body helper** (private function in same file):

```go
func buildMultipartBody(fileName string, fileContent []byte, prettyName string) (io.Reader, string, error) {
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)

    part, err := writer.CreateFormFile("file", fileName)
    if err != nil {
        return nil, "", err
    }
    if _, err := part.Write(fileContent); err != nil {
        return nil, "", err
    }
    if err := writer.WriteField("prettyName", prettyName); err != nil {
        return nil, "", err
    }
    if err := writer.Close(); err != nil {
        return nil, "", err
    }

    return &buf, writer.FormDataContentType(), nil
}
```

**IMPORTANT**: The interface endpoint uses `descriptorId openapi_types.UUID` directly (NOT the `DescriptorId` alias). Check the generated signatures carefully — the document endpoints use `DescriptorId` type while the interface endpoints use `openapi_types.UUID` for the descriptorId parameter.

### 4. Create `internal/client/api/descriptor_documents_test.go` — Contract Tests

Follow the parameterized test pattern from `descriptor_attributes_test.go`. Use UUIDs defined at package level to avoid redeclaration conflicts.

Canned Document JSON:
```json
{
  "id": "dd0e8400-e29b-41d4-a716-446655440020",
  "name": "openapi.yaml",
  "prettyName": "API Specification",
  "contentType": "application/yaml",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

| Test | Verifies |
|------|----------|
| `TestListDocuments_Contract` | GET path, offset/limit params, response parsing |
| `TestUploadDocument_Contract` | POST path, multipart body has file + prettyName fields, 201 response |
| `TestDeleteDocument_Contract` | DELETE path with documentId |
| `TestUploadInterface_Contract` | POST interface path, multipart body |
| `TestDeleteInterface_Contract` | DELETE interface path |
| `TestGetDocumentByID_Contract` | Lists and finds matching document by ID |
| `TestCheckInterfaceExists_Contract` | Returns true for 200, false for 404 |

For upload tests, verify multipart content in the handler by reading the request body. Don't use `r.ParseMultipartForm` in the test handler — just check the Content-Type header starts with `multipart/form-data` and that the body is non-empty, then return the canned response.

## Execution Steps

1. Read generated client to confirm exact method signatures
2. Add `DescriptorDocument` to `types.go`
3. Add `documentFromGenerated` to `conversions.go`
4. Create `descriptor_documents.go` with interface + implementation + multipart helper
5. Create `descriptor_documents_test.go` with all contract tests
6. Run `go test ./internal/client/api/... -v`
7. Run `go build ./...`

## Verification

```shell
go test ./internal/client/api/... -v
go build ./...
```

## Important Notes

- Do NOT break existing tests
- Variable names in test files must not conflict with other test files in the same package (e.g., don't redeclare `testEServiceID` if it exists in `descriptor_attributes_test.go`)
- The generated `UploadEServiceDescriptorDocumentWithBodyWithResponse` takes `contentType string` and `body io.Reader` — pass the multipart content type WITH boundary from `writer.FormDataContentType()`
- Download endpoints return raw `Body []byte`, not JSON — handle accordingly for `CheckInterfaceExists`
- For `GetDocumentByID`, iterate through pages (offset+limit) until found or all pages exhausted. When not found, return an error that `client.IsNotFound()` recognizes — wrap with the same status-code-based error pattern used elsewhere
