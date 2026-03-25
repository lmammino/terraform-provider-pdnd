# Agent 2: Transport / Auth Layer + Tests

## Objective

Implement the DPoP proof JWT generation, authenticated HTTP transport, retry logic, and error mapping for the PDND API. Write all transport unit tests. When done, `go test ./internal/client/... -v` must pass with all tests listed below.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

The repo already has Agent 1 output:
- Go module initialized with dependencies
- `internal/client/generated/client.gen.go` — generated API client
- `internal/provider/provider.go` — stub provider
- `Makefile`, etc.

## PDND Authentication Requirements

The PDND API uses DPoP (Demonstration of Proof-of-Possession) authentication per RFC 9449. Every request requires two headers:

1. **`Authorization: DPoP <access_token>`** — Bearer-style but with "DPoP" scheme
2. **`DPoP: <proof_jwt>`** — A fresh JWT proof for each request

The security schemes from the OpenAPI spec:
```yaml
securitySchemes:
  DPoPAuth:
    type: http
    scheme: DPoP
    description: Authorization header with DPoP scheme
  DPoPProofHeader:
    type: apiKey
    in: header
    name: DPoP
    description: DPoP proof JWT
```

Both are required globally on all endpoints.

## Files to Create

### 1. `internal/client/dpop.go` — DPoP Proof JWT Generator

The DPoP proof is a JWT signed with the provider's private key.

**JWT Header claims:**
- `typ`: `"dpop+jwt"` (literal string)
- `alg`: algorithm derived from key type:
  - RSA key → `"RS256"`
  - EC P-256 key → `"ES256"`
  - EC P-384 key → `"ES384"`
- `jwk`: the **public** key in JWK format (JSON object embedded in header)

**JWT Payload claims:**
- `jti`: UUID v4 string — unique per request, use `uuid.New().String()`
- `iat`: Unix timestamp (int64) — current time
- `htm`: HTTP method string — e.g., `"GET"`, `"POST"`, `"DELETE"`
- `htu`: HTTP URL — scheme + host + path ONLY (strip query string and fragment)

**Key types to support:**
- Parse PEM-encoded private keys: PKCS8 (`BEGIN PRIVATE KEY`), PKCS1 RSA (`BEGIN RSA PRIVATE KEY`), SEC1 EC (`BEGIN EC PRIVATE KEY`)
- Extract public key from private key for JWK header
- Build JWK representation: for RSA include `kty`, `n`, `e`; for EC include `kty`, `crv`, `x`, `y`

```go
package client

type DPoPProofGenerator struct {
    privateKey crypto.Signer
    keyID      string
    algorithm  string              // "RS256", "ES256", etc.
    publicJWK  map[string]interface{}
    clock      func() time.Time   // injectable for testing, defaults to time.Now
}

// NewDPoPProofGenerator parses a PEM-encoded private key and creates a generator.
// Returns error if PEM is invalid or key type is unsupported.
func NewDPoPProofGenerator(pemKey []byte, keyID string) (*DPoPProofGenerator, error)

// GenerateProof creates a signed DPoP proof JWT for the given HTTP method and URL.
// The URL is stripped of query string and fragment per RFC 9449.
func (g *DPoPProofGenerator) GenerateProof(method, rawURL string) (string, error)
```

Use `github.com/golang-jwt/jwt/v5` for JWT signing. The JWT must be signed with the private key using the appropriate signing method (`jwt.SigningMethodRS256`, `jwt.SigningMethodES256`, etc.).

**Important**: The `jwk` field goes in the JWT **header**, not payload. You'll need to use `jwt.Token` directly and set header fields via `token.Header["jwk"]` and `token.Header["typ"]`.

**htu construction**: Parse the URL, reconstruct as `scheme://host/path` only. Example: `https://api.example.com/v3/agreements?offset=0` → `https://api.example.com/v3/agreements`

### 2. `internal/client/transport.go` — DPoP HTTP Transport

An `http.RoundTripper` that attaches DPoP headers to every outgoing request.

