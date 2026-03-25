Below is a formal implementation specification you can hand to an autonomous coding agent.

It is written to optimize for iterative delivery, test-first behavior, and explicit completion criteria.

---

# Specification: Terraform Provider for PDND Interoperability API v3

## 1. Objective

Implement a production-grade Terraform provider named `terraform-provider-pdnd` for the **PDND Interoperability API v3**, using the supplied OpenAPI specification as the authoritative API contract.

The implementation must be developed incrementally, with a strong emphasis on:

* deterministic testability
* isolated transport/auth validation
* Terraform acceptance tests
* truthful lifecycle modeling for non-CRUD API workflows
* reproducible code generation and CI verification

The agent must continue iterating until all required tests are implemented and passing.

---

## 2. Primary Goals

The provider must:

1. expose a modern Terraform provider implemented in Go using the **Terraform Plugin Framework**
2. consume the provided OpenAPI spec to generate or partially generate a typed API client
3. implement a robust HTTP transport layer for PDND authentication and request signing
4. model PDND domain objects as Terraform resources and data sources where appropriate
5. handle non-CRUD lifecycle transitions explicitly and correctly
6. include a stateful fake PDND server for deterministic acceptance testing
7. include unit, contract, and acceptance tests
8. ensure the codebase can be regenerated and verified against the OpenAPI spec without drift

---

## 3. Non-Goals

The agent must not:

* blindly map every API operation to a Terraform resource
* expose operational or debugging endpoints as resources unless there is a strong Terraform use case
* fake successful destroy behavior for resources that the server does not allow deleting
* skip tests in order to ship code faster
* rely on live PDND environments for default CI
* hardcode behavior that should derive from the OpenAPI schemas or API responses

---

## 4. Technology Constraints

The implementation must use:

* **Go**
* **Terraform Plugin Framework**
* **terraform-plugin-testing**
* a reproducible OpenAPI-driven client generation flow
* standard Go testing tools
* idiomatic Go project structure

Preferred supporting tools:

* `oapi-codegen` for client/type generation
* `httptest` for transport and unit tests
* `golangci-lint`
* `make` for repeatable developer and CI workflows

---

## 5. Authoritative Inputs

The agent must treat the following as authoritative:

1. the user-supplied PDND OpenAPI v3 specification
2. the provider behavior described in this specification
3. Terraform resource lifecycle semantics
4. actual passing tests over assumptions

If the OpenAPI spec and this document conflict, the agent must prefer the OpenAPI contract for wire-level behavior and this document for Terraform modeling decisions.

---

## 6. Repository Layout

The agent must scaffold the repository with approximately this structure:

```text
terraform-provider-pdnd/
  cmd/
    terraform-provider-pdnd/
      main.go

  internal/
    provider/
      provider.go
      config.go

    client/
      generated/
      api/
      transport.go
      auth.go
      dpop.go
      retry.go
      errors.go

    resources/
      agreement_resource.go
      agreement_helpers.go
      agreement_state_machine.go
      agreement_import.go

    datasources/
      agreement_data_source.go
      agreements_data_source.go
      agreement_purposes_data_source.go

    models/
      agreement.go
      common.go

    testing/
      fakepdnd/
        server.go
        agreements.go
        state.go
      fixtures/
      golden/

  tools/
    tools.go

  examples/
    provider/
    resources/
    data-sources/

  scripts/
  docs/
  Makefile
  go.mod
  go.sum
```

The exact structure may evolve slightly, but separation of concerns must remain clear.

---

## 7. Delivery Strategy

The agent must implement the provider in milestones. Each milestone must leave the codebase in a passing state.

The agent must not attempt the full API surface in one pass.

## Milestone 1: Foundation + Agreements

Implement:

* provider skeleton
* transport/auth layer
* generated client wrapper for agreements endpoints
* fake PDND server for agreements
* `pdnd_agreement` resource
* `pdnd_agreement` data source
* `pdnd_agreements` data source
* `pdnd_agreement_purposes` data source
* full test suite for agreements behavior

This milestone is mandatory and defines the quality bar for later resources.

## Milestone 2: Clients and Keys

