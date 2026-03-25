# Agent 4: Fake PDND Server

## Objective

Build a deterministic, stateful, in-process fake PDND server for acceptance testing. When done, `go test ./internal/testing/fakepdnd/... -v` must pass with all smoke tests.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo already has:
- Generated client in `internal/client/generated/`
- Transport layer in `internal/client/`
- API wrapper in `internal/client/api/` with domain types

## Design

The fake server is an `httptest.Server` that:
- Stores agreements and purposes in memory
- Implements all agreement endpoints matching the OpenAPI spec
- Validates state transitions (returns 409 Conflict for invalid ones)
- Returns JSON responses matching the OpenAPI schema exactly
- Supports configurable approval policy (AUTOMATIC vs MANUAL)
- Supports seeding data for specific test scenarios

## Files to Create

### 1. `internal/testing/fakepdnd/server.go` — Server Setup

```go
package fakepdnd

import (
    "net/http"
    "net/http/httptest"
    "sync"
    "github.com/google/uuid"
)

// FakeServer is a deterministic, in-process fake PDND API server.
type FakeServer struct {
    mu             sync.RWMutex
    agreements     map[uuid.UUID]*StoredAgreement
    purposes       map[uuid.UUID][]StoredPurpose  // keyed by agreementID
    approvalPolicy string                          // "AUTOMATIC" (default) or "MANUAL"
    producerID     uuid.UUID                       // fixed per server instance
    consumerID     uuid.UUID                       // fixed per server instance
    mux            *http.ServeMux
}

// NewFakeServer creates a new fake PDND server with default settings.
func NewFakeServer() *FakeServer

// Start starts the server and returns an httptest.Server.
// The caller must call Close() on the returned server when done.
func (s *FakeServer) Start() *httptest.Server

// SetApprovalPolicy sets whether submit transitions to ACTIVE or PENDING.
// Valid values: "AUTOMATIC" (default), "MANUAL"
func (s *FakeServer) SetApprovalPolicy(policy string)

// SeedAgreement pre-populates an agreement in the store.
func (s *FakeServer) SeedAgreement(a StoredAgreement)

// SeedPurpose adds a purpose associated with an agreement.
func (s *FakeServer) SeedPurpose(agreementID uuid.UUID, p StoredPurpose)

// GetAgreement returns the current state of an agreement (for test assertions).
func (s *FakeServer) GetAgreement(id uuid.UUID) *StoredAgreement

// ProducerID returns the fixed producer ID used by this server.
func (s *FakeServer) ProducerID() uuid.UUID

// ConsumerID returns the fixed consumer ID used by this server.
func (s *FakeServer) ConsumerID() uuid.UUID
```

**Route registration** using Go 1.22+ `http.ServeMux` patterns:

```go
func (s *FakeServer) setupRoutes() {
    s.mux = http.NewServeMux()
    s.mux.HandleFunc("POST /agreements", s.handleCreateAgreement)
    s.mux.HandleFunc("GET /agreements", s.handleListAgreements)
    s.mux.HandleFunc("GET /agreements/{agreementId}", s.handleGetAgreement)
    s.mux.HandleFunc("DELETE /agreements/{agreementId}", s.handleDeleteAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/submit", s.handleSubmitAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/approve", s.handleApproveAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/reject", s.handleRejectAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/suspend", s.handleSuspendAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/unsuspend", s.handleUnsuspendAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/upgrade", s.handleUpgradeAgreement)
    s.mux.HandleFunc("POST /agreements/{agreementId}/clone", s.handleCloneAgreement)
    s.mux.HandleFunc("GET /agreements/{agreementId}/purposes", s.handleGetAgreementPurposes)
}
```

### 2. `internal/testing/fakepdnd/state.go` — State Types

```go
package fakepdnd

import (
    "time"
    "github.com/google/uuid"
)

// StoredAgreement represents an agreement in the fake server's state.
type StoredAgreement struct {
    ID                  uuid.UUID
    EServiceID          uuid.UUID
    DescriptorID        uuid.UUID
    ProducerID          uuid.UUID
    ConsumerID          uuid.UUID
    DelegationID        *uuid.UUID
    State               string
    SuspendedByConsumer bool
    SuspendedByProducer bool
    SuspendedByPlatform bool
    ConsumerNotes       string
    RejectionReason     string
    CreatedAt           time.Time
    UpdatedAt           *time.Time
    SuspendedAt         *time.Time
}

// StoredPurpose represents a purpose in the fake server's state.
type StoredPurpose struct {
    ID                  uuid.UUID
    EServiceID          uuid.UUID
    ConsumerID          uuid.UUID
    SuspendedByConsumer bool
    SuspendedByProducer bool
    Title               string
    Description         string
    CreatedAt           time.Time
    UpdatedAt           *time.Time
    IsRiskAnalysisValid bool
    IsFreeOfCharge      bool
    FreeOfChargeReason  string
    DelegationID        *uuid.UUID
}
```

