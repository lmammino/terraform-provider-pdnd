package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

func TestIsEServiceDraft_AllDraft(t *testing.T) {
	descriptors := []api.Descriptor{
		{ID: uuid.New(), State: "DRAFT", Version: "1"},
		{ID: uuid.New(), State: "DRAFT", Version: "2"},
	}
	if !isEServiceDraft(descriptors) {
		t.Fatal("expected isEServiceDraft to return true when all descriptors are DRAFT")
	}
}

func TestIsEServiceDraft_OnePublished(t *testing.T) {
	descriptors := []api.Descriptor{
		{ID: uuid.New(), State: "DRAFT", Version: "1"},
		{ID: uuid.New(), State: "PUBLISHED", Version: "2"},
	}
	if isEServiceDraft(descriptors) {
		t.Fatal("expected isEServiceDraft to return false when one descriptor is PUBLISHED")
	}
}

func TestIsEServiceDraft_Empty(t *testing.T) {
	if !isEServiceDraft([]api.Descriptor{}) {
		t.Fatal("expected isEServiceDraft to return true for empty list")
	}
}

func TestFindInitialDescriptor_ByVersion1(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	descriptors := []api.Descriptor{
		{ID: id2, State: "DRAFT", Version: "2"},
		{ID: id1, State: "PUBLISHED", Version: "1"},
	}
	result := findInitialDescriptor(descriptors)
	if result == nil {
		t.Fatal("expected to find initial descriptor")
	}
	if result.ID != id1 {
		t.Fatalf("expected descriptor with ID %s, got %s", id1, result.ID)
	}
}

func TestFindInitialDescriptor_Empty(t *testing.T) {
	result := findInitialDescriptor([]api.Descriptor{})
	if result != nil {
		t.Fatal("expected nil for empty list")
	}
}
