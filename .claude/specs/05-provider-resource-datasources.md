# Agent 5: Provider + Resource + Data Sources + Unit Tests

## Objective

Implement the full provider configuration, `pdnd_agreement` resource with desired-state lifecycle, 3 data sources, and all resource unit tests. When done, `go test ./internal/... -v -run 'Test[^A]'` must pass (all unit tests, excluding acceptance tests).

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo already has:
- Generated client: `internal/client/generated/client.gen.go`
- Transport/auth: `internal/client/dpop.go`, `transport.go`, `retry.go`, `errors.go`
- API wrapper: `internal/client/api/agreements.go`, `types.go`, `conversions.go`
- Fake server: `internal/testing/fakepdnd/`
- Stub provider: `internal/provider/provider.go`

## Files to Create/Modify

### Modified Files
- `internal/provider/provider.go` — replace stub with full implementation
- `cmd/terraform-provider-pdnd/main.go` — may need minor updates

### New Files
- `internal/provider/config.go`
- `internal/models/agreement.go`
- `internal/models/common.go`
- `internal/resources/agreement_resource.go`
- `internal/resources/agreement_helpers.go`
- `internal/resources/agreement_state_machine.go`
- `internal/resources/agreement_import.go`
- `internal/resources/agreement_state_machine_test.go`
- `internal/resources/agreement_helpers_test.go`
- `internal/resources/agreement_import_test.go`
- `internal/datasources/agreement_data_source.go`
- `internal/datasources/agreements_data_source.go`
- `internal/datasources/agreement_purposes_data_source.go`

---

## 1. Provider Configuration

### `internal/provider/provider.go`

Replace the stub with a full provider implementing Protocol Version 6.

```go
package provider

var _ provider.Provider = &pdndProvider{}

type pdndProvider struct {
    version string
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &pdndProvider{version: version}
    }
}
```

**Metadata**: TypeName = `"pdnd"`, Version = p.version

**Schema**:
```go
schema.Schema{
    Description: "Terraform provider for PDND Interoperability API v3",
    Attributes: map[string]schema.Attribute{
        "base_url": schema.StringAttribute{
            Description: "Base URL of the PDND API (e.g., https://api.interop.pagopa.it/v3)",
            Required:    true,
        },
        "access_token": schema.StringAttribute{
            Description: "Access token for PDND API authentication",
            Required:    true,
            Sensitive:   true,
        },
        "dpop_private_key": schema.StringAttribute{
            Description: "PEM-encoded private key for DPoP proof generation",
            Required:    true,
            Sensitive:   true,
        },
        "dpop_key_id": schema.StringAttribute{
            Description: "Key ID for the DPoP private key",
            Required:    true,
        },
        "request_timeout_s": schema.Int64Attribute{
            Description: "Request timeout in seconds (default: 30)",
            Optional:    true,
        },
    },
}
```

**Configure** (in `internal/provider/config.go`):

```go
type pdndProviderModel struct {
    BaseURL         types.String `tfsdk:"base_url"`
    AccessToken     types.String `tfsdk:"access_token"`
    DPoPPrivateKey  types.String `tfsdk:"dpop_private_key"`
    DPoPKeyID       types.String `tfsdk:"dpop_key_id"`
    RequestTimeoutS types.Int64  `tfsdk:"request_timeout_s"`
}
```

Validation steps (eager, in Configure):
1. Read config into `pdndProviderModel`
2. Validate `base_url` is non-empty and a valid URL
3. Validate `access_token` is non-empty
4. Parse `dpop_private_key` as PEM — create `DPoPProofGenerator`; fail with diagnostic if invalid
5. Validate `dpop_key_id` is non-empty
6. Set `request_timeout_s` default to 30 if not provided; validate > 0
7. Build: `DPoPProofGenerator` → `DPoPTransport` → `RetryTransport` → `http.Client` (with timeout)
8. Create `generated.ClientWithResponses` with base URL and custom http.Client
9. Create `api.NewAgreementsClient(generatedClient)` → `AgreementsAPI`
10. Store in provider data

**Provider data**: Store a struct containing `AgreementsAPI` so resources/data sources can access it via `req.ProviderData`.

```go
type ProviderData struct {
    AgreementsAPI api.AgreementsAPI
}
```

**Resources**:
```go
func (p *pdndProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        resources.NewAgreementResource,
    }
}
```

**DataSources**:
```go
func (p *pdndProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        datasources.NewAgreementDataSource,
        datasources.NewAgreementsDataSource,
        datasources.NewAgreementPurposesDataSource,
    }
}
```

---

## 2. Terraform Models

