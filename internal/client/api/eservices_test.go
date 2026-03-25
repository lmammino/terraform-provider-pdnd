package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const cannedEServiceJSON = `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "producerId": "880e8400-e29b-41d4-a716-446655440003",
  "name": "Test E-Service",
  "description": "A test e-service description",
  "technology": "REST",
  "mode": "DELIVER"
}`

const cannedDescriptorJSON = `{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "version": "1",
  "state": "DRAFT",
  "audience": ["api.example.com"],
  "voucherLifespan": 3600,
  "dailyCallsPerConsumer": 1000,
  "dailyCallsTotal": 10000,
  "agreementApprovalPolicy": "AUTOMATIC",
  "serverUrls": ["https://api.example.com"]
}`

func TestCreateEService_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedEServiceJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.CreateEService(context.Background(), EServiceSeed{
		Name:        "Test E-Service",
		Description: "A test e-service description",
		Technology:  "REST",
		Mode:        "DELIVER",
		Descriptor: DescriptorSeedForCreation{
			AgreementApprovalPolicy: "AUTOMATIC",
			Audience:                []string{"api.example.com"},
			DailyCallsPerConsumer:   1000,
			DailyCallsTotal:         10000,
			VoucherLifespan:         3600,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/eservices" {
		t.Errorf("expected path /eservices, got %s", gotPath)
	}
	if gotBody["name"] != "Test E-Service" {
		t.Errorf("expected name 'Test E-Service', got %v", gotBody["name"])
	}
	if gotBody["technology"] != "REST" {
		t.Errorf("expected technology 'REST', got %v", gotBody["technology"])
	}
	if result.Name != "Test E-Service" {
		t.Errorf("expected name 'Test E-Service', got %s", result.Name)
	}
	if result.ID != agreementID {
		t.Errorf("expected ID %s, got %s", agreementID, result.ID)
	}
}

func TestGetEService_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedEServiceJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.GetEService(context.Background(), agreementID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, agreementID.String()) {
		t.Errorf("expected path to contain %s, got %s", agreementID, gotPath)
	}
	if result.ID != agreementID {
		t.Errorf("expected ID %s, got %s", agreementID, result.ID)
	}
	if result.ProducerID != producerID {
		t.Errorf("expected ProducerID %s, got %s", producerID, result.ProducerID)
	}
	if result.Name != "Test E-Service" {
		t.Errorf("expected name 'Test E-Service', got %s", result.Name)
	}
	if result.Technology != "REST" {
		t.Errorf("expected technology 'REST', got %s", result.Technology)
	}
	if result.Mode != "DELIVER" {
		t.Errorf("expected mode 'DELIVER', got %s", result.Mode)
	}
}

