package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenProvider abstracts how an access token is obtained.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// StaticTokenProvider wraps a pre-obtained token string.
type StaticTokenProvider struct {
	token string
}

// NewStaticTokenProvider creates a TokenProvider that always returns the given token.
func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{token: token}
}

// Token returns the static token.
func (p *StaticTokenProvider) Token(_ context.Context) (string, error) {
	return p.token, nil
}

// AutoTokenProvider obtains and caches tokens from the PDND auth server.
type AutoTokenProvider struct {
	mu            sync.Mutex
	clientID      string
	purposeID     string
	tokenEndpoint string
	proofGen      *DPoPProofGenerator
	httpClient    *http.Client
	cachedToken   string
	tokenExpiry   time.Time
	clock         func() time.Time
}

// NewAutoTokenProvider creates a TokenProvider that automatically obtains tokens
// using client credentials and a signed JWT assertion.
func NewAutoTokenProvider(clientID, purposeID, tokenEndpoint string, proofGen *DPoPProofGenerator, httpClient *http.Client) *AutoTokenProvider {
	return &AutoTokenProvider{
		clientID:      clientID,
		purposeID:     purposeID,
		tokenEndpoint: tokenEndpoint,
		proofGen:      proofGen,
		httpClient:    httpClient,
		clock:         time.Now,
	}
}

// SetClock overrides the clock function used for token expiry calculations.
// This is intended for testing.
func (p *AutoTokenProvider) SetClock(clock func() time.Time) {
	p.clock = clock
}

// Token returns a cached access token or obtains a new one from the auth server.
func (p *AutoTokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if cached token is still valid (with 30s safety margin).
	if p.cachedToken != "" && p.clock().Before(p.tokenExpiry.Add(-30*time.Second)) {
		return p.cachedToken, nil
	}

	assertion, err := p.buildClientAssertion()
	if err != nil {
		return "", fmt.Errorf("building client assertion: %w", err)
	}

	formData := url.Values{
		"grant_type":            {"client_credentials"},
		"client_id":             {p.clientID},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {assertion},
	}

	resp, err := p.httpClient.PostForm(p.tokenEndpoint, formData)
	if err != nil {
		return "", fmt.Errorf("requesting token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	p.cachedToken = tokenResp.AccessToken
	p.tokenExpiry = p.clock().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return p.cachedToken, nil
}

// buildClientAssertion creates a signed JWT for client authentication.
func (p *AutoTokenProvider) buildClientAssertion() (string, error) {
	now := p.clock()

	claims := jwt.MapClaims{
		"iss":       p.clientID,
		"sub":       p.clientID,
		"aud":       p.tokenEndpoint + "/client-assertion",
		"jti":       uuid.New().String(),
		"iat":       now.Unix(),
		"exp":       now.Add(5 * time.Minute).Unix(),
		"purposeId": p.purposeID,
	}

	var signingMethod jwt.SigningMethod
	switch p.proofGen.algorithm {
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "ES256":
		signingMethod = jwt.SigningMethodES256
	case "ES384":
		signingMethod = jwt.SigningMethodES384
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", p.proofGen.algorithm)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = p.proofGen.keyID
	token.Header["typ"] = "JWT"

	return token.SignedString(p.proofGen.privateKey)
}
