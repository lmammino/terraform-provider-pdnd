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
	testDelegationID   = uuid.MustParse("dd0e8400-e29b-41d4-a716-446655440030")
	testDelegatorID    = uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440031")
	testDelegateID     = uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440032")
	testDelegationESID = uuid.MustParse("cc0e8400-e29b-41d4-a716-446655440033")
)

const cannedDelegationJSON = `{
  "id": "dd0e8400-e29b-41d4-a716-446655440030",
  "delegatorId": "aa0e8400-e29b-41d4-a716-446655440031",
  "delegateId": "bb0e8400-e29b-41d4-a716-446655440032",
  "eserviceId": "cc0e8400-e29b-41d4-a716-446655440033",
  "state": "WAITING_FOR_APPROVAL",
  "createdAt": "2024-01-01T00:00:00Z",
  "submittedAt": "2024-01-01T00:00:00Z"
}`

func TestCreateDelegation_Contract(t *testing.T) {
	for _, delegationType := range []string{"consumer", "producer"} {
		t.Run(delegationType, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]interface{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(cannedDelegationJSON))
			}))
			defer server.Close()

			api := NewDelegationsClient(newTestClient(t, server))
			result, err := api.CreateDelegation(context.Background(), delegationType, DelegationSeed{
				EServiceID: testDelegationESID,
				DelegateID: testDelegateID,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodPost {
				t.Errorf("expected method POST, got %s", gotMethod)
			}
			expectedPath := "/" + delegationType + "Delegations"
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			if gotBody["delegateId"] != testDelegateID.String() {
				t.Errorf("expected delegateId %s, got %v", testDelegateID, gotBody["delegateId"])
			}
			if gotBody["eserviceId"] != testDelegationESID.String() {
				t.Errorf("expected eserviceId %s, got %v", testDelegationESID, gotBody["eserviceId"])
			}
			if result.ID != testDelegationID {
				t.Errorf("expected ID %s, got %s", testDelegationID, result.ID)
			}
			if result.State != "WAITING_FOR_APPROVAL" {
				t.Errorf("expected state WAITING_FOR_APPROVAL, got %s", result.State)
			}
		})
	}
}

func TestGetDelegation_Contract(t *testing.T) {
	for _, delegationType := range []string{"consumer", "producer"} {
		t.Run(delegationType, func(t *testing.T) {
			var gotMethod, gotPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(cannedDelegationJSON))
			}))
			defer server.Close()

			api := NewDelegationsClient(newTestClient(t, server))
			result, err := api.GetDelegation(context.Background(), delegationType, testDelegationID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodGet {
				t.Errorf("expected method GET, got %s", gotMethod)
			}
			expectedPath := "/" + delegationType + "Delegations/" + testDelegationID.String()
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			if result.ID != testDelegationID {
				t.Errorf("expected ID %s, got %s", testDelegationID, result.ID)
			}
			if result.DelegatorID != testDelegatorID {
				t.Errorf("expected DelegatorID %s, got %s", testDelegatorID, result.DelegatorID)
			}
			if result.DelegateID != testDelegateID {
				t.Errorf("expected DelegateID %s, got %s", testDelegateID, result.DelegateID)
			}
			if result.EServiceID != testDelegationESID {
				t.Errorf("expected EServiceID %s, got %s", testDelegationESID, result.EServiceID)
			}
			if result.State != "WAITING_FOR_APPROVAL" {
				t.Errorf("expected state WAITING_FOR_APPROVAL, got %s", result.State)
			}
		})
	}
}

func TestListDelegations_Contract(t *testing.T) {
	for _, delegationType := range []string{"consumer", "producer"} {
		t.Run(delegationType, func(t *testing.T) {
			var gotMethod, gotPath, gotQuery string

			responseJSON := `{
				"results": [` + cannedDelegationJSON + `],
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

			api := NewDelegationsClient(newTestClient(t, server))
			result, err := api.ListDelegations(context.Background(), delegationType, ListDelegationsParams{
				Offset: 0,
				Limit:  50,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodGet {
				t.Errorf("expected method GET, got %s", gotMethod)
			}
			expectedPath := "/" + delegationType + "Delegations"
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
			if result.Results[0].ID != testDelegationID {
				t.Errorf("expected delegation ID %s, got %s", testDelegationID, result.Results[0].ID)
			}
			if result.Pagination.TotalCount != 1 {
				t.Errorf("expected totalCount 1, got %d", result.Pagination.TotalCount)
			}
		})
	}
}

func TestAcceptDelegation_Contract(t *testing.T) {
	for _, delegationType := range []string{"consumer", "producer"} {
		t.Run(delegationType, func(t *testing.T) {
			var gotMethod, gotPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(cannedDelegationJSON))
			}))
			defer server.Close()

			api := NewDelegationsClient(newTestClient(t, server))
			result, err := api.AcceptDelegation(context.Background(), delegationType, testDelegationID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodPost {
				t.Errorf("expected method POST, got %s", gotMethod)
			}
			expectedPath := "/" + delegationType + "Delegations/" + testDelegationID.String() + "/accept"
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			if result.ID != testDelegationID {
				t.Errorf("expected ID %s, got %s", testDelegationID, result.ID)
			}
		})
	}
}

func TestRejectDelegation_Contract(t *testing.T) {
	for _, delegationType := range []string{"consumer", "producer"} {
		t.Run(delegationType, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]interface{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(cannedDelegationJSON))
			}))
			defer server.Close()

			api := NewDelegationsClient(newTestClient(t, server))
			result, err := api.RejectDelegation(context.Background(), delegationType, testDelegationID, DelegationRejection{
				RejectionReason: "Not meeting requirements",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodPost {
				t.Errorf("expected method POST, got %s", gotMethod)
			}
			expectedPath := "/" + delegationType + "Delegations/" + testDelegationID.String() + "/reject"
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			if gotBody["rejectionReason"] != "Not meeting requirements" {
				t.Errorf("expected rejectionReason in body, got %v", gotBody["rejectionReason"])
			}
			if result.ID != testDelegationID {
				t.Errorf("expected ID %s, got %s", testDelegationID, result.ID)
			}
		})
	}
}
