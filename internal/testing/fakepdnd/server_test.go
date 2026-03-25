package fakepdnd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// helper to do HTTP requests and decode JSON response.
func doRequest(t *testing.T, client *http.Client, method, url string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck // test helper

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result
}

// createAgreementViaAPI creates a DRAFT agreement via the API and returns its ID.
func createAgreementViaAPI(t *testing.T, client *http.Client, baseURL string) string {
	t.Helper()
	status, resp := doRequest(t, client, "POST", baseURL+"/agreements", map[string]string{
		"eserviceId":   uuid.New().String(),
		"descriptorId": uuid.New().String(),
	})
	if status != 201 {
		t.Fatalf("create agreement: expected 201, got %d", status)
	}
	id, ok := resp["id"].(string)
	if !ok {
		t.Fatal("response missing id field")
	}
	return id
}

func TestFakeServer_CreateAgreement(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements", map[string]string{
		"eserviceId":   uuid.New().String(),
		"descriptorId": uuid.New().String(),
	})

	if status != 201 {
		t.Fatalf("expected 201, got %d", status)
	}
	if resp["state"] != "DRAFT" {
		t.Fatalf("expected state DRAFT, got %v", resp["state"])
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestFakeServer_GetAgreement(t *testing.T) {
	fake := NewFakeServer()
	now := time.Now().UTC()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:           id,
		EServiceID:   uuid.New(),
		DescriptorID: uuid.New(),
		ProducerID:   fake.ProducerID(),
		ConsumerID:   fake.ConsumerID(),
		State:        "DRAFT",
		CreatedAt:    now,
	})

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/agreements/"+id.String(), nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["id"] != id.String() {
		t.Fatalf("expected id %s, got %v", id, resp["id"])
	}
	if resp["state"] != "DRAFT" {
		t.Fatalf("expected state DRAFT, got %v", resp["state"])
	}
}

func TestFakeServer_GetAgreement_NotFound(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	status, _ := doRequest(t, ts.Client(), "GET", ts.URL+"/agreements/"+uuid.New().String(), nil)
	if status != 404 {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestFakeServer_SubmitAutomatic(t *testing.T) {
	fake := NewFakeServer()
	fake.SetApprovalPolicy("AUTOMATIC")
	ts := fake.Start()
	defer ts.Close()

	agID := createAgreementViaAPI(t, ts.Client(), ts.URL)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+agID+"/submit", map[string]string{
		"consumerNotes": "test notes",
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "ACTIVE" {
		t.Fatalf("expected state ACTIVE, got %v", resp["state"])
	}
}

func TestFakeServer_SubmitManual(t *testing.T) {
	fake := NewFakeServer()
	fake.SetApprovalPolicy("MANUAL")
	ts := fake.Start()
	defer ts.Close()

	agID := createAgreementViaAPI(t, ts.Client(), ts.URL)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+agID+"/submit", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "PENDING" {
		t.Fatalf("expected state PENDING, got %v", resp["state"])
	}
}

func TestFakeServer_ApproveAgreement(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "PENDING", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/approve", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "ACTIVE" {
		t.Fatalf("expected state ACTIVE, got %v", resp["state"])
	}
}

func TestFakeServer_RejectAgreement(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "PENDING", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	reason := "This agreement does not meet the required criteria for approval at this time"
	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/reject", map[string]string{
		"reason": reason,
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "REJECTED" {
		t.Fatalf("expected state REJECTED, got %v", resp["state"])
	}
	if resp["rejectionReason"] != reason {
		t.Fatalf("expected rejectionReason %q, got %v", reason, resp["rejectionReason"])
	}
}

func TestFakeServer_SuspendActive(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "ACTIVE", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/suspend", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "SUSPENDED" {
		t.Fatalf("expected state SUSPENDED, got %v", resp["state"])
	}
	if resp["suspendedByConsumer"] != true {
		t.Fatalf("expected suspendedByConsumer true, got %v", resp["suspendedByConsumer"])
	}
}

func TestFakeServer_UnsuspendToActive(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "SUSPENDED", SuspendedByConsumer: true,
		CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/unsuspend", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "ACTIVE" {
		t.Fatalf("expected state ACTIVE, got %v", resp["state"])
	}
}