```go
package client

type DPoPTransport struct {
    Base        http.RoundTripper  // underlying transport (e.g., http.DefaultTransport)
    AccessToken string
    ProofGen    *DPoPProofGenerator
}

func (t *DPoPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // 1. Generate proof for this request's method and URL
    // 2. Clone the request (don't modify the original)
    // 3. Set Authorization: DPoP <access_token>
    // 4. Set DPoP: <proof_jwt>
    // 5. Forward to base transport
}
```

Clone the request before modifying headers (standard Go http.RoundTripper practice).

### 3. `internal/client/retry.go` — Retry Transport

Wraps another `http.RoundTripper` with retry logic for transient errors.

```go
package client

type RetryTransport struct {
    Base       http.RoundTripper
    MaxRetries int           // default: 3
    MinBackoff time.Duration // default: 500ms
    MaxBackoff time.Duration // default: 10s
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error)
```

**Retry behavior:**
- Retry on: 429 (Too Many Requests), 502, 503, 504
- Do NOT retry on: 400, 401, 403, 404, 409, 500, 501, or any 2xx
- Exponential backoff with jitter between retries
- Buffer the request body before first attempt so it can be replayed on retries
- On retry, the inner DPoPTransport generates a fresh proof (new jti, new iat)
- Read and discard the previous response body before retrying

### 4. `internal/client/errors.go` — API Error Mapping

Parse PDND API error responses (application/problem+json) into structured Go errors.

```go
package client

// APIError represents a PDND API error response (RFC 7807 Problem Details)
type APIError struct {
    StatusCode    int
    ProblemType   string       // "type" field from Problem JSON
    Title         string
    Detail        string
    CorrelationID string
    Errors        []ProblemError
}

type ProblemError struct {
    Code   string  // pattern: "^[0-9]{3}-[0-9]{4}$"
    Detail string
}

func (e *APIError) Error() string  // format: "PDND API error (HTTP {status}): {title} - {detail}"

// IsNotFound returns true if the error is a 404 Not Found
func IsNotFound(err error) bool

// IsConflict returns true if the error is a 409 Conflict
func IsConflict(err error) bool

// CheckResponse reads the response and returns an APIError if status is not 2xx.
// The response body is read and closed.
// For 2xx responses, returns nil and leaves the body unread.
func CheckResponse(resp *http.Response) error
```

The Problem JSON schema from the OpenAPI spec:
```json
{
  "type": "string (URI)",          // required
  "status": "integer (100-600)",    // required
  "title": "string (max 64)",       // required
  "correlationId": "string",        // optional
  "detail": "string (max 4096)",    // optional
  "errors": [{"code": "123-4567", "detail": "..."}]  // optional
}
```

Content-Type for errors is `application/problem+json`. Parse the body as JSON. If body is not valid JSON, construct an APIError from the status code alone.

## Test Files to Create

### 5. `internal/client/dpop_test.go`

```go
package client_test
```

**Required tests:**

| Test Function | What to Assert |
|---------------|---------------|
| `TestNewDPoPProofGenerator_RSA` | Successfully creates generator from RSA PEM key |
| `TestNewDPoPProofGenerator_EC` | Successfully creates generator from EC P-256 PEM key |
| `TestNewDPoPProofGenerator_InvalidPEM` | Returns error for invalid PEM |
| `TestGenerateProof_ContainsCorrectMethod` | Decode proof JWT, verify `htm` claim matches request method (test GET, POST, DELETE) |
| `TestGenerateProof_ContainsCorrectURL` | Decode proof JWT, verify `htu` = scheme+host+path (no query). Test URL `https://api.example.com/v3/agreements?offset=0` → `htu` = `https://api.example.com/v3/agreements` |
| `TestGenerateProof_UniqueJTI` | Generate two proofs, verify `jti` values are different |
| `TestGenerateProof_IATIsRecent` | Verify `iat` claim is within ±5 seconds of current time |
| `TestGenerateProof_HeaderTyp` | Verify JWT header `typ` = `"dpop+jwt"` |
| `TestGenerateProof_HeaderJWK` | Verify JWT header contains `jwk` with valid public key |
| `TestGenerateProof_SignatureValid` | Verify the JWT signature is valid using the public key |

