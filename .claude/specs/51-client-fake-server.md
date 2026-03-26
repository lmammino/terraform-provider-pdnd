# Spec 51: Fake Server for Client Key Management

## Objective

Extend the fake PDND server with client, key, and client-purpose endpoints. When done, `go build ./...` must pass and all existing tests must continue to pass.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/testing/fakepdnd/delegations.go` — recent parameterized handler pattern
2. `internal/testing/fakepdnd/state.go` — storage type definitions
3. `internal/testing/fakepdnd/server.go` — FakeServer struct, NewFakeServer, setupRoutes, Seed/Get methods
4. `internal/testing/fakepdnd/helpers.go` — writeJSON, writeProblem, parseUUID, parseIntDefault (in agreements.go)
5. `internal/testing/fakepdnd/descriptor_documents.go` — handler pattern with multipart (reference for structure)

## Files to Create/Modify

### 1. Extend `internal/testing/fakepdnd/state.go` — Add types

```go
// StoredClient represents a client in the fake server's state.
type StoredClient struct {
    ID          uuid.UUID
    ConsumerID  uuid.UUID
    Name        string
    Description string
    CreatedAt   time.Time
}

// StoredClientKey represents a JWK key on a client.
type StoredClientKey struct {
    Kid      string
    Kty      string
    Alg      string
    Use      string
    Name     string
    Key      string // raw PEM key (stored for reference)
}
```

### 2. Extend `internal/testing/fakepdnd/server.go`

**Add to FakeServer struct:**
```go
clients          map[uuid.UUID]*StoredClient
clientKeys       map[uuid.UUID][]StoredClientKey       // clientID -> keys
clientPurposes   map[uuid.UUID]map[uuid.UUID]bool      // clientID -> purposeID -> exists
```

**Add to NewFakeServer initialization:**
```go
clients:        make(map[uuid.UUID]*StoredClient),
clientKeys:     make(map[uuid.UUID][]StoredClientKey),
clientPurposes: make(map[uuid.UUID]map[uuid.UUID]bool),
```

**Add seed/getter methods:**
```go
func (s *FakeServer) SeedClient(c StoredClient) {
    s.mu.Lock(); defer s.mu.Unlock()
    s.clients[c.ID] = &c
}

func (s *FakeServer) GetClient(id uuid.UUID) *StoredClient {
    s.mu.RLock(); defer s.mu.RUnlock()
    return s.clients[id]
}

func (s *FakeServer) SeedClientKey(clientID uuid.UUID, key StoredClientKey) {
    s.mu.Lock(); defer s.mu.Unlock()
    s.clientKeys[clientID] = append(s.clientKeys[clientID], key)
}

func (s *FakeServer) GetClientKeys(clientID uuid.UUID) []StoredClientKey {
    s.mu.RLock(); defer s.mu.RUnlock()
    return s.clientKeys[clientID]
}

func (s *FakeServer) SeedClientPurpose(clientID, purposeID uuid.UUID) {
    s.mu.Lock(); defer s.mu.Unlock()
    if s.clientPurposes[clientID] == nil {
        s.clientPurposes[clientID] = make(map[uuid.UUID]bool)
    }
    s.clientPurposes[clientID][purposeID] = true
}

func (s *FakeServer) GetClientPurposes(clientID uuid.UUID) map[uuid.UUID]bool {
    s.mu.RLock(); defer s.mu.RUnlock()
    return s.clientPurposes[clientID]
}
```

**Add routes in setupRoutes** (before delegation routes):
```go
// Client routes.
s.mux.HandleFunc("GET /clients", s.handleListClients)
s.mux.HandleFunc("GET /clients/{clientId}", s.handleGetClient)
s.mux.HandleFunc("GET /clients/{clientId}/keys", s.handleListClientKeys)
s.mux.HandleFunc("POST /clients/{clientId}/keys", s.handleCreateClientKey)
s.mux.HandleFunc("DELETE /clients/{clientId}/keys/{keyId}", s.handleDeleteClientKey)
s.mux.HandleFunc("POST /clients/{clientId}/purposes", s.handleAddClientPurpose)
s.mux.HandleFunc("DELETE /clients/{clientId}/purposes/{purposeId}", s.handleRemoveClientPurpose)
```

### 3. Create `internal/testing/fakepdnd/clients.go` — Handlers

**`clientToJSON(c *StoredClient) map[string]interface{}`:**
Return FullClient-compatible JSON:
```go
m := map[string]interface{}{
    "id": c.ID.String(), "consumerId": c.ConsumerID.String(),
    "name": c.Name, "createdAt": c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
}
if c.Description != "" { m["description"] = c.Description }
return m
```

**`clientKeyToJWK(k *StoredClientKey) map[string]interface{}`:**
```go
return map[string]interface{}{
    "kid": k.Kid, "kty": k.Kty, "alg": k.Alg, "use": k.Use,
}
```

**`handleListClients`:** Filter by name (case-insensitive contains) and consumerId. Paginate. Return `{"results": [...], "pagination": {...}}`.

**`handleGetClient`:** Look up by ID, return 200 or 404.

**`handleListClientKeys`:** Get keys for clientID, paginate, return `{"results": [...JWK...], "pagination": {...}}`.

**`handleCreateClientKey`:** Parse body (`KeySeed`: key, use, alg, name). Validate client exists. Generate kid as `uuid.New().String()`. Store key. Return 200 with `{"clientId": "...", "jwk": {...}}`.

**`handleDeleteClientKey`:** Parse clientId (UUID) and keyId (string, path value "keyId"). Find and remove key by kid. Return 200 or 404.

**`handleAddClientPurpose`:** Parse body `{"purposeId": "..."}`. Validate client exists. Add to `clientPurposes`. Return 200 with `{}`.

**`handleRemoveClientPurpose`:** Parse clientId and purposeId. Remove from `clientPurposes`. Return 200 with `{}`.

## Execution Steps

1. Read all referenced files
2. Add `StoredClient` and `StoredClientKey` to `state.go`
3. Add storage fields, initialization, seed/getter methods, routes to `server.go`
4. Create `clients.go` with all handlers
5. Run `go build ./...`
6. Run `go test ./internal/testing/fakepdnd/... -v`

## Verification

```shell
go build ./...
go test ./internal/testing/fakepdnd/... -v
```

## Important Notes

- Do NOT break existing tests or handlers
- Client IDs are UUIDs, but key IDs (`kid`) are strings (not UUIDs)
- `handleDeleteClientKey`: the `keyId` path parameter is a string, NOT a UUID — use `r.PathValue("keyId")` directly, don't parse as UUID
- `handleCreateClientKey` returns status 200 (not 201) — match the generated client response
- The Client union type in the real API wraps in `{"union": ...}`, but for the fake server, just return the FullClient JSON directly — the generated client's `AsFullClient()` will parse it correctly
- `parseIntDefault` is defined in `agreements.go` (line 472)
