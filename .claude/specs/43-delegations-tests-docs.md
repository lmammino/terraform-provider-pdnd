# Spec 43: Delegation Tests + Documentation

## Objective

Add acceptance tests for 2 delegation resources and 4 data sources, plus doc templates, examples, and README update. When done, all tests pass and `make docs` generates documentation for all new resources/data sources.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/agreement_resource_acc_test.go` — resource acc test pattern, `testAccProviderConfig` (line 41), `testAccProtoV6ProviderFactories` (line 35), tracker cleanup pattern
2. `internal/resources/purpose_resource_acc_test.go` — resource acc test with fake server seeding
3. `internal/datasources/purpose_data_source_acc_test.go` — singular data source test with seeded data
4. `internal/datasources/purposes_data_source_acc_test.go` — plural data source test with filters
5. `internal/testing/fakepdnd/server.go` — `SeedDelegation`, `GetDelegation` methods (added in Spec 41)
6. `internal/testing/fakepdnd/state.go` — `StoredDelegation` type
7. `templates/resources/agreement.md.tmpl` — doc template format
8. `templates/data-sources/purpose.md.tmpl` — data source doc template format
9. `examples/resources/pdnd_agreement/resource.tf` — example format
10. `examples/data-sources/pdnd_purpose/data-source.tf` — data source example format
11. `README.md` — resources and data sources tables to update

Also check how acc test helpers are available:
12. `internal/datasources/agreement_data_source_acc_test.go` — `testAccProviderConfig` and `testAccProtoV6ProviderFactories` in `datasources_test` package

## Files to Create/Modify

### 1. Create `internal/resources/consumer_delegation_resource_acc_test.go`

**Tests:**

**`TestAccConsumerDelegation_Create`:** Create a consumer delegation, verify state=WAITING_FOR_APPROVAL, computed fields populated.
```go
fake := fakepdnd.NewFakeServer()
ts := fake.Start()
defer ts.Close()

delegateID := uuid.New()
esID := uuid.New()
fake.SeedEService(fakepdnd.StoredEService{
    ID: esID, ProducerID: fake.ProducerID(), Name: "Test",
    Description: "test", Technology: "REST", Mode: "DELIVER",
})

resource.Test(t, resource.TestCase{
    ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
    Steps: []resource.TestStep{
        {
            Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_consumer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
            Check: resource.ComposeAggregateTestCheckFunc(
                resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "id"),
                resource.TestCheckResourceAttr("pdnd_consumer_delegation.test", "state", "WAITING_FOR_APPROVAL"),
                resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "delegator_id"),
                resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "created_at"),
                resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "submitted_at"),
            ),
        },
    },
})
```

**`TestAccConsumerDelegation_ExternalAccept`:** Create delegation, then externally accept via fake server, refresh, verify state=ACTIVE.
```go
// Step 1: Create delegation
// Step 2: PreConfig function that calls fake.GetDelegation and mutates state to ACTIVE
// Step 3: RefreshState: true, verify state=ACTIVE
```

**`TestAccConsumerDelegation_Import`:** Create, then import by ID.
```go
{
    ResourceName:            "pdnd_consumer_delegation.test",
    ImportState:             true,
    ImportStateVerify:       true,
}
```

**`TestAccConsumerDelegation_Destroy`:** Verify destroy removes from state without error (no API delete).

### 2. Create `internal/resources/producer_delegation_resource_acc_test.go`

Minimal — same shared logic. Just test Create + Import:

**`TestAccProducerDelegation_Create`:** Verify state=WAITING_FOR_APPROVAL.
**`TestAccProducerDelegation_Import`:** Import by ID.

### 3. Create `internal/datasources/consumer_delegation_data_source_acc_test.go`

**`TestAccConsumerDelegationDataSource`:** Seed a consumer delegation, fetch by ID, verify fields.

### 4. Create `internal/datasources/consumer_delegations_data_source_acc_test.go`

**`TestAccConsumerDelegationsDataSource`:** Seed multiple, list all, verify count.
**`TestAccConsumerDelegationsDataSource_FilterByState`:** Seed ACTIVE + WAITING_FOR_APPROVAL, filter by states=["ACTIVE"], verify only 1 result.

### 5. Create `internal/datasources/producer_delegation_data_source_acc_test.go`

**`TestAccProducerDelegationDataSource`:** Seed + fetch by ID.

### 6. Create `internal/datasources/producer_delegations_data_source_acc_test.go`

**`TestAccProducerDelegationsDataSource`:** Seed + list.

### 7. Create doc templates (6 files)

**`templates/resources/consumer_delegation.md.tmpl`:**
```markdown
---
page_title: "pdnd_consumer_delegation Resource - PDND"
description: |-
  Manages a PDND consumer delegation.