### 3. `internal/testing/fakepdnd/agreements.go` — Agreement Handlers

Each handler must:
1. Parse the request (path params, query params, JSON body as applicable)
2. Validate the request (required fields, UUID format)
3. Check state transitions (return 409 if invalid)
4. Update internal state
5. Return JSON response matching the OpenAPI schema

**Handler implementations:**

#### `handleCreateAgreement` (POST /agreements)
- Parse body: `{"eserviceId":"uuid","descriptorId":"uuid","delegationId":"uuid"}`
- Create agreement with state "DRAFT", assign new UUID, set producerID/consumerID from server config
- Set createdAt to now
- Return 201 with Agreement JSON

#### `handleGetAgreement` (GET /agreements/{agreementId})
- Extract agreementId from path: `r.PathValue("agreementId")`
- If not found: return 404 Problem JSON
- Return 200 with Agreement JSON

#### `handleListAgreements` (GET /agreements)
- Parse query params: `states`, `producerIds`, `consumerIds`, `descriptorIds`, `eserviceIds`, `offset`, `limit`
- Filter agreements matching all provided filters
- Apply pagination (offset/limit slicing)
- Return 200 with `{"results":[...],"pagination":{"offset":N,"limit":N,"totalCount":N}}`

#### `handleDeleteAgreement` (DELETE /agreements/{agreementId})
- Only allowed when state is: DRAFT, PENDING, MISSING_CERTIFIED_ATTRIBUTES
- If state is anything else: return 409 Problem JSON
- Remove agreement from store
- Return 200 with `{}`

#### `handleSubmitAgreement` (POST /agreements/{agreementId}/submit)
- Only allowed when state is DRAFT
- Parse body: `{"consumerNotes":"string"}`
- If `approvalPolicy == "AUTOMATIC"`: transition to ACTIVE
- If `approvalPolicy == "MANUAL"`: transition to PENDING
- Update consumerNotes if provided
- Set updatedAt to now
- Return 200 with updated Agreement JSON

#### `handleApproveAgreement` (POST /agreements/{agreementId}/approve)
- Only allowed when state is PENDING
- Parse optional body: `{"delegationId":"uuid"}`
- Transition to ACTIVE
- Set updatedAt to now
- Return 200 with updated Agreement JSON

#### `handleRejectAgreement` (POST /agreements/{agreementId}/reject)
- Only allowed when state is PENDING
- Parse body: `{"reason":"string"}` — reason is required, 20-1000 chars
- Transition to REJECTED, set rejectionReason
- Set updatedAt to now
- Return 200 with updated Agreement JSON

#### `handleSuspendAgreement` (POST /agreements/{agreementId}/suspend)
- Only allowed when state is ACTIVE or SUSPENDED
- Parse optional body: `{"delegationId":"uuid"}`
- Transition to SUSPENDED
- Set suspendedByConsumer = true (simulating consumer-initiated suspension)
- Set suspendedAt to now, updatedAt to now
- Return 200 with updated Agreement JSON

#### `handleUnsuspendAgreement` (POST /agreements/{agreementId}/unsuspend)
- Only allowed when state is SUSPENDED
- Parse optional body: `{"delegationId":"uuid"}`
- Set suspendedByConsumer = false
- If no other suspend flags (suspendedByProducer, suspendedByPlatform) are set: transition to ACTIVE, clear suspendedAt
- If other suspend flags still set: remain SUSPENDED
- Set updatedAt to now
- Return 200 with updated Agreement JSON

#### `handleUpgradeAgreement` (POST /agreements/{agreementId}/upgrade)
- Only allowed when state is ACTIVE or SUSPENDED
- Archive the current agreement (set state to ARCHIVED)
- Create a NEW agreement with:
  - New UUID
  - Same eserviceId but with a new descriptorId (simulate new version available - just generate a new UUID for descriptorId)
  - Same producerId, consumerId
  - Same state as the original (ACTIVE or SUSPENDED)
  - New createdAt
- Store the new agreement
- Return 200 with the NEW Agreement JSON (not the archived one)

#### `handleCloneAgreement` (POST /agreements/{agreementId}/clone)
- Only allowed when state is REJECTED
- Create new DRAFT agreement with new UUID, same eserviceId/descriptorId
- Return 200 with new Agreement JSON

### 4. `internal/testing/fakepdnd/purposes.go` — Purposes Handler

#### `handleGetAgreementPurposes` (GET /agreements/{agreementId}/purposes)
- Extract agreementId, verify agreement exists (404 if not)
- Parse query: offset, limit
- Return paginated purposes list for that agreement
- Return 200 with `{"results":[...],"pagination":{...}}`

### 5. `internal/testing/fakepdnd/helpers.go` — Response Helpers