func TestFakeServer_UnsuspendRemainSuspended(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "SUSPENDED", SuspendedByConsumer: true, SuspendedByProducer: true,
		CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/unsuspend", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "SUSPENDED" {
		t.Fatalf("expected state SUSPENDED, got %v", resp["state"])
	}
	if resp["suspendedByConsumer"] != false {
		t.Fatalf("expected suspendedByConsumer false, got %v", resp["suspendedByConsumer"])
	}
	if resp["suspendedByProducer"] != true {
		t.Fatalf("expected suspendedByProducer true, got %v", resp["suspendedByProducer"])
	}
}

func TestFakeServer_DeleteDraft(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	agID := createAgreementViaAPI(t, ts.Client(), ts.URL)

	status, _ := doRequest(t, ts.Client(), "DELETE", ts.URL+"/agreements/"+agID, nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}

	// Verify gone.
	status, _ = doRequest(t, ts.Client(), "GET", ts.URL+"/agreements/"+agID, nil)
	if status != 404 {
		t.Fatalf("expected 404 after delete, got %d", status)
	}
}

func TestFakeServer_DeleteActive_Fails(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "ACTIVE", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, _ := doRequest(t, ts.Client(), "DELETE", ts.URL+"/agreements/"+id.String(), nil)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestFakeServer_DeletePending(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "PENDING", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, _ := doRequest(t, ts.Client(), "DELETE", ts.URL+"/agreements/"+id.String(), nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}

	// Verify gone.
	status, _ = doRequest(t, ts.Client(), "GET", ts.URL+"/agreements/"+id.String(), nil)
	if status != 404 {
		t.Fatalf("expected 404 after delete, got %d", status)
	}
}

func TestFakeServer_UpgradeActive(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	esID := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: esID, DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "ACTIVE", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/upgrade", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	// New agreement should have different ID.
	if resp["id"] == id.String() {
		t.Fatal("expected new agreement with different ID")
	}
	if resp["state"] != "ACTIVE" {
		t.Fatalf("expected state ACTIVE, got %v", resp["state"])
	}
	if resp["eserviceId"] != esID.String() {
		t.Fatalf("expected same eserviceId %s, got %v", esID, resp["eserviceId"])
	}

	// Old agreement should be ARCHIVED.
	old := fake.GetAgreement(id)
	if old == nil || old.State != "ARCHIVED" {
		t.Fatalf("expected old agreement to be ARCHIVED, got %v", old)
	}
}

func TestFakeServer_UpgradeSuspended(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	now := time.Now().UTC()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "SUSPENDED", SuspendedByConsumer: true,
		CreatedAt: now, SuspendedAt: &now,
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/upgrade", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "SUSPENDED" {
		t.Fatalf("expected state SUSPENDED, got %v", resp["state"])
	}
}

func TestFakeServer_CloneRejected(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	esID := uuid.New()
	descID := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: esID, DescriptorID: descID,
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "REJECTED", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/clone", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["id"] == id.String() {
		t.Fatal("expected new agreement with different ID")
	}
	if resp["state"] != "DRAFT" {
		t.Fatalf("expected state DRAFT, got %v", resp["state"])
	}
	if resp["eserviceId"] != esID.String() {
		t.Fatalf("expected same eserviceId")
	}
	if resp["descriptorId"] != descID.String() {
		t.Fatalf("expected same descriptorId")
	}
}

func TestFakeServer_InvalidTransition_SubmitActive(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         id,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "ACTIVE", CreatedAt: time.Now().UTC(),
	})
	ts := fake.Start()
	defer ts.Close()

	status, _ := doRequest(t, ts.Client(), "POST", ts.URL+"/agreements/"+id.String()+"/submit", nil)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestFakeServer_ListAgreements(t *testing.T) {
	fake := NewFakeServer()
	// Seed 3 agreements: 2 ACTIVE, 1 DRAFT.
	for i := 0; i < 2; i++ {
		fake.SeedAgreement(StoredAgreement{
			ID:         uuid.New(),
			EServiceID: uuid.New(), DescriptorID: uuid.New(),
			ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
			State: "ACTIVE", CreatedAt: time.Now().UTC(),
		})
	}
	fake.SeedAgreement(StoredAgreement{
		ID:         uuid.New(),
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "DRAFT", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/agreements?states=ACTIVE", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	results, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("response missing results field")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 ACTIVE agreements, got %d", len(results))
	}
}