---

# pdnd_consumer_delegation (Resource)

Creates a PDND consumer delegation. The delegation starts in WAITING_FOR_APPROVAL state and must be accepted by the delegate.

Note: Delegations cannot be deleted via the API. Destroying this resource removes it from Terraform state only.

## Example Usage

{{ tffile "examples/resources/pdnd_consumer_delegation/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "shell" "examples/resources/pdnd_consumer_delegation/import.sh" }}
```

**`templates/resources/producer_delegation.md.tmpl`:** Same pattern for producer.

**`templates/data-sources/consumer_delegation.md.tmpl`**
**`templates/data-sources/consumer_delegations.md.tmpl`**
**`templates/data-sources/producer_delegation.md.tmpl`**
**`templates/data-sources/producer_delegations.md.tmpl`**

### 8. Create examples (8 files in 6 directories)

**`examples/resources/pdnd_consumer_delegation/resource.tf`:**
```hcl
resource "pdnd_consumer_delegation" "example" {
  eservice_id = pdnd_eservice.example.id
  delegate_id = "550e8400-e29b-41d4-a716-446655440000"
}
```

**`examples/resources/pdnd_consumer_delegation/import.sh`:**
```shell
terraform import pdnd_consumer_delegation.example 990e8400-e29b-41d4-a716-446655440000
```

**`examples/resources/pdnd_producer_delegation/resource.tf`** + `import.sh`

**`examples/data-sources/pdnd_consumer_delegation/data-source.tf`:**
```hcl
data "pdnd_consumer_delegation" "example" {
  id = "990e8400-e29b-41d4-a716-446655440000"
}
```

**`examples/data-sources/pdnd_consumer_delegations/data-source.tf`:**
```hcl
data "pdnd_consumer_delegations" "active" {
  states = ["ACTIVE"]
}
```

**`examples/data-sources/pdnd_producer_delegation/data-source.tf`**
**`examples/data-sources/pdnd_producer_delegations/data-source.tf`**

### 9. Modify `README.md`

Add to resources table:
```
| `pdnd_consumer_delegation`   | Manages a PDND consumer delegation            |
| `pdnd_producer_delegation`   | Manages a PDND producer delegation            |
```

Add to data sources table:
```
| `pdnd_consumer_delegation`   | Fetches a single consumer delegation by ID    |
| `pdnd_consumer_delegations`  | Lists consumer delegations with filters       |
| `pdnd_producer_delegation`   | Fetches a single producer delegation by ID    |
| `pdnd_producer_delegations`  | Lists producer delegations with filters       |
```

## Execution Steps

1. Read all referenced files
2. Create example directories (`mkdir -p`) and files
3. Create doc templates
4. Create resource acceptance tests
5. Create data source acceptance tests
6. Update README.md
7. Run `TF_ACC=1 go test ./internal/resources/... -v -run "Delegation" -timeout 5m`
8. Run `TF_ACC=1 go test ./internal/datasources/... -v -run "Delegation" -timeout 5m`
9. Run `go test ./...`
10. Run `make lint`
11. Run `make docs`

## Verification

```shell
TF_ACC=1 go test ./internal/resources/... -v -run "Delegation" -timeout 5m
TF_ACC=1 go test ./internal/datasources/... -v -run "Delegation" -timeout 5m
go test ./...
make lint
make docs
```

## Important Notes

- Do NOT break existing tests
- `testAccProviderConfig` and `testAccProtoV6ProviderFactories` are already defined — use them, don't redefine
- For the ExternalAccept test: in the PreConfig func, get the delegation from the fake server, mutate its State to "ACTIVE" and set ActivatedAt, then re-seed it. This simulates the delegate accepting outside Terraform.
- Destroy tests: delegation destroy should succeed without errors. The resource removes from state and emits a warning. The test verifies the step completes without error.
- For data source seeding: use `fake.SeedDelegation("consumer", StoredDelegation{...})` — remember to set all required fields (ID, DelegatorID, DelegateID, EServiceID, State, CreatedAt, SubmittedAt)
- Create example directories with `mkdir -p` before writing files
