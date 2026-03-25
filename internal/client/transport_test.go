package client_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
)

func TestDPoPTransport_HeaderAttached(t *testing.T) {
	_, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedDPoP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedDPoP = r.Header.Get("DPoP")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: "test-token",
		ProofGen:    gen,
	}

	httpClient := &http.Client{Transport: transport}
	_, err = httpClient.Get(server.URL + "/v3/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedDPoP == "" {
		t.Error("DPoP header was not attached to the request")
	}
}

func TestDPoPTransport_AuthorizationHeaderFormat(t *testing.T) {
	_, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: "test-token",
		ProofGen:    gen,
	}

	httpClient := &http.Client{Transport: transport}
	_, err = httpClient.Get(server.URL + "/v3/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "DPoP test-token"
	if receivedAuth != expected {
		t.Errorf("expected Authorization=%q, got %q", expected, receivedAuth)
	}
}

func TestDPoPTransport_ProofMethodMatchesRequest(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedDPoP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedDPoP = r.Header.Get("DPoP")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: "test-token",
		ProofGen:    gen,
	}
	httpClient := &http.Client{Transport: transport}

	methods := []string{"GET", "POST"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, _ := http.NewRequest(method, server.URL+"/v3/test", nil)
			_, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			parser := jwt.NewParser(jwt.WithoutClaimsValidation())
			token, err := parser.Parse(receivedDPoP, func(token *jwt.Token) (interface{}, error) {
				return &key.PublicKey, nil
			})
			if err != nil {
				t.Fatalf("failed to parse DPoP token: %v", err)
			}

			claims, _ := token.Claims.(jwt.MapClaims)
			htm, _ := claims["htm"].(string)
			if htm != method {
				t.Errorf("expected htm=%q, got %q", method, htm)
			}
		})
	}
}

func TestDPoPTransport_ProofURLMatchesRequest(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedDPoP string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedDPoP = r.Header.Get("DPoP")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: "test-token",
		ProofGen:    gen,
	}
	httpClient := &http.Client{Transport: transport}

	_, err = httpClient.Get(server.URL + "/v3/agreements?offset=0&limit=10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.Parse(receivedDPoP, func(token *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse DPoP token: %v", err)
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	htu, _ := claims["htu"].(string)

	// htu should be the server URL + path, without query string.
	expectedPrefix := server.URL + "/v3/agreements"
	if htu != expectedPrefix {
		t.Errorf("expected htu=%q, got %q", expectedPrefix, htu)
	}
	if strings.Contains(htu, "?") {
		t.Errorf("htu should not contain query string, got %q", htu)
	}
}

func TestDPoPTransport_EmptyToken(t *testing.T) {
	_, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: "",
		ProofGen:    gen,
	}

	httpClient := &http.Client{Transport: transport}
	_, err = httpClient.Get(server.URL + "/v3/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With empty token, the header should still start with "DPoP".
	if receivedAuth != "DPoP " && receivedAuth != "DPoP" {
		t.Errorf("expected Authorization to start with 'DPoP', got %q", receivedAuth)
	}
}
