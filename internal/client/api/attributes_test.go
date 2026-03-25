package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var (
	certifiedAttrID = uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440010")
	declaredAttrID  = uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440011")
	verifiedAttrID  = uuid.MustParse("ee0e8400-e29b-41d4-a716-446655440012")
)

const cannedCertifiedAttributeJSON = `{
  "id": "cc0e8400-e29b-41d4-a716-446655440010",
  "name": "Comuni",
  "description": "Certified attribute for municipalities",
  "code": "L6",
  "origin": "IPA",
  "createdAt": "2024-01-15T10:30:00Z"
}`

const cannedDeclaredAttributeJSON = `{
  "id": "dd0e8400-e29b-41d4-a716-446655440011",
  "name": "ISO 27001",
  "description": "Declared ISO 27001 compliance",
  "createdAt": "2024-01-15T10:30:00Z"
}`

const cannedVerifiedAttributeJSON = `{
  "id": "ee0e8400-e29b-41d4-a716-446655440012",
  "name": "SPID",
  "description": "Verified SPID identity provider",
  "createdAt": "2024-01-15T10:30:00Z"
}`

func TestGetCertifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedCertifiedAttributeJSON))
	}))
	defer server.Close()

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.GetCertifiedAttribute(context.Background(), certifiedAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, certifiedAttrID.String()) {
		t.Errorf("expected path to contain %s, got %s", certifiedAttrID, gotPath)
	}
	if result.ID != certifiedAttrID {
		t.Errorf("expected ID %s, got %s", certifiedAttrID, result.ID)
	}
	if result.Name != "Comuni" {
		t.Errorf("expected name Comuni, got %s", result.Name)
	}
	if result.Code != "L6" {
		t.Errorf("expected code L6, got %s", result.Code)
	}
	if result.Origin != "IPA" {
		t.Errorf("expected origin IPA, got %s", result.Origin)
	}
	if result.Description != "Certified attribute for municipalities" {
		t.Errorf("expected description 'Certified attribute for municipalities', got %s", result.Description)
	}
}

func TestListCertifiedAttributes_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedCertifiedAttributeJSON + `],
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

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.ListCertifiedAttributes(context.Background(), PaginationParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/certifiedAttributes" {
		t.Errorf("expected path /certifiedAttributes, got %s", gotPath)
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
	if result.Results[0].ID != certifiedAttrID {
		t.Errorf("expected attribute ID %s, got %s", certifiedAttrID, result.Results[0].ID)
	}
	if result.Results[0].Name != "Comuni" {
		t.Errorf("expected name Comuni, got %s", result.Results[0].Name)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestGetDeclaredAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDeclaredAttributeJSON))
	}))
	defer server.Close()

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.GetDeclaredAttribute(context.Background(), declaredAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, declaredAttrID.String()) {
		t.Errorf("expected path to contain %s, got %s", declaredAttrID, gotPath)
	}
	if result.ID != declaredAttrID {
		t.Errorf("expected ID %s, got %s", declaredAttrID, result.ID)
	}
	if result.Name != "ISO 27001" {
		t.Errorf("expected name 'ISO 27001', got %s", result.Name)
	}
	if result.Description != "Declared ISO 27001 compliance" {
		t.Errorf("expected description 'Declared ISO 27001 compliance', got %s", result.Description)
	}
}

func TestListDeclaredAttributes_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedDeclaredAttributeJSON + `],
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

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.ListDeclaredAttributes(context.Background(), PaginationParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/declaredAttributes" {
		t.Errorf("expected path /declaredAttributes, got %s", gotPath)
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
	if result.Results[0].ID != declaredAttrID {
		t.Errorf("expected attribute ID %s, got %s", declaredAttrID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestGetVerifiedAttribute_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedVerifiedAttributeJSON))
	}))
	defer server.Close()

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.GetVerifiedAttribute(context.Background(), verifiedAttrID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, verifiedAttrID.String()) {
		t.Errorf("expected path to contain %s, got %s", verifiedAttrID, gotPath)
	}
	if result.ID != verifiedAttrID {
		t.Errorf("expected ID %s, got %s", verifiedAttrID, result.ID)
	}
	if result.Name != "SPID" {
		t.Errorf("expected name SPID, got %s", result.Name)
	}
	if result.Description != "Verified SPID identity provider" {
		t.Errorf("expected description 'Verified SPID identity provider', got %s", result.Description)
	}
}

func TestListVerifiedAttributes_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedVerifiedAttributeJSON + `],
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

	api := NewAttributesClient(newTestClient(t, server))
	result, err := api.ListVerifiedAttributes(context.Background(), PaginationParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/verifiedAttributes" {
		t.Errorf("expected path /verifiedAttributes, got %s", gotPath)
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
	if result.Results[0].ID != verifiedAttrID {
		t.Errorf("expected attribute ID %s, got %s", verifiedAttrID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}