### `internal/models/agreement.go`

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AgreementResourceModel is the Terraform state model for pdnd_agreement resource.
type AgreementResourceModel struct {
    ID                  types.String `tfsdk:"id"`
    EServiceID          types.String `tfsdk:"eservice_id"`
    DescriptorID        types.String `tfsdk:"descriptor_id"`
    DelegationID        types.String `tfsdk:"delegation_id"`
    DesiredState        types.String `tfsdk:"desired_state"`
    ConsumerNotes       types.String `tfsdk:"consumer_notes"`
    AllowPending        types.Bool   `tfsdk:"allow_pending"`
    State               types.String `tfsdk:"state"`
    ProducerID          types.String `tfsdk:"producer_id"`
    ConsumerID          types.String `tfsdk:"consumer_id"`
    SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
    SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
    SuspendedByPlatform types.Bool   `tfsdk:"suspended_by_platform"`
    RejectionReason     types.String `tfsdk:"rejection_reason"`
    CreatedAt           types.String `tfsdk:"created_at"`
    UpdatedAt           types.String `tfsdk:"updated_at"`
    SuspendedAt         types.String `tfsdk:"suspended_at"`
}

// AgreementDataSourceModel is the Terraform state model for pdnd_agreement data source.
type AgreementDataSourceModel struct {
    ID                  types.String `tfsdk:"id"`
    EServiceID          types.String `tfsdk:"eservice_id"`
    DescriptorID        types.String `tfsdk:"descriptor_id"`
    ProducerID          types.String `tfsdk:"producer_id"`
    ConsumerID          types.String `tfsdk:"consumer_id"`
    DelegationID        types.String `tfsdk:"delegation_id"`
    State               types.String `tfsdk:"state"`
    SuspendedByConsumer types.Bool   `tfsdk:"suspended_by_consumer"`
    SuspendedByProducer types.Bool   `tfsdk:"suspended_by_producer"`
    SuspendedByPlatform types.Bool   `tfsdk:"suspended_by_platform"`
    ConsumerNotes       types.String `tfsdk:"consumer_notes"`
    RejectionReason     types.String `tfsdk:"rejection_reason"`
    CreatedAt           types.String `tfsdk:"created_at"`
    UpdatedAt           types.String `tfsdk:"updated_at"`
    SuspendedAt         types.String `tfsdk:"suspended_at"`
}
```

---

## 3. `pdnd_agreement` Resource

### `internal/resources/agreement_resource.go`

```go
package resources

var _ resource.Resource = &agreementResource{}
var _ resource.ResourceWithImportState = &agreementResource{}

type agreementResource struct {
    client api.AgreementsAPI
}

func NewAgreementResource() resource.Resource {
    return &agreementResource{}
}
```

**Metadata**: TypeName = `"pdnd_agreement"`

**Schema attributes:**

| Attribute | Type | Required/Optional/Computed | Plan Modifiers | Validators | Notes |
|-----------|------|--------------------------|----------------|------------|-------|
| `id` | String | Computed | UseStateForUnknown | | Agreement UUID |
| `eservice_id` | String | Required | RequiresReplace | IsValidUUID | ForceNew |
| `descriptor_id` | String | Required | | IsValidUUID | Change triggers upgrade |
| `delegation_id` | String | Optional | RequiresReplace | IsValidUUID | ForceNew |
| `desired_state` | String | Required | | OneOf("DRAFT","ACTIVE","SUSPENDED") | Intent field |
| `consumer_notes` | String | Optional | | StringLenBetween(0,1000) | Sent on submit |
| `allow_pending` | Bool | Optional | | | Default: false |
| `state` | String | Computed | | | Observed server state |
| `producer_id` | String | Computed | UseStateForUnknown | | |
| `consumer_id` | String | Computed | UseStateForUnknown | | |
| `suspended_by_consumer` | Bool | Computed | | | |
| `suspended_by_producer` | Bool | Computed | | | |
| `suspended_by_platform` | Bool | Computed | | | |
| `rejection_reason` | String | Computed | | | |
| `created_at` | String | Computed | UseStateForUnknown | | |
| `updated_at` | String | Computed | | | |
| `suspended_at` | String | Computed | | | |

**Configure**: Extract `ProviderData` from `req.ProviderData`, get `AgreementsAPI`.

### Create

```
1. Read plan into AgreementResourceModel
2. Build AgreementSeed from plan (eservice_id, descriptor_id, delegation_id)
3. Call CreateAgreement → returns DRAFT agreement
4. If desired_state == "DRAFT":
   - Populate model from API response, save state, done
5. If desired_state == "ACTIVE":
   a. Build AgreementSubmission (consumer_notes if set)
   b. Call SubmitAgreement
   c. If response.State == "ACTIVE": populate model, save state, done
   d. If response.State == "PENDING":
      - If allow_pending == true: populate model, save state, done
      - If allow_pending == false:
        - Call DeleteAgreement to clean up
        - Add error diagnostic: "Agreement entered PENDING state but allow_pending is false"
        - Return
