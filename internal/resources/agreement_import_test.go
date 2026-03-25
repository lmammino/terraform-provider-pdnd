package resources

import (
	"testing"
)

func TestInferDesiredState_Active(t *testing.T) {
	state, err := inferDesiredState("ACTIVE")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", state)
	}
}

func TestInferDesiredState_Suspended(t *testing.T) {
	state, err := inferDesiredState("SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "SUSPENDED" {
		t.Fatalf("expected SUSPENDED, got %s", state)
	}
}

func TestInferDesiredState_Draft(t *testing.T) {
	state, err := inferDesiredState("DRAFT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "DRAFT" {
		t.Fatalf("expected DRAFT, got %s", state)
	}
}

func TestInferDesiredState_Pending(t *testing.T) {
	_, err := inferDesiredState("PENDING")
	if err == nil {
		t.Fatal("expected error for PENDING state")
	}
}

func TestInferDesiredState_Rejected(t *testing.T) {
	_, err := inferDesiredState("REJECTED")
	if err == nil {
		t.Fatal("expected error for REJECTED state")
	}
}

func TestInferDesiredState_Archived(t *testing.T) {
	_, err := inferDesiredState("ARCHIVED")
	if err == nil {
		t.Fatal("expected error for ARCHIVED state")
	}
}

func TestInferDesiredState_MissingCertifiedAttributes(t *testing.T) {
	_, err := inferDesiredState("MISSING_CERTIFIED_ATTRIBUTES")
	if err == nil {
		t.Fatal("expected error for MISSING_CERTIFIED_ATTRIBUTES state")
	}
}
