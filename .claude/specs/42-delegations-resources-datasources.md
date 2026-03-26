# Spec 42: Delegation Resources + Data Sources + Provider Wiring

## Objective

Implement 2 delegation resources (consumer + producer, shared parameterized logic) and 4 data sources (singular + plural for each type). Wire into provider. When done, `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/descriptor_attributes_shared.go` — shared resource logic parameterized by type string
2. `internal/resources/descriptor_certified_attributes_resource.go` — thin wrapper pattern
3. `internal/resources/agreement_resource.go` — CRUD resource with computed state, import
4. `internal/resources/agreement_helpers.go` — `populateModelFromAgreement` pattern
5. `internal/resources/agreement_import.go` — simple ID passthrough import
6. `internal/datasources/purpose_data_source.go` — singular data source pattern
7. `internal/datasources/purposes_data_source.go` — plural data source with filters + auto-pagination
8. `internal/models/agreement.go` — resource + data source model pattern
9. `internal/models/purpose_datasource.go` — data source model pattern
10. `internal/providerdata/providerdata.go` — ProviderData struct
11. `internal/provider/provider.go` — Resources(), DataSources(), Configure()
12. `internal/client/api/delegations.go` — DelegationsAPI interface (created in Spec 40)

## Files to Create/Modify

### 1. Create `internal/models/delegation.go` — Terraform models

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// DelegationResourceModel is the Terraform state model for delegation resources.
type DelegationResourceModel struct {
    ID              types.String `tfsdk:"id"`
    EServiceID      types.String `tfsdk:"eservice_id"`
    DelegateID      types.String `tfsdk:"delegate_id"`
    // Computed
    DelegatorID     types.String `tfsdk:"delegator_id"`
    State           types.String `tfsdk:"state"`
    CreatedAt       types.String `tfsdk:"created_at"`
    SubmittedAt     types.String `tfsdk:"submitted_at"`
    UpdatedAt       types.String `tfsdk:"updated_at"`
    ActivatedAt     types.String `tfsdk:"activated_at"`
    RejectedAt      types.String `tfsdk:"rejected_at"`
    RevokedAt       types.String `tfsdk:"revoked_at"`
    RejectionReason types.String `tfsdk:"rejection_reason"`
}

