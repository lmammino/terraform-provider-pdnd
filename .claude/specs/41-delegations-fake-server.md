# Spec 41: Fake Server for Delegations

## Objective

Extend the fake PDND server with consumer and producer delegation endpoints using parameterized handlers. When done, `go build ./...` must pass and all existing tests must continue to pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/testing/fakepdnd/descriptor_attributes.go` — parameterized handler pattern (`makeHandle*` functions taking type string)
2. `internal/testing/fakepdnd/state.go` — storage type definitions (StoredAgreement, StoredPurpose, etc.)
3. `internal/testing/fakepdnd/server.go` — FakeServer struct, NewFakeServer, setupRoutes, Seed/Get pattern, `getAttrGroupStore` private helper
4. `internal/testing/fakepdnd/helpers.go` — `writeJSON`, `writeProblem`, `parseUUID`, `parseIntDefault` (in agreements.go line 472), JSON helper functions
5. `internal/testing/fakepdnd/agreements.go` — handler patterns for create/get/list/accept/reject

## Files to Create/Modify

### 1. Extend `internal/testing/fakepdnd/state.go` — Add StoredDelegation

Add after `StoredDocument`:

```go
// StoredDelegation represents a consumer or producer delegation in the fake server's state.
type StoredDelegation struct {
    ID              uuid.UUID
    DelegatorID     uuid.UUID
    DelegateID      uuid.UUID
    EServiceID      uuid.UUID
    State           string // WAITING_FOR_APPROVAL, ACTIVE, REJECTED, REVOKED
    CreatedAt       time.Time
    SubmittedAt     time.Time
    UpdatedAt       *time.Time
    ActivatedAt     *time.Time
    RejectedAt      *time.Time
    RevokedAt       *time.Time
    RejectionReason *string
}
```

### 2. Extend `internal/testing/fakepdnd/server.go`

**Add to FakeServer struct:**
```go
consumerDelegations map[uuid.UUID]*StoredDelegation
producerDelegations map[uuid.UUID]*StoredDelegation
```

**Add to NewFakeServer initialization:**
```go
consumerDelegations: make(map[uuid.UUID]*StoredDelegation),
producerDelegations: make(map[uuid.UUID]*StoredDelegation),
```

**Add seed/getter methods** (before `setupRoutes`):

```go
// SeedDelegation pre-populates a delegation.
func (s *FakeServer) SeedDelegation(delegationType string, d StoredDelegation) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.getDelegationStore(delegationType)[d.ID] = &d
}

// GetDelegation returns a stored delegation for test assertions.
func (s *FakeServer) GetDelegation(delegationType string, id uuid.UUID) *StoredDelegation {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.getDelegationStore(delegationType)[id]
}

func (s *FakeServer) getDelegationStore(delegationType string) map[uuid.UUID]*StoredDelegation {
    switch delegationType {
    case "consumer":
        return s.consumerDelegations
    case "producer":
        return s.producerDelegations
    default:
        return nil
    }
}
```

**Add routes in setupRoutes** (before `// Descriptor document routes`):

```go
// Delegation routes.
for _, dt := range []string{"consumer", "producer"} {
    delegationType := dt // capture for closure
    prefix := "/" + delegationType + "Delegations"
    s.mux.HandleFunc("POST "+prefix, s.makeHandleCreateDelegation(delegationType))
    s.mux.HandleFunc("GET "+prefix, s.makeHandleListDelegations(delegationType))
    s.mux.HandleFunc("GET "+prefix+"/{delegationId}", s.makeHandleGetDelegation(delegationType))
    s.mux.HandleFunc("POST "+prefix+"/{delegationId}/accept", s.makeHandleAcceptDelegation(delegationType))
    s.mux.HandleFunc("POST "+prefix+"/{delegationId}/reject", s.makeHandleRejectDelegation(delegationType))
}
```

### 3. Create `internal/testing/fakepdnd/delegations.go` — Parameterized handlers

5 handler factory functions + 1 JSON helper:

**`delegationToJSON(d *StoredDelegation) map[string]interface{}`:**
Return map with all fields. Handle optional timestamps/string with nil checks.

**`makeHandleCreateDelegation(delegationType)`:**
- Parse body: `{"eserviceId": "...", "delegateId": "..."}`
- Lock, create StoredDelegation with new UUID, state=WAITING_FOR_APPROVAL
- Set `DelegatorID` from `s.consumerID` if consumer, `s.producerID` if producer
- Set `CreatedAt` and `SubmittedAt` to now
- Store in appropriate map, respond 201

**`makeHandleGetDelegation(delegationType)`:**
- Parse `{delegationId}`, look up in store, respond 200 or 404

**`makeHandleListDelegations(delegationType)`:**
- Parse query params: offset, limit, states[], delegatorIds[], delegateIds[], eserviceIds[]
- Filter delegations by matching criteria (using set-based lookup like purposes handler)
- Paginate and respond

**`makeHandleAcceptDelegation(delegationType)`:**
- Parse `{delegationId}`, validate state is WAITING_FOR_APPROVAL
- Set state=ACTIVE, activatedAt=now, updatedAt=now
- Respond 200

**`makeHandleRejectDelegation(delegationType)`:**
- Parse `{delegationId}` and body `{"rejectionReason": "..."}`
- Validate state is WAITING_FOR_APPROVAL
- Set state=REJECTED, rejectedAt=now, rejectionReason, updatedAt=now
- Respond 200

## Execution Steps

1. Read all referenced files
2. Add `StoredDelegation` to `state.go`
3. Add storage fields, initialization, seed/getter methods, routes to `server.go`
4. Create `delegations.go` with parameterized handlers + `delegationToJSON`
5. Run `go build ./...`
6. Run `go test ./internal/testing/fakepdnd/... -v`

## Verification

```shell
go build ./...
go test ./internal/testing/fakepdnd/... -v
```

## Important Notes

- Do NOT break existing tests or handlers
- The delegator is the authenticated user: use `s.consumerID` for consumer delegations, `s.producerID` for producer delegations
- Only WAITING_FOR_APPROVAL → ACTIVE (accept) and WAITING_FOR_APPROVAL → REJECTED (reject) transitions
- Query param arrays: parse with `r.URL.Query()["states"]`, `r.URL.Query()["delegatorIds"]`, etc.
- `parseIntDefault` is in `agreements.go` (line 472), `parseUUID` is in `helpers.go`
- Use `s.mu.Lock()` for writes, `s.mu.RLock()` for reads
