# Agent 2: Fake Server — E-Services and Descriptors

## Objective

Extend the fake PDND server with e-service and descriptor endpoints, including state machine for descriptors. When done, `go test ./internal/testing/fakepdnd/... -v` must pass all existing + new tests.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read

1. `internal/testing/fakepdnd/server.go` — existing server structure, route setup, public API
2. `internal/testing/fakepdnd/state.go` — existing StoredAgreement, StoredPurpose
3. `internal/testing/fakepdnd/agreements.go` — handler pattern
4. `internal/testing/fakepdnd/helpers.go` — writeJSON, writeProblem, parseUUID, agreementToJSON
5. `internal/testing/fakepdnd/server_test.go` — existing test pattern

## Files to Modify

### 1. `internal/testing/fakepdnd/state.go` — Add Types

```go
type StoredEService struct {
    ID                      uuid.UUID
    ProducerID              uuid.UUID
    Name                    string
    Description             string
    Technology              string // REST, SOAP
    Mode                    string // RECEIVE, DELIVER
    IsSignalHubEnabled      bool
    IsConsumerDelegable     bool
    IsClientAccessDelegable bool
    PersonalData            bool
}

type StoredDescriptor struct {
    ID                      uuid.UUID
    EServiceID              uuid.UUID
    Version                 string
    State                   string // DRAFT, PUBLISHED, DEPRECATED, SUSPENDED, ARCHIVED, WAITING_FOR_APPROVAL
    AgreementApprovalPolicy string
    Audience                []string
    DailyCallsPerConsumer   int32
    DailyCallsTotal         int32
    VoucherLifespan         int32
    ServerUrls              []string
    Description             string
    CreatedAt               time.Time
    PublishedAt             *time.Time
    SuspendedAt             *time.Time
    DeprecatedAt            *time.Time
    ArchivedAt              *time.Time
}
```

### 2. `internal/testing/fakepdnd/server.go` — Add State and Routes

Add to FakeServer struct:
```go
eservices        map[uuid.UUID]*StoredEService
descriptors      map[uuid.UUID]map[uuid.UUID]*StoredDescriptor // eserviceID -> descriptorID -> descriptor
descriptorCount  map[uuid.UUID]int // eserviceID -> next version number
```

Add seed/getter methods:
```go
func (s *FakeServer) SeedEService(e StoredEService)
func (s *FakeServer) SeedDescriptor(d StoredDescriptor)
func (s *FakeServer) GetEService(id uuid.UUID) *StoredEService
func (s *FakeServer) GetDescriptor(eserviceID, descriptorID uuid.UUID) *StoredDescriptor
```

Add routes using Go 1.22+ patterns:
```
POST   /eservices
GET    /eservices
GET    /eservices/{eserviceId}
PATCH  /eservices/{eserviceId}
DELETE /eservices/{eserviceId}
PATCH  /eservices/{eserviceId}/name
PATCH  /eservices/{eserviceId}/description
PATCH  /eservices/{eserviceId}/delegation
PATCH  /eservices/{eserviceId}/signalHub
POST   /eservices/{eserviceId}/descriptors
GET    /eservices/{eserviceId}/descriptors
GET    /eservices/{eserviceId}/descriptors/{descriptorId}
PATCH  /eservices/{eserviceId}/descriptors/{descriptorId}
DELETE /eservices/{eserviceId}/descriptors/{descriptorId}
PATCH  /eservices/{eserviceId}/descriptors/{descriptorId}/quotas
POST   /eservices/{eserviceId}/descriptors/{descriptorId}/publish
POST   /eservices/{eserviceId}/descriptors/{descriptorId}/suspend
POST   /eservices/{eserviceId}/descriptors/{descriptorId}/unsuspend
POST   /eservices/{eserviceId}/descriptors/{descriptorId}/approve
```

### 3. Create `internal/testing/fakepdnd/eservices.go` — E-Service Handlers

