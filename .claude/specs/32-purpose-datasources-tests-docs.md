# Spec 32: Tests and Documentation for Purpose Data Sources

## Objective

Add acceptance tests for `pdnd_purpose` and `pdnd_purposes` data sources, documentation templates, and examples. When done, `TF_ACC=1 go test ./internal/datasources/... -v -run "TestAccPurpose"` must pass and `make docs` must generate docs for both.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/datasources/eservice_data_source_acc_test.go` — singular data source test pattern
2. `internal/datasources/eservices_data_source_acc_test.go` — plural data source test pattern
3. `internal/datasources/agreement_purposes_data_source_acc_test.go` — existing purpose-related test (for seeding pattern)
4. `internal/testing/fakepdnd/server.go` — `SeedStandalonePurpose`, `standalonePurposes`
5. `internal/testing/fakepdnd/state.go` — `StoredPurpose`, `StoredPurposeVersion`
6. `internal/datasources/purpose_data_source.go` — singular data source (created in Spec 31)
7. `internal/datasources/purposes_data_source.go` — plural data source (created in Spec 31)
8. `templates/data-sources/eservice.md.tmpl` — doc template pattern for singular
9. `templates/data-sources/eservices.md.tmpl` — doc template pattern for plural
10. `examples/data-sources/pdnd_eservice/data-source.tf` — example pattern for singular
11. `examples/data-sources/pdnd_eservices/data-source.tf` — example pattern for plural
12. `README.md` — data sources table to update

Also check how test provider config and factories are defined:
13. `internal/datasources/agreement_data_source_acc_test.go` — look for `testAccProtoV6ProviderFactories` and `testAccProviderConfig` definitions in the datasources_test package

## Files to Create/Modify

### 1. Create `internal/datasources/purpose_data_source_acc_test.go`

Seed a standalone purpose with versions, then fetch by ID.

```go
func TestAccPurposeDataSource(t *testing.T) {
    fake := fakepdnd.NewFakeServer()
    purposeID := uuid.New()
    versionID := uuid.New()
    now := time.Now().UTC()

    fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
        ID:                  purposeID,
        EServiceID:          uuid.New(),
        ConsumerID:          fake.ConsumerID(),
        Title:               "Test Purpose",
        Description:         "A test purpose for data source",
        IsFreeOfCharge:      false,
        IsRiskAnalysisValid: true,
        CreatedAt:           now,
        Versions: []fakepdnd.StoredPurposeVersion{
            {
                ID:        versionID,
                State:     "ACTIVE",
                DailyCalls: 1000,
                CreatedAt: now,
                FirstActivationAt: &now,
            },
        },
    })

    ts := fake.Start()
    defer ts.Close()

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_purpose" "test" {
  id = %q
}
`, purposeID),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.pdnd_purpose.test", "title", "Test Purpose"),
                    resource.TestCheckResourceAttr("data.pdnd_purpose.test", "state", "ACTIVE"),
                    resource.TestCheckResourceAttr("data.pdnd_purpose.test", "daily_calls", "1000"),
                ),
            },
        },
    })
}
```

### 2. Create `internal/datasources/purposes_data_source_acc_test.go`

Seed multiple purposes, then list with filters.

```go
func TestAccPurposesDataSource(t *testing.T) {
    fake := fakepdnd.NewFakeServer()
    esID := uuid.New()
    now := time.Now().UTC()

    // Seed two purposes, different states
    fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
        ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
        Title: "Active Purpose", Description: "An active purpose",
        IsFreeOfCharge: false, IsRiskAnalysisValid: true, CreatedAt: now,
        Versions: []fakepdnd.StoredPurposeVersion{
            {ID: uuid.New(), State: "ACTIVE", DailyCalls: 1000, CreatedAt: now},
        },
    })
    fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
        ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
        Title: "Draft Purpose", Description: "A draft purpose",
        IsFreeOfCharge: true, IsRiskAnalysisValid: true, CreatedAt: now,
        Versions: []fakepdnd.StoredPurposeVersion{
            {ID: uuid.New(), State: "DRAFT", DailyCalls: 500, CreatedAt: now},
        },
    })

    ts := fake.Start()
    defer ts.Close()

    // Test: list all (no filters)
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: testAccProviderConfig(ts.URL) + `
data "pdnd_purposes" "test" {}
`,
                Check: resource.TestCheckResourceAttr("data.pdnd_purposes.test", "purposes.#", "2"),
            },
        },
    })
}

func TestAccPurposesDataSource_FilterByState(t *testing.T) {
    // Seed both ACTIVE and DRAFT, filter by states = ["ACTIVE"]
    // Verify only 1 result returned
}
```

### 3. Create `templates/data-sources/purpose.md.tmpl`

```markdown
---
page_title: "pdnd_purpose Data Source - PDND"
description: |-
  Fetches a single PDND purpose by ID.
---

# pdnd_purpose (Data Source)

Fetches a single PDND purpose by its UUID.

## Example Usage

{{ tffile "examples/data-sources/pdnd_purpose/data-source.tf" }}

{{ .SchemaMarkdown | trimspace }}
```

### 4. Create `templates/data-sources/purposes.md.tmpl`

```markdown
---
page_title: "pdnd_purposes Data Source - PDND"
description: |-
  Lists PDND purposes with optional filters.
---

# pdnd_purposes (Data Source)

Lists PDND purposes with optional filters by e-service IDs, title, consumer IDs, and states.

## Example Usage

{{ tffile "examples/data-sources/pdnd_purposes/data-source.tf" }}

{{ .SchemaMarkdown | trimspace }}
```

### 5. Create `examples/data-sources/pdnd_purpose/data-source.tf`

```hcl
data "pdnd_purpose" "example" {
  id = "990e8400-e29b-41d4-a716-446655440000"
}

output "purpose_title" {
  value = data.pdnd_purpose.example.title
}
```

### 6. Create `examples/data-sources/pdnd_purposes/data-source.tf`

```hcl
# List all active purposes for a specific e-service
data "pdnd_purposes" "active" {
  eservice_ids = [pdnd_eservice.example.id]
  states       = ["ACTIVE"]
}

output "active_purpose_count" {
  value = length(data.pdnd_purposes.active.purposes)
}
```

### 7. Modify `README.md` — Add rows to data sources table

Add after `pdnd_agreement_purposes`:
```
| `pdnd_purpose`               | Fetches a single purpose by ID                   |
| `pdnd_purposes`              | Lists purposes with optional filters             |
```

## Execution Steps

1. Read all referenced files for patterns
2. Create example directories and files
3. Create doc templates
4. Create `purpose_data_source_acc_test.go`
5. Create `purposes_data_source_acc_test.go`
6. Update `README.md`
7. Run `TF_ACC=1 go test ./internal/datasources/... -v -run "TestAccPurpose"`
8. Run full test suite: `go test ./...`
9. Run `make lint`
10. Run `make docs`

## Verification

```shell
TF_ACC=1 go test ./internal/datasources/... -v -run "TestAccPurpose"
go test ./...
make lint
make docs
```

## Important Notes

- Do NOT break existing tests
- `testAccProtoV6ProviderFactories` and `testAccProviderConfig` are defined in the `datasources_test` package (check `agreement_data_source_acc_test.go` or similar for the exact definition)
- `SeedStandalonePurpose` takes a `StoredPurpose` struct which must include `Versions` for version-dependent fields to be populated
- Use `fake.ConsumerID()` for the consumer ID (consistent with other tests)
- For the plural data source test with state filter, the fake server's `handleListPurposes` derives state from versions using `deriveStoredPurposeState`
- Create example directories with `mkdir -p` before writing files
