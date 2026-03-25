# Agent 1: Project Init + Code Generation

## Objective

Scaffold a Go project for `terraform-provider-pdnd` and set up reproducible OpenAPI code generation. When you're done, `make generate && make build` must succeed and `make drift-check` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo already has:
- `SPEC.md` — project specification (do not modify)
- `interop-api-v3-spec.yaml` — OpenAPI 3.0.3 spec (13k lines, do not modify)
- `.git/` — initialized git repo

## Files to Create

### 1. `go.mod`

```
module github.com/pagopa/terraform-provider-pdnd
go 1.22
```

Dependencies to add (use latest stable versions):
- `github.com/hashicorp/terraform-plugin-framework`
- `github.com/hashicorp/terraform-plugin-testing`
- `github.com/hashicorp/terraform-plugin-go`
- `github.com/oapi-codegen/runtime`
- `github.com/golang-jwt/jwt/v5`
- `github.com/google/uuid`

Tool dependencies (in tools/tools.go):
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen`
- `github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`

### 2. `tools/tools.go`

```go
//go:build tools

package tools

import (
    _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
    _ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
```

### 3. `cmd/terraform-provider-pdnd/main.go`

Minimal entry point. Use `providerserver.Serve` with Protocol Version 6 (default).

```go
package main

import (
    "context"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/pagopa/terraform-provider-pdnd/internal/provider"
)

var version = "dev"

func main() {
    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/pagopa/pdnd",
    }
    err := providerserver.Serve(context.Background(), provider.New(version), opts)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. `internal/provider/provider.go`

Minimal stub provider that compiles. Will be fully implemented later.

```go
package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &pdndProvider{}

type pdndProvider struct {
    version string
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &pdndProvider{version: version}
    }
}

func (p *pdndProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "pdnd"
    resp.Version = p.version
}

func (p *pdndProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Terraform provider for PDND Interoperability API v3",
    }
}

func (p *pdndProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *pdndProvider) Resources(_ context.Context) []func() resource.Resource {
    return nil
}

func (p *pdndProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return nil
}
```

### 5. `internal/client/generated/oapi-codegen.cfg.yaml`

Configuration for oapi-codegen v2. Generate a client with response types and model types.

```yaml
package: generated
output: client.gen.go
generate:
  client: true
  models: true
  embedded-spec: false
output-options:
  skip-prune: true
```

Run generation with:
```
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config internal/client/generated/oapi-codegen.cfg.yaml interop-api-v3-spec.yaml
```

The output file `internal/client/generated/client.gen.go` will be large (full API). This is intentional — the wrapper layer (Agent 3) will isolate resources from it.

### 6. `Makefile`

```makefile
.PHONY: generate build test testacc lint drift-check docs

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config internal/client/generated/oapi-codegen.cfg.yaml \
		interop-api-v3-spec.yaml

build:
	go build ./...

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./internal/... -v -run TestAcc -timeout 10m

lint:
	golangci-lint run ./...

drift-check: generate
	git diff --exit-code internal/client/generated/

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
```

### 7. `.golangci.yml`

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused

linters-settings:
  errcheck:
    check-type-assertions: true

issues:
  exclude-dirs:
    - internal/client/generated
```

### 8. `.gitignore`

```
# Binaries
terraform-provider-pdnd
*.exe
dist/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Terraform
.terraform/
*.tfstate
*.tfstate.backup
crash.log
```

## Execution Steps

1. Create all files listed above
2. Run `go mod init github.com/pagopa/terraform-provider-pdnd`
3. Add all the dependencies with `go get`
4. Run `go mod tidy`
5. Run `make generate` to produce `internal/client/generated/client.gen.go`
6. Run `make build` to verify everything compiles
7. Verify `make drift-check` passes (generate then check no diff)
8. If anything fails, fix and iterate

## Verification

All of these must succeed:
- `make generate` — produces client.gen.go without errors
- `make build` — compiles all packages
- `make drift-check` — no uncommitted generated code changes

## Important Notes

- Use Go 1.22+ (required for http.ServeMux pattern matching used later)
- The generated file will be large — this is expected
- Do NOT modify the OpenAPI spec file
- The provider stub just needs to compile; full implementation comes later
- Commit all generated code (it's checked into the repo and drift-checked in CI)
