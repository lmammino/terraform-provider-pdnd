# Spec 31: Purpose Data Sources (pdnd_purpose + pdnd_purposes)

## Objective

Implement two new Terraform data sources: `pdnd_purpose` (fetch by ID) and `pdnd_purposes` (list with filters). When done, `go build ./...` must succeed and the provider must register both data sources.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/datasources/eservice_data_source.go` — singular data source pattern (fetch by ID, populate model)
2. `internal/datasources/eservices_data_source.go` — plural data source pattern (filters, auto-pagination, nested model)
3. `internal/datasources/agreement_purposes_data_source.go` — existing purpose-related data source with `purposeNestedModel` and `purposeToNestedModel` conversion
4. `internal/models/eservice.go` — data source model pattern (`EServiceDataSourceModel` vs `EServiceResourceModel`)
5. `internal/models/purpose.go` — existing `PurposeResourceModel` (reference for fields, NOT the data source model)
6. `internal/client/api/purposes.go` — `PurposesAPI` interface with `GetPurpose` and `ListPurposes` (added in Spec 30)
7. `internal/client/api/types.go` — `Purpose`, `PurposeVersion`, `PurposesPage`, `ListPurposesParams`
8. `internal/resources/purpose_helpers.go` — `derivePurposeState` function (reuse for state derivation)
9. `internal/providerdata/providerdata.go` — ProviderData struct (already has `PurposesAPI`)
10. `internal/provider/provider.go` — DataSources() method (add new data sources here)

## Files to Create/Modify

### 1. Create `internal/models/purpose_datasource.go` — Data source models

Follow the pattern from `eservice.go` which has separate resource and data source models.

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// PurposeDataSourceModel is the Terraform state model for pdnd_purpose data source.
type PurposeDataSourceModel struct {
    ID                  types.String `tfsdk:"id"`
    EServiceID          types.String `tfsdk:"eservice_id"`
    ConsumerID          types.String `tfsdk:"consumer_id"`
    Title               types.String `tfsdk:"title"`
    Description         types.String `tfsdk:"description"`
    DailyCalls          types.Int64  `tfsdk:"daily_calls"`
    State               types.String `tfsdk:"state"`
    IsFreeOfCharge      types.Bool   `tfsdk:"is_free_of_charge"`
    FreeOfChargeReason  types.String `tfsdk:"free_of_charge_reason"`
    IsRiskAnalysisValid types.Bool   `tfsdk:"is_risk_analysis_valid"`
    SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
    SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
    DelegationID        types.String `tfsdk:"delegation_id"`
    VersionID           types.String `tfsdk:"version_id"`
    CreatedAt           types.String `tfsdk:"created_at"`
    UpdatedAt           types.String `tfsdk:"updated_at"`
}

// PurposesDataSourceModel is the Terraform state model for pdnd_purposes data source.
type PurposesDataSourceModel struct {
    EServiceIDs types.List   `tfsdk:"eservice_ids"`
    Title       types.String `tfsdk:"title"`
    ConsumerIDs types.List   `tfsdk:"consumer_ids"`
    States      types.List   `tfsdk:"states"`
    Purposes    types.List   `tfsdk:"purposes"`
}
```

### 2. Create `internal/datasources/purpose_data_source.go` — Singular data source

Follow `eservice_data_source.go` pattern exactly.

```go
type purposeDataSource struct {
    client api.PurposesAPI
}

func NewPurposeDataSource() datasource.DataSource {
    return &purposeDataSource{}
}
```

**Metadata:** TypeName = `req.ProviderTypeName + "_purpose"`

**Schema:** All attributes computed except `id` which is Required + UUID validated.

Fields:
- `id` — Required, UUID validator
- `eservice_id`, `consumer_id`, `title`, `description`, `state`, `version_id`, `created_at`, `updated_at`, `free_of_charge_reason`, `delegation_id` — Computed string
- `daily_calls` — Computed int64
- `is_free_of_charge`, `is_risk_analysis_valid`, `suspended_by_consumer`, `suspended_by_producer` — Computed bool

**Configure:** Get `PurposesAPI` from `providerdata.ProviderData.PurposesAPI`

**Read:**
1. Parse `id` from config
2. Call `d.client.GetPurpose(ctx, id)`
3. Populate model using a `populatePurposeDataSourceModel` helper
4. Set state