// DelegationDataSourceModel is the Terraform state model for singular delegation data sources.
type DelegationDataSourceModel struct {
    ID              types.String `tfsdk:"id"`
    EServiceID      types.String `tfsdk:"eservice_id"`
    DelegateID      types.String `tfsdk:"delegate_id"`
    DelegatorID     types.String `tfsdk:"delegator_id"`
    State           types.String `tfsdk:"state"`
    CreatedAt       types.String `tfsdk:"created_at"`
    SubmittedAt     types.String `tfsdk:"submitted_at"`
    UpdatedAt       types.String `tfsdk:"updated_at"`
    ActivatedAt     types.String `tfsdk:"activated_at"`
    RejectedAt      types.String `tfsdk:"rejected_at"`
    RevokedAt       types.String `tfsdk:"revoked_at"`
    RejectionReason types.String `tfsdk:"rejection_reason"`
}
```

### 2. Create `internal/resources/delegation_shared.go` — Shared resource logic

Parameterized CRUD functions following `descriptor_attributes_shared.go` pattern:

**`delegationSchema(delegationType string) schema.Schema`:**
- `id` — Computed, UseStateForUnknown
- `eservice_id` — Required, RequiresReplace, UUID validator
- `delegate_id` — Required, RequiresReplace, UUID validator
- `delegator_id` — Computed, UseStateForUnknown
- `state` — Computed
- `created_at`, `submitted_at` — Computed, UseStateForUnknown
- `updated_at`, `activated_at`, `rejected_at`, `revoked_at`, `rejection_reason` — Computed

**`delegationConfigure(req, resp) api.DelegationsAPI`:**
Extract from `providerdata.ProviderData.DelegationsAPI`

**`delegationCreate(ctx, req, resp, client, delegationType)`:**
1. Parse plan into `DelegationResourceModel`
2. Build `api.DelegationSeed{EServiceID, DelegateID}`
3. Call `client.CreateDelegation(ctx, delegationType, seed)`
4. Populate model with `populateDelegationModel`
5. Set state

**`delegationRead(ctx, req, resp, client, delegationType)`:**
1. Parse state
2. Call `client.GetDelegation(ctx, delegationType, id)`
3. If not found (404), `resp.State.RemoveResource(ctx)`
4. Populate model

**`delegationDelete(ctx, req, resp, client, delegationType)`:**
No API delete endpoint. Remove from state only. Add warning diagnostic:
```go
resp.Diagnostics.AddWarning(
    "Delegation Not Deleted on Platform",
    "The delegation has been removed from Terraform state but continues to exist on the PDND platform. Delegations cannot be deleted via the API.",
)
```

**`delegationImportState(ctx, req, resp)`:**
Simple ID passthrough using existing `idPath`:
```go
resource.ImportStatePassthroughID(ctx, idPath, req, resp)
```

**`populateDelegationModel(model *models.DelegationResourceModel, d *api.Delegation)`:**
Map all fields. Handle nullable timestamps/strings with `types.StringNull()`.

### 3. Create `internal/resources/consumer_delegation_resource.go` — Thin wrapper

```go
type consumerDelegationResource struct { client api.DelegationsAPI }
func NewConsumerDelegationResource() resource.Resource { return &consumerDelegationResource{} }
// Metadata: TypeName = req.ProviderTypeName + "_consumer_delegation"
// Schema: delegationSchema("consumer")
// Configure: r.client = delegationConfigure(req, resp)
// Create/Read/Update/Delete: delegate to shared functions with "consumer"
// ImportState: delegationImportState(ctx, req, resp)
```

### 4. Create `internal/resources/producer_delegation_resource.go` — Thin wrapper

Same as consumer with `"producer"` everywhere. TypeName = `"_producer_delegation"`.

### 5. Create `internal/datasources/delegation_shared.go` — Shared data source logic

Contains shared helpers for data sources (to avoid duplicating across 4 files):

**`populateDelegationDataSourceModel(model *models.DelegationDataSourceModel, d *api.Delegation)`:**
Map all fields, handle nullable with `types.StringNull()`.

### 6. Create `internal/datasources/consumer_delegation_data_source.go` — Singular

Follow `purpose_data_source.go` pattern:
- TypeName: `"_consumer_delegation"`
- Schema: `id` Required (UUID validated), everything else Computed
- Configure: Get `DelegationsAPI` from providerdata
- Read: Call `client.GetDelegation(ctx, "consumer", id)`, populate model

### 7. Create `internal/datasources/consumer_delegations_data_source.go` — Plural

Follow `purposes_data_source.go` pattern:
- TypeName: `"_consumer_delegations"`
- Filters: `states` (list of strings, OneOf validator), `delegator_ids` (list of UUID strings), `delegate_ids`, `eservice_ids`
- Read: Build `ListDelegationsParams`, auto-paginate, convert to nested model list

**Nested model** (local struct):
```go
type delegationItemModel struct {
    ID, EServiceID, DelegateID, DelegatorID, State types.String
    CreatedAt, SubmittedAt, UpdatedAt types.String
    ActivatedAt, RejectedAt, RevokedAt, RejectionReason types.String
}
```

### 8. Create `internal/datasources/producer_delegation_data_source.go` — Singular

Same as consumer with `"producer"`. TypeName: `"_producer_delegation"`.

### 9. Create `internal/datasources/producer_delegations_data_source.go` — Plural

Same as consumer with `"producer"`. TypeName: `"_producer_delegations"`.

### 10. Modify `internal/providerdata/providerdata.go` — Add field

```go
DelegationsAPI api.DelegationsAPI
```

### 11. Modify `internal/provider/provider.go` — Wire everything

In Configure:
```go
delegationsAPI := api.NewDelegationsClient(genClient)
```
Add to pd: `DelegationsAPI: delegationsAPI`

In Resources:
```go
resources.NewConsumerDelegationResource,
resources.NewProducerDelegationResource,
```

In DataSources:
```go
datasources.NewConsumerDelegationDataSource,
datasources.NewConsumerDelegationsDataSource,
datasources.NewProducerDelegationDataSource,
datasources.NewProducerDelegationsDataSource,
```

## Execution Steps

1. Read all referenced files
2. Create `models/delegation.go`
3. Create `resources/delegation_shared.go`
4. Create `resources/consumer_delegation_resource.go`
5. Create `resources/producer_delegation_resource.go`
6. Create `datasources/delegation_shared.go`
7. Create 4 data source files
8. Modify `providerdata.go` and `provider.go`
9. Run `go build ./...`

## Verification

```shell
go build ./...
```

## Important Notes

- Do NOT break existing resources, data sources, or tests
- No `desired_state` for delegations — they are fire-and-forget from the creator's perspective
- Delete only removes from Terraform state (add warning diagnostic)
- Update should never be called (all user fields have RequiresReplace). Implement as error just in case.
- `uuidRegex` is defined in `internal/resources/agreement_resource.go` and is package-accessible in `resources` package. For `datasources` package, you may need to define a local UUID regex or skip validation on filter fields (check how existing data sources handle this — `eservices_data_source.go` does NOT validate UUID format in filters).
- `idPath` is defined in `agreement_resource.go` and shared across `resources` package
- The `DelegationsAPI` is already in providerdata — the resources and data sources just need to read it from `pd.DelegationsAPI`
