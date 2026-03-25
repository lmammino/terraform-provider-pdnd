# Agent 6: Acceptance Tests + Docs + CI

## Objective

Write all Terraform acceptance tests against the fake PDND server, add documentation (templates + examples for tfplugindocs), and CI workflow. When done:
- `TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m` — all 11+ acceptance tests pass
- `go build ./...` — compiles
- All examples are valid HCL

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo has all implementation complete:
- Provider: `internal/provider/`
- Resource: `internal/resources/agreement_resource.go`
- Data sources: `internal/datasources/`
- Fake server: `internal/testing/fakepdnd/`
- API wrapper: `internal/client/api/`
- Transport: `internal/client/`

## Files to Create

### Acceptance Test Files
- `internal/resources/agreement_resource_acc_test.go`
- `internal/datasources/agreement_data_source_acc_test.go`
- `internal/datasources/agreements_data_source_acc_test.go`
- `internal/datasources/agreement_purposes_data_source_acc_test.go`

### Documentation Templates (for tfplugindocs)
- `templates/index.md.tmpl`
- `templates/resources/agreement.md.tmpl`
- `templates/data-sources/agreement.md.tmpl`
- `templates/data-sources/agreements.md.tmpl`
- `templates/data-sources/agreement_purposes.md.tmpl`

### Examples (used by tfplugindocs and as standalone examples)
- `examples/provider/provider.tf`
- `examples/resources/pdnd_agreement/resource.tf`
- `examples/resources/pdnd_agreement/import.sh`
- `examples/data-sources/pdnd_agreement/data-source.tf`
- `examples/data-sources/pdnd_agreements/data-source.tf`
- `examples/data-sources/pdnd_agreement_purposes/data-source.tf`

### Other
- `README.md`
- `.github/workflows/ci.yml`

---

## 1. Acceptance Test Infrastructure

### Test Helper Setup

Read the existing provider and fake server code first:
- `internal/provider/provider.go` — to understand `New()` function signature
- `internal/testing/fakepdnd/server.go` — to understand `FakeServer` API

```go
package resources_test  // or datasources_test

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/pagopa/terraform-provider-pdnd/internal/provider"
    "github.com/pagopa/terraform-provider-pdnd/internal/testing/fakepdnd"
)

// testAccProtoV6ProviderFactories returns provider factories for acceptance tests.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "pdnd": providerserver.NewProtocol6WithError(provider.New("test")()),
    }
}
```

### Test RSA Key

Generate a fixed test RSA key for DPoP in tests. You can generate it at init time or use a constant PEM string:

```go
var testPrivateKeyPEM string

func init() {
    key, _ := rsa.GenerateKey(rand.Reader, 2048)
    pemBytes := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(key),
    })
    testPrivateKeyPEM = string(pemBytes)
}
```

### Provider Config Helper

```go
func testAccProviderConfig(serverURL string) string {
    return fmt.Sprintf(`
provider "pdnd" {
  base_url         = %q
  access_token     = "test-access-token"
  dpop_private_key = <<-EOT
%sEOT
  dpop_key_id      = "test-key-id"
}
`, serverURL, testPrivateKeyPEM)
}
```

### Common Test Pattern

Each acceptance test:
1. Creates a `fakepdnd.NewFakeServer()`
2. Configures it (approval policy, seeds, etc.)
3. Starts it: `ts := fake.Start(); defer ts.Close()`
4. Runs `resource.Test(t, resource.TestCase{...})` with steps

The fake server needs to be accessible from outside the test function. Since `resource.Test` runs the provider in a separate goroutine, the fake server must be started before and stay running throughout.

**Important**: The fake server doesn't validate DPoP headers, so our transport layer will send them but they won't be checked — this is correct for acceptance testing.

---

## 2. Required Acceptance Tests

### `internal/resources/agreement_resource_acc_test.go`

#### TestAccAgreement_CreateDraft

