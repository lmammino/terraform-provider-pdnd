# Agent 1: E-Service API Wrapper + Contract Tests

## Objective

Add `EServicesAPI` interface and implementation to the existing API wrapper layer, with contract tests. When done, `go test ./internal/client/api/... -v` must pass all existing + new tests.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/agreements.go` — pattern to follow (interface + `agreementsClient` struct)
2. `internal/client/api/types.go` — existing domain types
3. `internal/client/api/conversions.go` — existing conversion helpers
4. `internal/client/api/agreements_test.go` — contract test pattern
5. `internal/client/generated/client.gen.go` — generated client (find exact method names for eservice endpoints)
6. `internal/client/errors.go` — `CheckResponse`, `IsNotFound`, `IsConflict`

## Files to Create/Modify

### 1. Extend `internal/client/api/types.go` — Add E-Service Domain Types

```go
// EService represents a PDND e-service.
type EService struct {
    ID                      uuid.UUID
    ProducerID              uuid.UUID
    Name                    string
    Description             string
    Technology              string   // "REST" or "SOAP"
    Mode                    string   // "RECEIVE" or "DELIVER"
    IsSignalHubEnabled      *bool
    IsConsumerDelegable     *bool
    IsClientAccessDelegable *bool
    PersonalData            *bool
    TemplateID              *uuid.UUID
}

type EServiceSeed struct {
    Name                    string
    Description             string
    Technology              string
    Mode                    string
    IsSignalHubEnabled      *bool
    IsConsumerDelegable     *bool
    IsClientAccessDelegable *bool
    PersonalData            *bool
    Descriptor              DescriptorSeedForCreation
}

type DescriptorSeedForCreation struct {
    AgreementApprovalPolicy string
    Audience                []string
    DailyCallsPerConsumer   int32
    DailyCallsTotal         int32
    VoucherLifespan         int32
    Description             *string
}

type EServiceDraftUpdate struct {
    Name                    *string
    Description             *string
    Technology              *string
    Mode                    *string
    IsSignalHubEnabled      *bool
    IsConsumerDelegable     *bool
    IsClientAccessDelegable *bool
    PersonalData            *bool
}

type EServiceDelegationUpdate struct {
    IsConsumerDelegable     *bool
    IsClientAccessDelegable *bool
}

type Descriptor struct {
    ID                      uuid.UUID
    Version                 string
    State                   string   // DRAFT, PUBLISHED, DEPRECATED, SUSPENDED, ARCHIVED, WAITING_FOR_APPROVAL
    AgreementApprovalPolicy string
    Audience                []string
    DailyCallsPerConsumer   int32
    DailyCallsTotal         int32
    VoucherLifespan         int32
    ServerUrls              []string
    Description             *string
    PublishedAt             *time.Time
    SuspendedAt             *time.Time
    DeprecatedAt            *time.Time
    ArchivedAt              *time.Time
}

type DescriptorSeed struct {
    AgreementApprovalPolicy string
    Audience                []string
    DailyCallsPerConsumer   int32
    DailyCallsTotal         int32
    VoucherLifespan         int32
    Description             *string
}

type DescriptorDraftUpdate struct {
    AgreementApprovalPolicy *string
    Audience                []string   // nil means no change
    DailyCallsPerConsumer   *int32
    DailyCallsTotal         *int32
    VoucherLifespan         *int32
    Description             *string
}

type DescriptorQuotasUpdate struct {
    DailyCallsPerConsumer *int32
    DailyCallsTotal       *int32
    VoucherLifespan       *int32
}

type ListEServicesParams struct {
    ProducerIDs []uuid.UUID
    Name        *string
    Technology  *string
    Mode        *string
    Offset      int32
    Limit       int32
}

type EServicesPage struct {
    Results    []EService
    Pagination Pagination
}

type ListDescriptorsParams struct {
    State  *string
    Offset int32
    Limit  int32
}

