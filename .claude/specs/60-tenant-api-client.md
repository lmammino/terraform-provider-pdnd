# Spec 60: Tenant API Client + Contract Tests

## Objective

Add `TenantsAPI` interface covering tenant reads and attribute assign/revoke operations for all 3 types. When done, `go test ./internal/client/api/... -v` and `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/clients.go` — recent API client pattern (interface + impl)
2. `internal/client/api/delegations.go` — parameterized pattern (reference only)
3. `internal/client/api/types.go` — domain types inventory
4. `internal/client/api/conversions.go` — conversion patterns
5. `internal/client/api/clients_test.go` — recent contract test pattern
6. `internal/client/generated/client.gen.go` — find:
   - `type Tenant struct` (~line 2408): id, externalId, createdAt, updatedAt, name, kind, onboardedAt, subUnitType
   - `type Tenants struct`: Results []Tenant, Pagination
   - `type TenantKind string` — PA, PRIVATE, GSP, SCP
   - `type TenantCertifiedAttribute struct`: Id, AssignedAt, RevokedAt*
   - `type TenantCertifiedAttributeSeed struct`: Id
   - `type TenantCertifiedAttributes struct`: Results, Pagination
   - `type TenantDeclaredAttribute struct`: Id, AssignedAt, RevokedAt*, DelegationId*
   - `type TenantDeclaredAttributeSeed struct`: Id, DelegationId*
   - `type TenantDeclaredAttributes struct`: Results, Pagination
   - `type TenantVerifiedAttribute struct`: Id, AssignedAt
   - `type TenantVerifiedAttributeSeed struct`: Id, AgreementId, ExpirationDate*
   - `type TenantVerifiedAttributes struct`: Results, Pagination
   - `GetTenantsParams`: IPACode*, TaxCode*, Offset, Limit
   - `RevokeTenantVerifiedAttributeParams`: AgreementId (required query param)
   - Response types: Assign returns JSON200 (not JSON201!), Get returns JSON200, Revoke returns JSON200
   - `type ExternalId struct`: Origin string, Value string
   - WithResponse methods: `GetTenantWithResponse`, `GetTenantsWithResponse`, `AssignTenantCertifiedAttributeWithResponse`, `RevokeTenantCertifiedAttributeWithResponse`, etc.
7. `internal/client/errors.go` — `CheckResponse`, `IsNotFound`

## Files to Create/Modify

### 1. Extend `internal/client/api/types.go` — Add tenant domain types

```go
// TenantInfo represents a PDND tenant.
type TenantInfo struct {
    ID          uuid.UUID
    Name        string
    Kind        *string
    ExternalID  *TenantExternalID
    CreatedAt   time.Time
    UpdatedAt   *time.Time
    OnboardedAt *time.Time
    SubUnitType *string
}

// TenantExternalID represents a tenant's external identity.
type TenantExternalID struct {
    Origin string
    Value  string
}

// TenantsPage is a paginated list of tenants.
type TenantsPage struct {
    Results    []TenantInfo
    Pagination Pagination
}

// ListTenantsParams contains filter parameters for listing tenants.
type ListTenantsParams struct {
    IPACode *string
    TaxCode *string
    Offset  int32
    Limit   int32
}

// TenantCertifiedAttr represents a certified attribute assigned to a tenant.
type TenantCertifiedAttr struct {
    ID         uuid.UUID
    AssignedAt time.Time
    RevokedAt  *time.Time
}

// TenantDeclaredAttr represents a declared attribute assigned to a tenant.
type TenantDeclaredAttr struct {
    ID           uuid.UUID
    AssignedAt   time.Time
    RevokedAt    *time.Time
    DelegationID *uuid.UUID
}

// TenantVerifiedAttr represents a verified attribute assigned to a tenant.
type TenantVerifiedAttr struct {
    ID         uuid.UUID
    AssignedAt time.Time
}

// TenantCertifiedAttrsPage is a paginated list.
type TenantCertifiedAttrsPage struct {
    Results    []TenantCertifiedAttr
    Pagination Pagination
}

// TenantDeclaredAttrsPage is a paginated list.
type TenantDeclaredAttrsPage struct {
    Results    []TenantDeclaredAttr
    Pagination Pagination
}

// TenantVerifiedAttrsPage is a paginated list.
type TenantVerifiedAttrsPage struct {
    Results    []TenantVerifiedAttr
    Pagination Pagination
}
```

### 2. Extend `internal/client/api/conversions.go`