```go
func TestAccAgreement_CreateDraft(t *testing.T) {
    fake := fakepdnd.NewFakeServer()
    ts := fake.Start()
    defer ts.Close()

    // Use two fixed UUIDs for eservice and descriptor
    eserviceID := uuid.New().String()
    descriptorID := uuid.New().String()

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "DRAFT"
}
`, eserviceID, descriptorID),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "DRAFT"),
                    resource.TestCheckResourceAttr("pdnd_agreement.test", "desired_state", "DRAFT"),
                    resource.TestCheckResourceAttrSet("pdnd_agreement.test", "id"),
                    resource.TestCheckResourceAttrSet("pdnd_agreement.test", "producer_id"),
                    resource.TestCheckResourceAttrSet("pdnd_agreement.test", "consumer_id"),
                    resource.TestCheckResourceAttrSet("pdnd_agreement.test", "created_at"),
                ),
            },
        },
    })
}
```

#### TestAccAgreement_CreateActive_HappyPath
- Fake server with AUTOMATIC approval policy (default)
- Create with desired_state="ACTIVE"
- Check state="ACTIVE"

#### TestAccAgreement_CreateActive_Pending_Allowed
- Set fake server approval policy to MANUAL
- Create with desired_state="ACTIVE", allow_pending=true
- Check state="PENDING" (no error)

#### TestAccAgreement_CreateActive_Pending_Forbidden
- Set fake server approval policy to MANUAL
- Create with desired_state="ACTIVE", allow_pending=false (default)
- Use `ExpectError: regexp.MustCompile("PENDING")` to expect failure

#### TestAccAgreement_UpdateActiveToSuspended
- Step 1: Create ACTIVE agreement
- Step 2: Change desired_state to "SUSPENDED"
- Check state="SUSPENDED" after step 2

```go
Steps: []resource.TestStep{
    {
        Config: /* desired_state = "ACTIVE" */,
        Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
    },
    {
        Config: /* desired_state = "SUSPENDED" */,
        Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "SUSPENDED"),
    },
},
```

#### TestAccAgreement_UpdateSuspendedToActive
- Step 1: Create ACTIVE
- Step 2: Change to SUSPENDED
- Step 3: Change back to ACTIVE
- Check state="ACTIVE" after step 3

#### TestAccAgreement_DestroyDraft
- Create DRAFT agreement
- Let the test framework destroy it (Terraform will call Delete on cleanup)
- Verify no error during destroy

This should just work — the test framework destroys resources after the last step. Use `CheckDestroy` to verify the agreement is gone.

```go
CheckDestroy: func(s *terraform.State) error {
    // Verify agreement no longer exists in fake server
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "pdnd_agreement" { continue }
        id, _ := uuid.Parse(rs.Primary.ID)
        if fake.GetAgreement(id) != nil {
            return fmt.Errorf("agreement %s still exists", rs.Primary.ID)
        }
    }
    return nil
},
```

#### TestAccAgreement_DestroyActiveFails
- Create ACTIVE agreement
- Attempt to destroy — this should fail because ACTIVE agreements can't be deleted
- Use `ExpectError` on a destroy step, or handle via a separate test that checks error on destroy

**Pattern for testing destroy failure**: Use a multi-step test where the last step tries to destroy by setting the config to empty, with `ExpectError`.

Actually, the simpler approach: use `resource.Test` with `Steps` that end with an `ExpectDestroyError: true` or catch it through the `CheckDestroy` returning an error. However, the standard Terraform test framework runs `Destroy` after all steps automatically.

Better approach: Create the active agreement, then in a second step remove it from config — this triggers destroy during the apply, which should fail:

```go
Steps: []resource.TestStep{
    {
        Config: /* create ACTIVE agreement */,
        Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
    },
    {
        Config: testAccProviderConfig(ts.URL), // empty config - triggers destroy
        ExpectError: regexp.MustCompile(`Cannot delete agreement`),
    },
},
```

**Note**: This test will leave the resource in state. You may need `RefreshState: true` or handle cleanup differently.

#### TestAccAgreement_Import
- Step 1: Create an ACTIVE agreement
- Step 2: Import it by ID

```go
Steps: []resource.TestStep{
    {
        Config: /* create ACTIVE */,
    },
    {
        ResourceName:      "pdnd_agreement.test",
        ImportState:       true,
        ImportStateVerify: true,
        // Fields that won't match exactly after import:
        ImportStateVerifyIgnore: []string{"allow_pending", "consumer_notes"},
    },
},
```

### `internal/datasources/agreement_data_source_acc_test.go`

(can be combined into agreements_data_source_acc_test.go if preferred)

### `internal/datasources/agreements_data_source_acc_test.go`

#### TestAccAgreementsDataSource
- Create 2 agreements via the resource (one ACTIVE, one DRAFT)
- Use data source to list with states=["ACTIVE"]
- Check only 1 result returned

```go
Steps: []resource.TestStep{
    {
        Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_agreement" "active" {
  eservice_id   = "..."
  descriptor_id = "..."
  desired_state = "ACTIVE"
}

resource "pdnd_agreement" "draft" {
  eservice_id   = "..."
  descriptor_id = "..."
  desired_state = "DRAFT"
}

data "pdnd_agreements" "active_only" {
  states = ["ACTIVE"]
  depends_on = [pdnd_agreement.active, pdnd_agreement.draft]
}
`,
        Check: resource.ComposeAggregateTestCheckFunc(
            resource.TestCheckResourceAttr("data.pdnd_agreements.active_only", "agreements.#", "1"),
        ),
    },
},
```

### `internal/datasources/agreement_purposes_data_source_acc_test.go`

#### TestAccAgreementPurposesDataSource
- Seed a purpose in the fake server
- Create an agreement via resource
- Use data source to read purposes

Since purposes are seeded in the fake server and the agreement ID is only known after creation, we need a two-step approach:
1. Create the agreement, capture its ID
2. Seed purposes for that agreement ID
3. Read via data source

Alternative: Seed the agreement directly in the fake server, then use the data source to read it.

```go
func TestAccAgreementPurposesDataSource(t *testing.T) {
    fake := fakepdnd.NewFakeServer()

    // Pre-seed an agreement and its purposes
    agreementID := uuid.New()
    fake.SeedAgreement(fakepdnd.StoredAgreement{
        ID: agreementID,
        EServiceID: uuid.New(),
        DescriptorID: uuid.New(),
        ProducerID: fake.ProducerID(),
        ConsumerID: fake.ConsumerID(),
        State: "ACTIVE",
        CreatedAt: time.Now(),
    })
    fake.SeedPurpose(agreementID, fakepdnd.StoredPurpose{
        ID: uuid.New(),
        EServiceID: uuid.New(),
        ConsumerID: fake.ConsumerID(),
        Title: "Test Purpose",
        Description: "A test purpose",
        CreatedAt: time.Now(),
        IsRiskAnalysisValid: true,
        IsFreeOfCharge: false,
    })

    ts := fake.Start()
    defer ts.Close()

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_agreement_purposes" "test" {
  agreement_id = %q
}
`, agreementID.String()),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.pdnd_agreement_purposes.test", "purposes.#", "1"),
                    resource.TestCheckResourceAttr("data.pdnd_agreement_purposes.test", "purposes.0.title", "Test Purpose"),
                ),
            },
        },
    })
}
```

---

## 3. Documentation

### `templates/index.md.tmpl`

```markdown
---
page_title: "PDND Provider"
description: |-
  The PDND provider enables Terraform to manage resources on the PDND Interoperability platform.
