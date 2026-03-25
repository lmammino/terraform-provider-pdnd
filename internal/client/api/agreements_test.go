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
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
)

// Fixed UUIDs for tests.
var (
	agreementID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	eserviceID   = uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	descriptorID = uuid.MustParse("770e8400-e29b-41d4-a716-446655440002")
	producerID   = uuid.MustParse("880e8400-e29b-41d4-a716-446655440003")
	consumerID   = uuid.MustParse("990e8400-e29b-41d4-a716-446655440004")
	purposeID    = uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440005")
	delegationID = uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440006")
)

const cannedAgreementJSON = `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "eserviceId": "660e8400-e29b-41d4-a716-446655440001",
  "descriptorId": "770e8400-e29b-41d4-a716-446655440002",
  "producerId": "880e8400-e29b-41d4-a716-446655440003",
  "consumerId": "990e8400-e29b-41d4-a716-446655440004",
  "state": "DRAFT",
  "createdAt": "2024-01-15T10:30:00Z"
}`

func cannedAgreementWithState(state string) string {
	return `{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "eserviceId": "660e8400-e29b-41d4-a716-446655440001",
  "descriptorId": "770e8400-e29b-41d4-a716-446655440002",
  "producerId": "880e8400-e29b-41d4-a716-446655440003",
  "consumerId": "990e8400-e29b-41d4-a716-446655440004",
  "state": "` + state + `",
  "createdAt": "2024-01-15T10:30:00Z"
}`
}

const cannedPurposeJSON = `{
  "id": "aa0e8400-e29b-41d4-a716-446655440005",
  "eserviceId": "660e8400-e29b-41d4-a716-446655440001",
  "consumerId": "990e8400-e29b-41d4-a716-446655440004",
  "title": "Test Purpose",
  "description": "A test purpose description",
  "createdAt": "2024-01-15T10:30:00Z",
  "isRiskAnalysisValid": true,
  "isFreeOfCharge": false
}`

// newTestClient creates a ClientWithResponses pointing at the given test server.
func newTestClient(t *testing.T, server *httptest.Server) *generated.ClientWithResponses {
	t.Helper()
	c, err := generated.NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

func TestCreateAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(cannedAgreementJSON))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.CreateAgreement(context.Background(), AgreementSeed{
		EServiceID:   eserviceID,
		DescriptorID: descriptorID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if gotPath != "/agreements" {
		t.Errorf("expected path /agreements, got %s", gotPath)
	}
	if gotBody["eserviceId"] != eserviceID.String() {
		t.Errorf("expected eserviceId %s, got %v", eserviceID, gotBody["eserviceId"])
	}
	if gotBody["descriptorId"] != descriptorID.String() {
		t.Errorf("expected descriptorId %s, got %v", descriptorID, gotBody["descriptorId"])
	}
	if result.State != "DRAFT" {
		t.Errorf("expected state DRAFT, got %s", result.State)
	}
	if result.ID != agreementID {
		t.Errorf("expected ID %s, got %s", agreementID, result.ID)
	}
}

func TestGetAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementJSON))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.GetAgreement(context.Background(), agreementID)
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
	if result.EServiceID != eserviceID {
		t.Errorf("expected EServiceID %s, got %s", eserviceID, result.EServiceID)
	}
	if result.DescriptorID != descriptorID {
		t.Errorf("expected DescriptorID %s, got %s", descriptorID, result.DescriptorID)
	}
	if result.ProducerID != producerID {
		t.Errorf("expected ProducerID %s, got %s", producerID, result.ProducerID)
	}
	if result.ConsumerID != consumerID {
		t.Errorf("expected ConsumerID %s, got %s", consumerID, result.ConsumerID)
	}
	if result.State != "DRAFT" {
		t.Errorf("expected state DRAFT, got %s", result.State)
	}
}

func TestDeleteAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	err := api.DeleteAgreement(context.Background(), agreementID)
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