func TestFakeServer_ListAgreementPurposes(t *testing.T) {
	fake := NewFakeServer()
	agID := uuid.New()
	fake.SeedAgreement(StoredAgreement{
		ID:         agID,
		EServiceID: uuid.New(), DescriptorID: uuid.New(),
		ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
		State: "ACTIVE", CreatedAt: time.Now().UTC(),
	})
	fake.SeedPurpose(agID, StoredPurpose{
		ID: uuid.New(), EServiceID: uuid.New(), ConsumerID: fake.ConsumerID(),
		Title: "Purpose 1", Description: "Desc 1", CreatedAt: time.Now().UTC(),
	})
	fake.SeedPurpose(agID, StoredPurpose{
		ID: uuid.New(), EServiceID: uuid.New(), ConsumerID: fake.ConsumerID(),
		Title: "Purpose 2", Description: "Desc 2", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/agreements/"+agID.String()+"/purposes", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	results2, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("response missing results field")
	}
	if len(results2) != 2 {
		t.Fatalf("expected 2 purposes, got %d", len(results2))
	}
}

func TestFakeServer_ListPagination(t *testing.T) {
	fake := NewFakeServer()
	// Seed 5 agreements.
	for i := 0; i < 5; i++ {
		fake.SeedAgreement(StoredAgreement{
			ID:         uuid.New(),
			EServiceID: uuid.New(), DescriptorID: uuid.New(),
			ProducerID: fake.ProducerID(), ConsumerID: fake.ConsumerID(),
			State: "ACTIVE", CreatedAt: time.Now().UTC(),
		})
	}

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", fmt.Sprintf("%s/agreements?offset=2&limit=2", ts.URL), nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	results3, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("response missing results field")
	}
	if len(results3) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results3))
	}
	pagination, ok := resp["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing pagination field")
	}
	totalCount, ok := pagination["totalCount"].(float64)
	if !ok {
		t.Fatal("pagination missing totalCount")
	}
	if int(totalCount) != 5 {
		t.Fatalf("expected totalCount 5, got %v", totalCount)
	}
}

// --- E-Service and Descriptor Tests ---

// createEServiceViaAPI creates an e-service via the API and returns its ID.
func createEServiceViaAPI(t *testing.T, client *http.Client, baseURL string) string {
	t.Helper()
	status, resp := doRequest(t, client, "POST", baseURL+"/eservices", map[string]interface{}{
		"name":        "Test EService",
		"description": "A test eservice",
		"technology":  "REST",
		"mode":        "DELIVER",
	})
	if status != 201 {
		t.Fatalf("create eservice: expected 201, got %d", status)
	}
	id, ok := resp["id"].(string)
	if !ok {
		t.Fatal("response missing id field")
	}
	return id
}

// getFirstDescriptorID returns the first descriptor ID for an eservice by listing descriptors.
func getFirstDescriptorID(t *testing.T, client *http.Client, baseURL, esID string) string {
	t.Helper()
	status, resp := doRequest(t, client, "GET", baseURL+"/eservices/"+esID+"/descriptors", nil)
	if status != 200 {
		t.Fatalf("list descriptors: expected 200, got %d", status)
	}
	results, ok := resp["results"].([]interface{})
	if !ok || len(results) == 0 {
		t.Fatal("expected at least one descriptor")
	}
	first, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatal("descriptor is not a map")
	}
	id, ok := first["id"].(string)
	if !ok {
		t.Fatal("descriptor missing id")
	}
	return id
}

func TestFakeServer_CreateEService(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices", map[string]interface{}{
		"name":        "My EService",
		"description": "Description",
		"technology":  "REST",
		"mode":        "DELIVER",
	})

	if status != 201 {
		t.Fatalf("expected 201, got %d", status)
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Fatal("expected non-empty id")
	}
	if resp["name"] != "My EService" {
		t.Fatalf("expected name 'My EService', got %v", resp["name"])
	}
}

func TestFakeServer_GetEService(t *testing.T) {
	fake := NewFakeServer()
	id := uuid.New()
	fake.SeedEService(StoredEService{
		ID:          id,
		ProducerID:  fake.ProducerID(),
		Name:        "Seeded",
		Description: "Desc",
		Technology:  "REST",
		Mode:        "DELIVER",
	})

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+id.String(), nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["id"] != id.String() {
		t.Fatalf("expected id %s, got %v", id, resp["id"])
	}
	if resp["name"] != "Seeded" {
		t.Fatalf("expected name 'Seeded', got %v", resp["name"])
	}
}

func TestFakeServer_GetEService_NotFound(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	status, _ := doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+uuid.New().String(), nil)
	if status != 404 {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestFakeServer_ListEServices(t *testing.T) {
	fake := NewFakeServer()
	for i := 0; i < 3; i++ {
		fake.SeedEService(StoredEService{
			ID:         uuid.New(),
			ProducerID: fake.ProducerID(),
			Name:       fmt.Sprintf("ES-%d", i),
			Technology: "REST",
			Mode:       "DELIVER",
		})
	}

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/eservices?technology=REST", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	results, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("response missing results field")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 eservices, got %d", len(results))
	}
}