type DescriptorsPage struct {
    Results    []Descriptor
    Pagination Pagination
}
```

### 2. Create `internal/client/api/eservices.go` — Interface + Implementation

```go
type EServicesAPI interface {
    // E-Service CRUD
    CreateEService(ctx context.Context, seed EServiceSeed) (*EService, error)
    GetEService(ctx context.Context, id uuid.UUID) (*EService, error)
    ListEServices(ctx context.Context, params ListEServicesParams) (*EServicesPage, error)
    DeleteEService(ctx context.Context, id uuid.UUID) error
    UpdateDraftEService(ctx context.Context, id uuid.UUID, seed EServiceDraftUpdate) (*EService, error)

    // Published e-service updates (per-field endpoints)
    UpdatePublishedEServiceName(ctx context.Context, id uuid.UUID, name string) (*EService, error)
    UpdatePublishedEServiceDescription(ctx context.Context, id uuid.UUID, description string) (*EService, error)
    UpdatePublishedEServiceDelegation(ctx context.Context, id uuid.UUID, seed EServiceDelegationUpdate) (*EService, error)
    UpdatePublishedEServiceSignalHub(ctx context.Context, id uuid.UUID, enabled bool) (*EService, error)

    // Descriptor CRUD
    CreateDescriptor(ctx context.Context, eserviceID uuid.UUID, seed DescriptorSeed) (*Descriptor, error)
    GetDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) (*Descriptor, error)
    ListDescriptors(ctx context.Context, eserviceID uuid.UUID, params ListDescriptorsParams) (*DescriptorsPage, error)
    DeleteDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
    UpdateDraftDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorDraftUpdate) (*Descriptor, error)
    UpdatePublishedDescriptorQuotas(ctx context.Context, eserviceID, descriptorID uuid.UUID, seed DescriptorQuotasUpdate) (*Descriptor, error)

    // Descriptor state transitions
    PublishDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
    SuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
    UnsuspendDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
    ApproveDelegatedDescriptor(ctx context.Context, eserviceID, descriptorID uuid.UUID) error
}
```

Concrete implementation: `eservicesClient` struct holding `*generated.APIClientWithResponses`.

```go
func NewEServicesClient(client *generated.APIClientWithResponses) EServicesAPI {
    return &eservicesClient{client: client}
}
```

**API Endpoint Mapping** (check generated code for exact method names):

| Method | HTTP | Path | Body | Success |
|--------|------|------|------|---------|
| CreateEService | POST | /eservices | EServiceSeed JSON | 200 |
| GetEService | GET | /eservices/{id} | none | 200 |
| ListEServices | GET | /eservices | query params | 200 |
| DeleteEService | DELETE | /eservices/{id} | none | 200 |
| UpdateDraftEService | PATCH | /eservices/{id} | merge-patch JSON | 200 |
| UpdatePublishedEServiceName | PATCH | /eservices/{id}/name | merge-patch JSON | 200 |
| UpdatePublishedEServiceDescription | PATCH | /eservices/{id}/description | merge-patch JSON | 200 |
| UpdatePublishedEServiceDelegation | PATCH | /eservices/{id}/delegation | merge-patch JSON | 200 |
| UpdatePublishedEServiceSignalHub | PATCH | /eservices/{id}/signalHub | merge-patch JSON | 200 |
| CreateDescriptor | POST | /eservices/{id}/descriptors | DescriptorSeed JSON | 200 |
| GetDescriptor | GET | /eservices/{id}/descriptors/{did} | none | 200 |
| ListDescriptors | GET | /eservices/{id}/descriptors | query params | 200 |
| DeleteDraftDescriptor | DELETE | /eservices/{id}/descriptors/{did} | none | 200 |
| UpdateDraftDescriptor | PATCH | /eservices/{id}/descriptors/{did} | merge-patch JSON | 200 |
| UpdatePublishedDescriptorQuotas | PATCH | /eservices/{id}/descriptors/{did}/quotas | merge-patch JSON | 200 |
| PublishDescriptor | POST | /eservices/{id}/descriptors/{did}/publish | none | 204 or 200 |
| SuspendDescriptor | POST | /eservices/{id}/descriptors/{did}/suspend | none | 204 or 200 |
| UnsuspendDescriptor | POST | /eservices/{id}/descriptors/{did}/unsuspend | none | 204 or 200 |
| ApproveDelegatedDescriptor | POST | /eservices/{id}/descriptors/{did}/approve | none | 204 or 200 |

**IMPORTANT**: The PATCH endpoints use `application/merge-patch+json` content type. The generated client will have methods like `UpdateDraftEServiceWithApplicationMergePatchPlusJSONBody` or similar. Read the generated code carefully to find the correct method variants.

### 3. Extend `internal/client/api/conversions.go` — Add Conversion Functions

```go
func eserviceFromGenerated(g *generated.EService) *EService
func eserviceSeedToGenerated(s EServiceSeed) generated.EServiceSeed
func descriptorFromGenerated(g *generated.EServiceDescriptor) *Descriptor
func descriptorSeedToGenerated(s DescriptorSeed) generated.EServiceDescriptorSeed
func eserviceDraftUpdateToGenerated(s EServiceDraftUpdate) generated.EServiceDraftUpdateSeed
func descriptorDraftUpdateToGenerated(s DescriptorDraftUpdate) generated.EServiceDescriptorDraftUpdateSeed
func descriptorQuotasUpdateToGenerated(s DescriptorQuotasUpdate) generated.EServiceDescriptorQuotasUpdateSeed
```

Handle UUID ↔ openapi_types.UUID, pointers, enum string types. Follow the exact same pattern as the agreement conversions.

### 4. Create `internal/client/api/eservices_test.go` — Contract Tests

Follow the exact pattern from `agreements_test.go`: httptest.Server records request, returns canned response.

**Required tests:**

| Test | Method | Path | Body |
|------|--------|------|------|
| TestCreateEService_Contract | POST | /eservices | EServiceSeed JSON |
| TestGetEService_Contract | GET | /eservices/{uuid} | none |
| TestListEServices_Contract | GET | /eservices?offset=0&limit=50 | query params |
| TestDeleteEService_Contract | DELETE | /eservices/{uuid} | none |
| TestUpdateDraftEService_Contract | PATCH | /eservices/{uuid} | merge-patch JSON |
| TestUpdatePublishedEServiceName_Contract | PATCH | /eservices/{uuid}/name | name JSON |
| TestUpdatePublishedEServiceDescription_Contract | PATCH | /eservices/{uuid}/description | description JSON |
| TestCreateDescriptor_Contract | POST | /eservices/{uuid}/descriptors | DescriptorSeed JSON |
| TestGetDescriptor_Contract | GET | /eservices/{uuid}/descriptors/{uuid} | none |
| TestListDescriptors_Contract | GET | /eservices/{uuid}/descriptors?offset=0&limit=50 | query params |
| TestDeleteDraftDescriptor_Contract | DELETE | /eservices/{uuid}/descriptors/{uuid} | none |
| TestPublishDescriptor_Contract | POST | /eservices/{uuid}/descriptors/{uuid}/publish | none |
| TestSuspendDescriptor_Contract | POST | /eservices/{uuid}/descriptors/{uuid}/suspend | none |
| TestUnsuspendDescriptor_Contract | POST | /eservices/{uuid}/descriptors/{uuid}/unsuspend | none |

**Canned EService JSON:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "producerId": "880e8400-e29b-41d4-a716-446655440003",
  "name": "Test E-Service",
  "description": "A test e-service description",
  "technology": "REST",
  "mode": "DELIVER"
}
```

**Canned Descriptor JSON:**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "version": "1",
  "state": "DRAFT",
  "audience": ["api.example.com"],
  "voucherLifespan": 3600,
  "dailyCallsPerConsumer": 1000,
  "dailyCallsTotal": 10000,
  "agreementApprovalPolicy": "AUTOMATIC",
  "serverUrls": ["https://api.example.com"]
}
```

## Execution Steps

1. Read the generated client code to find exact method names for eservice/descriptor endpoints
2. Add domain types to `types.go`
3. Add conversions to `conversions.go`
4. Create `eservices.go` with interface + implementation
5. Create `eservices_test.go` with all contract tests
6. Run `go mod tidy`
7. Run `TMPDIR=/tmp go test ./internal/client/api/... -v`
8. Iterate until ALL tests pass (both old agreement tests and new eservice tests)
9. Run `TMPDIR=/tmp go build ./...`

## Important Notes

- Do NOT break existing agreement tests
- PATCH endpoints use `application/merge-patch+json` — use the correct generated method variants
- The state transition endpoints (publish, suspend, unsuspend, approve) may return void or the updated descriptor — check the generated code
- All existing patterns from `agreements.go` apply: domain types → generated types → API call → check response → convert back
