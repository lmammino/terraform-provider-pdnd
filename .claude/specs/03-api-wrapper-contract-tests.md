# Agent 3: API Wrapper + Contract Tests

## Objective

Build a thin `AgreementsAPI` interface and concrete implementation over the generated client. Write contract tests for every API method. When done, `go test ./internal/client/api/... -v` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo already has:
- Generated client in `internal/client/generated/client.gen.go`
- Transport layer in `internal/client/` (dpop.go, transport.go, retry.go, errors.go)

## Design Principles

1. Resource/data source code MUST NOT import `internal/client/generated` directly
2. The `AgreementsAPI` interface is the seam — it's mockable for unit tests
3. Domain types are plain Go structs (not generated types)
4. The concrete implementation converts between domain and generated types

## Files to Create

### 1. `internal/client/api/types.go` — Domain Types

```go
package api

import (
    "time"
    "github.com/google/uuid"
)

// Agreement represents a PDND agreement with all fields from the API.
type Agreement struct {
    ID                  uuid.UUID
    EServiceID          uuid.UUID
    DescriptorID        uuid.UUID
    ProducerID          uuid.UUID
    ConsumerID          uuid.UUID
    DelegationID        *uuid.UUID
    State               string   // One of: DRAFT, ACTIVE, ARCHIVED, PENDING, SUSPENDED, MISSING_CERTIFIED_ATTRIBUTES, REJECTED
    SuspendedByConsumer *bool
    SuspendedByProducer *bool
    SuspendedByPlatform *bool
    ConsumerNotes       *string
    RejectionReason     *string
    CreatedAt           time.Time
    UpdatedAt           *time.Time
    SuspendedAt         *time.Time
}

// AgreementSeed contains fields for creating a new agreement.
// Maps to POST /agreements request body.
type AgreementSeed struct {
    EServiceID   uuid.UUID   // required
    DescriptorID uuid.UUID   // required
    DelegationID *uuid.UUID  // optional
}

// AgreementSubmission contains fields for submitting a draft agreement.
// Maps to POST /agreements/{id}/submit request body.
type AgreementSubmission struct {
    ConsumerNotes *string  // optional, max 1000 chars
}

// AgreementRejection contains fields for rejecting a pending agreement.
// Maps to POST /agreements/{id}/reject request body.
type AgreementRejection struct {
    Reason string  // required, 20-1000 chars
}

// DelegationRef identifies a delegation for delegate operations.
// Used by approve, suspend, unsuspend endpoints.
type DelegationRef struct {
    DelegationID uuid.UUID  // required
}

// ListAgreementsParams contains filter parameters for listing agreements.
type ListAgreementsParams struct {
    States        []string     // AgreementState values to filter by
    ProducerIDs   []uuid.UUID
    ConsumerIDs   []uuid.UUID
    DescriptorIDs []uuid.UUID
    EServiceIDs   []uuid.UUID
    Offset        int32        // required, min 0
    Limit         int32        // required, 1-50
}

// PaginationParams for paginated list endpoints.
type PaginationParams struct {
    Offset int32
    Limit  int32
}

// AgreementsPage is a paginated list of agreements.
type AgreementsPage struct {
    Results    []Agreement
    Pagination Pagination
}

// PurposesPage is a paginated list of purposes.
type PurposesPage struct {
    Results    []Purpose
    Pagination Pagination
}

// Pagination contains pagination metadata from list responses.
type Pagination struct {
    Offset     int32
    Limit      int32
    TotalCount int32
}

// Purpose represents a PDND purpose associated with an agreement.
type Purpose struct {
    ID                  uuid.UUID
    EServiceID          uuid.UUID
    ConsumerID          uuid.UUID
    SuspendedByConsumer *bool
    SuspendedByProducer *bool
    Title               string
    Description         string
    CreatedAt           time.Time
    UpdatedAt           *time.Time
    IsRiskAnalysisValid bool
    IsFreeOfCharge      bool
    FreeOfChargeReason  *string
    DelegationID        *uuid.UUID
}
```

### 2. `internal/client/api/agreements.go` — Interface + Implementation

```go
package api

import (
    "context"
    "github.com/google/uuid"
)

// AgreementsAPI defines operations on PDND agreements.
// This interface is the boundary between Terraform resource code and the HTTP client.
type AgreementsAPI interface {
    CreateAgreement(ctx context.Context, seed AgreementSeed) (*Agreement, error)
    GetAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
    ListAgreements(ctx context.Context, params ListAgreementsParams) (*AgreementsPage, error)
    DeleteAgreement(ctx context.Context, id uuid.UUID) error
    SubmitAgreement(ctx context.Context, id uuid.UUID, payload AgreementSubmission) (*Agreement, error)
    ApproveAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    RejectAgreement(ctx context.Context, id uuid.UUID, payload AgreementRejection) (*Agreement, error)
    SuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    UnsuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    UpgradeAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
    CloneAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
    ListAgreementPurposes(ctx context.Context, id uuid.UUID, params PaginationParams) (*PurposesPage, error)
}
```

