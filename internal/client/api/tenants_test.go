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
	testTenantID     = uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440060")
	testTenantAttrID = uuid.MustParse("da0e8400-e29b-41d4-a716-446655440061")
	testTenantAgrID  = uuid.MustParse("a90e8400-e29b-41d4-a716-446655440062")
)

const cannedTenantJSON = `{
  "id": "dd0e8400-e29b-41d4-a716-446655440060",
  "name": "Test Tenant",
  "kind": "PA",
  "externalId": {"origin": "IPA", "value": "abc123"},
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-06-01T00:00:00Z",
  "onboardedAt": "2024-01-02T00:00:00Z"
}`

const cannedCertifiedAttrJSON = `{
  "id": "da0e8400-e29b-41d4-a716-446655440061",
  "assignedAt": "2024-03-01T00:00:00Z",
  "revokedAt": null
}`

const cannedDeclaredAttrJSON = `{
  "id": "da0e8400-e29b-41d4-a716-446655440061",
  "assignedAt": "2024-03-01T00:00:00Z",
  "revokedAt": null
}`

const cannedVerifiedAttrJSON = `{
  "id": "da0e8400-e29b-41d4-a716-446655440061",
  "assignedAt": "2024-03-01T00:00:00Z"
}`

func TestGetTenant_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedTenantJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.GetTenant(context.Background(), testTenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if result.ID != testTenantID {
		t.Errorf("expected ID %s, got %s", testTenantID, result.ID)
	}
	if result.Name != "Test Tenant" {
		t.Errorf("expected name 'Test Tenant', got %s", result.Name)
	}
	if result.Kind == nil || *result.Kind != "PA" {
		t.Errorf("expected kind 'PA', got %v", result.Kind)
	}
	if result.ExternalID == nil || result.ExternalID.Origin != "IPA" || result.ExternalID.Value != "abc123" {
		t.Errorf("expected externalId {IPA, abc123}, got %v", result.ExternalID)
	}
}

func TestListTenants_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedTenantJSON + `],
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

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.ListTenants(context.Background(), ListTenantsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/tenants" {
		t.Errorf("expected path /tenants, got %s", gotPath)
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
	if result.Results[0].ID != testTenantID {
		t.Errorf("expected tenant ID %s, got %s", testTenantID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestAssignTenantCertifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedCertifiedAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.AssignTenantCertifiedAttribute(context.Background(), testTenantID, testTenantAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/certifiedAttributes"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotBody["id"] != testTenantAttrID.String() {
		t.Errorf("expected body id %s, got %v", testTenantAttrID, gotBody["id"])
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}

func TestRevokeTenantCertifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedCertifiedAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.RevokeTenantCertifiedAttribute(context.Background(), testTenantID, testTenantAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/certifiedAttributes/" + testTenantAttrID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}

func TestAssignTenantDeclaredAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDeclaredAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.AssignTenantDeclaredAttribute(context.Background(), testTenantID, testTenantAttrID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/declaredAttributes"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotBody["id"] != testTenantAttrID.String() {
		t.Errorf("expected body id %s, got %v", testTenantAttrID, gotBody["id"])
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}

func TestRevokeTenantDeclaredAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDeclaredAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.RevokeTenantDeclaredAttribute(context.Background(), testTenantID, testTenantAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/declaredAttributes/" + testTenantAttrID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}

func TestAssignTenantVerifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedVerifiedAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.AssignTenantVerifiedAttribute(context.Background(), testTenantID, testTenantAttrID, testTenantAgrID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/verifiedAttributes"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotBody["id"] != testTenantAttrID.String() {
		t.Errorf("expected body id %s, got %v", testTenantAttrID, gotBody["id"])
	}
	if gotBody["agreementId"] != testTenantAgrID.String() {
		t.Errorf("expected body agreementId %s, got %v", testTenantAgrID, gotBody["agreementId"])
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}

func TestRevokeTenantVerifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedVerifiedAttrJSON))
	}))
	defer server.Close()

	api := NewTenantsClient(newTestClient(t, server))
	result, err := api.RevokeTenantVerifiedAttribute(context.Background(), testTenantID, testTenantAttrID, testTenantAgrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/tenants/" + testTenantID.String() + "/verifiedAttributes/" + testTenantAttrID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if !strings.Contains(gotQuery, "agreementId="+testTenantAgrID.String()) {
		t.Errorf("expected query to contain agreementId=%s, got %s", testTenantAgrID, gotQuery)
	}
	if result.ID != testTenantAttrID {
		t.Errorf("expected ID %s, got %s", testTenantAttrID, result.ID)
	}
}