**Helper function** `populatePurposeDataSourceModel(model *models.PurposeDataSourceModel, p *api.Purpose)`:
- Import `derivePurposeState` from `resources` package? No — data sources are in `datasources` package and can't import `resources`. Duplicate the derivation inline or create a shared utility.
- **Best approach:** Inline the state derivation logic (3 lines: check currentVersion, then waitingForApproval, else DRAFT). This avoids cross-package dependencies.
- Set `DailyCalls` from `currentVersion.DailyCalls` or `waitingForApprovalVersion.DailyCalls`
- Set `VersionID` from `currentVersion.ID` or `waitingForApprovalVersion.ID`
- Handle nullable fields with `types.StringNull()` / `types.BoolNull()`

### 3. Create `internal/datasources/purposes_data_source.go` — Plural data source

Follow `eservices_data_source.go` pattern exactly.

```go
type purposesDataSource struct {
    client api.PurposesAPI
}

func NewPurposesDataSource() datasource.DataSource {
    return &purposesDataSource{}
}
```

**Metadata:** TypeName = `req.ProviderTypeName + "_purposes"`

**Schema:**
- `eservice_ids` — Optional, List of strings (UUID validated)
- `title` — Optional, string
- `consumer_ids` — Optional, List of strings (UUID validated)
- `states` — Optional, List of strings with `stringvalidator.OneOf("DRAFT", "ACTIVE", "SUSPENDED", "ARCHIVED", "WAITING_FOR_APPROVAL", "REJECTED")`
- `purposes` — Computed, ListNestedAttribute with nested schema containing all purpose fields

**Nested schema** `purposeNestedSchema()`:
All computed attributes: id, eservice_id, consumer_id, title, description, daily_calls, state, is_free_of_charge, free_of_charge_reason, is_risk_analysis_valid, suspended_by_consumer, suspended_by_producer, delegation_id, version_id, created_at, updated_at.

**Nested model** (local struct in same file):
```go
type purposeItemModel struct {
    ID                  types.String `tfsdk:"id"`
    EServiceID          types.String `tfsdk:"eservice_id"`
    // ... same fields as PurposeDataSourceModel minus the filter fields
}
```

**Read:**
1. Parse filters from config
2. Build `api.ListPurposesParams` (parse UUID lists, extract optional title/states)
3. Auto-paginate: loop calling `d.client.ListPurposes(ctx, params)` until all fetched
4. Convert each `api.Purpose` to `purposeItemModel` using a `purposeToItemModel` helper
5. Set purposes list in state

### 4. Modify `internal/provider/provider.go` — Register data sources

Add to `DataSources()` returned slice:
```go
datasources.NewPurposeDataSource,
datasources.NewPurposesDataSource,
```

## Execution Steps

1. Read all referenced files
2. Create `models/purpose_datasource.go`
3. Create `datasources/purpose_data_source.go`
4. Create `datasources/purposes_data_source.go`
5. Modify `provider.go` to register both
6. Run `go build ./...`

## Verification

```shell
go build ./...
```

## Important Notes

- Do NOT break existing data sources or tests
- The `PurposesAPI` already has `GetPurpose` and `ListPurposes` — do NOT add new methods
- The `ProviderData` already has `PurposesAPI` field — do NOT modify `providerdata.go`
- State derivation: inline the logic (don't import from `resources` package). Check `currentVersion != nil` → use its state, else check `waitingForApprovalVersion` → "WAITING_FOR_APPROVAL", else "DRAFT"
- For daily_calls in the plural data source: if `currentVersion` is nil and `waitingForApprovalVersion` is nil, set `daily_calls` to `types.Int64Null()`
- The existing `agreement_purposes_data_source.go` has a `purposeToNestedModel` function. Do NOT reuse it — it doesn't handle version fields (daily_calls, state, version_id). Create new conversion functions.
- Use `stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-f]{8}-..."), "must be a valid UUID")` for UUID validation in filter lists. Actually, look at how the existing data sources validate UUIDs in filter lists — `eservices_data_source.go` uses `listvalidator.ValueStringsAre(stringvalidator.RegexMatches(...))`. Follow that pattern.
- The UUID regex is defined in `internal/resources/agreement_resource.go` as `uuidRegex`. Since data sources are in a different package (`datasources`), you may need to define a local regex or import it. Check how existing data sources handle this — they likely define their own or don't validate UUIDs in filters.