To generate test RSA keys:
```go
func generateTestRSAKey() (*rsa.PrivateKey, []byte) {
    key, _ := rsa.GenerateKey(rand.Reader, 2048)
    pemBytes := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(key),
    })
    return key, pemBytes
}
```

To decode and verify proof claims in tests, parse the JWT with the known public key.

### 6. `internal/client/transport_test.go`

```go
package client_test
```

Use `httptest.NewServer` to capture requests.

**Required tests:**

| Test Function | What to Assert |
|---------------|---------------|
| `TestDPoPTransport_HeaderAttached` | Start httptest server, make request via DPoPTransport, verify server received `DPoP` header |
| `TestDPoPTransport_AuthorizationHeaderFormat` | Verify server received `Authorization` header = `DPoP test-token` (exact format) |
| `TestDPoPTransport_ProofMethodMatchesRequest` | Make GET request, decode DPoP header JWT, verify `htm` = `GET`. Make POST, verify `htm` = `POST` |
| `TestDPoPTransport_ProofURLMatchesRequest` | Make request to httptest server URL + `/v3/agreements`, verify `htu` matches (without query) |

### 7. `internal/client/retry_test.go`

```go
package client_test
```

**Required tests:**

| Test Function | What to Assert |
|---------------|---------------|
| `TestRetryTransport_RetryOn429` | Mock transport returns 429 twice then 200. Verify 3 total calls. |
| `TestRetryTransport_RetryOn502` | Mock returns 502 then 200. Verify 2 calls. |
| `TestRetryTransport_RetryOn503` | Mock returns 503 then 200. Verify 2 calls. |
| `TestRetryTransport_RetryOn504` | Mock returns 504 then 200. Verify 2 calls. |
| `TestRetryTransport_NoRetryOn200` | Mock returns 200. Verify exactly 1 call. |
| `TestRetryTransport_NoRetryOn400` | Mock returns 400. Verify exactly 1 call. |
| `TestRetryTransport_NoRetryOn401` | Mock returns 401. Verify exactly 1 call. |
| `TestRetryTransport_NoRetryOn404` | Mock returns 404. Verify exactly 1 call. |
| `TestRetryTransport_NoRetryOn500` | Mock returns 500. Verify exactly 1 call. |
| `TestRetryTransport_MaxRetriesExhausted` | Mock always returns 429. Verify MaxRetries+1 total calls, returns 429 error. |
| `TestRetryTransport_BodyReplayable` | POST with body, mock returns 429 then 200. Verify second request has the same body. |

For mock transport in retry tests, create a simple `roundTripFunc` type:
```go
type roundTripFunc func(*http.Request) (*http.Response, error)
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
```

### 8. Additional: Config Validation Tests

Add these in `dpop_test.go` or `transport_test.go`:

| Test Function | What to Assert |
|---------------|---------------|
| `TestNewDPoPProofGenerator_NilKey` | Returns error when PEM bytes are nil/empty |
| `TestDPoPTransport_EmptyToken` | Verify error or empty auth header when token is "" |

## Execution Steps

1. Read existing code in `internal/client/generated/` to understand what's already there
2. Create all 4 source files: `dpop.go`, `transport.go`, `retry.go`, `errors.go`
3. Create all 3 test files: `dpop_test.go`, `transport_test.go`, `retry_test.go`
4. Run `go mod tidy`
5. Run `go test ./internal/client/... -v`
6. Fix any failures and iterate until ALL tests pass
7. Run `go build ./...` to verify nothing is broken

## Verification

- `go test ./internal/client/... -v` — ALL tests pass
- `go build ./...` — compiles successfully

## Important Notes

- All files go in `internal/client/` package (NOT `internal/client/generated/`)
- Tests should be in `internal/client/` package using `package client_test` (external test package) or `package client` (internal)
- Use `httptest.NewServer` for transport tests, mock `RoundTripper` for retry tests
- The DPoP proof generator must be deterministic given a fixed clock — inject clock function for testing
- Do NOT modify any files in `internal/client/generated/`
- Do NOT modify `internal/provider/provider.go`