6. If desired_state == "SUSPENDED":
   - Add error diagnostic: "Cannot create agreement directly in SUSPENDED state"
   - Return
```

### Read

```
1. Read current state to get ID
2. Call GetAgreement(id)
3. If error is NotFound: resp.State.RemoveResource(), return
4. Populate all computed fields from API response
5. Preserve desired_state and allow_pending from current state (not from API)
6. Save state
```

### Update

```
1. Read plan and current state
2. Determine what changed:
   a. If descriptor_id changed AND state is ACTIVE or SUSPENDED:
      - Call UpgradeAgreement → returns NEW agreement with NEW ID
      - Update model with new ID and all new fields
      - Save state, done
   b. Compute state transitions via ComputeTransitions(currentState, desiredState)
   c. Execute transitions in order
   d. After transitions, call GetAgreement to refresh
   e. Save state
```

### Delete

```
1. Read current state to get ID and state
2. If canDelete(state):
   - Call DeleteAgreement
3. Else:
   - Add error diagnostic: "Cannot delete agreement {id} in state {state}.
     Only DRAFT, PENDING, and MISSING_CERTIFIED_ATTRIBUTES agreements can be deleted."
```

### ImportState

```
1. Extract ID from req.ID
2. Validate UUID format
3. Use resource.ImportStatePassthroughID for the "id" attribute
4. Read will fill all fields
5. After Read, infer desired_state from observed state:
   - ACTIVE → desired_state = "ACTIVE"
   - SUSPENDED → desired_state = "SUSPENDED"
   - DRAFT → desired_state = "DRAFT"
   - Others → error diagnostic
```

Note: The import logic for inferring `desired_state` should be handled in the Read method by checking if `desired_state` is null/unknown (which it will be after ImportState passthrough), and then setting it from the observed state.

### `internal/resources/agreement_helpers.go`

Conversion functions between `api.Agreement` and `models.AgreementResourceModel`:

```go
func populateModelFromAgreement(model *models.AgreementResourceModel, a *api.Agreement)
func canDelete(state string) bool
func inferDesiredState(observedState string) (string, error)
```

### `internal/resources/agreement_state_machine.go`

```go
package resources

type TransitionType string

const (
    TransitionSubmit    TransitionType = "submit"
    TransitionSuspend   TransitionType = "suspend"
    TransitionUnsuspend TransitionType = "unsuspend"
    TransitionUpgrade   TransitionType = "upgrade"
)

type Transition struct {
    Type TransitionType
}