func TestFakeServer_DeleteDraftEService(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)

	status, _ := doRequest(t, ts.Client(), "DELETE", ts.URL+"/eservices/"+esID, nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}

	// Verify gone.
	status, _ = doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+esID, nil)
	if status != 404 {
		t.Fatalf("expected 404 after delete, got %d", status)
	}
}

func TestFakeServer_DeletePublishedEService_Fails(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish the descriptor.
	status, _ := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)
	if status != 200 {
		t.Fatalf("publish: expected 200, got %d", status)
	}

	// Try to delete eservice.
	status, _ = doRequest(t, ts.Client(), "DELETE", ts.URL+"/eservices/"+esID, nil)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestFakeServer_UpdateDraftEService(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)

	status, resp := doRequest(t, ts.Client(), "PATCH", ts.URL+"/eservices/"+esID, map[string]interface{}{
		"name": "Updated Name",
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["name"] != "Updated Name" {
		t.Fatalf("expected name 'Updated Name', got %v", resp["name"])
	}
}

func TestFakeServer_UpdatePublishedEServiceName(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish the descriptor.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	status, resp := doRequest(t, ts.Client(), "PATCH", ts.URL+"/eservices/"+esID+"/name", map[string]interface{}{
		"name": "New Name",
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["name"] != "New Name" {
		t.Fatalf("expected name 'New Name', got %v", resp["name"])
	}
}

func TestFakeServer_CreateDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish the existing draft so we can create a new one.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors", map[string]interface{}{
		"audience":              []string{"api/v1"},
		"voucherLifespan":      600,
		"dailyCallsPerConsumer": 1000,
		"dailyCallsTotal":      10000,
	})
	if status != 201 {
		t.Fatalf("expected 201, got %d", status)
	}
	if resp["state"] != "DRAFT" {
		t.Fatalf("expected state DRAFT, got %v", resp["state"])
	}
	if resp["version"] != "2" {
		t.Fatalf("expected version '2', got %v", resp["version"])
	}
}

func TestFakeServer_GetDescriptor(t *testing.T) {
	fake := NewFakeServer()
	esID := uuid.New()
	descID := uuid.New()
	now := time.Now().UTC()
	fake.SeedEService(StoredEService{
		ID:         esID,
		ProducerID: fake.ProducerID(),
		Name:       "ES",
		Technology: "REST",
		Mode:       "DELIVER",
	})
	fake.SeedDescriptor(StoredDescriptor{
		ID:         descID,
		EServiceID: esID,
		Version:    "1",
		State:      "DRAFT",
		CreatedAt:  now,
	})

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+esID.String()+"/descriptors/"+descID.String(), nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["id"] != descID.String() {
		t.Fatalf("expected id %s, got %v", descID, resp["id"])
	}
	if resp["version"] != "1" {
		t.Fatalf("expected version '1', got %v", resp["version"])
	}
}

func TestFakeServer_ListDescriptors(t *testing.T) {
	fake := NewFakeServer()
	esID := uuid.New()
	now := time.Now().UTC()
	fake.SeedEService(StoredEService{
		ID:         esID,
		ProducerID: fake.ProducerID(),
		Name:       "ES",
		Technology: "REST",
		Mode:       "DELIVER",
	})
	for i := 1; i <= 3; i++ {
		fake.SeedDescriptor(StoredDescriptor{
			ID:         uuid.New(),
			EServiceID: esID,
			Version:    fmt.Sprintf("%d", i),
			State:      "DRAFT",
			CreatedAt:  now,
		})
	}

	ts := fake.Start()
	defer ts.Close()

	status, resp := doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+esID.String()+"/descriptors", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	results, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("response missing results field")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 descriptors, got %d", len(results))
	}
}

func TestFakeServer_PublishDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "PUBLISHED" {
		t.Fatalf("expected state PUBLISHED, got %v", resp["state"])
	}
}

func TestFakeServer_SuspendDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish first.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/suspend", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "SUSPENDED" {
		t.Fatalf("expected state SUSPENDED, got %v", resp["state"])
	}
}

func TestFakeServer_UnsuspendDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish then suspend.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/suspend", nil)

	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/unsuspend", nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["state"] != "PUBLISHED" {
		t.Fatalf("expected state PUBLISHED, got %v", resp["state"])
	}
}