**handleCreateEService** (POST /eservices):
- Parse body: `{"name":"...","description":"...","technology":"REST","mode":"DELIVER","descriptor":{...}}`
- Create EService with new UUID, set producerID from server config
- Create initial descriptor in DRAFT state with version "1"
- Return 200 with EService JSON (include descriptors array? — check OpenAPI response. The EService schema doesn't include descriptors inline in the response, so just return the EService)

**handleGetEService** (GET /eservices/{eserviceId}):
- Lookup by ID, 404 if not found
- Return 200 with EService JSON

**handleListEServices** (GET /eservices):
- Parse query: producerIds, name, technology, mode, offset, limit
- Filter, paginate, return `{results: [...], pagination: {...}}`

**handleDeleteEService** (DELETE /eservices/{eserviceId}):
- Check all descriptors are DRAFT — 409 if any non-DRAFT
- Remove eservice and all its descriptors
- Return 200 with `{}`

**handleUpdateDraftEService** (PATCH /eservices/{eserviceId}):
- Check all descriptors are DRAFT — 409 if not
- Parse merge-patch body, update only provided fields
- Return 200 with updated EService

**handleUpdatePublishedEServiceName** (PATCH /eservices/{eserviceId}/name):
- Parse body: `{"name":"..."}`
- Update name
- Return 200 with updated EService

**handleUpdatePublishedEServiceDescription** (PATCH /eservices/{eserviceId}/description):
- Parse body: `{"description":"..."}`
- Update description
- Return 200 with updated EService

**handleUpdatePublishedEServiceDelegation** (PATCH /eservices/{eserviceId}/delegation):
- Parse body: `{"isConsumerDelegable":true, "isClientAccessDelegable":false}`
- Update delegation flags
- Return 200 with updated EService

**handleUpdatePublishedEServiceSignalHub** (PATCH /eservices/{eserviceId}/signalHub):
- Parse body: `{"isSignalHubEnabled":true}`
- Update flag
- Return 200 with updated EService

### 4. Create `internal/testing/fakepdnd/descriptors.go` — Descriptor Handlers

**handleCreateDescriptor** (POST /eservices/{eserviceId}/descriptors):
- Check eservice exists (404)
- Check no existing DRAFT descriptor (409 if one exists)
- Parse body: descriptor seed
- Create descriptor with new UUID, version = next version number, state = DRAFT
- Return 200 with Descriptor JSON

**handleGetDescriptor** (GET /eservices/{eserviceId}/descriptors/{descriptorId}):
- Lookup, 404 if not found

**handleListDescriptors** (GET /eservices/{eserviceId}/descriptors):
- Parse query: state (optional filter), offset, limit
- Return paginated results

**handleDeleteDraftDescriptor** (DELETE /eservices/{eserviceId}/descriptors/{descriptorId}):
- Check state is DRAFT (409 otherwise)
- Check not last descriptor (409 if it is)
- Remove descriptor
- Return 200 with `{}`

**handleUpdateDraftDescriptor** (PATCH /eservices/{eserviceId}/descriptors/{descriptorId}):
- Check state is DRAFT (409 otherwise)
- Parse merge-patch body, update provided fields
- Return 200 with updated Descriptor

**handleUpdatePublishedDescriptorQuotas** (PATCH .../quotas):
- Check state is PUBLISHED, SUSPENDED, or DEPRECATED (409 otherwise)
- Parse body: `{"dailyCallsPerConsumer":..., "dailyCallsTotal":..., "voucherLifespan":...}`
- Update quotas
- Return 200

**handlePublishDescriptor** (POST .../publish):
- Check state is DRAFT (409 otherwise)
- Transition to PUBLISHED, set publishedAt
- Deprecate any previously PUBLISHED descriptor for this eservice (set to DEPRECATED, set deprecatedAt)
- Return 204 or 200

**handleSuspendDescriptor** (POST .../suspend):
- Check state is PUBLISHED or DEPRECATED (409 otherwise)
- Transition to SUSPENDED, set suspendedAt
- Return 204 or 200

**handleUnsuspendDescriptor** (POST .../unsuspend):
- Check state is SUSPENDED (409 otherwise)
- Transition back to PUBLISHED (or DEPRECATED if it was deprecated before — simplify to PUBLISHED for fake)
- Clear suspendedAt
- Return 204 or 200

**handleApproveDelegatedDescriptor** (POST .../approve):
- Check state is WAITING_FOR_APPROVAL (409 otherwise)
- Transition to PUBLISHED, set publishedAt
- Return 204 or 200

### 5. Extend `internal/testing/fakepdnd/helpers.go`

```go
func eserviceToJSON(e *StoredEService) map[string]interface{} { ... }
func descriptorToJSON(d *StoredDescriptor) map[string]interface{} { ... }
```

Use camelCase field names matching OpenAPI schema.

### 6. Extend `internal/testing/fakepdnd/server_test.go` — Smoke Tests

Add tests (keep all existing tests):

| Test | Action | Assert |
|------|--------|--------|
| TestFakeServer_CreateEService | POST /eservices | 200, has id, name matches |
| TestFakeServer_GetEService | Seed + GET | 200, fields match |
| TestFakeServer_GetEService_NotFound | GET random id | 404 |
| TestFakeServer_ListEServices | Seed 3, GET with filter | correct count |
| TestFakeServer_DeleteDraftEService | Create, DELETE | 200, gone |
| TestFakeServer_DeletePublishedEService_Fails | Create, publish descriptor, DELETE | 409 |
| TestFakeServer_UpdateDraftEService | Create, PATCH | 200, name updated |
| TestFakeServer_UpdatePublishedEServiceName | Publish, PATCH name | 200, name updated |
| TestFakeServer_CreateDescriptor | Create eservice, POST descriptor | 200, DRAFT |
| TestFakeServer_GetDescriptor | Seed + GET | 200, fields match |
| TestFakeServer_ListDescriptors | Seed multiple, GET | correct count |
| TestFakeServer_PublishDescriptor | Create DRAFT, POST publish | PUBLISHED |
| TestFakeServer_SuspendDescriptor | Publish, POST suspend | SUSPENDED |
| TestFakeServer_UnsuspendDescriptor | Suspend, POST unsuspend | PUBLISHED |
| TestFakeServer_DeleteDraftDescriptor | Create, DELETE | 200, gone |
| TestFakeServer_DeletePublishedDescriptor_Fails | Publish, DELETE | 409 |
| TestFakeServer_PublishDeprecatesPrevious | Publish v1, create+publish v2 | v1=DEPRECATED, v2=PUBLISHED |
| TestFakeServer_UpdateDraftDescriptor | Create, PATCH | quotas updated |
| TestFakeServer_UpdatePublishedDescriptorQuotas | Publish, PATCH quotas | quotas updated |
| TestFakeServer_InvalidTransition_PublishPublished | Publish, POST publish again | 409 |

## Execution Steps

1. Read all existing fake server code
2. Modify `state.go` — add StoredEService, StoredDescriptor
3. Modify `server.go` — add maps, routes, seed methods, init in NewFakeServer
4. Create `eservices.go` — all e-service handlers
5. Create `descriptors.go` — all descriptor handlers
6. Modify `helpers.go` — add eserviceToJSON, descriptorToJSON
7. Add new smoke tests to `server_test.go`
8. Run `TMPDIR=/tmp go test ./internal/testing/fakepdnd/... -v`
9. Iterate until ALL tests pass (old + new)

## Important Notes

- Do NOT break existing agreement tests
- Thread safety: all state mutations under mutex
- JSON field names: camelCase (eserviceId, descriptorId, etc.)
- The EService JSON response does NOT include descriptors inline — descriptors are fetched separately
- When publishing a descriptor, auto-deprecate any previously PUBLISHED descriptor for the same eservice
- Version numbering: first descriptor = "1", second = "2", etc.