// ComputeTransitions determines the API calls needed to move from current state to desired state.
func ComputeTransitions(currentState, desiredState string, descriptorChanged bool) ([]Transition, error)
```

**Transition rules:**

| Current State | Desired State | Descriptor Changed | Result |
|--------------|--------------|-------------------|--------|
| DRAFT | DRAFT | no | [] (no-op) |
| DRAFT | ACTIVE | no | [submit] |
| ACTIVE | ACTIVE | no | [] (no-op) |
| ACTIVE | ACTIVE | yes | [upgrade] |
| ACTIVE | SUSPENDED | no | [suspend] |
| SUSPENDED | SUSPENDED | no | [] (no-op) |
| SUSPENDED | SUSPENDED | yes | [upgrade] |
| SUSPENDED | ACTIVE | no | [unsuspend] |
| DRAFT | SUSPENDED | * | error: "Cannot transition from DRAFT to SUSPENDED" |
| PENDING | * | * | error: "Cannot transition from PENDING state" |
| REJECTED | * | * | error: "Cannot transition from REJECTED state" |
| ARCHIVED | * | * | error: "Cannot transition from ARCHIVED state" |

---

## 4. Data Sources

### `internal/datasources/agreement_data_source.go`

**Type**: `pdnd_agreement`

```hcl
data "pdnd_agreement" "example" {
  id = "agreement-uuid"
}
```

Schema: `id` is Required string. All other Agreement fields are Computed.
Read: Call `GetAgreement(id)`, populate all fields.

### `internal/datasources/agreements_data_source.go`

**Type**: `pdnd_agreements`

```hcl
data "pdnd_agreements" "example" {
  states       = ["ACTIVE"]
  producer_ids = ["uuid"]
}
```

Schema:
- `states` — Optional, list of strings
- `producer_ids` — Optional, list of strings
- `consumer_ids` — Optional, list of strings
- `eservice_ids` — Optional, list of strings
- `descriptor_ids` — Optional, list of strings
- `agreements` — Computed, list of objects (nested agreement attributes)

Read: Auto-paginate — loop calling `ListAgreements` with increasing offset until all results fetched.

### `internal/datasources/agreement_purposes_data_source.go`

**Type**: `pdnd_agreement_purposes`

```hcl
data "pdnd_agreement_purposes" "example" {
  agreement_id = "agreement-uuid"
}
```

Schema:
- `agreement_id` — Required string
- `purposes` — Computed, list of purpose objects

Purpose object attributes: `id`, `eservice_id`, `consumer_id`, `title`, `description`, `created_at`, `updated_at`, `is_risk_analysis_valid`, `is_free_of_charge`, `free_of_charge_reason`, `suspended_by_consumer`, `suspended_by_producer`, `delegation_id`

Read: Auto-paginate `ListAgreementPurposes`.

---

## 5. Unit Tests

### `internal/resources/agreement_state_machine_test.go`

Table-driven tests:

| Test | Current | Desired | DescChanged | Expected |
|------|---------|---------|------------|----------|
| `TestComputeTransitions_DraftToActive` | DRAFT | ACTIVE | false | [submit] |
| `TestComputeTransitions_ActiveToSuspended` | ACTIVE | SUSPENDED | false | [suspend] |
| `TestComputeTransitions_SuspendedToActive` | SUSPENDED | ACTIVE | false | [unsuspend] |
| `TestComputeTransitions_ActiveUpgrade` | ACTIVE | ACTIVE | true | [upgrade] |
| `TestComputeTransitions_SuspendedUpgrade` | SUSPENDED | SUSPENDED | true | [upgrade] |
| `TestComputeTransitions_DraftToSuspended_Error` | DRAFT | SUSPENDED | false | error |
| `TestComputeTransitions_PendingToAny_Error` | PENDING | ACTIVE | false | error |
| `TestComputeTransitions_NoOp_DraftDraft` | DRAFT | DRAFT | false | [] |
| `TestComputeTransitions_NoOp_ActiveActive` | ACTIVE | ACTIVE | false | [] |

### `internal/resources/agreement_helpers_test.go`

| Test | What to Assert |
|------|---------------|
| `TestCanDelete_Draft` | true |
| `TestCanDelete_Pending` | true |
| `TestCanDelete_MissingCertifiedAttributes` | true |
| `TestCanDelete_Active` | false |
| `TestCanDelete_Suspended` | false |
| `TestCanDelete_Archived` | false |
| `TestCanDelete_Rejected` | false |

### `internal/resources/agreement_import_test.go`

| Test | What to Assert |
|------|---------------|
| `TestInferDesiredState_Active` | "ACTIVE", nil |
| `TestInferDesiredState_Suspended` | "SUSPENDED", nil |
| `TestInferDesiredState_Draft` | "DRAFT", nil |
| `TestInferDesiredState_Pending` | "", error |
| `TestInferDesiredState_Rejected` | "", error |
| `TestInferDesiredState_Archived` | "", error |
| `TestInferDesiredState_MissingCertifiedAttributes` | "", error |

---

## Execution Steps

1. Read existing code: `internal/provider/provider.go`, `internal/client/api/agreements.go`, `internal/client/dpop.go`, `internal/client/transport.go`, `internal/client/errors.go`
2. Create `internal/models/agreement.go` and `internal/models/common.go`
3. Create resource files: `agreement_resource.go`, `agreement_helpers.go`, `agreement_state_machine.go`, `agreement_import.go`
4. Create data source files
5. Update `internal/provider/provider.go` with full configuration and resource/datasource registration
6. Create `internal/provider/config.go`
7. Create all test files
8. Run `go mod tidy`
9. Run `go test ./internal/resources/... -v` — unit tests
10. Run `go test ./internal/... -v -run 'Test[^A]'` — all non-acceptance tests
11. Run `go build ./...`
12. Fix and iterate until all pass

## Verification

- `go test ./internal/resources/... -v` — all resource unit tests pass
- `go test ./internal/... -v -run 'Test[^A]'` — all unit + contract tests pass
- `go build ./...` — compiles

## Important Notes

- Use Terraform Plugin Framework types (`types.String`, `types.Bool`, `types.Int64`)
- Use `UseStateForUnknown` plan modifier for computed fields that don't change (id, producer_id, consumer_id, created_at)
- The `id` attribute MUST be computed at the schema root level (framework requirement)
- `allow_pending` should default to `false` — use `booldefault.StaticBool(false)` and mark as `Computed: true` + `Optional: true`
- For UUID validation, use `stringvalidator` from `terraform-plugin-framework-validators`
- Import `github.com/hashicorp/terraform-plugin-framework-validators` if needed
- Read the `api.AgreementsAPI` interface to understand exact method signatures
- The provider must create the `generated.ClientWithResponses` correctly — read the generated code to see how