func TestFakeServer_DeleteDraftDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish first descriptor so we can create a second.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	// Create a second descriptor.
	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors", map[string]interface{}{
		"dailyCallsPerConsumer": 100,
		"dailyCallsTotal":      1000,
		"voucherLifespan":      300,
	})
	if status != 201 {
		t.Fatalf("create descriptor: expected 201, got %d", status)
	}
	newDescID, ok := resp["id"].(string)
	if !ok {
		t.Fatal("response missing id field")
	}

	// Delete the new draft descriptor.
	status, _ = doRequest(t, ts.Client(), "DELETE", ts.URL+"/eservices/"+esID+"/descriptors/"+newDescID, nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}

	// Verify gone.
	status, _ = doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+esID+"/descriptors/"+newDescID, nil)
	if status != 404 {
		t.Fatalf("expected 404 after delete, got %d", status)
	}
}

func TestFakeServer_DeletePublishedDescriptor_Fails(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	status, _ := doRequest(t, ts.Client(), "DELETE", ts.URL+"/eservices/"+esID+"/descriptors/"+descID, nil)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestFakeServer_PublishDeprecatesPrevious(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	v1DescID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish v1.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+v1DescID+"/publish", nil)

	// Create and publish v2.
	status, resp := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors", map[string]interface{}{
		"dailyCallsPerConsumer": 100,
		"dailyCallsTotal":      1000,
		"voucherLifespan":      300,
	})
	if status != 201 {
		t.Fatalf("create v2: expected 201, got %d", status)
	}
	v2DescID, ok := resp["id"].(string)
	if !ok {
		t.Fatal("response missing id field")
	}

	status, resp = doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+v2DescID+"/publish", nil)
	if status != 200 {
		t.Fatalf("publish v2: expected 200, got %d", status)
	}
	if resp["state"] != "PUBLISHED" {
		t.Fatalf("expected v2 state PUBLISHED, got %v", resp["state"])
	}

	// Check v1 is now DEPRECATED.
	status, resp = doRequest(t, ts.Client(), "GET", ts.URL+"/eservices/"+esID+"/descriptors/"+v1DescID, nil)
	if status != 200 {
		t.Fatalf("get v1: expected 200, got %d", status)
	}
	if resp["state"] != "DEPRECATED" {
		t.Fatalf("expected v1 state DEPRECATED, got %v", resp["state"])
	}
}

func TestFakeServer_UpdateDraftDescriptor(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	status, resp := doRequest(t, ts.Client(), "PATCH", ts.URL+"/eservices/"+esID+"/descriptors/"+descID, map[string]interface{}{
		"dailyCallsPerConsumer": 5000,
		"dailyCallsTotal":      50000,
		"voucherLifespan":      900,
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["dailyCallsPerConsumer"] != float64(5000) {
		t.Fatalf("expected dailyCallsPerConsumer 5000, got %v", resp["dailyCallsPerConsumer"])
	}
	if resp["dailyCallsTotal"] != float64(50000) {
		t.Fatalf("expected dailyCallsTotal 50000, got %v", resp["dailyCallsTotal"])
	}
	if resp["voucherLifespan"] != float64(900) {
		t.Fatalf("expected voucherLifespan 900, got %v", resp["voucherLifespan"])
	}
}

func TestFakeServer_UpdatePublishedDescriptorQuotas(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish first.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	status, resp := doRequest(t, ts.Client(), "PATCH", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/quotas", map[string]interface{}{
		"dailyCallsPerConsumer": 2000,
		"dailyCallsTotal":      20000,
		"voucherLifespan":      1200,
	})
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	if resp["dailyCallsPerConsumer"] != float64(2000) {
		t.Fatalf("expected dailyCallsPerConsumer 2000, got %v", resp["dailyCallsPerConsumer"])
	}
	if resp["dailyCallsTotal"] != float64(20000) {
		t.Fatalf("expected dailyCallsTotal 20000, got %v", resp["dailyCallsTotal"])
	}
}

func TestFakeServer_InvalidTransition_PublishPublished(t *testing.T) {
	fake := NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := createEServiceViaAPI(t, ts.Client(), ts.URL)
	descID := getFirstDescriptorID(t, ts.Client(), ts.URL, esID)

	// Publish once.
	doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)

	// Try to publish again.
	status, _ := doRequest(t, ts.Client(), "POST", ts.URL+"/eservices/"+esID+"/descriptors/"+descID+"/publish", nil)
	if status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}