---

# PDND Provider

The PDND provider enables Terraform to manage resources on the [PDND Interoperability](https://developers.interop.pagopa.it/) platform (API v3).

## Authentication

The provider uses DPoP (Demonstration of Proof-of-Possession) authentication. You need:

- An **access token** for the PDND API
- A **PEM-encoded private key** for DPoP proof generation
- A **key ID** identifying your DPoP key

## Example Usage

{{ tffile "examples/provider/provider.tf" }}

{{ .SchemaMarkdown | trimspace }}
```

### `templates/resources/agreement.md.tmpl`

```markdown
---
page_title: "pdnd_agreement Resource - PDND"
description: |-
  Manages a PDND Agreement.
---

# pdnd_agreement (Resource)

Manages a PDND Agreement lifecycle, including creation, submission, suspension, and upgrade.

The `desired_state` attribute controls the intended lifecycle state. The provider will automatically
perform the necessary API operations (submit, suspend, unsuspend) to reach the desired state.

## Example Usage

{{ tffile "examples/resources/pdnd_agreement/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

Agreements can be imported using their UUID:

{{ codefile "shell" "examples/resources/pdnd_agreement/import.sh" }}

After import, `desired_state` is inferred from the current server state.
Only agreements in DRAFT, ACTIVE, or SUSPENDED state can be imported.
```

### Data source templates follow same pattern with `{{ .SchemaMarkdown }}` and `{{ tffile ... }}`.

### `examples/provider/provider.tf`

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  access_token     = var.pdnd_access_token
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

### `examples/resources/pdnd_agreement/resource.tf`

```hcl
# Create a draft agreement
resource "pdnd_agreement" "draft" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "660e8400-e29b-41d4-a716-446655440001"
  desired_state = "DRAFT"
}

# Create and activate an agreement
resource "pdnd_agreement" "active" {
  eservice_id    = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id  = "660e8400-e29b-41d4-a716-446655440001"
  desired_state  = "ACTIVE"
  consumer_notes = "Requesting access to the service"
}

# Create an agreement that tolerates pending approval
resource "pdnd_agreement" "with_pending" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "660e8400-e29b-41d4-a716-446655440001"
  desired_state = "ACTIVE"
  allow_pending = true
}
```

### `examples/resources/pdnd_agreement/import.sh`

```shell
terraform import pdnd_agreement.example 550e8400-e29b-41d4-a716-446655440000
```

### Data source examples — straightforward HCL showing usage.

---

## 4. README.md

Cover:
- Provider purpose and maturity (Milestone 1, agreements only)
- Prerequisites (Go 1.22+, Terraform 1.0+)
- Authentication setup (DPoP, access token)
- Provider configuration
- Quick example
- Available resources and data sources
- Testing instructions (`make test`, `make testacc`)
- Code generation (`make generate`, `make drift-check`)
- Contributing / development setup

---

## 5. CI Workflow

### `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./... -v -timeout 5m

  acceptance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m

  drift-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make generate
      - run: git diff --exit-code internal/client/generated/
```

---

## Execution Steps

1. Read all existing code: provider, resource, data sources, fake server, API wrapper
2. Create acceptance test helper functions (provider factories, config helper, test key)
3. Write all acceptance tests
4. Run `TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m`
5. Fix failures and iterate until ALL tests pass
6. Create documentation templates and examples
7. Create README.md
8. Create CI workflow
9. Run `go build ./...` to verify final state

## Verification

- `TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m` — ALL acceptance tests pass
- `go test ./... -v` — ALL tests pass (unit + contract + acceptance)
- `go build ./...` — compiles
- All example files are valid HCL

## Important Notes

- Acceptance tests MUST use `TF_ACC=1` environment variable (Terraform testing convention)
- Each test creates its own fake server instance (test isolation)
- The fake server doesn't validate auth — this is correct for acceptance tests
- Use `resource.ComposeAggregateTestCheckFunc` to check multiple attributes
- Use `regexp.MustCompile` for `ExpectError` patterns
- Test functions must start with `TestAcc` prefix
- For data source tests, use `depends_on` to ensure resources are created before data sources read
- The `ImportStateVerifyIgnore` list should include `allow_pending` and `consumer_notes` (user-specified fields not in API response)
- DO NOT modify implementation files — only create test, doc, and CI files
