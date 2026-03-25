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

var (
	testEServiceID   = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testDescriptorID = uuid.MustParse("770e8400-e29b-41d4-a716-446655440002")
	testAttributeID1 = uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440010")
	testAttributeID2 = uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440011")
)

func TestListDescriptorAttributes_Contract(t *testing.T) {
	for _, attrType := range []string{"certified", "declared", "verified"} {
		t.Run(attrType, func(t *testing.T) {
			var gotMethod, gotPath, gotOffset, gotLimit string

			cannedResponse := buildListResponse(attrType)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotOffset = r.URL.Query().Get("offset")
				gotLimit = r.URL.Query().Get("limit")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(cannedResponse)
			}))
			defer server.Close()

			api := NewDescriptorAttributesClient(newTestClient(t, server))
			entries, pagination, err := api.ListDescriptorAttributes(context.Background(), testEServiceID, testDescriptorID, attrType, 0, 50)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodGet {
				t.Errorf("expected method GET, got %s", gotMethod)
			}
			expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/" + attrType + "Attributes"
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			if gotOffset != "0" {
				t.Errorf("expected offset 0, got %s", gotOffset)
			}
			if gotLimit != "50" {
				t.Errorf("expected limit 50, got %s", gotLimit)
			}
			if len(entries) != 2 {
				t.Fatalf("expected 2 entries, got %d", len(entries))
			}
			if entries[0].AttributeID != testAttributeID1 {
				t.Errorf("expected attribute ID %s, got %s", testAttributeID1, entries[0].AttributeID)
			}
			if entries[0].GroupIndex != 0 {
				t.Errorf("expected group index 0, got %d", entries[0].GroupIndex)
			}
			if entries[1].AttributeID != testAttributeID2 {
				t.Errorf("expected attribute ID %s, got %s", testAttributeID2, entries[1].AttributeID)
			}
			if entries[1].GroupIndex != 1 {
				t.Errorf("expected group index 1, got %d", entries[1].GroupIndex)
			}
			if pagination.TotalCount != 2 {
				t.Errorf("expected total count 2, got %d", pagination.TotalCount)
			}
		})
	}
}

func TestCreateDescriptorAttributeGroup_Contract(t *testing.T) {
	for _, attrType := range []string{"certified", "declared", "verified"} {
		t.Run(attrType, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]interface{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"attributes":[]}`))
			}))
			defer server.Close()

			api := NewDescriptorAttributesClient(newTestClient(t, server))
			err := api.CreateDescriptorAttributeGroup(context.Background(), testEServiceID, testDescriptorID, attrType, []uuid.UUID{testAttributeID1, testAttributeID2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodPost {
				t.Errorf("expected method POST, got %s", gotMethod)
			}
			expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/" + attrType + "Attributes/groups"
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
			ids, ok := gotBody["attributeIds"].([]interface{})
			if !ok || len(ids) != 2 {
				t.Errorf("expected 2 attributeIds in body, got %v", gotBody["attributeIds"])
			}
		})
	}
}

func TestDeleteDescriptorAttributeFromGroup_Contract(t *testing.T) {
	for _, attrType := range []string{"certified", "declared", "verified"} {
		t.Run(attrType, func(t *testing.T) {
			var gotMethod, gotPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			api := NewDescriptorAttributesClient(newTestClient(t, server))
			err := api.DeleteDescriptorAttributeFromGroup(context.Background(), testEServiceID, testDescriptorID, attrType, 0, testAttributeID1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMethod != http.MethodDelete {
				t.Errorf("expected method DELETE, got %s", gotMethod)
			}
			expectedPath := "/eservices/" + testEServiceID.String() + "/descriptors/" + testDescriptorID.String() + "/" + attrType + "Attributes/groups/0/attributes/" + testAttributeID1.String()
			if gotPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, gotPath)
			}
		})
	}
}

// buildListResponse creates a canned JSON response for the list endpoint.
// The attribute structure differs by type (certified has code/origin, others don't).
func buildListResponse(attrType string) []byte {
	attr1 := buildAttributeJSON(attrType, testAttributeID1, "Attr 1")
	attr2 := buildAttributeJSON(attrType, testAttributeID2, "Attr 2")

	resp := map[string]interface{}{
		"results": []map[string]interface{}{
			{"attribute": attr1, "groupIndex": 0},
			{"attribute": attr2, "groupIndex": 1},
		},
		"pagination": map[string]interface{}{
			"offset":     0,
			"limit":      50,
			"totalCount": 2,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

func buildAttributeJSON(attrType string, id uuid.UUID, name string) map[string]interface{} {
	attr := map[string]interface{}{
		"id":          id.String(),
		"name":        name,
		"description": "Test " + name,
		"createdAt":   "2024-01-01T00:00:00Z",
	}
	if attrType == "certified" {
		attr["code"] = "TEST_CODE"
		attr["origin"] = "IPA"
	}
	return attr
}
