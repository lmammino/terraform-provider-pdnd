package resources

import (
	"testing"
)

func TestCanDeleteDescriptor_Draft(t *testing.T) {
	if !canDeleteDescriptor("DRAFT") {
		t.Fatal("expected canDeleteDescriptor(DRAFT) to be true")
	}
}

func TestCanDeleteDescriptor_Published(t *testing.T) {
	if canDeleteDescriptor("PUBLISHED") {
		t.Fatal("expected canDeleteDescriptor(PUBLISHED) to be false")
	}
}

func TestCanDeleteDescriptor_Suspended(t *testing.T) {
	if canDeleteDescriptor("SUSPENDED") {
		t.Fatal("expected canDeleteDescriptor(SUSPENDED) to be false")
	}
}

func TestInferDescriptorDesiredState_Published(t *testing.T) {
	state, err := inferDescriptorDesiredState("PUBLISHED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "PUBLISHED" {
		t.Fatalf("expected PUBLISHED, got %s", state)
	}
}

func TestInferDescriptorDesiredState_Suspended(t *testing.T) {
	state, err := inferDescriptorDesiredState("SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "SUSPENDED" {
		t.Fatalf("expected SUSPENDED, got %s", state)
	}
}

func TestInferDescriptorDesiredState_Draft(t *testing.T) {
	state, err := inferDescriptorDesiredState("DRAFT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if state != "DRAFT" {
		t.Fatalf("expected DRAFT, got %s", state)
	}
}

func TestInferDescriptorDesiredState_Deprecated(t *testing.T) {
	_, err := inferDescriptorDesiredState("DEPRECATED")
	if err == nil {
		t.Fatal("expected error for DEPRECATED state")
	}
}

func TestInferDescriptorDesiredState_Archived(t *testing.T) {
	_, err := inferDescriptorDesiredState("ARCHIVED")
	if err == nil {
		t.Fatal("expected error for ARCHIVED state")
	}
}