func TestListEServices_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedEServiceJSON + `],
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

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.ListEServices(context.Background(), ListEServicesParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/eservices" {
		t.Errorf("expected path /eservices, got %s", gotPath)
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
	if result.Results[0].ID != agreementID {
		t.Errorf("expected eservice ID %s, got %s", agreementID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestDeleteEService_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	err := api.DeleteEService(context.Background(), agreementID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, agreementID.String()) {
		t.Errorf("expected path to contain %s, got %s", agreementID, gotPath)
	}
}

func TestUpdateDraftEService_Contract(t *testing.T) {
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
		_, _ = w.Write([]byte(cannedEServiceJSON))
	}))
	defer server.Close()

	name := "Updated Name"
	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.UpdateDraftEService(context.Background(), agreementID, EServiceDraftUpdate{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected method PATCH, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, agreementID.String()) {
		t.Errorf("expected path to contain %s, got %s", agreementID, gotPath)
	}
	if !strings.Contains(gotContentType, "application/merge-patch+json") {
		t.Errorf("expected content-type application/merge-patch+json, got %s", gotContentType)
	}
	if gotBody["name"] != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %v", gotBody["name"])
	}
	if result.Name != "Test E-Service" {
		t.Errorf("expected name 'Test E-Service', got %s", result.Name)
	}
}

func TestUpdatePublishedEServiceName_Contract(t *testing.T) {
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
		_, _ = w.Write([]byte(cannedEServiceJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.UpdatePublishedEServiceName(context.Background(), agreementID, "New Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected method PATCH, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/name") {
		t.Errorf("expected path to end with /name, got %s", gotPath)
	}
	if !strings.Contains(gotContentType, "application/merge-patch+json") {
		t.Errorf("expected content-type application/merge-patch+json, got %s", gotContentType)
	}
	if gotBody["name"] != "New Name" {
		t.Errorf("expected name 'New Name', got %v", gotBody["name"])
	}
	if result.Name != "Test E-Service" {
		t.Errorf("expected name 'Test E-Service', got %s", result.Name)
	}
}

func TestUpdatePublishedEServiceDescription_Contract(t *testing.T) {
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
		_, _ = w.Write([]byte(cannedEServiceJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.UpdatePublishedEServiceDescription(context.Background(), agreementID, "New Description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("expected method PATCH, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/description") {
		t.Errorf("expected path to end with /description, got %s", gotPath)
	}
	if !strings.Contains(gotContentType, "application/merge-patch+json") {
		t.Errorf("expected content-type application/merge-patch+json, got %s", gotContentType)
	}
	if gotBody["description"] != "New Description" {
		t.Errorf("expected description 'New Description', got %v", gotBody["description"])
	}
	if result.Description != "A test e-service description" {
		t.Errorf("expected description 'A test e-service description', got %s", result.Description)
	}
}

func TestCreateDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedDescriptorJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.CreateDescriptor(context.Background(), eserviceID, DescriptorSeed{
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.example.com"},
		DailyCallsPerConsumer:   1000,
		DailyCallsTotal:         10000,
		VoucherLifespan:         3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/descriptors") {
		t.Errorf("expected path to end with /descriptors, got %s", gotPath)
	}
	if gotBody["agreementApprovalPolicy"] != "AUTOMATIC" {
		t.Errorf("expected agreementApprovalPolicy 'AUTOMATIC', got %v", gotBody["agreementApprovalPolicy"])
	}
	if result.ID != descriptorID {
		t.Errorf("expected ID %s, got %s", descriptorID, result.ID)
	}
	if result.State != "DRAFT" {
		t.Errorf("expected state DRAFT, got %s", result.State)
	}
}

func TestGetDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDescriptorJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.GetDescriptor(context.Background(), eserviceID, descriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, eserviceID.String()) {
		t.Errorf("expected path to contain %s, got %s", eserviceID, gotPath)
	}
	if !strings.Contains(gotPath, descriptorID.String()) {
		t.Errorf("expected path to contain %s, got %s", descriptorID, gotPath)
	}
	if result.ID != descriptorID {
		t.Errorf("expected ID %s, got %s", descriptorID, result.ID)
	}
	if result.Version != "1" {
		t.Errorf("expected version '1', got %s", result.Version)
	}
	if result.VoucherLifespan != 3600 {
		t.Errorf("expected voucherLifespan 3600, got %d", result.VoucherLifespan)
	}
}

func TestListDescriptors_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedDescriptorJSON + `],
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

	api := NewEServicesClient(newTestClient(t, server))
	result, err := api.ListDescriptors(context.Background(), eserviceID, ListDescriptorsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/descriptors") {
		t.Errorf("expected path to contain /descriptors, got %s", gotPath)
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
	if result.Results[0].ID != descriptorID {
		t.Errorf("expected descriptor ID %s, got %s", descriptorID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestDeleteDraftDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	err := api.DeleteDraftDescriptor(context.Background(), eserviceID, descriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected method DELETE, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, eserviceID.String()) {
		t.Errorf("expected path to contain %s, got %s", eserviceID, gotPath)
	}
	if !strings.Contains(gotPath, descriptorID.String()) {
		t.Errorf("expected path to contain %s, got %s", descriptorID, gotPath)
	}
}

func TestPublishDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDescriptorJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	err := api.PublishDescriptor(context.Background(), eserviceID, descriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/publish") {
		t.Errorf("expected path to end with /publish, got %s", gotPath)
	}
}

func TestSuspendDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDescriptorJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	err := api.SuspendDescriptor(context.Background(), eserviceID, descriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/suspend") {
		t.Errorf("expected path to end with /suspend, got %s", gotPath)
	}
}

func TestUnsuspendDescriptor_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedDescriptorJSON))
	}))
	defer server.Close()

	api := NewEServicesClient(newTestClient(t, server))
	err := api.UnsuspendDescriptor(context.Background(), eserviceID, descriptorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/unsuspend") {
		t.Errorf("expected path to end with /unsuspend, got %s", gotPath)
	}
}
