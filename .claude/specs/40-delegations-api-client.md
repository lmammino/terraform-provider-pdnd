# Spec 40: Delegations API Client + Contract Tests

## Objective

Add `DelegationsAPI` interface parameterized by `delegationType` ("consumer"/"producer"), domain types, conversions, and contract tests. When done, `go test ./internal/client/api/... -v` and `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/descriptor_attributes.go` — parameterized API pattern with switch dispatch
2. `internal/client/api/agreements.go` — API interface pattern with Create/Get/List/state-transition methods
3. `internal/client/api/types.go` — domain types; has `DelegationRef` (~line 178), `PurposesPage` pattern
4. `internal/client/api/conversions.go` — conversion functions; has `delegationRefToGenerated` (~line 69)
5. `internal/client/api/descriptor_attributes_test.go` — contract test pattern with parameterized subtests
6. `internal/client/api/agreements_test.go` — `newTestClient` helper (lines 61-68)
7. `internal/client/generated/client.gen.go` — find these generated types and methods:
   - `type ConsumerDelegation struct` — fields: Id, DelegatorId, DelegateId, EserviceId, State, CreatedAt, SubmittedAt, UpdatedAt, ActivatedAt, RejectedAt, RevokedAt, RejectionReason
   - `type ProducerDelegation struct` — identical fields to ConsumerDelegation
   - `type DelegationSeed struct` — DelegateId, EserviceId
   - `type DelegationRejection struct` — RejectionReason
   - `type DelegationState string` — ACTIVE, REJECTED, REVOKED, WAITING_FOR_APPROVAL
   - `type ConsumerDelegations struct` — Results []ConsumerDelegation, Pagination
   - `type ProducerDelegations struct` — Results []ProducerDelegation, Pagination
   - `GetConsumerDelegationsParams` — States, DelegatorIds, DelegateIds, EserviceIds, Offset, Limit
   - `GetProducerDelegationsParams` — same fields
   - WithResponse methods: `CreateConsumerDelegationWithResponse` (JSON201), `GetConsumerDelegationWithResponse` (JSON200), `GetConsumerDelegationsWithResponse` (JSON200), `AcceptConsumerDelegationWithResponse` (JSON200), `RejectConsumerDelegationWithResponse` (JSON200) — and Producer equivalents
8. `internal/client/errors.go` — `CheckResponse`, `IsNotFound`

## Files to Create/Modify

### 1. Extend `internal/client/api/types.go` — Add delegation domain types

Add after `DelegationRef` type:

```go
// Delegation represents a consumer or producer delegation.
type Delegation struct {
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

// DelegationSeed contains fields for creating a new delegation.
type DelegationSeed struct {
    EServiceID uuid.UUID
    DelegateID uuid.UUID
}

// DelegationRejection contains fields for rejecting a delegation.
type DelegationRejection struct {
    RejectionReason string
}

// DelegationsPage is a paginated list of delegations.
type DelegationsPage struct {
    Results    []Delegation
    Pagination Pagination
}

// ListDelegationsParams contains filter parameters for listing delegations.
type ListDelegationsParams struct {
    States       []string
    DelegatorIDs []uuid.UUID
    DelegateIDs  []uuid.UUID
    EServiceIDs  []uuid.UUID
    Offset       int32
    Limit        int32
}
```

### 2. Extend `internal/client/api/conversions.go` — Add delegation conversions

```go
// consumerDelegationFromGenerated converts a generated ConsumerDelegation to a domain Delegation.
func consumerDelegationFromGenerated(g *generated.ConsumerDelegation) *Delegation {
    if g == nil { return nil }
    d := &Delegation{
        ID: uuid.UUID(g.Id), DelegatorID: uuid.UUID(g.DelegatorId),
        DelegateID: uuid.UUID(g.DelegateId), EServiceID: uuid.UUID(g.EserviceId),
        State: string(g.State), CreatedAt: g.CreatedAt, SubmittedAt: g.SubmittedAt,
        UpdatedAt: g.UpdatedAt, ActivatedAt: g.ActivatedAt,
        RejectedAt: g.RejectedAt, RevokedAt: g.RevokedAt,
        RejectionReason: g.RejectionReason,
    }
    return d
}

// producerDelegationFromGenerated converts a generated ProducerDelegation to a domain Delegation.
func producerDelegationFromGenerated(g *generated.ProducerDelegation) *Delegation {
    // Same structure as consumer, just different generated type
    if g == nil { return nil }
    d := &Delegation{
        ID: uuid.UUID(g.Id), DelegatorID: uuid.UUID(g.DelegatorId),
        DelegateID: uuid.UUID(g.DelegateId), EServiceID: uuid.UUID(g.EserviceId),
        State: string(g.State), CreatedAt: g.CreatedAt, SubmittedAt: g.SubmittedAt,
        UpdatedAt: g.UpdatedAt, ActivatedAt: g.ActivatedAt,
        RejectedAt: g.RejectedAt, RevokedAt: g.RevokedAt,
        RejectionReason: g.RejectionReason,
    }
    return d
}

// delegationSeedToGenerated converts a domain DelegationSeed to a generated DelegationSeed.
func delegationSeedToGenerated(s DelegationSeed) generated.DelegationSeed {
    return generated.DelegationSeed{
        EserviceId: openapi_types.UUID(s.EServiceID),
        DelegateId: openapi_types.UUID(s.DelegateID),
    }
}

// delegationRejectionToGenerated converts a domain DelegationRejection to a generated DelegationRejection.
func delegationRejectionToGenerated(r DelegationRejection) generated.DelegationRejection {
    return generated.DelegationRejection{
        RejectionReason: r.RejectionReason,
    }
}
```

