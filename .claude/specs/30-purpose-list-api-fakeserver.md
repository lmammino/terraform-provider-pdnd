# Spec 30: Add ListPurposes to API Client + Fake Server Handler

## Objective

Add `ListPurposes` method to the `PurposesAPI` interface and implementation, a `GET /purposes` handler to the fake server, and a contract test. When done, `go test ./internal/client/api/... -v` and `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/purposes.go` — existing PurposesAPI interface (add ListPurposes to it)
2. `internal/client/api/types.go` — existing types. Has `PurposesPage` (line ~207), `Purpose` (line ~278), `PurposeVersion` (line ~270). Add `ListPurposesParams`.
3. `internal/client/api/agreements.go` — `ListAgreements` implementation pattern (lines 82-141). Follow this exactly for ListPurposes.
4. `internal/client/api/conversions.go` — `purposeFromGenerated` (line ~97). Already handles version fields.
5. `internal/client/api/purposes_test.go` — existing contract tests. Add `TestListPurposes_Contract`.
6. `internal/client/generated/client.gen.go`:
   - `GetPurposesParams` (line 3208) — fields: EserviceIds (`*[]openapi_types.UUID`), Title (`*string`), ConsumerIds (`*[]openapi_types.UUID`), States (`*[]PurposeVersionState`), Offset, Limit
   - `GetPurposesWithResponse` (line 26817) — returns `*GetPurposesResponse`
   - `GetPurposesResponse` — has `JSON200 *Purposes`
   - `Purposes` struct — has `Results []Purpose` and `Pagination Pagination`
   - `PurposeVersionState` — enum: ACTIVE, ARCHIVED, DRAFT, REJECTED, SUSPENDED, WAITING_FOR_APPROVAL
7. `internal/testing/fakepdnd/purposes.go` — existing purpose handlers. Add `handleListPurposes`.
8. `internal/testing/fakepdnd/server.go` — route registration (add `GET /purposes`), `standalonePurposes` map, `fullPurposeToJSON` helper in `helpers.go`.
9. `internal/testing/fakepdnd/helpers.go` — `fullPurposeToJSON`, `deriveStoredPurposeState` (in purposes.go)

## Files to Modify

### 1. `internal/client/api/types.go` — Add ListPurposesParams

Add after the existing `PurposeVersionSeed` type:

```go
// ListPurposesParams contains filter parameters for listing purposes.
type ListPurposesParams struct {
    EServiceIDs []uuid.UUID
    Title       *string
    ConsumerIDs []uuid.UUID
    States      []string
    Offset      int32
    Limit       int32
}
```

### 2. `internal/client/api/purposes.go` — Add ListPurposes to interface + implementation

Add to the `PurposesAPI` interface:
```go
ListPurposes(ctx context.Context, params ListPurposesParams) (*PurposesPage, error)
```

Add implementation following the `ListAgreements` pattern:
```go
func (p *purposesClient) ListPurposes(ctx context.Context, params ListPurposesParams) (*PurposesPage, error) {
    genParams := &generated.GetPurposesParams{
        Offset: params.Offset,
        Limit:  params.Limit,
    }
    if len(params.EServiceIDs) > 0 {
        ids := uuidsToOpenAPI(params.EServiceIDs)
        genParams.EserviceIds = &ids
    }
    if params.Title != nil {
        genParams.Title = params.Title
    }
    if len(params.ConsumerIDs) > 0 {
        ids := uuidsToOpenAPI(params.ConsumerIDs)
        genParams.ConsumerIds = &ids
    }
    if len(params.States) > 0 {
        states := make([]generated.PurposeVersionState, len(params.States))
        for i, s := range params.States {
            states[i] = generated.PurposeVersionState(s)
        }
        genParams.States = &states
    }

    resp, err := p.client.GetPurposesWithResponse(ctx, genParams)
    // ... CheckResponse, JSON200 nil check ...
    // Convert results using purposeFromGenerated
    // Return &PurposesPage{Results, Pagination}
}
```

### 3. `internal/client/api/purposes_test.go` — Add contract test

Add `TestListPurposes_Contract`:
- Verify GET method, path `/purposes`, query params (offset, limit, optional filters)
- Use a canned response with `Results` containing one purpose with `currentVersion`
- Verify parsed results

Canned response:
```json
{
  "results": [{
    "id": "990e8400-e29b-41d4-a716-446655440000",
    "eserviceId": "550e8400-e29b-41d4-a716-446655440000",
    "consumerId": "880e8400-e29b-41d4-a716-446655440003",
    "title": "Test Purpose",
    "description": "A test purpose",
    "createdAt": "2024-01-01T00:00:00Z",
    "isRiskAnalysisValid": true,
    "isFreeOfCharge": false,
    "currentVersion": {
      "id": "aa0e8400-e29b-41d4-a716-446655440010",
      "state": "ACTIVE",
      "dailyCalls": 1000,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  }],
  "pagination": {"offset": 0, "limit": 50, "totalCount": 1}
}
```

### 4. `internal/testing/fakepdnd/purposes.go` — Add handleListPurposes

Add handler that:
1. Reads query params: offset, limit, eserviceIds (comma-separated or repeated), title, consumerIds, states
2. Iterates `s.standalonePurposes`, applies filters
3. For `states` filter: derive state using `deriveStoredPurposeState(p)` and check match
4. For `title` filter: case-insensitive contains match
5. For `eserviceIds`/`consumerIds`: check membership
6. Paginates and returns `{"results": [...], "pagination": {...}}`
7. Use `fullPurposeToJSON` for each result

Note: Query params for array values in the generated client use repeated params (e.g., `?eserviceIds=uuid1&eserviceIds=uuid2`). Parse with `r.URL.Query()["eserviceIds"]`.

### 5. `internal/testing/fakepdnd/server.go` — Add route

Add before the `GET /purposes/{purposeId}` route:
```go
s.mux.HandleFunc("GET /purposes", s.handleListPurposes)
```

**IMPORTANT**: This must be registered BEFORE `GET /purposes/{purposeId}` to avoid the more specific pattern matching first. Actually in Go 1.22 ServeMux, `GET /purposes` and `GET /purposes/{purposeId}` are distinct patterns and won't conflict. But ensure the order is correct.

## Execution Steps

1. Read all referenced files
2. Add `ListPurposesParams` to `types.go`
3. Add `ListPurposes` to interface and implementation in `purposes.go`
4. Add `TestListPurposes_Contract` to `purposes_test.go`
5. Add `handleListPurposes` to `purposes.go` (fakepdnd)
6. Add route to `server.go`
7. Run `go test ./internal/client/api/... -v`
8. Run `go build ./...`

## Verification

```shell
go test ./internal/client/api/... -v
go test ./internal/testing/fakepdnd/... -v
go build ./...
```

## Important Notes

- Do NOT break existing tests
- Variable names in test files must not conflict with other test files in the same package
- The `purposeFromGenerated` function already handles `CurrentVersion`, `WaitingForApprovalVersion`, `RejectedVersion` fields
- `PurposesPage` type already exists in `types.go`
- `uuidsToOpenAPI` helper already exists in `conversions.go`
