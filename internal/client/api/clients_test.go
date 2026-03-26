package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var (
	testClientID  = uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440040")
	testClientKID = "test-key-id-001"
)

const cannedClientJSON = `{
  "id": "cc0e8400-e29b-41d4-a716-446655440040",
  "consumerId": "880e8400-e29b-41d4-a716-446655440003",
  "name": "Test Client",
  "description": "A test client",
  "createdAt": "2024-01-01T00:00:00Z"
}`

const cannedKeyResponseJSON = `{
  "clientId": "cc0e8400-e29b-41d4-a716-446655440040",
  "jwk": {"kid": "test-key-id-001", "kty": "RSA", "alg": "RS256", "use": "sig"}
}`

func TestGetClient_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedClientJSON))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	result, err := api.GetClient(context.Background(), testClientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if result.ID != testClientID {
		t.Errorf("expected ID %s, got %s", testClientID, result.ID)
	}
	if result.Name != "Test Client" {
		t.Errorf("expected name 'Test Client', got %s", result.Name)
	}
	if result.Description == nil || *result.Description != "A test client" {
		t.Errorf("expected description 'A test client', got %v", result.Description)
	}
	expectedConsumerID := uuid.MustParse("880e8400-e29b-41d4-a716-446655440003")
	if result.ConsumerID != expectedConsumerID {
		t.Errorf("expected ConsumerID %s, got %s", expectedConsumerID, result.ConsumerID)
	}
}

func TestListClients_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedClientJSON + `],
		"pagination": {"offset": 0, "limit": 50, "totalCount": 1}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	testName := "Test"
	api := NewClientsClient(newTestClient(t, server))
	result, err := api.ListClients(context.Background(), ListClientsParams{
		Name:   &testName,
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/clients" {
		t.Errorf("expected path /clients, got %s", gotPath)
	}
	if !strings.Contains(gotQuery, "offset=0") {
		t.Errorf("expected query to contain offset=0, got %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=50") {
		t.Errorf("expected query to contain limit=50, got %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "name=Test") {
		t.Errorf("expected query to contain name=Test, got %s", gotQuery)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].ID != testClientID {
		t.Errorf("expected client ID %s, got %s", testClientID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestListClientKeys_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [{"kid": "test-key-id-001", "kty": "RSA", "alg": "RS256", "use": "sig"}],
		"pagination": {"offset": 0, "limit": 50, "totalCount": 1}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	result, err := api.ListClientKeys(context.Background(), testClientID, 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String() + "/keys"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if !strings.Contains(gotQuery, "offset=0") {
		t.Errorf("expected query to contain offset=0, got %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=50") {
		t.Errorf("expected query to contain limit=50, got %s", gotQuery)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].Kid != testClientKID {
		t.Errorf("expected kid %s, got %s", testClientKID, result.Results[0].Kid)
	}
	if result.Results[0].Kty != "RSA" {
		t.Errorf("expected kty RSA, got %s", result.Results[0].Kty)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestCreateClientKey_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedKeyResponseJSON))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	result, err := api.CreateClientKey(context.Background(), testClientID, ClientKeySeed{
		Key:  "base64-pem-key",
		Use:  "SIG",
		Alg:  "RS256",
		Name: "my-test-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String() + "/keys"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotBody["key"] != "base64-pem-key" {
		t.Errorf("expected key 'base64-pem-key', got %v", gotBody["key"])
	}
	if gotBody["use"] != "SIG" {
		t.Errorf("expected use 'SIG', got %v", gotBody["use"])
	}
	if gotBody["alg"] != "RS256" {
		t.Errorf("expected alg 'RS256', got %v", gotBody["alg"])
	}
	if gotBody["name"] != "my-test-key" {
		t.Errorf("expected name 'my-test-key', got %v", gotBody["name"])
	}
	if result.ClientID != testClientID {
		t.Errorf("expected ClientID %s, got %s", testClientID, result.ClientID)
	}
	if result.Key.Kid != testClientKID {
		t.Errorf("expected kid %s, got %s", testClientKID, result.Key.Kid)
	}
}

func TestDeleteClientKey_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	err := api.DeleteClientKey(context.Background(), testClientID, testClientKID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String() + "/keys/" + testClientKID
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
}

func TestAddClientPurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	purposeID := uuid.MustParse("ff0e8400-e29b-41d4-a716-446655440050")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	err := api.AddClientPurpose(context.Background(), testClientID, purposeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String() + "/purposes"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotBody["purposeId"] != purposeID.String() {
		t.Errorf("expected purposeId %s, got %v", purposeID, gotBody["purposeId"])
	}
}

func TestRemoveClientPurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string

	purposeID := uuid.MustParse("ff0e8400-e29b-41d4-a716-446655440050")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewClientsClient(newTestClient(t, server))
	err := api.RemoveClientPurpose(context.Background(), testClientID, purposeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/clients/" + testClientID.String() + "/purposes/" + purposeID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
}