**Concrete implementation**: `agreementsClient` struct holding `*generated.ClientWithResponses`.

```go
type agreementsClient struct {
    client *generated.ClientWithResponses
}

func NewAgreementsClient(client *generated.ClientWithResponses) AgreementsAPI {
    return &agreementsClient{client: client}
}
```

Each method must:
1. Convert domain types → generated request types
2. Call the generated client method (e.g., `client.CreateAgreementWithResponse(ctx, body)`)
3. Check for HTTP errors using `CheckResponse` from `internal/client/errors.go`
4. Convert generated response types → domain types
5. Return domain types or error

**API endpoint mapping** (from OpenAPI spec):

| Interface Method | Generated Method | HTTP | Path | Request Body | Success Code |
|-----------------|-----------------|------|------|-------------|-------------|
| CreateAgreement | `CreateAgreementWithResponse` | POST | `/agreements` | AgreementSeed JSON | 201 |
| GetAgreement | `GetAgreementWithResponse` | GET | `/agreements/{agreementId}` | none | 200 |
| ListAgreements | `GetAgreementsWithResponse` | GET | `/agreements` | query params | 200 |
| DeleteAgreement | `DeleteAgreementWithResponse` | DELETE | `/agreements/{agreementId}` | none | 200 |
| SubmitAgreement | `SubmitAgreementWithResponse` | POST | `/agreements/{agreementId}/submit` | AgreementSubmission JSON | 200 |
| ApproveAgreement | `ApproveAgreementWithResponse` | POST | `/agreements/{agreementId}/approve` | DelegationRef JSON (optional body) | 200 |
| RejectAgreement | `RejectAgreementWithResponse` | POST | `/agreements/{agreementId}/reject` | AgreementRejection JSON | 200 |
| SuspendAgreement | `SuspendAgreementWithResponse` | POST | `/agreements/{agreementId}/suspend` | DelegationRef JSON (optional body) | 200 |
| UnsuspendAgreement | `UnsuspendAgreementWithResponse` | POST | `/agreements/{agreementId}/unsuspend` | DelegationRef JSON (optional body) | 200 |
| UpgradeAgreement | `UpgradeAgreementWithResponse` | POST | `/agreements/{agreementId}/upgrade` | none | 200 |
| CloneAgreement | `CloneAgreementWithResponse` | POST | `/agreements/{agreementId}/clone` | none | 200 |
| ListAgreementPurposes | `GetAgreementPurposesWithResponse` | GET | `/agreements/{agreementId}/purposes` | query: offset, limit | 200 |

**Query param notes for ListAgreements:**
- `states`, `producerIds`, `consumerIds`, `descriptorIds`, `eserviceIds` are arrays with `explode: false` (comma-separated)
- `offset` and `limit` are required int32 query params
- Check the generated client to see the exact parameter types — they may use pointer types or custom array types

**Important**: Read the generated code in `internal/client/generated/client.gen.go` to understand the exact function signatures, parameter types, and response types. The generated types will have specific names and structures that you need to adapt to.

### 3. `internal/client/api/conversions.go` — Type Conversion Helpers

Helper functions to convert between generated types and domain types:

```go
func agreementFromGenerated(g *generated.Agreement) *Agreement { ... }
func agreementSeedToGenerated(s AgreementSeed) generated.AgreementSeed { ... }
func purposeFromGenerated(g *generated.Purpose) *Purpose { ... }
// etc.
```

Handle nullable/optional fields carefully — the generated code may use pointers or `*openapi_types.UUID`.

### 4. `internal/client/api/agreements_test.go` — Contract Tests

Each test:
1. Starts an `httptest.Server` that records the incoming request
2. Returns a canned JSON response
3. Creates a `generated.ClientWithResponses` pointing at the test server
4. Wraps it with `NewAgreementsClient()`
5. Calls the method
6. Asserts: correct HTTP method, correct path, correct request body (if any)
7. Asserts: response correctly deserialized into domain types

**Required Contract Tests:**

#### TestCreateAgreement_Contract
- Request: POST to `/agreements`
- Body: `{"eserviceId":"<uuid>","descriptorId":"<uuid>"}`
- Response: 201, Agreement JSON with state="DRAFT"
- Assert: method=POST, path="/agreements", body fields match, result.State=="DRAFT"

