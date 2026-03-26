# Spec 50: Client API Client + Contract Tests

## Objective

Add `ClientsAPI` interface and implementation covering client read, key CRUD, and purpose linking. Add domain types, conversions, and contract tests. When done, `go test ./internal/client/api/... -v` and `go build ./...` must pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/client/api/delegations.go` — recent API client pattern with interface + implementation
2. `internal/client/api/agreements.go` — ListAgreements pattern for list with filters + pagination
3. `internal/client/api/types.go` — domain types inventory
4. `internal/client/api/conversions.go` — conversion patterns
5. `internal/client/api/delegations_test.go` — recent contract test pattern
6. `internal/client/generated/client.gen.go` — find these types and methods:
   - `type FullClient struct` (line 1799) — id, consumerId, name, description*, createdAt
   - `type PartialClient struct` (line 1923) — id, consumerId
   - `type Client struct { union json.RawMessage }` (line 1079) — has `AsFullClient()` (line 3572) and `AsPartialClient()` (line 3598)
   - `type Clients struct` (line 1089) — Results []Client, Pagination
   - `type Key struct` (line 1854) — ClientId, Jwk JWK
   - `type KeySeed struct` (line 1883) — Key, Use, Alg, Name
   - `type KeyUse string` (line 1898) — SIG, ENC
   - `type JWK struct` (line 1820) — Kid, Kty, Alg, Use, + many optional fields
   - `type JWKs struct` (line 1848) — Results []JWK, Pagination
   - `type ClientAddPurpose struct` (line 1084) — PurposeId
   - `GetClientsParams` — Name*, ConsumerId*, Offset, Limit
   - `GetClientKeysParams` — Offset, Limit
   - WithResponse methods: `GetClientsWithResponse`, `GetClientWithResponse`, `GetClientKeysWithResponse`, `CreateClientKeyWithResponse`, `DeleteClientKeyByIdWithResponse`, `AddClientPurposeWithResponse`, `RemoveClientPurposeWithResponse`
   - Note: `CreateClientKeyResponse` has `JSON200 *Key` (NOT JSON201)
   - Note: `type Kid = string` (line 1880)
7. `internal/client/errors.go` — `CheckResponse`, `IsNotFound`

## Files to Create/Modify

### 1. Extend `internal/client/api/types.go` — Add client domain types

```go
// ClientInfo represents a PDND client (full visibility).
type ClientInfo struct {
    ID          uuid.UUID
    ConsumerID  uuid.UUID
    Name        string
    Description *string
    CreatedAt   time.Time
}

// ClientKey represents a JWK public key on a client.
type ClientKey struct {
    Kid string
    Kty string
    Alg *string
    Use *string
}

// ClientKeyDetail represents a key with its associated client ID (from create response).
type ClientKeyDetail struct {
    ClientID uuid.UUID
    Key      ClientKey
}

// ClientKeySeed contains fields for creating a new client key.
type ClientKeySeed struct {
    Key  string // Base64 PEM public key
    Use  string // SIG or ENC
    Alg  string
    Name string // 5-60 chars
}

// ClientsPage is a paginated list of clients.
type ClientsPage struct {
    Results    []ClientInfo
    Pagination Pagination
}

// ClientKeysPage is a paginated list of client keys.
type ClientKeysPage struct {
    Results    []ClientKey
    Pagination Pagination
}

// ListClientsParams contains filter parameters for listing clients.
type ListClientsParams struct {
    Name       *string
    ConsumerID *uuid.UUID
    Offset     int32
    Limit      int32
}
```

### 2. Extend `internal/client/api/conversions.go` — Add client conversions

```go
func clientInfoFromGenerated(g *generated.Client) (*ClientInfo, error) {
    // Try FullClient first
    full, err := g.AsFullClient()
    if err == nil {
        return &ClientInfo{
            ID: uuid.UUID(full.Id), ConsumerID: uuid.UUID(full.ConsumerId),
            Name: full.Name, Description: full.Description, CreatedAt: full.CreatedAt,
        }, nil
    }
    // Fall back to PartialClient
    partial, err := g.AsPartialClient()
    if err != nil {
        return nil, fmt.Errorf("failed to parse client: %w", err)
    }
    return &ClientInfo{
        ID: uuid.UUID(partial.Id), ConsumerID: uuid.UUID(partial.ConsumerId),
    }, nil
}

func clientKeyFromJWK(g *generated.JWK) ClientKey {
    return ClientKey{Kid: string(g.Kid), Kty: g.Kty, Alg: g.Alg, Use: g.Use}
}

func clientKeyDetailFromGenerated(g *generated.Key) *ClientKeyDetail {
    return &ClientKeyDetail{
        ClientID: uuid.UUID(g.ClientId),
        Key: clientKeyFromJWK(&g.Jwk),
    }
}

