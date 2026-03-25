package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
)

func TestStaticTokenProvider_ReturnsToken(t *testing.T) {
	p := client.NewStaticTokenProvider("my-token")
	tok, err := p.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "my-token" {
		t.Errorf("expected %q, got %q", "my-token", tok)
	}
}

func setupAutoProvider(t *testing.T, handler http.HandlerFunc) (*client.AutoTokenProvider, *httptest.Server, func(time.Time)) {
	t.Helper()
	_, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server := httptest.NewServer(handler)

	provider := client.NewAutoTokenProvider("client-123", "purpose-456", server.URL, gen, server.Client())

	// Set up injectable clock
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	provider.SetClock(func() time.Time { return now })

	setClock := func(newTime time.Time) {
		now = newTime
		provider.SetClock(func() time.Time { return now })
	}

	return provider, server, setClock
}

func tokenHandler(requestCount *atomic.Int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok123",
			"token_type":   "DPoP",
			"expires_in":   600,
		})
	}
}

func TestAutoTokenProvider_ObtainsToken(t *testing.T) {
	var requestCount atomic.Int32
	var receivedMethod string
	var receivedContentType string
	var receivedGrantType string
	var receivedClientID string
	var receivedAssertionType string
	var receivedAssertion string

	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		receivedGrantType = r.FormValue("grant_type")
		receivedClientID = r.FormValue("client_id")
		receivedAssertionType = r.FormValue("client_assertion_type")
		receivedAssertion = r.FormValue("client_assertion")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok123",
			"token_type":   "DPoP",
			"expires_in":   600,
		})
	}

	provider, server, _ := setupAutoProvider(t, handler)
	defer server.Close()

	tok, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "tok123" {
		t.Errorf("expected %q, got %q", "tok123", tok)
	}
	if receivedMethod != "POST" {
		t.Errorf("expected POST, got %q", receivedMethod)
	}
	if receivedContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content type, got %q", receivedContentType)
	}
	if receivedGrantType != "client_credentials" {
		t.Errorf("expected grant_type=client_credentials, got %q", receivedGrantType)
	}
	if receivedClientID != "client-123" {
		t.Errorf("expected client_id=client-123, got %q", receivedClientID)
	}
	if receivedAssertionType != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		t.Errorf("expected jwt-bearer assertion type, got %q", receivedAssertionType)
	}
	if receivedAssertion == "" {
		t.Error("client_assertion was empty")
	}
}

func TestAutoTokenProvider_ClientAssertionFormat(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedAssertion string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		receivedAssertion = r.FormValue("client_assertion")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok123",
			"token_type":   "DPoP",
			"expires_in":   600,
		})
	}))
	defer server.Close()

	provider := client.NewAutoTokenProvider("client-123", "purpose-456", server.URL, gen, server.Client())
	provider.SetClock(func() time.Time {
		return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	})

	_, err = provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the assertion JWT
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.Parse(receivedAssertion, func(token *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse client assertion: %v", err)
	}

	// Check header
	if token.Header["typ"] != "JWT" {
		t.Errorf("expected typ=JWT, got %v", token.Header["typ"])
	}
	if token.Header["alg"] != "RS256" {
		t.Errorf("expected alg=RS256, got %v", token.Header["alg"])
	}
	if token.Header["kid"] != "test-key-id" {
		t.Errorf("expected kid=test-key-id, got %v", token.Header["kid"])
	}

	// Check claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("unexpected claims type")
	}
	if claims["iss"] != "client-123" {
		t.Errorf("expected iss=client-123, got %v", claims["iss"])
	}
	if claims["sub"] != "client-123" {
		t.Errorf("expected sub=client-123, got %v", claims["sub"])
	}
	aud, ok := claims["aud"].(string)
	if !ok {
		t.Fatal("aud claim missing or not a string")
	}
	expectedAud := server.URL + "/client-assertion"
	if aud != expectedAud {
		t.Errorf("expected aud=%q, got %q", expectedAud, aud)
	}
	if claims["purposeId"] != "purpose-456" {
		t.Errorf("expected purposeId=purpose-456, got %v", claims["purposeId"])
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		t.Error("jti claim missing or empty")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		t.Fatal("iat claim missing")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim missing")
	}
	// exp should be iat + 5 minutes
	if int64(exp)-int64(iat) != 300 {
		t.Errorf("expected exp-iat=300, got %v", int64(exp)-int64(iat))
	}
}

func TestAutoTokenProvider_CachesToken(t *testing.T) {
	var requestCount atomic.Int32
	provider, server, _ := setupAutoProvider(t, tokenHandler(&requestCount))
	defer server.Close()

	tok1, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok2, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tok1 != tok2 {
		t.Errorf("expected same token, got %q and %q", tok1, tok2)
	}
	if requestCount.Load() != 1 {
		t.Errorf("expected 1 request, got %d", requestCount.Load())
	}
}

func TestAutoTokenProvider_RefreshesExpired(t *testing.T) {
	var requestCount atomic.Int32
	callNum := atomic.Int32{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		n := callNum.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok" + string(rune('0'+n)),
			"token_type":   "DPoP",
			"expires_in":   60,
		})
	}

	provider, server, setClock := setupAutoProvider(t, handler)
	defer server.Close()

	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// First call at t=0
	setClock(baseTime)
	_, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Advance clock to t=90 (past expiry of 60s)
	setClock(baseTime.Add(90 * time.Second))
	_, err = provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount.Load() != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount.Load())
	}
}

func TestAutoTokenProvider_SafetyMargin(t *testing.T) {
	var requestCount atomic.Int32
	handler := func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok123",
			"token_type":   "DPoP",
			"expires_in":   60,
		})
	}

	provider, server, setClock := setupAutoProvider(t, handler)
	defer server.Close()

	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// First call at t=0, token expires at t=60
	setClock(baseTime)
	_, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At t=25, well within cache period (60 - 30 = 30s effective), should be cached
	setClock(baseTime.Add(25 * time.Second))
	_, err = provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount.Load() != 1 {
		t.Errorf("at t=25: expected 1 request (cached), got %d", requestCount.Load())
	}

	// At t=35, past the safety margin (60 - 30 = 30s), should refresh
	setClock(baseTime.Add(35 * time.Second))
	_, err = provider.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount.Load() != 2 {
		t.Errorf("at t=35: expected 2 requests (refreshed), got %d", requestCount.Load())
	}
}

func TestAutoTokenProvider_HandlesErrorResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}

	provider, server, _ := setupAutoProvider(t, handler)
	defer server.Close()

	_, err := provider.Token(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAutoTokenProvider_ConcurrentAccess(t *testing.T) {
	var requestCount atomic.Int32
	provider, server, _ := setupAutoProvider(t, tokenHandler(&requestCount))
	defer server.Close()

	const goroutines = 10
	var wg sync.WaitGroup
	tokens := make([]string, goroutines)
	errs := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			tokens[idx], errs[idx] = provider.Token(context.Background())
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d got error: %v", i, err)
		}
	}

	for i, tok := range tokens {
		if tok != "tok123" {
			t.Errorf("goroutine %d got token %q, expected %q", i, tok, "tok123")
		}
	}

	if requestCount.Load() != 1 {
		t.Errorf("expected 1 request, got %d", requestCount.Load())
	}
}
