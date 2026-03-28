# Spec 63: Tenant Tests + Documentation

## Objective

Add acceptance tests for 3 tenant attribute resources and 2 tenant data sources, plus doc templates, examples, and README update. When done, all tests pass and `make docs` generates documentation.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/client_key_resource_acc_test.go` — recent immutable resource test
2. `internal/resources/client_purpose_resource_acc_test.go` — link/unlink resource test
3. `internal/datasources/client_data_source_acc_test.go` — singular data source test
4. `internal/datasources/clients_data_source_acc_test.go` — plural data source test
5. `internal/testing/fakepdnd/server.go` — SeedTenant, SeedTenantCertifiedAttr, etc. (added in Spec 61)
6. `internal/testing/fakepdnd/state.go` — StoredTenant, StoredTenantCertifiedAttr, etc.
7. `internal/resources/tenant_certified_attribute_resource.go` — created in Spec 62
8. `internal/resources/tenant_declared_attribute_resource.go` — created in Spec 62
9. `internal/resources/tenant_verified_attribute_resource.go` — created in Spec 62
10. `templates/resources/client_key.md.tmpl` — recent doc template format
11. `README.md` — tables to update

Also check test helpers:
12. `internal/resources/agreement_resource_acc_test.go` — `testAccProviderConfig`, `testAccProtoV6ProviderFactories`
13. `internal/datasources/agreement_data_source_acc_test.go` — data source test helpers

## Files to Create/Modify

### 1. Create `internal/resources/tenant_certified_attribute_resource_acc_test.go`

**`TestAccTenantCertifiedAttribute_Create`:** Seed tenant + certified attribute, assign, verify assigned_at.
```go
fake := fakepdnd.NewFakeServer()
tenantID := uuid.New()
attrID := uuid.New()
fake.SeedTenant(fakepdnd.StoredTenant{
    ID: tenantID, Name: "Test Tenant", Kind: "PA",
    ExternalOrigin: "IPA", ExternalValue: "test123",
    CreatedAt: time.Now().UTC(),
})
// Also seed the certified attribute in the attributes store
fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
    ID: attrID, Name: "Test Attr", Description: "test",
    Code: "TA1", Origin: "IPA", CreatedAt: time.Now().UTC(),
})
ts := fake.Start()
defer ts.Close()

// Apply and verify
```

**`TestAccTenantCertifiedAttribute_Delete`:** Assign then destroy, verify revoked.

**`TestAccTenantCertifiedAttribute_Import`:** Import via `tenant_id/attribute_id`.

### 2. Create `internal/resources/tenant_declared_attribute_resource_acc_test.go`

**`TestAccTenantDeclaredAttribute_Create`:** With optional delegation_id.
**`TestAccTenantDeclaredAttribute_Delete`:** Verify revoked.

### 3. Create `internal/resources/tenant_verified_attribute_resource_acc_test.go`

**`TestAccTenantVerifiedAttribute_Create`:** With agreement_id.
**`TestAccTenantVerifiedAttribute_Delete`:** Verify revoked (passes agreementId).
**`TestAccTenantVerifiedAttribute_Import`:** Import, ImportStateVerifyIgnore: `["agreement_id", "expiration_date"]`.

### 4. Create `internal/datasources/tenant_data_source_acc_test.go`

**`TestAccTenantDataSource`:** Seed tenant, fetch by ID, verify name/kind.

### 5. Create `internal/datasources/tenants_data_source_acc_test.go`

**`TestAccTenantsDataSource`:** Seed multiple tenants, list, verify count.

### 6. Create doc templates (5 files)

- `templates/resources/tenant_certified_attribute.md.tmpl`
- `templates/resources/tenant_declared_attribute.md.tmpl`
- `templates/resources/tenant_verified_attribute.md.tmpl`
- `templates/data-sources/tenant.md.tmpl`
- `templates/data-sources/tenants.md.tmpl`

### 7. Create examples (7 files in 5 directories)

**`examples/resources/pdnd_tenant_certified_attribute/resource.tf`:**
```hcl
resource "pdnd_tenant_certified_attribute" "ipa" {
  tenant_id    = data.pdnd_tenant.example.id
  attribute_id = data.pdnd_certified_attribute.ipa.id
}
```

**`examples/resources/pdnd_tenant_certified_attribute/import.sh`:**
```shell
terraform import pdnd_tenant_certified_attribute.ipa <tenant-uuid>/<attribute-uuid>
```

Similar for declared and verified.

**`examples/data-sources/pdnd_tenant/data-source.tf`**
**`examples/data-sources/pdnd_tenants/data-source.tf`**

### 8. Modify `README.md`

Add to resources table:
```
| `pdnd_tenant_certified_attribute` | Assigns a certified attribute to a tenant    |
| `pdnd_tenant_declared_attribute`  | Assigns a declared attribute to a tenant     |
| `pdnd_tenant_verified_attribute`  | Assigns a verified attribute to a tenant     |
```

Add to data sources table:
```
| `pdnd_tenant`                | Fetches a single tenant by ID                |
| `pdnd_tenants`               | Lists tenants with optional filters          |
```

## Execution Steps

1. Read all referenced files
2. Create example directories (`mkdir -p`) and files
3. Create doc templates
4. Create resource acceptance tests
5. Create data source acceptance tests
6. Update README.md
7. Run `TF_ACC=1 go test ./internal/resources/... -v -run "Tenant" -timeout 5m`
8. Run `TF_ACC=1 go test ./internal/datasources/... -v -run "Tenant" -timeout 5m`
9. Run `go test ./...`
10. Run `make lint`
11. Run `make docs`

## Verification

```shell
TF_ACC=1 go test ./internal/resources/... -v -run "Tenant" -timeout 5m
TF_ACC=1 go test ./internal/datasources/... -v -run "Tenant" -timeout 5m
go test ./...
make lint
make docs
```

## Important Notes

- Do NOT break existing tests
- `testAccProviderConfig` and `testAccProtoV6ProviderFactories` are already defined
- For verified attribute import: `agreement_id` and `expiration_date` cannot be recovered from API — add to `ImportStateVerifyIgnore`
- For certified/declared import: `revoked_at` is computed, should be fine
- Seed tenants with `SeedTenant`, seed attributes with existing `SeedCertifiedAttribute`/`SeedDeclaredAttribute`/`SeedVerifiedAttribute`
- Create example directories with `mkdir -p` before writing files
