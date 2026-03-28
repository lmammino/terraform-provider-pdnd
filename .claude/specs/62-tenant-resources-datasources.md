# Spec 62: Tenant Attribute Resources + Data Sources + Provider Wiring

## Objective

Implement 3 tenant attribute resources and 2 tenant data sources. Wire into provider. When done, `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/client_purpose_resource.go` — link/unlink resource pattern (assign=create, revoke=delete)
2. `internal/resources/client_key_resource.go` — immutable resource with composite ID + import
3. `internal/resources/descriptor_attributes_shared.go` — shared resource with composite ID pattern
4. `internal/datasources/client_data_source.go` — singular data source
5. `internal/datasources/clients_data_source.go` — plural data source with filters
6. `internal/models/client.go` — model pattern
7. `internal/providerdata/providerdata.go` — ProviderData struct
8. `internal/provider/provider.go` — Resources(), DataSources(), Configure()
9. `internal/client/api/tenants.go` — TenantsAPI interface (created in Spec 60)

## Files to Create/Modify

### 1. Create `internal/models/tenant.go`

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// TenantCertifiedAttrResourceModel for pdnd_tenant_certified_attribute.
type TenantCertifiedAttrResourceModel struct {
    ID          types.String `tfsdk:"id"`          // composite: tenant_id/attribute_id
    TenantID    types.String `tfsdk:"tenant_id"`
    AttributeID types.String `tfsdk:"attribute_id"`
    AssignedAt  types.String `tfsdk:"assigned_at"`
    RevokedAt   types.String `tfsdk:"revoked_at"`
}

// TenantDeclaredAttrResourceModel for pdnd_tenant_declared_attribute.
type TenantDeclaredAttrResourceModel struct {
    ID           types.String `tfsdk:"id"`
    TenantID     types.String `tfsdk:"tenant_id"`
    AttributeID  types.String `tfsdk:"attribute_id"`
    DelegationID types.String `tfsdk:"delegation_id"`
    AssignedAt   types.String `tfsdk:"assigned_at"`
    RevokedAt    types.String `tfsdk:"revoked_at"`
}

// TenantVerifiedAttrResourceModel for pdnd_tenant_verified_attribute.
type TenantVerifiedAttrResourceModel struct {
    ID             types.String `tfsdk:"id"`
    TenantID       types.String `tfsdk:"tenant_id"`
    AttributeID    types.String `tfsdk:"attribute_id"`
    AgreementID    types.String `tfsdk:"agreement_id"`
    ExpirationDate types.String `tfsdk:"expiration_date"`
    AssignedAt     types.String `tfsdk:"assigned_at"`
}

// TenantDataSourceModel for pdnd_tenant data source.
type TenantDataSourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    Kind        types.String `tfsdk:"kind"`
    CreatedAt   types.String `tfsdk:"created_at"`
    UpdatedAt   types.String `tfsdk:"updated_at"`
    OnboardedAt types.String `tfsdk:"onboarded_at"`
}
```

### 2. Create `internal/resources/tenant_certified_attribute_resource.go`

Resource: `pdnd_tenant_certified_attribute`

**Schema:**
- `id` — Computed, UseStateForUnknown (composite `tenant_id/attribute_id`)
- `tenant_id` — Required, RequiresReplace, UUID validator
- `attribute_id` — Required, RequiresReplace, UUID validator
- `assigned_at` — Computed, UseStateForUnknown
- `revoked_at` — Computed

**Create:** Call `client.AssignTenantCertifiedAttribute(ctx, tenantID, attributeID)`. Set composite ID + assignedAt.

**Read:** Call `client.ListTenantCertifiedAttributes(ctx, tenantID, 0, 50)` and paginate to find matching attributeID. If not found, RemoveResource. If found and RevokedAt is set, also RemoveResource (attribute was revoked externally).

**Delete:** Call `client.RevokeTenantCertifiedAttribute(ctx, tenantID, attributeID)`. Handle 404 gracefully.

**ImportState:** Parse composite `tenant_id/attribute_id` using a helper (reuse or create `parseTenantAttributeCompositeID`).

### 3. Create `internal/resources/tenant_declared_attribute_resource.go`

Same as certified plus:
- `delegation_id` — Optional, RequiresReplace, UUID validator
- Create passes delegationID to `AssignTenantDeclaredAttribute`

### 4. Create `internal/resources/tenant_verified_attribute_resource.go`

Different from certified:
- `agreement_id` — Required, RequiresReplace, UUID validator
- `expiration_date` — Optional (RFC3339 timestamp)
- Create passes agreementID and expirationDate to `AssignTenantVerifiedAttribute`
- Delete passes agreementID to `RevokeTenantVerifiedAttribute`
- No `revoked_at` field (verified attributes don't have it in the response)

### 5. Create `internal/resources/tenant_attribute_import.go`

```go
func parseTenantAttributeCompositeID(id string) (tenantID, attributeID string, err error) {
    // Same pattern as parseDescriptorCompositeID — split on "/", validate both UUIDs
}
```

### 6. Create `internal/datasources/tenant_data_source.go` — Singular

Fetch tenant by ID. Populate name, kind, created_at, etc.

### 7. Create `internal/datasources/tenants_data_source.go` — Plural

List with optional filters: `ipa_code`, `tax_code`. Auto-paginate.

### 8. Modify `internal/providerdata/providerdata.go`

Add: `TenantsAPI api.TenantsAPI`

### 9. Modify `internal/provider/provider.go`

In Configure: create `tenantsAPI := api.NewTenantsClient(genClient)`, add to pd.
In Resources: add 3 tenant attribute resources.
In DataSources: add 2 tenant data sources.

## Execution Steps

1. Read all referenced files
2. Create `models/tenant.go`
3. Create `resources/tenant_attribute_import.go`
4. Create 3 tenant attribute resource files
5. Create 2 data source files
6. Modify `providerdata.go` and `provider.go`
7. Run `go build ./...`

## Verification

```shell
go build ./...
```

## Important Notes

- All assign operations return `JSON200` not `JSON201`
- Certified and declared: Read checks for `revokedAt != nil` — if revoked externally, treat as deleted
- Verified: no `revokedAt` in response type, so just check existence
- `RevokeTenantVerifiedAttribute` needs `agreementID` — store it in state from create
- `uuidRegex` is available from `agreement_resource.go` in the `resources` package
- Composite ID: `tenant_id/attribute_id` — same 2-part pattern as descriptor composite IDs