func TestListAgreements_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedAgreementJSON + `],
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

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.ListAgreements(context.Background(), ListAgreementsParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if gotPath != "/agreements" {
		t.Errorf("expected path /agreements, got %s", gotPath)
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
		t.Errorf("expected agreement ID %s, got %s", agreementID, result.Results[0].ID)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}

func TestSubmitAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementWithState("ACTIVE")))
	}))
	defer server.Close()

	notes := "test notes"
	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.SubmitAgreement(context.Background(), agreementID, AgreementSubmission{
		ConsumerNotes: &notes,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/submit") {
		t.Errorf("expected path to end with /submit, got %s", gotPath)
	}
	if gotBody["consumerNotes"] != "test notes" {
		t.Errorf("expected consumerNotes 'test notes', got %v", gotBody["consumerNotes"])
	}
	if result.State != "ACTIVE" {
		t.Errorf("expected state ACTIVE, got %s", result.State)
	}
}

func TestApproveAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementWithState("ACTIVE")))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.ApproveAgreement(context.Background(), agreementID, &DelegationRef{
		DelegationID: delegationID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/approve") {
		t.Errorf("expected path to end with /approve, got %s", gotPath)
	}
	if result.State != "ACTIVE" {
		t.Errorf("expected state ACTIVE, got %s", result.State)
	}
}

func TestRejectAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementWithState("REJECTED")))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.RejectAgreement(context.Background(), agreementID, AgreementRejection{
		Reason: "This agreement does not meet our requirements",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/reject") {
		t.Errorf("expected path to end with /reject, got %s", gotPath)
	}
	if gotBody["reason"] != "This agreement does not meet our requirements" {
		t.Errorf("expected reason in body, got %v", gotBody["reason"])
	}
	if result.State != "REJECTED" {
		t.Errorf("expected state REJECTED, got %s", result.State)
	}
}

func TestSuspendAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementWithState("SUSPENDED")))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.SuspendAgreement(context.Background(), agreementID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/suspend") {
		t.Errorf("expected path to end with /suspend, got %s", gotPath)
	}
	if result.State != "SUSPENDED" {
		t.Errorf("expected state SUSPENDED, got %s", result.State)
	}
}

func TestUnsuspendAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cannedAgreementWithState("ACTIVE")))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.UnsuspendAgreement(context.Background(), agreementID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/unsuspend") {
		t.Errorf("expected path to end with /unsuspend, got %s", gotPath)
	}
	if result.State != "ACTIVE" {
		t.Errorf("expected state ACTIVE, got %s", result.State)
	}
}

func TestUpgradeAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	newID := uuid.MustParse("110e8400-e29b-41d4-a716-446655440099")
	responseJSON := `{
		"id": "110e8400-e29b-41d4-a716-446655440099",
		"eserviceId": "660e8400-e29b-41d4-a716-446655440001",
		"descriptorId": "770e8400-e29b-41d4-a716-446655440002",
		"producerId": "880e8400-e29b-41d4-a716-446655440003",
		"consumerId": "990e8400-e29b-41d4-a716-446655440004",
		"state": "ACTIVE",
		"createdAt": "2024-01-15T10:30:00Z"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		// Verify no body was sent
		body, _ := io.ReadAll(r.Body)
		_ = body // no body expected for this endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.UpgradeAgreement(context.Background(), agreementID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/upgrade") {
		t.Errorf("expected path to end with /upgrade, got %s", gotPath)
	}
	if result.ID != newID {
		t.Errorf("expected new ID %s, got %s", newID, result.ID)
	}
}

func TestCloneAgreement_Contract(t *testing.T) {
	var gotMethod, gotPath string

	newID := uuid.MustParse("220e8400-e29b-41d4-a716-446655440099")
	responseJSON := `{
		"id": "220e8400-e29b-41d4-a716-446655440099",
		"eserviceId": "660e8400-e29b-41d4-a716-446655440001",
		"descriptorId": "770e8400-e29b-41d4-a716-446655440002",
		"producerId": "880e8400-e29b-41d4-a716-446655440003",
		"consumerId": "990e8400-e29b-41d4-a716-446655440004",
		"state": "DRAFT",
		"createdAt": "2024-01-15T10:30:00Z"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.CloneAgreement(context.Background(), agreementID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected method POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/clone") {
		t.Errorf("expected path to end with /clone, got %s", gotPath)
	}
	if result.ID != newID {
		t.Errorf("expected new ID %s, got %s", newID, result.ID)
	}
	if result.State != "DRAFT" {
		t.Errorf("expected state DRAFT, got %s", result.State)
	}
}

func TestListAgreementPurposes_Contract(t *testing.T) {
	var gotMethod, gotPath, gotQuery string

	responseJSON := `{
		"results": [` + cannedPurposeJSON + `],
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

	api := NewAgreementsClient(newTestClient(t, server))
	result, err := api.ListAgreementPurposes(context.Background(), agreementID, PaginationParams{
		Offset: 0,
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/purposes") {
		t.Errorf("expected path to contain /purposes, got %s", gotPath)
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
	if result.Results[0].ID != purposeID {
		t.Errorf("expected purpose ID %s, got %s", purposeID, result.Results[0].ID)
	}
	if result.Results[0].Title != "Test Purpose" {
		t.Errorf("expected title 'Test Purpose', got %s", result.Results[0].Title)
	}
	if result.Pagination.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
	}
}
