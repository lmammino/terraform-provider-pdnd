# Spec 61: Fake Server for Tenant Attribute Management

## Objective

Extend the fake PDND server with tenant and tenant attribute endpoints. When done, `go build ./...` must pass and all existing tests must continue to pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/testing/fakepdnd/clients.go` — recent handler pattern (CRUD + link/unlink)
2. `internal/testing/fakepdnd/state.go` — storage type definitions
3. `internal/testing/fakepdnd/server.go` — FakeServer struct, NewFakeServer, setupRoutes, Seed/Get pattern
4. `internal/testing/fakepdnd/helpers.go` — writeJSON, writeProblem, parseUUID, parseIntDefault (in agreements.go)
5. `internal/testing/fakepdnd/delegations.go` — handler factory pattern with closures

## Files to Create/Modify

### 1. Extend `internal/testing/fakepdnd/state.go`

```go
// StoredTenant represents a tenant in the fake server's state.
type StoredTenant struct {
    ID             uuid.UUID
    Name           string
    Kind           string // PA, PRIVATE, GSP, SCP
    ExternalOrigin string
    ExternalValue  string
    CreatedAt      time.Time
    UpdatedAt      *time.Time
    OnboardedAt    *time.Time
}

// StoredTenantCertifiedAttr represents a certified attribute assignment on a tenant.
type StoredTenantCertifiedAttr struct {
    ID         uuid.UUID
    AssignedAt time.Time
    RevokedAt  *time.Time
}

// StoredTenantDeclaredAttr represents a declared attribute assignment on a tenant.
type StoredTenantDeclaredAttr struct {
    ID           uuid.UUID
    AssignedAt   time.Time
    RevokedAt    *time.Time
    DelegationID *uuid.UUID
}

// StoredTenantVerifiedAttr represents a verified attribute assignment on a tenant.
type StoredTenantVerifiedAttr struct {
    ID         uuid.UUID
    AssignedAt time.Time
}
```

### 2. Extend `internal/testing/fakepdnd/server.go`

**Add to FakeServer struct:**
```go
tenants                  map[uuid.UUID]*StoredTenant
tenantCertifiedAttrs     map[uuid.UUID][]StoredTenantCertifiedAttr  // tenantID -> attrs
tenantDeclaredAttrs      map[uuid.UUID][]StoredTenantDeclaredAttr
tenantVerifiedAttrs      map[uuid.UUID][]StoredTenantVerifiedAttr
```

**Initialize in NewFakeServer.**

**Add seed/getter methods:**
```go
func (s *FakeServer) SeedTenant(t StoredTenant)
func (s *FakeServer) GetTenant(id uuid.UUID) *StoredTenant
func (s *FakeServer) SeedTenantCertifiedAttr(tenantID uuid.UUID, attr StoredTenantCertifiedAttr)
func (s *FakeServer) GetTenantCertifiedAttrs(tenantID uuid.UUID) []StoredTenantCertifiedAttr
func (s *FakeServer) SeedTenantDeclaredAttr(tenantID uuid.UUID, attr StoredTenantDeclaredAttr)
func (s *FakeServer) GetTenantDeclaredAttrs(tenantID uuid.UUID) []StoredTenantDeclaredAttr
func (s *FakeServer) SeedTenantVerifiedAttr(tenantID uuid.UUID, attr StoredTenantVerifiedAttr)
func (s *FakeServer) GetTenantVerifiedAttrs(tenantID uuid.UUID) []StoredTenantVerifiedAttr
```

**Add routes in setupRoutes** (before client routes):
```go
// Tenant routes.
s.mux.HandleFunc("GET /tenants", s.handleListTenants)
s.mux.HandleFunc("GET /tenants/{tenantId}", s.handleGetTenant)
s.mux.HandleFunc("GET /tenants/{tenantId}/certifiedAttributes", s.handleListTenantCertifiedAttrs)
s.mux.HandleFunc("POST /tenants/{tenantId}/certifiedAttributes", s.handleAssignTenantCertifiedAttr)
s.mux.HandleFunc("DELETE /tenants/{tenantId}/certifiedAttributes/{attributeId}", s.handleRevokeTenantCertifiedAttr)
s.mux.HandleFunc("GET /tenants/{tenantId}/declaredAttributes", s.handleListTenantDeclaredAttrs)
s.mux.HandleFunc("POST /tenants/{tenantId}/declaredAttributes", s.handleAssignTenantDeclaredAttr)
s.mux.HandleFunc("DELETE /tenants/{tenantId}/declaredAttributes/{attributeId}", s.handleRevokeTenantDeclaredAttr)
s.mux.HandleFunc("GET /tenants/{tenantId}/verifiedAttributes", s.handleListTenantVerifiedAttrs)
s.mux.HandleFunc("POST /tenants/{tenantId}/verifiedAttributes", s.handleAssignTenantVerifiedAttr)
s.mux.HandleFunc("DELETE /tenants/{tenantId}/verifiedAttributes/{attributeId}", s.handleRevokeTenantVerifiedAttr)
```

### 3. Create `internal/testing/fakepdnd/tenants.go` — Handlers

**Tenant handlers:**
- `handleListTenants` — filter by IPACode/taxCode (match ExternalOrigin/ExternalValue), paginate
- `handleGetTenant` — lookup by ID, return tenant JSON

**Tenant JSON helper:**
```go
func tenantToJSON(t *StoredTenant) map[string]interface{} {
    m := map[string]interface{}{
        "id": t.ID.String(), "name": t.Name,
        "createdAt": t.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
        "externalId": map[string]interface{}{"origin": t.ExternalOrigin, "value": t.ExternalValue},
    }
    if t.Kind != "" { m["kind"] = t.Kind }
    // handle optional timestamps...
    return m
}
```

**Certified attribute handlers:**
- `handleListTenantCertifiedAttrs` — list + paginate
- `handleAssignTenantCertifiedAttr` — parse body `{id}`, create StoredTenantCertifiedAttr with AssignedAt=now, respond 200
- `handleRevokeTenantCertifiedAttr` — find attr by ID, set RevokedAt=now, respond 200

**Declared attribute handlers:**
- `handleAssignTenantDeclaredAttr` — parse body `{id, delegationId?}`, store, respond 200
- `handleRevokeTenantDeclaredAttr` — set RevokedAt, respond 200

**Verified attribute handlers:**
- `handleAssignTenantVerifiedAttr` — parse body `{id, agreementId, expirationDate?}`, store, respond 200
- `handleRevokeTenantVerifiedAttr` — parse `agreementId` from QUERY param, respond 200

## Execution Steps

1. Read all referenced files
2. Add stored types to `state.go`
3. Add storage, init, seed/getter, routes to `server.go`
4. Create `tenants.go` with all handlers
5. Run `go build ./...`

## Verification

```shell
go build ./...
go test ./internal/testing/fakepdnd/... -v
```

## Important Notes

- Do NOT break existing tests
- `handleRevokeTenantVerifiedAttr` reads `agreementId` from `r.URL.Query().Get("agreementId")` — it's a required query param, NOT a path param
- Assign responses use status 200 (not 201)
- The `GetTenant` method name conflicts with the existing `GetTenant` on FakeServer used for test assertions of Stored tenants. Use `GetStoredTenant` or check if there's already a method named `GetTenant` — actually check `server.go` first. If no conflict exists, use `SeedTenant`/`GetTenantByID` or just `GetTenant`.