Implement:

* `pdnd_client` resource and data source
* `pdnd_key` or `pdnd_client_key` resource
* any necessary association semantics
* full unit and acceptance tests

## Milestone 3: Purposes

Implement:

* `pdnd_purpose` resource and data source
* stateful transitions as appropriate
* acceptance tests

## Milestone 4: E-services and Attributes

Implement:

* `pdnd_eservice`
* `pdnd_eservice_version` if warranted by spec structure
* `pdnd_attribute`
* tests

## Later Milestones

Evaluate:

* delegations
* templates
* producer keychains
* tenants
* users
* events as data sources only if useful

---

## 8. Provider Configuration Requirements

The provider must support configuration for at least:

* `base_url`
* access token or token source configuration
* DPoP private key material
* DPoP key ID if needed
* request timeout
* retry settings

Initial schema may look like:

```hcl
provider "pdnd" {
  base_url          = "https://api.interop.pagopa.it/v3"
  access_token      = var.pdnd_access_token
  dpop_private_key  = file(var.dpop_private_key_path)
  dpop_key_id       = var.dpop_key_id
  request_timeout_s = 30
}
```

The exact final shape may vary, but it must be documented and tested.

The provider must validate configuration eagerly and return clear diagnostics.

---

## 9. Transport and Authentication Requirements

The agent must implement a dedicated transport layer.

Responsibilities:

* attach authorization header(s)
* generate DPoP proof header(s) per request
* ensure proof claims use the correct method and URL
* ensure a unique `jti` per request
* handle clock-based claims correctly
* support retry/backoff for `429` and selected `5xx`
* map PDND API errors into structured Go errors
* isolate auth/signing logic from resource code

The transport layer must be fully unit tested.

If the API requires nonce handling for DPoP, the implementation must support challenge-response retry behavior. If the provided spec does not define nonce semantics, the code may defer this behind a clearly isolated extension point, but the transport must be designed to accommodate it.

---

## 10. OpenAPI Client Requirements

The agent must generate or derive strongly typed request/response models from the OpenAPI spec.

Requirements:

* pin the exact spec file in the repository
* make code generation reproducible via `make generate`
* do not let Terraform resource code depend directly on raw generated code everywhere
* define a thin internal API wrapper layer over generated types and calls

Example internal interface shape:

```go
type AgreementsAPI interface {
    CreateAgreement(ctx context.Context, seed AgreementSeed) (*Agreement, error)
    GetAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
    DeleteAgreement(ctx context.Context, id uuid.UUID) error
    SubmitAgreement(ctx context.Context, id uuid.UUID, payload AgreementSubmission) (*Agreement, error)
    ApproveAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    RejectAgreement(ctx context.Context, id uuid.UUID, payload AgreementRejection) (*Agreement, error)
    SuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    UnsuspendAgreement(ctx context.Context, id uuid.UUID, payload *DelegationRef) (*Agreement, error)
    UpgradeAgreement(ctx context.Context, id uuid.UUID) (*Agreement, error)
}
```

The wrapper must be mockable for tests.

---

## 11. Terraform Modeling Rules

The following modeling rules are mandatory.

### 11.1 Prefer one primary resource per domain object

For example, agreements must be modeled primarily as:

* `pdnd_agreement`

Do not create action-only resources such as:

* `pdnd_agreement_submit`
* `pdnd_agreement_approve`
* `pdnd_agreement_suspend`

unless a later milestone proves that a separate lifecycle identity is necessary.

### 11.2 Model lifecycle transitions through desired state

Where appropriate, resources must expose a desired-state or intent-based field rather than raw action operations.

For agreements, the resource must support a desired lifecycle intent, for example:

* `DRAFT`
* `ACTIVE`
* `SUSPENDED`

The exact field name may be `desired_state`.

The resource must also expose observed server state as a computed field, for example `state`.

### 11.3 Do not lie about deletion

If the API allows deletion only in specific states, destroy must:

* call delete only when allowed
* fail with a clear diagnostic when deletion is not allowed

The implementation must not silently remove state if the upstream object still exists.

### 11.4 Lists become data sources

