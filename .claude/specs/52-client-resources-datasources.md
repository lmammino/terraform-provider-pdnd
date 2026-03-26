# Spec 52: Client Resources + Data Sources + Provider Wiring

## Objective

Implement 2 resources (`pdnd_client_key`, `pdnd_client_purpose`) and 3 data sources (`pdnd_client`, `pdnd_clients`, `pdnd_client_keys`). Wire into provider. When done, `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/descriptor_document_resource.go` — immutable resource pattern (all RequiresReplace, no Update)
2. `internal/resources/descriptor_attributes_shared.go` — composite ID pattern, Configure from providerdata
3. `internal/resources/agreement_import.go` — simple ID passthrough import
4. `internal/resources/eservice_descriptor_import.go` — composite ID import (`parseDescriptorCompositeID`)
5. `internal/datasources/purpose_data_source.go` — singular data source
6. `internal/datasources/purposes_data_source.go` — plural data source with filters + auto-pagination
7. `internal/datasources/consumer_delegations_data_source.go` — recent plural data source pattern
8. `internal/models/delegation.go` — model pattern
9. `internal/providerdata/providerdata.go` — ProviderData struct
10. `internal/provider/provider.go` — Resources(), DataSources(), Configure()
11. `internal/client/api/clients.go` — ClientsAPI interface (created in Spec 50)

## Files to Create/Modify

### 1. Create `internal/models/client.go` — Terraform models

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// ClientKeyResourceModel is the Terraform state for pdnd_client_key resource.
type ClientKeyResourceModel struct {
    ID        types.String `tfsdk:"id"`         // Composite: client_id/kid
    ClientID  types.String `tfsdk:"client_id"`
    Key       types.String `tfsdk:"key"`         // Base64 PEM public key
    Use       types.String `tfsdk:"use"`          // SIG or ENC
    Alg       types.String `tfsdk:"alg"`
    Name      types.String `tfsdk:"name"`         // 5-60 chars
    // Computed from response
    Kid       types.String `tfsdk:"kid"`          // Server-assigned key ID
    Kty       types.String `tfsdk:"kty"`          // Key type (RSA, EC, etc.)
}

// ClientPurposeResourceModel is the Terraform state for pdnd_client_purpose resource.
type ClientPurposeResourceModel struct {
    ID        types.String `tfsdk:"id"`           // Composite: client_id/purpose_id
    ClientID  types.String `tfsdk:"client_id"`
    PurposeID types.String `tfsdk:"purpose_id"`
}

// ClientDataSourceModel is the Terraform state for pdnd_client data source.
type ClientDataSourceModel struct {
    ID          types.String `tfsdk:"id"`
    ConsumerID  types.String `tfsdk:"consumer_id"`
    Name        types.String `tfsdk:"name"`
    Description types.String `tfsdk:"description"`
    CreatedAt   types.String `tfsdk:"created_at"`
}