Add conversion functions:
- `tenantInfoFromGenerated(g *generated.Tenant) *TenantInfo`
- `tenantCertifiedAttrFromGenerated(g *generated.TenantCertifiedAttribute) *TenantCertifiedAttr`
- `tenantDeclaredAttrFromGenerated(g *generated.TenantDeclaredAttribute) *TenantDeclaredAttr`
- `tenantVerifiedAttrFromGenerated(g *generated.TenantVerifiedAttribute) *TenantVerifiedAttr`

Handle ExternalId (nested struct), Kind (pointer to string), nullable timestamps.

### 3. Create `internal/client/api/tenants.go` — Interface + implementation

```go
type TenantsAPI interface {
    // Tenant reads
    GetTenant(ctx context.Context, id uuid.UUID) (*TenantInfo, error)
    ListTenants(ctx context.Context, params ListTenantsParams) (*TenantsPage, error)

    // Certified attributes
    ListTenantCertifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantCertifiedAttrsPage, error)
    AssignTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error)
    RevokeTenantCertifiedAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantCertifiedAttr, error)

    // Declared attributes
    ListTenantDeclaredAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantDeclaredAttrsPage, error)
    AssignTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID, delegationID *uuid.UUID) (*TenantDeclaredAttr, error)
    RevokeTenantDeclaredAttribute(ctx context.Context, tenantID, attributeID uuid.UUID) (*TenantDeclaredAttr, error)

    // Verified attributes
    ListTenantVerifiedAttributes(ctx context.Context, tenantID uuid.UUID, offset, limit int32) (*TenantVerifiedAttrsPage, error)
    AssignTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID, expirationDate *time.Time) (*TenantVerifiedAttr, error)
    RevokeTenantVerifiedAttribute(ctx context.Context, tenantID, attributeID, agreementID uuid.UUID) (*TenantVerifiedAttr, error)
}
```

Key details:
- All assign endpoints return `JSON200` (not JSON201)
- `RevokeTenantVerifiedAttribute` must pass `agreementId` via `RevokeTenantVerifiedAttributeParams`
- `AssignTenantVerifiedAttribute` body includes `agreementId` and optional `expirationDate`
- `AssignTenantDeclaredAttribute` body includes optional `delegationId`

### 4. Create `internal/client/api/tenants_test.go` — Contract tests

Use unique UUIDs:
```go
var (
    testTenantID     = uuid.MustParse("tt0e8400-e29b-41d4-a716-446655440060")
    testTenantAttrID = uuid.MustParse("ta0e8400-e29b-41d4-a716-446655440061")
    testTenantAgrID  = uuid.MustParse("ag0e8400-e29b-41d4-a716-446655440062")
)
```

| Test | Verifies |
|------|----------|
| `TestGetTenant_Contract` | GET `/tenants/{id}`, parses tenant fields |
| `TestListTenants_Contract` | GET `/tenants` with offset/limit |
| `TestAssignTenantCertifiedAttribute_Contract` | POST `/tenants/{id}/certifiedAttributes`, body with id |
| `TestRevokeTenantCertifiedAttribute_Contract` | DELETE `/tenants/{id}/certifiedAttributes/{attrId}` |
| `TestAssignTenantDeclaredAttribute_Contract` | POST `/tenants/{id}/declaredAttributes`, body with id |
| `TestRevokeTenantDeclaredAttribute_Contract` | DELETE `/tenants/{id}/declaredAttributes/{attrId}` |
| `TestAssignTenantVerifiedAttribute_Contract` | POST `/tenants/{id}/verifiedAttributes`, body with id+agreementId |
| `TestRevokeTenantVerifiedAttribute_Contract` | DELETE `/tenants/{id}/verifiedAttributes/{attrId}?agreementId=...` |

## Execution Steps

1. Read generated client for exact signatures
2. Add types to `types.go`
3. Add conversions to `conversions.go`
4. Create `tenants.go`
5. Create `tenants_test.go`
6. Run `go test ./internal/client/api/... -v`
7. Run `go build ./...`

## Verification

```shell
go test ./internal/client/api/... -v
go build ./...
```

## Important Notes

- All assign operations return `JSON200` not `JSON201`
- `RevokeTenantVerifiedAttribute` requires `agreementId` as a QUERY parameter (via `RevokeTenantVerifiedAttributeParams`), not in the path
- The `Tenant` generated type has `ExternalId ExternalId` (nested struct with Origin + Value)
- `TenantKind` is optional (`*TenantKind`) — may be nil
- UUIDs for test variables must not conflict with others in the package — use the `tt`/`ta`/`ag` prefixes
