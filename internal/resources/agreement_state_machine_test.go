package resources

import (
	"testing"
)

func TestComputeTransitions_DraftToActive(t *testing.T) {
	transitions, err := ComputeTransitions("DRAFT", "ACTIVE", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != TransitionSubmit {
		t.Fatalf("expected submit transition, got %s", transitions[0].Type)
	}
}

func TestComputeTransitions_ActiveToSuspended(t *testing.T) {
	transitions, err := ComputeTransitions("ACTIVE", "SUSPENDED", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != TransitionSuspend {
		t.Fatalf("expected suspend transition, got %s", transitions[0].Type)
	}
}

func TestComputeTransitions_SuspendedToActive(t *testing.T) {
	transitions, err := ComputeTransitions("SUSPENDED", "ACTIVE", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != TransitionUnsuspend {
		t.Fatalf("expected unsuspend transition, got %s", transitions[0].Type)
	}
}

func TestComputeTransitions_ActiveUpgrade(t *testing.T) {
	transitions, err := ComputeTransitions("ACTIVE", "ACTIVE", true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != TransitionUpgrade {
		t.Fatalf("expected upgrade transition, got %s", transitions[0].Type)
	}
}

func TestComputeTransitions_SuspendedUpgrade(t *testing.T) {
	transitions, err := ComputeTransitions("SUSPENDED", "SUSPENDED", true)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].Type != TransitionUpgrade {
		t.Fatalf("expected upgrade transition, got %s", transitions[0].Type)
	}
}

func TestComputeTransitions_DraftToSuspended_Error(t *testing.T) {
	_, err := ComputeTransitions("DRAFT", "SUSPENDED", false)
	if err == nil {
		t.Fatal("expected error for DRAFT -> SUSPENDED transition")
	}
}

func TestComputeTransitions_PendingToAny_Error(t *testing.T) {
	_, err := ComputeTransitions("PENDING", "ACTIVE", false)
	if err == nil {
		t.Fatal("expected error for PENDING -> ACTIVE transition")
	}
}

func TestComputeTransitions_NoOp_DraftDraft(t *testing.T) {
	transitions, err := ComputeTransitions("DRAFT", "DRAFT", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(transitions))
	}
}

func TestComputeTransitions_NoOp_ActiveActive(t *testing.T) {
	transitions, err := ComputeTransitions("ACTIVE", "ACTIVE", false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(transitions))
	}
}