List and filter endpoints must generally become data sources, not resources.

### 11.5 Relationship endpoints should become explicit association resources only when necessary

Do not prematurely create association resources unless the relationship has meaningful independent lifecycle semantics.

---

## 12. Required Agreements Resource Behavior

The `pdnd_agreement` resource is mandatory in Milestone 1.

The agent must derive the exact schema fields from the OpenAPI spec, but the behavior must follow these rules.

### 12.1 Create behavior

On create:

1. call `POST /agreements`
2. if desired state is `DRAFT`, stop
3. if desired state is `ACTIVE`, call `submit`
4. if submit yields `ACTIVE`, persist active state
5. if submit yields `PENDING`:

   * succeed if `allow_pending = true`
   * fail if `allow_pending = false`

### 12.2 Update behavior

Support transitions that the API explicitly allows.

Examples:

* `ACTIVE` to `SUSPENDED` calls suspend
* `SUSPENDED` to `ACTIVE` calls unsuspend

The agent must implement transitions only when justified by the API and tests.

### 12.3 Upgrade behavior

The API indicates agreement upgrade archives the current agreement and creates a new one.

The provider must model this carefully.

Acceptable approaches:

* explicit `target_descriptor_id` update logic that invokes upgrade
* clear replacement/state migration semantics

The provider must not treat upgrade as a simple in-place update unless the API semantics truly allow that.

### 12.4 Delete behavior

Delete must only succeed when the server allows it.

### 12.5 Read behavior

Read must refresh all stable fields and observed state.

If the object no longer exists, remove it from state.

---

## 13. Required Data Sources for Milestone 1

Implement:

* `pdnd_agreement`
* `pdnd_agreements`
* `pdnd_agreement_purposes`

These must support filtering or lookup behavior consistent with the spec.

The exact Terraform schemas must be derived from the OpenAPI fields.

---

## 14. Fake PDND Server Requirements

The agent must implement a deterministic, stateful fake PDND server for tests.

This server is mandatory.

It must:

* run in-process in Go tests
* store entities in memory
* emulate agreements endpoints and state transitions
* validate methods, paths, and selected request payloads
* return JSON compatible with the OpenAPI schema
* simulate valid and invalid transitions
* simulate upgrade creating a successor agreement
* support list/filter behavior for implemented data sources

At minimum, for agreements the fake server must model these transitions:

* create -> `DRAFT`
* submit from `DRAFT` -> `ACTIVE` or `PENDING`
* approve from `PENDING` -> `ACTIVE`
* reject from `PENDING` -> `REJECTED`
* suspend from `ACTIVE` or `SUSPENDED` -> `SUSPENDED`
* unsuspend from `SUSPENDED` -> `ACTIVE` or remain `SUSPENDED` when applicable
* delete only from allowed states
* upgrade archives current and produces a new agreement identity

The fake server must be used in Terraform acceptance tests by default.

---

## 15. Test Requirements

Testing is a hard requirement. The agent must not stop until all tests are implemented and passing.

The required test pyramid is below.

## 15.1 Transport Unit Tests

Mandatory tests include:

* DPoP header is attached to every request
* authorization header is attached correctly
* proof contains correct method
* proof contains correct URL
* proof uses unique `jti`
* retry on `429`
* retry on selected `5xx`
* config validation errors are clear

These tests may use `httptest.Server`.

## 15.2 Client Contract Tests

Mandatory tests include:

* create agreement sends correct method, path, and body
* get agreement uses correct path
* delete agreement uses correct method/path
* submit uses correct path/body
* approve uses correct path/body
* reject uses correct path/body
* suspend uses correct path/body
* unsuspend uses correct path/body
* upgrade uses correct path

These tests must verify serialization and response decoding.

## 15.3 Resource Unit Tests

Mandatory tests include logic around:

* desired-state translation
* pending handling
* deletion eligibility
* upgrade semantics
* schema plan modifiers as needed
* import parsing if custom import format is used

## 15.4 Terraform Acceptance Tests

Mandatory acceptance tests for agreements:

* create draft agreement
* create active agreement happy path
* create active agreement when submit yields pending and `allow_pending = true`
* create active agreement when submit yields pending and `allow_pending = false`
* update active agreement to suspended
* update suspended agreement to active
* destroy draft agreement
* destroy forbidden active agreement and verify expected failure
* import agreement
* list agreements data source
* agreement purposes data source

These tests must run against the fake PDND server.

## 15.5 Regeneration and Drift Tests

The repository must include a mechanism to verify that generated client code matches the pinned OpenAPI spec.

Minimum requirement:

* `make generate`
* CI fails if generated artifacts differ from committed state

Optional but preferred:

* a spec-diff check between current and prior pinned versions

---

## 16. CI Requirements

The repository must include a CI workflow that runs at least:

* `go test ./...`
* acceptance tests
* `golangci-lint`
* generation drift check
* `terraform fmt` for examples if applicable

The default CI path must not depend on external PDND infrastructure.

Live environment smoke tests may exist, but must be opt-in and excluded from default CI.

---

## 17. Documentation Requirements

The agent must produce:

* a README describing provider purpose, maturity, and limitations
* setup instructions
* auth configuration documentation
* example provider configuration
* example `pdnd_agreement` resource usage
* import examples
* test instructions
* code generation instructions

Examples must be valid and testable where feasible.

---

## 18. Iteration Protocol for the Agent

The agent must work in the following loop:

1. scaffold or modify code for the current milestone
2. add or update tests first or alongside implementation
3. run tests
4. inspect failures
5. fix code or tests
6. repeat until all tests pass
7. only then proceed to the next milestone

The agent must never declare completion while required tests are missing or failing.

If blocked by ambiguity in the OpenAPI schema, the agent must:

* inspect the schema
* derive the closest correct implementation
* prefer explicit TODO markers only when a feature is outside the current milestone
* avoid leaving untested stubs in required milestone scope

---

## 19. Definition of Done

The implementation is complete for a milestone only when all of the following are true:

1. code compiles
2. lint passes
3. generation is reproducible
4. all unit tests pass
5. all contract tests pass
6. all acceptance tests pass
7. examples are present
8. documentation is updated
9. there are no placeholder stubs for required milestone features

The overall project is complete only when the milestone scope requested by the user has been implemented under the same criteria.

---

## 20. Initial Acceptance Criteria for Milestone 1

Milestone 1 is complete only when all of the following are true:

* provider configuration is implemented and validated
* DPoP/auth transport is implemented and tested
* agreements client wrapper is implemented
* fake PDND server supports agreements workflows
* `pdnd_agreement` resource is implemented
* `pdnd_agreement` data source is implemented
* `pdnd_agreements` data source is implemented
* `pdnd_agreement_purposes` data source is implemented
* all required agreement tests are implemented and passing
* README includes usage and testing instructions

---

## 21. Suggested Task List for the Agent

The agent may use this as the concrete execution order:

1. initialize Go module and provider skeleton
2. pin the OpenAPI spec into the repo
3. add codegen tooling and `make generate`
4. generate base client/types
5. implement internal API wrapper for agreements
6. implement transport/auth layer with DPoP support
7. write transport unit tests
8. write agreements client contract tests
9. implement fake PDND server for agreements
10. implement `pdnd_agreement` Terraform model and schema
11. write acceptance tests for agreement create/read/update/delete/import
12. implement `pdnd_agreement` resource until acceptance tests pass
13. implement agreements data sources and tests
14. add docs and examples
15. add lint and generation drift checks
16. verify all required tests pass

---

## 22. Quality Rules

The agent must follow these quality rules:

* keep resource code thin
* keep transport/auth code isolated
* avoid duplicating generated schema logic by hand unless necessary
* prefer explicit error messages
* prefer deterministic tests
* avoid flaky time-based tests
* use table-driven tests where helpful
* ensure import behavior is documented and tested
* ensure state transitions are visible in Terraform state

---

## 23. Output Expectations for the Agent

At the end of each iteration or milestone, the agent should be able to report:

* what was implemented
* which tests were added
* which tests are passing
* what remains for the milestone

But the agent must prioritize actual code and passing tests over narrative status reports.