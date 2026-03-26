package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	client "github.com/lmammino/terraform-provider-pdnd/internal/client"
)

var (
	testDocumentID = uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440020")
)

const cannedDocumentJSON = `{
  "id": "dd0e8400-e29b-41d4-a716-446655440020",
  "name": "openapi.yaml",
  "prettyName": "API Specification",
  "contentType": "application/yaml",
  "createdAt": "2024-01-15T10:30:00Z"
}`

const cannedDocumentsListJSON = `{
  "results": [{
    "id": "dd0e8400-e29b-41d4-a716-446655440020",
    "name": "openapi.yaml",
    "prettyName": "API Specification",
    "contentType": "application/yaml",
    "createdAt": "2024-01-15T10:30:00Z"
  }],
  "pagination": {
    "offset": 0,
    "limit": 50,
    "totalCount": 1
  }
}`

func TestListDocuments_Contract(t *testing.T) {
	var gotMethod, gotPath, gotOffset, gotLimit string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotOffset = r.URL.Query().Get("offset")
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDocumentsListJSON))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))
	docs, pagination, err := api.ListDocuments(context.Background(), testEServiceID, testDescriptorID, 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/documents"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if gotOffset != "0" {
		t.Errorf("expected offset 0, got %s", gotOffset)
	}
	if gotLimit != "50" {
		t.Errorf("expected limit 50, got %s", gotLimit)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if docs[0].ID != testDocumentID {
		t.Errorf("expected document ID %s, got %s", testDocumentID, docs[0].ID)
	}
	if docs[0].Name != "openapi.yaml" {
		t.Errorf("expected name openapi.yaml, got %s", docs[0].Name)
	}
	if docs[0].PrettyName != "API Specification" {
		t.Errorf("expected prettyName 'API Specification', got %s", docs[0].PrettyName)
	}
	if docs[0].ContentType != "application/yaml" {
		t.Errorf("expected contentType application/yaml, got %s", docs[0].ContentType)
	}
	if pagination.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", pagination.TotalCount)
	}
}

func TestUploadDocument_Contract(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBodyLen int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBodyLen = len(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedDocumentJSON))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))
	doc, err := api.UploadDocument(context.Background(), testEServiceID, testDescriptorID, "openapi.yaml", []byte("openapi: 3.0.0"), "API Specification")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/documents"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got %s", gotContentType)
	}
	if gotBodyLen == 0 {
		t.Error("expected non-empty body")
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.ID != testDocumentID {
		t.Errorf("expected document ID %s, got %s", testDocumentID, doc.ID)
	}
}

func TestDeleteDocument_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))
	err := api.DeleteDocument(context.Background(), testEServiceID, testDescriptorID, testDocumentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/documents/" + testDocumentID.String()
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
}

func TestUploadInterface_Contract(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBodyLen int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBodyLen = len(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedDocumentJSON))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))
	doc, err := api.UploadInterface(context.Background(), testEServiceID, testDescriptorID, "openapi.yaml", []byte("openapi: 3.0.0"), "API Specification")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/interface"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got %s", gotContentType)
	}
	if gotBodyLen == 0 {
		t.Error("expected non-empty body")
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.ID != testDocumentID {
		t.Errorf("expected document ID %s, got %s", testDocumentID, doc.ID)
	}
}

func TestDeleteInterface_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))
	err := api.DeleteInterface(context.Background(), testEServiceID, testDescriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/interface"
	if gotPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, gotPath)
	}
}

func TestGetDocumentByID_Contract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDocumentsListJSON))
	}))
	defer server.Close()

	api := NewDescriptorDocumentsClient(newTestClient(t, server))

	// Test finding an existing document
	doc, err := api.GetDocumentByID(context.Background(), testEServiceID, testDescriptorID, testDocumentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.ID != testDocumentID {
		t.Errorf("expected document ID %s, got %s", testDocumentID, doc.ID)
	}
	if doc.Name != "openapi.yaml" {
		t.Errorf("expected name openapi.yaml, got %s", doc.Name)
	}

	// Test not finding a document
	missingID := uuid.MustParse("ff0e8400-e29b-41d4-a716-446655440099")
	_, err = api.GetDocumentByID(context.Background(), testEServiceID, testDescriptorID, missingID)
	if err == nil {
		t.Fatal("expected error for missing document")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestCheckInterfaceExists_Contract(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("file-content-bytes"))
		}))
		defer server.Close()

		api := NewDescriptorDocumentsClient(newTestClient(t, server))
		exists, err := api.CheckInterfaceExists(context.Background(), testEServiceID, testDescriptorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected interface to exist")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"not-found","status":404,"title":"Not Found","detail":"interface not found"}`))
		}))
		defer server.Close()

		api := NewDescriptorDocumentsClient(newTestClient(t, server))
		exists, err := api.CheckInterfaceExists(context.Background(), testEServiceID, testDescriptorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected interface to not exist")
		}
	})
}
