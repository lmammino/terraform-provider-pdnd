package resources

import (
	"testing"
)

func TestComputeDescriptorTransitions_DraftToPublished(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("DRAFT", "PUBLISHED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != DescriptorTransitionPublish {
		t.Fatalf("expected publish transition, got %s", transitions[0].Type)
	}
}

func TestComputeDescriptorTransitions_PublishedToSuspended(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("PUBLISHED", "SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != DescriptorTransitionSuspend {
		t.Fatalf("expected suspend transition, got %s", transitions[0].Type)
	}
}

func TestComputeDescriptorTransitions_SuspendedToPublished(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("SUSPENDED", "PUBLISHED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != DescriptorTransitionUnsuspend {
		t.Fatalf("expected unsuspend transition, got %s", transitions[0].Type)
	}
}

func TestComputeDescriptorTransitions_DraftToDraft(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("DRAFT", "DRAFT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(transitions))
	}
}

func TestComputeDescriptorTransitions_PublishedToPublished(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("PUBLISHED", "PUBLISHED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(transitions))
	}
}

func TestComputeDescriptorTransitions_SuspendedToSuspended(t *testing.T) {
	transitions, err := ComputeDescriptorTransitions("SUSPENDED", "SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(transitions))
	}
}

func TestComputeDescriptorTransitions_DraftToSuspended_Error(t *testing.T) {
	_, err := ComputeDescriptorTransitions("DRAFT", "SUSPENDED")
	if err == nil {
		t.Fatal("expected error for DRAFT -> SUSPENDED transition")
	}
}

func TestComputeDescriptorTransitions_DeprecatedToAny_Error(t *testing.T) {
	_, err := ComputeDescriptorTransitions("DEPRECATED", "PUBLISHED")
	if err == nil {
		t.Fatal("expected error for DEPRECATED -> PUBLISHED transition")
	}
}

func TestComputeDescriptorTransitions_ArchivedToAny_Error(t *testing.T) {
	_, err := ComputeDescriptorTransitions("ARCHIVED", "DRAFT")
	if err == nil {
		t.Fatal("expected error for ARCHIVED -> DRAFT transition")
	}
}

func TestComputeDescriptorTransitions_WaitingToAny_Error(t *testing.T) {
	_, err := ComputeDescriptorTransitions("WAITING_FOR_APPROVAL", "PUBLISHED")
	if err == nil {
		t.Fatal("expected error for WAITING_FOR_APPROVAL -> PUBLISHED transition")
	}
}