```go
// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{})

// writeProblem writes a Problem+JSON error response.
func writeProblem(w http.ResponseWriter, status int, title, detail string)

// parseUUID parses a UUID string, writes 400 error if invalid.
func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool)

// agreementToJSON converts a StoredAgreement to a JSON-serializable map/struct
// matching the OpenAPI Agreement schema field names (camelCase).
func agreementToJSON(a *StoredAgreement) map[string]interface{}
```

**Agreement JSON format** (must match OpenAPI schema exactly):
```json
{
  "id": "uuid-string",
  "eserviceId": "uuid-string",
  "descriptorId": "uuid-string",
  "producerId": "uuid-string",
  "consumerId": "uuid-string",
  "state": "DRAFT",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T11:00:00Z",
  "suspendedByConsumer": false,
  "suspendedByProducer": false,
  "suspendedByPlatform": false
}
```

Note: only include optional fields when they have values. The `delegationId`, `consumerNotes`, `rejectionReason`, `suspendedAt` fields should be omitted from JSON when empty/nil (use `omitempty` in struct tags or construct the map manually).

**Problem JSON format** (for errors):
```json
{
  "type": "about:blank",
  "status": 409,
  "title": "Conflict",
  "detail": "Cannot delete agreement in ACTIVE state"
}
```

### 6. `internal/testing/fakepdnd/server_test.go` — Smoke Tests

Test the fake server directly (not through Terraform). Use `http.Client` to make requests.

**Required smoke tests:**

| Test | Setup | Action | Assert |
|------|-------|--------|--------|
| `TestFakeServer_CreateAgreement` | Start server | POST /agreements with seed | 201, state=DRAFT |
| `TestFakeServer_GetAgreement` | Seed agreement | GET /agreements/{id} | 200, correct fields |
| `TestFakeServer_GetAgreement_NotFound` | Empty server | GET /agreements/{random-id} | 404 |
| `TestFakeServer_SubmitAutomatic` | Create DRAFT, policy=AUTOMATIC | POST submit | state=ACTIVE |
| `TestFakeServer_SubmitManual` | Create DRAFT, policy=MANUAL | POST submit | state=PENDING |
| `TestFakeServer_ApproveAgreement` | Create PENDING | POST approve | state=ACTIVE |
| `TestFakeServer_RejectAgreement` | Create PENDING | POST reject with reason | state=REJECTED |
| `TestFakeServer_SuspendActive` | Create ACTIVE | POST suspend | state=SUSPENDED, suspendedByConsumer=true |
| `TestFakeServer_UnsuspendToActive` | Create SUSPENDED (consumer only) | POST unsuspend | state=ACTIVE |
| `TestFakeServer_UnsuspendRemainsuspended` | SUSPENDED (both consumer+producer) | POST unsuspend | state=SUSPENDED (producer flag still set) |
| `TestFakeServer_DeleteDraft` | Create DRAFT | DELETE | 200, agreement gone |
| `TestFakeServer_DeleteActive_Fails` | Create ACTIVE | DELETE | 409 |
| `TestFakeServer_DeletePending` | Create PENDING | DELETE | 200, agreement gone |
| `TestFakeServer_UpgradeActive` | Create ACTIVE | POST upgrade | 200, new agreement with new ID, old is ARCHIVED |
| `TestFakeServer_UpgradeSuspended` | Create SUSPENDED | POST upgrade | 200, new agreement SUSPENDED |
| `TestFakeServer_CloneRejected` | Create REJECTED | POST clone | 200, new DRAFT with new ID |
| `TestFakeServer_InvalidTransition_SubmitActive` | Create ACTIVE | POST submit | 409 |
| `TestFakeServer_ListAgreements` | Seed 3 agreements (2 ACTIVE, 1 DRAFT) | GET with states=ACTIVE | 2 results |
| `TestFakeServer_ListAgreementPurposes` | Seed agreement + 2 purposes | GET purposes | 2 purposes returned |
| `TestFakeServer_ListPagination` | Seed 5 agreements | GET with offset=2, limit=2 | 2 results, totalCount=5 |

## Execution Steps

1. Read the domain types in `internal/client/api/types.go` for reference
2. Create all files in `internal/testing/fakepdnd/`
3. Run `go mod tidy`
4. Run `go test ./internal/testing/fakepdnd/... -v`
5. Fix failures and iterate until ALL tests pass
6. Run `go build ./...` to verify nothing is broken

## Verification

- `go test ./internal/testing/fakepdnd/... -v` — ALL smoke tests pass
- `go build ./...` — compiles successfully

## Important Notes

- The fake server must be thread-safe (use `sync.RWMutex`)
- Use Go 1.22+ `http.ServeMux` patterns for routing (e.g., `"POST /agreements/{agreementId}/submit"`)
- Extract path values with `r.PathValue("agreementId")`
- JSON field names must be camelCase to match the OpenAPI spec
- The server does NOT need to validate auth headers — the fake server trusts all requests
- Time values should use UTC and RFC 3339 format
- DO NOT import from `internal/client/api/` — the fake server has its own types to avoid circular dependencies
- DO NOT modify any existing files