#### TestGetAgreement_Contract
- Request: GET to `/agreements/{uuid}`
- Response: 200, Agreement JSON
- Assert: method=GET, path contains UUID, result fields populated

#### TestDeleteAgreement_Contract
- Request: DELETE to `/agreements/{uuid}`
- Response: 200, `{}`
- Assert: method=DELETE, path contains UUID, no error returned

#### TestListAgreements_Contract
- Request: GET to `/agreements` with query params
- Response: 200, `{"results":[...],"pagination":{"offset":0,"limit":50,"totalCount":1}}`
- Assert: method=GET, query contains `offset=0&limit=50`, results parsed correctly

#### TestSubmitAgreement_Contract
- Request: POST to `/agreements/{uuid}/submit`
- Body: `{"consumerNotes":"test notes"}`
- Response: 200, Agreement JSON with state="ACTIVE"
- Assert: method=POST, path ends with `/submit`, body has consumerNotes

#### TestApproveAgreement_Contract
- Request: POST to `/agreements/{uuid}/approve`
- Body: `{"delegationId":"<uuid>"}` or empty body when nil
- Response: 200, Agreement JSON with state="ACTIVE"
- Assert: method=POST, path ends with `/approve`

#### TestRejectAgreement_Contract
- Request: POST to `/agreements/{uuid}/reject`
- Body: `{"reason":"This agreement does not meet our requirements"}`
- Response: 200, Agreement JSON with state="REJECTED"
- Assert: method=POST, path ends with `/reject`, body has reason

#### TestSuspendAgreement_Contract
- Request: POST to `/agreements/{uuid}/suspend`
- Response: 200, Agreement JSON with state="SUSPENDED"
- Assert: method=POST, path ends with `/suspend`

#### TestUnsuspendAgreement_Contract
- Request: POST to `/agreements/{uuid}/unsuspend`
- Response: 200, Agreement JSON with state="ACTIVE"
- Assert: method=POST, path ends with `/unsuspend`

#### TestUpgradeAgreement_Contract
- Request: POST to `/agreements/{uuid}/upgrade`
- Response: 200, Agreement JSON (new agreement with new ID)
- Assert: method=POST, path ends with `/upgrade`, no body sent

#### TestCloneAgreement_Contract
- Request: POST to `/agreements/{uuid}/clone`
- Response: 200, Agreement JSON with state="DRAFT" and new ID
- Assert: method=POST, path ends with `/clone`

#### TestListAgreementPurposes_Contract
- Request: GET to `/agreements/{uuid}/purposes?offset=0&limit=50`
- Response: 200, `{"results":[purpose objects],"pagination":{...}}`
- Assert: method=GET, path contains `/purposes`, query has offset/limit

**Canned Agreement JSON for tests:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "eserviceId": "660e8400-e29b-41d4-a716-446655440001",
  "descriptorId": "770e8400-e29b-41d4-a716-446655440002",
  "producerId": "880e8400-e29b-41d4-a716-446655440003",
  "consumerId": "990e8400-e29b-41d4-a716-446655440004",
  "state": "DRAFT",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

**Canned Purpose JSON:**
```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440005",
  "eserviceId": "660e8400-e29b-41d4-a716-446655440001",
  "consumerId": "990e8400-e29b-41d4-a716-446655440004",
  "title": "Test Purpose",
  "description": "A test purpose description",
  "createdAt": "2024-01-15T10:30:00Z",
  "isRiskAnalysisValid": true,
  "isFreeOfCharge": false
}
```

## Execution Steps

1. **Read the generated client code** in `internal/client/generated/client.gen.go` to understand exact function signatures, type names, and parameter structures
2. Read `internal/client/errors.go` to understand the `CheckResponse` function
3. Create `internal/client/api/types.go`
4. Create `internal/client/api/agreements.go` with interface and implementation
5. Create `internal/client/api/conversions.go` with type conversion helpers
6. Create `internal/client/api/agreements_test.go` with all contract tests
7. Run `go mod tidy`
8. Run `go test ./internal/client/api/... -v`
9. Fix failures and iterate

## Verification

- `go test ./internal/client/api/... -v` — ALL contract tests pass
- `go build ./...` — compiles successfully

## Important Notes

- Read the generated code FIRST — the exact type names may differ from what I've described
- The generated client uses `*generated.ClientWithResponses` — read its method signatures
- Query params for ListAgreements may be typed as `*GetAgreementsParams` — check the generated code
- For optional request bodies (approve, suspend, unsuspend with optional DelegationRef), check how the generated client handles nil bodies
- UUID fields in generated types may use `openapi_types.UUID` or `string` — handle conversion
- Timestamp fields may be `time.Time` or `string` — handle accordingly
- Do NOT modify files in `internal/client/` (transport layer) or `internal/client/generated/`