### 3. Create `internal/client/api/delegations.go` — Interface + implementation

```go
// DelegationsAPI defines operations on PDND delegations.
// The delegationType parameter must be "consumer" or "producer".
type DelegationsAPI interface {
    CreateDelegation(ctx context.Context, delegationType string, seed DelegationSeed) (*Delegation, error)
    GetDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error)
    ListDelegations(ctx context.Context, delegationType string, params ListDelegationsParams) (*DelegationsPage, error)
    AcceptDelegation(ctx context.Context, delegationType string, id uuid.UUID) (*Delegation, error)
    RejectDelegation(ctx context.Context, delegationType string, id uuid.UUID, rejection DelegationRejection) (*Delegation, error)
}
```

Implementation: `delegationsClient` struct with `*generated.ClientWithResponses`.
Constructor: `NewDelegationsClient(c *generated.ClientWithResponses) DelegationsAPI`

Each method uses a switch on `delegationType`:

- **CreateDelegation**: switch to `CreateConsumerDelegationWithResponse` or `CreateProducerDelegationWithResponse`. Body: `delegationSeedToGenerated(seed)`. Check `JSON201`.
- **GetDelegation**: switch to `GetConsumerDelegationWithResponse` or `GetProducerDelegationWithResponse`. Check `JSON200`.
- **ListDelegations**: Build params struct (`GetConsumerDelegationsParams` or `GetProducerDelegationsParams`). Convert States to `[]generated.DelegationState`, IDs to `[]openapi_types.UUID` using `uuidsToOpenAPI`. Check `JSON200`. Convert results using appropriate `*DelegationFromGenerated`.
- **AcceptDelegation**: switch to `AcceptConsumerDelegationWithResponse` or `AcceptProducerDelegationWithResponse`. Check `JSON200`.
- **RejectDelegation**: switch. Body: `delegationRejectionToGenerated(rejection)`. Check `JSON200`.

### 4. Create `internal/client/api/delegations_test.go` — Contract tests

Parameterized tests iterating over `[]string{"consumer", "producer"}`:

Canned delegation JSON:
```json
{
  "id": "dd0e8400-e29b-41d4-a716-446655440030",
  "delegatorId": "aa0e8400-e29b-41d4-a716-446655440031",
  "delegateId": "bb0e8400-e29b-41d4-a716-446655440032",
  "eserviceId": "cc0e8400-e29b-41d4-a716-446655440033",
  "state": "WAITING_FOR_APPROVAL",
  "createdAt": "2024-01-01T00:00:00Z",
  "submittedAt": "2024-01-01T00:00:00Z"
}
```

| Test | Verifies |
|------|----------|
| `TestCreateDelegation_Contract` | POST `/{type}Delegations`, body has delegateId+eserviceId, 201 response |
| `TestGetDelegation_Contract` | GET `/{type}Delegations/{id}`, 200 response, all fields parsed |
| `TestListDelegations_Contract` | GET `/{type}Delegations` with offset/limit, paginated response |
| `TestAcceptDelegation_Contract` | POST `/{type}Delegations/{id}/accept`, 200 response |
| `TestRejectDelegation_Contract` | POST `/{type}Delegations/{id}/reject`, body has rejectionReason |

Use unique test UUIDs to avoid conflicts with other test files:
```go
var (
    testDelegationID  = uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440030")
    testDelegatorID   = uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440031")
    testDelegateID    = uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440032")
    testDelegationESID = uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440033")
)
```

## Execution Steps

1. Read generated client to confirm exact method signatures and type names
2. Add domain types to `types.go`
3. Add conversion functions to `conversions.go`
4. Create `delegations.go`
5. Create `delegations_test.go`
6. Run `go test ./internal/client/api/... -v`
7. Run `go build ./...`

## Verification

```shell
go test ./internal/client/api/... -v
go build ./...
```

## Important Notes

- Do NOT break existing tests
- The API paths are camelCase: `/consumerDelegations`, NOT `/consumer-delegations`
- `ConsumerDelegation` and `ProducerDelegation` are separate Go types with identical fields — two conversion functions are required
- The `DelegationState` enum values are: `ACTIVE`, `REJECTED`, `REVOKED`, `WAITING_FOR_APPROVAL` (note WAITING_FOR_APPROVAL not WAITING-FOR-APPROVAL)
- AcceptDelegation takes NO body — just POST to the accept endpoint
- RejectDelegation takes body with `rejectionReason` (required string)
- The generated `GetConsumerDelegationsParams` and `GetProducerDelegationsParams` have identical fields but are separate types — handle each in the switch
