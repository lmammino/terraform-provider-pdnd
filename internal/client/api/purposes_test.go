package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

const cannedStandalonePurposeJSON = `{
  "id": "990e8400-e29b-41d4-a716-446655440000",
  "eserviceId": "550e8400-e29b-41d4-a716-446655440000",
  "consumerId": "880e8400-e29b-41d4-a716-446655440003",
  "title": "Test Purpose",
  "description": "A test purpose description",
  "createdAt": "2024-01-01T00:00:00Z",
  "isRiskAnalysisValid": true,
  "isFreeOfCharge": false,
  "currentVersion": {
    "id": "aa0e8400-e29b-41d4-a716-446655440010",
    "state": "DRAFT",
    "dailyCalls": 1000,
    "createdAt": "2024-01-01T00:00:00Z"
  }
}`

const cannedPurposeVersionJSON = `{
  "id": "aa0e8400-e29b-41d4-a716-446655440010",
  "state": "ACTIVE",
  "dailyCalls": 2000,
  "createdAt": "2024-01-01T00:00:00Z"
}`

var testPurposeID = uuid.MustParse("990e8400-e29b-41d4-a716-446655440000")

func TestCreatePurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedStandalonePurposeJSON))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	result, err := api.CreatePurpose(context.Background(), PurposeSeed{
		EServiceID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Title:          "Test Purpose",
		Description:    "A test purpose description",
		DailyCalls:     1000,
		IsFreeOfCharge: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/purposes" {
		t.Errorf("expected path /purposes, got %s", gotPath)
	}
	if gotBody["title"] != "Test Purpose" {
		t.Errorf("expected title 'Test Purpose', got %v", gotBody["title"])
	}
	if result.Title != "Test Purpose" {
		t.Errorf("expected title 'Test Purpose', got %s", result.Title)
	}
	if result.CurrentVersion == nil {
		t.Error("expected currentVersion to be set")
	}
}

func TestGetPurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedStandalonePurposeJSON))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	result, err := api.GetPurpose(context.Background(), testPurposeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String() {
		t.Errorf("expected path /purposes/%s, got %s", testPurposeID, gotPath)
	}
	if result.ID != testPurposeID {
		t.Errorf("expected ID %s, got %s", testPurposeID, result.ID)
	}
}

func TestDeletePurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	err := api.DeletePurpose(context.Background(), testPurposeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String() {
		t.Errorf("expected path /purposes/%s, got %s", testPurposeID, gotPath)
	}
}

func TestUpdateDraftPurpose_Contract(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedStandalonePurposeJSON))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	title := "Updated Title"
	_, err := api.UpdateDraftPurpose(context.Background(), testPurposeID, PurposeDraftUpdate{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected method PATCH, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String() {
		t.Errorf("expected path /purposes/%s, got %s", testPurposeID, gotPath)
	}
	if gotContentType != "application/merge-patch+json" {
		t.Errorf("expected content type application/merge-patch+json, got %s", gotContentType)
	}
	if gotBody["title"] != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %v", gotBody["title"])
	}
}

func TestActivatePurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string

	activePurpose := `{
		"id": "990e8400-e29b-41d4-a716-446655440000",
		"eserviceId": "550e8400-e29b-41d4-a716-446655440000",
		"consumerId": "880e8400-e29b-41d4-a716-446655440003",
		"title": "Test Purpose",
		"description": "A test purpose",
		"createdAt": "2024-01-01T00:00:00Z",
		"isRiskAnalysisValid": true,
		"isFreeOfCharge": false,
		"currentVersion": {"id": "aa0e8400-e29b-41d4-a716-446655440010", "state": "ACTIVE", "dailyCalls": 1000, "createdAt": "2024-01-01T00:00:00Z"}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(activePurpose))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	result, err := api.ActivatePurpose(context.Background(), testPurposeID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String()+"/activate" {
		t.Errorf("expected path /purposes/%s/activate, got %s", testPurposeID, gotPath)
	}
	if result.CurrentVersion == nil || result.CurrentVersion.State != "ACTIVE" {
		t.Error("expected currentVersion.State to be ACTIVE")
	}
}

func TestSuspendPurpose_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedStandalonePurposeJSON))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	_, err := api.SuspendPurpose(context.Background(), testPurposeID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String()+"/suspend" {
		t.Errorf("expected path ending in /suspend, got %s", gotPath)
	}
}

func TestCreatePurposeVersion_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedPurposeVersionJSON))
	}))
	defer server.Close()

	api := NewPurposesClient(newTestClient(t, server))
	result, err := api.CreatePurposeVersion(context.Background(), testPurposeID, PurposeVersionSeed{DailyCalls: 2000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/purposes/"+testPurposeID.String()+"/versions" {
		t.Errorf("expected path ending in /versions, got %s", gotPath)
	}
	if gotBody["dailyCalls"] != float64(2000) {
		t.Errorf("expected dailyCalls 2000, got %v", gotBody["dailyCalls"])
	}
	if result.DailyCalls != 2000 {
		t.Errorf("expected result dailyCalls 2000, got %d", result.DailyCalls)
	}
}