// ClientKeyDataSourceModel is for the nested key items in pdnd_client_keys.
type ClientKeyDataSourceModel struct {
    Kid types.String `tfsdk:"kid"`
    Kty types.String `tfsdk:"kty"`
    Alg types.String `tfsdk:"alg"`
    Use types.String `tfsdk:"use"`
}
```

### 2. Create `internal/resources/client_key_resource.go`

Resource: `pdnd_client_key`

**Schema:**
- `id` — Computed, UseStateForUnknown (composite `client_id/kid`)
- `client_id` — Required, RequiresReplace, UUID validator
- `key` — Required, RequiresReplace, Sensitive (the PEM public key)
- `use` — Required, RequiresReplace, OneOf("SIG", "ENC")
- `alg` — Required, RequiresReplace
- `name` — Required, RequiresReplace, LengthBetween(5, 60)
- `kid` — Computed, UseStateForUnknown
- `kty` — Computed, UseStateForUnknown

**Configure:** Get `ClientsAPI` from `providerdata.ProviderData.ClientsAPI`

**Create:**
1. Build `ClientKeySeed` from plan
2. Call `client.CreateClientKey(ctx, clientID, seed)`
3. Set `kid`, `kty` from response
4. Set composite `id` = `clientId/kid`

**Read:**
1. Parse composite ID to get `clientID` and `kid`
2. Call `client.ListClientKeys(ctx, clientID, 0, 50)` and paginate to find the key by `kid`
3. If not found, `RemoveResource`
4. Preserve `key`, `use`, `alg`, `name` from state (not returned by list)

**Update:** No-op — all user fields have RequiresReplace

**Delete:** Call `client.DeleteClientKey(ctx, clientID, kid)`. Handle 404 gracefully.

**ImportState:** Parse composite ID `client_id/kid`. Set `client_id`, `kid`, `id`. Note: `key`, `use`, `alg`, `name` cannot be recovered from API — add to ImportStateVerifyIgnore in tests.

**Import helper** — add `parseClientKeyCompositeID(id string) (clientID, kid string, err error)` that splits on `/`, validates first part is UUID, returns both.

### 3. Create `internal/resources/client_purpose_resource.go`

Resource: `pdnd_client_purpose`

**Schema:**
- `id` — Computed, UseStateForUnknown (composite `client_id/purpose_id`)
- `client_id` — Required, RequiresReplace, UUID validator
- `purpose_id` — Required, RequiresReplace, UUID validator

**Create:** Call `client.AddClientPurpose(ctx, clientID, purposeID)`. Set composite `id`.

**Read:** Call `client.ListClientKeys` won't work here — need to verify the purpose is still linked. Since there's no "get single client-purpose" endpoint, we can check via listing. But the `GetClientPurposes` endpoint returns purposes (not a simple existence check). For simplicity: attempt to read and if the purpose is no longer linked, remove resource. Actually, the simplest approach: just keep state as-is (the link is immutable). If the user manually removes the link, the next plan will show a diff.

Alternative: Just trust the state. The link is a simple association — if deleted externally, Terraform will try to create it again on next apply.

**Update:** No-op — all fields RequiresReplace

**Delete:** Call `client.RemoveClientPurpose(ctx, clientID, purposeID)`. Handle 404 gracefully.

**ImportState:** Parse composite ID `client_id/purpose_id` using a `parseClientPurposeCompositeID` helper.

### 4. Create `internal/datasources/client_data_source.go` — Singular

Fetch client by ID. Use `client.GetClient(ctx, id)`. Populate model.

### 5. Create `internal/datasources/clients_data_source.go` — Plural

List with optional filters: `name` (string), `consumer_id` (string UUID). Auto-paginate.

### 6. Create `internal/datasources/client_keys_data_source.go`

List keys for a client. Required input: `client_id`. Returns nested list of JWK objects.

### 7. Modify `internal/providerdata/providerdata.go`

Add: `ClientsAPI api.ClientsAPI`

### 8. Modify `internal/provider/provider.go`

In Configure: `clientsAPI := api.NewClientsClient(genClient)`, add to pd.
In Resources: `resources.NewClientKeyResource`, `resources.NewClientPurposeResource`
In DataSources: `datasources.NewClientDataSource`, `datasources.NewClientsDataSource`, `datasources.NewClientKeysDataSource`

## Execution Steps

1. Read all referenced files
2. Create `models/client.go`
3. Create `resources/client_key_resource.go` (with import helper)
4. Create `resources/client_purpose_resource.go` (with import helper)
5. Create 3 data source files
6. Modify `providerdata.go` and `provider.go`
7. Run `go build ./...`

## Verification

```shell
go build ./...
```

## Important Notes

- `key` field should be marked `Sensitive: true` (it's a public key but may contain formatting users don't want in logs)
- `kid` is a string, NOT a UUID — no UUID validation on it
- `CreateClientKey` returns `JSON200` (not `JSON201`) — this is already handled in the API client
- Composite IDs: `client_key` uses `client_id/kid`, `client_purpose` uses `client_id/purpose_id`
- For `client_key` Read: we paginate through `ListClientKeys` looking for the matching `kid`. If not found, resource was deleted externally.
- For `client_purpose` Read: trust the state (no efficient check). If deleted externally, the next apply will re-create the link.
- `uuidRegex` is available in the `resources` package from `agreement_resource.go`
