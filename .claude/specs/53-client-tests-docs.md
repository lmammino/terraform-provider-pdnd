# Spec 53: Client Tests + Documentation

## Objective

Add acceptance tests for 2 resources and 3 data sources, doc templates, examples, and README update. When done, all tests pass and `make docs` generates documentation.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/descriptor_document_resource_acc_test.go` — immutable resource test pattern
2. `internal/resources/consumer_delegation_resource_acc_test.go` — recent resource acc test
3. `internal/datasources/consumer_delegation_data_source_acc_test.go` — singular data source test
4. `internal/datasources/consumer_delegations_data_source_acc_test.go` — plural data source test
5. `internal/testing/fakepdnd/server.go` — SeedClient, SeedClientKey, SeedClientPurpose methods
6. `internal/testing/fakepdnd/state.go` — StoredClient, StoredClientKey
7. `internal/resources/client_key_resource.go` — created in Spec 52
8. `internal/resources/client_purpose_resource.go` — created in Spec 52
9. `templates/resources/consumer_delegation.md.tmpl` — recent doc template format
10. `README.md` — tables to update

Also check test helpers:
11. `internal/resources/agreement_resource_acc_test.go` — `testAccProviderConfig`, `testAccProtoV6ProviderFactories`
12. `internal/datasources/agreement_data_source_acc_test.go` — data source test helpers in `datasources_test` package

## Files to Create/Modify

### 1. Create `internal/resources/client_key_resource_acc_test.go`

**Tests:**

**`TestAccClientKey_Create`:** Seed a client, create a key, verify kid/kty are set.
```go
fake := fakepdnd.NewFakeServer()
clientID := uuid.New()
fake.SeedClient(fakepdnd.StoredClient{
    ID: clientID, ConsumerID: fake.ConsumerID(),
    Name: "Test Client", CreatedAt: time.Now().UTC(),
})
ts := fake.Start()
defer ts.Close()

// Use a test RSA public key (can generate one or use a static PEM)
resource.Test(t, resource.TestCase{
    ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
    Steps: []resource.TestStep{
        {
            Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_client_key" "test" {
  client_id = %q
  key       = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."
  use       = "SIG"
  alg       = "RS256"
  name      = "Test Key"
}
`, clientID),
            Check: resource.ComposeAggregateTestCheckFunc(
                resource.TestCheckResourceAttrSet("pdnd_client_key.test", "kid"),
                resource.TestCheckResourceAttrSet("pdnd_client_key.test", "kty"),
                resource.TestCheckResourceAttrSet("pdnd_client_key.test", "id"),
            ),
        },
    },
})
```

**`TestAccClientKey_Delete`:** Create then destroy, verify key removed via `fake.GetClientKeys()`.

**`TestAccClientKey_Import`:** Create, then import via `client_id/kid` composite ID. Use `ImportStateIdFunc` to extract from state. `ImportStateVerifyIgnore: []string{"key", "use", "alg", "name"}` (not recoverable from API).

### 2. Create `internal/resources/client_purpose_resource_acc_test.go`

**`TestAccClientPurpose_Create`:** Seed client, create purpose link, verify id is set.

**`TestAccClientPurpose_Delete`:** Create then destroy, verify purpose unlinked.

### 3. Create `internal/datasources/client_data_source_acc_test.go`

**`TestAccClientDataSource`:** Seed client, fetch by ID, verify name/consumer_id.

### 4. Create `internal/datasources/clients_data_source_acc_test.go`

**`TestAccClientsDataSource`:** Seed multiple clients, list all, verify count.

### 5. Create `internal/datasources/client_keys_data_source_acc_test.go`

**`TestAccClientKeysDataSource`:** Seed client with keys, list keys, verify count and kid.

### 6. Create doc templates (5 files)

**`templates/resources/client_key.md.tmpl`:**
```markdown
---
page_title: "pdnd_client_key Resource - PDND"
description: |-
  Manages a public key on a PDND client.
---

# pdnd_client_key (Resource)

Creates a public key (JWK) on a PDND client for API authentication. Keys are immutable — any change triggers replacement.

## Example Usage

{{ tffile "examples/resources/pdnd_client_key/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

{{ codefile "shell" "examples/resources/pdnd_client_key/import.sh" }}
```

**`templates/resources/client_purpose.md.tmpl`** — similar for purpose linking.

**`templates/data-sources/client.md.tmpl`**
**`templates/data-sources/clients.md.tmpl`**
**`templates/data-sources/client_keys.md.tmpl`**

### 7. Create examples (7 files in 5 directories)

**`examples/resources/pdnd_client_key/resource.tf`:**
```hcl
resource "pdnd_client_key" "api_key" {
  client_id = data.pdnd_client.example.id
  key       = filebase64(var.public_key_path)
  use       = "SIG"
  alg       = "RS256"
  name      = "Production API Key"
}
```

**`examples/resources/pdnd_client_key/import.sh`:**
```shell
terraform import pdnd_client_key.api_key cc0e8400-e29b-41d4-a716-446655440040/test-key-id-001
```

**`examples/resources/pdnd_client_purpose/resource.tf`:**
```hcl
resource "pdnd_client_purpose" "link" {
  client_id  = data.pdnd_client.example.id
  purpose_id = pdnd_purpose.example.id
}
```

**`examples/resources/pdnd_client_purpose/import.sh`:**
```shell
terraform import pdnd_client_purpose.link cc0e8400-e29b-41d4-a716-446655440040/990e8400-e29b-41d4-a716-446655440000
```

**`examples/data-sources/pdnd_client/data-source.tf`**
**`examples/data-sources/pdnd_clients/data-source.tf`**
**`examples/data-sources/pdnd_client_keys/data-source.tf`**

### 8. Modify `README.md`

Add to resources table:
```
| `pdnd_client_key`     | Manages a public key on a PDND client       |
| `pdnd_client_purpose` | Links a purpose to a PDND client            |
```

Add to data sources table:
```
| `pdnd_client`          | Fetches a single client by ID                |
| `pdnd_clients`         | Lists clients with optional filters          |
| `pdnd_client_keys`     | Lists keys on a client                       |
```

## Execution Steps

1. Read all referenced files
2. Create example directories (`mkdir -p`) and files
3. Create doc templates
4. Create resource acceptance tests
5. Create data source acceptance tests
6. Update README.md
7. Run `TF_ACC=1 go test ./internal/resources/... -v -run "Client" -timeout 5m`
8. Run `TF_ACC=1 go test ./internal/datasources/... -v -run "Client" -timeout 5m`
9. Run `go test ./...`
10. Run `make lint`
11. Run `make docs`

## Verification

```shell
TF_ACC=1 go test ./internal/resources/... -v -run "Client" -timeout 5m
TF_ACC=1 go test ./internal/datasources/... -v -run "Client" -timeout 5m
go test ./...
make lint
make docs
```

## Important Notes

- Do NOT break existing tests
- For the `key` field in tests: use a short base64-encoded string (doesn't need to be a real PEM key for the fake server — it's just stored and returned)
- `kid` is a string, not UUID — the import composite ID uses `client_id/kid` where `client_id` is UUID and `kid` is an opaque string
- Import tests for `client_key`: `key`, `use`, `alg`, `name` cannot be recovered from the API — add to `ImportStateVerifyIgnore`
- `testAccProviderConfig` and `testAccProtoV6ProviderFactories` are already defined in both `resources_test` and `datasources_test` packages
- Create example directories with `mkdir -p` before writing files