func clientKeySeedToGenerated(s ClientKeySeed) generated.KeySeed {
    return generated.KeySeed{
        Key: s.Key, Use: generated.KeyUse(s.Use), Alg: s.Alg, Name: s.Name,
    }
}
```

### 3. Create `internal/client/api/clients.go` — Interface + implementation

```go
type ClientsAPI interface {
    GetClient(ctx context.Context, id uuid.UUID) (*ClientInfo, error)
    ListClients(ctx context.Context, params ListClientsParams) (*ClientsPage, error)
    ListClientKeys(ctx context.Context, clientID uuid.UUID, offset, limit int32) (*ClientKeysPage, error)
    CreateClientKey(ctx context.Context, clientID uuid.UUID, seed ClientKeySeed) (*ClientKeyDetail, error)
    DeleteClientKey(ctx context.Context, clientID uuid.UUID, kid string) error
    AddClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error
    RemoveClientPurpose(ctx context.Context, clientID, purposeID uuid.UUID) error
}
```

Implementation: `clientsClient` struct, `NewClientsClient()` constructor.

Key details:
- **GetClient**: calls `GetClientWithResponse`, uses `clientInfoFromGenerated` to handle union type
- **ListClients**: builds `GetClientsParams`, converts ConsumerID to `*openapi_types.UUID`, calls `GetClientsWithResponse`, iterates results using `clientInfoFromGenerated` (skip parse errors for partial clients)
- **ListClientKeys**: calls `GetClientKeysWithResponse`, converts JWK array
- **CreateClientKey**: calls `CreateClientKeyWithResponse`, checks `JSON200` (NOT JSON201!), converts via `clientKeyDetailFromGenerated`
- **DeleteClientKey**: calls `DeleteClientKeyByIdWithResponse` with `clientID` as UUID and `kid` as string
- **AddClientPurpose**: calls `AddClientPurposeWithResponse` with `ClientAddPurpose{PurposeId}`
- **RemoveClientPurpose**: calls `RemoveClientPurposeWithResponse`

### 4. Create `internal/client/api/clients_test.go` — Contract tests

Use unique UUIDs:
```go
var (
    testClientID  = uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440040")
    testClientKID = "test-key-id-001"
)
```

Canned client JSON (FullClient):
```json
{
  "id": "cc0e8400-e29b-41d4-a716-446655440040",
  "consumerId": "880e8400-e29b-41d4-a716-446655440003",
  "name": "Test Client",
  "description": "A test client",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

Canned key response JSON:
```json
{
  "clientId": "cc0e8400-e29b-41d4-a716-446655440040",
  "jwk": {"kid": "test-key-id-001", "kty": "RSA", "alg": "RS256", "use": "sig"}
}
```

| Test | Verifies |
|------|----------|
| `TestGetClient_Contract` | GET `/clients/{clientId}`, parses union to FullClient |
| `TestListClients_Contract` | GET `/clients` with offset/limit/name params |
| `TestListClientKeys_Contract` | GET `/clients/{clientId}/keys` with pagination |
| `TestCreateClientKey_Contract` | POST `/clients/{clientId}/keys`, body with key/use/alg/name, `JSON200` |
| `TestDeleteClientKey_Contract` | DELETE `/clients/{clientId}/keys/{keyId}` |
| `TestAddClientPurpose_Contract` | POST `/clients/{clientId}/purposes` with purposeId body |
| `TestRemoveClientPurpose_Contract` | DELETE `/clients/{clientId}/purposes/{purposeId}` |

Note: The `GetClient` response wraps in the Client union. The test server should return raw FullClient JSON (the generated client's `AsFullClient` will parse it from the union's `json.RawMessage`). However, the `Clients` list response wraps each item in the union. For the test, return a `{"results": [{...fullClient...}], "pagination": {...}}` and verify parsing works.

## Execution Steps

1. Read generated client to confirm exact method signatures
2. Add domain types to `types.go`
3. Add conversion functions to `conversions.go`
4. Create `clients.go`
5. Create `clients_test.go`
6. Run `go test ./internal/client/api/... -v`
7. Run `go build ./...`

## Verification

```shell
go test ./internal/client/api/... -v
go build ./...
```

## Important Notes

- `CreateClientKeyResponse` uses `JSON200` not `JSON201` — this is unusual, check carefully
- The `Client` type is a union (`json.RawMessage`). Use `AsFullClient()` first, fallback to `AsPartialClient()`. The `Clients` list response contains `[]Client` where each is a union.
- Key IDs (`kid`) are strings, NOT UUIDs
- `DeleteClientKeyByIdWithResponse` takes `clientId ClientId` (UUID) and `keyId Kid` (string)
- The `conversions.go` file needs `"fmt"` import for the `clientInfoFromGenerated` error case
