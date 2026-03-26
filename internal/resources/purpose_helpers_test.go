package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

func TestDerivePurposeState_CurrentVersion(t *testing.T) {
	p := &api.Purpose{
		CurrentVersion: &api.PurposeVersion{State: "ACTIVE"},
	}
	if got := derivePurposeState(p); got != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", got)
	}
}

func TestDerivePurposeState_WaitingForApproval(t *testing.T) {
	p := &api.Purpose{
		WaitingForApprovalVersion: &api.PurposeVersion{State: "WAITING_FOR_APPROVAL"},
	}
	if got := derivePurposeState(p); got != "WAITING_FOR_APPROVAL" {
		t.Errorf("expected WAITING_FOR_APPROVAL, got %s", got)
	}
}

func TestDerivePurposeState_Draft(t *testing.T) {
	p := &api.Purpose{}
	if got := derivePurposeState(p); got != "DRAFT" {
		t.Errorf("expected DRAFT, got %s", got)
	}
}

func TestDerivePurposeState_Suspended(t *testing.T) {
	p := &api.Purpose{
		CurrentVersion: &api.PurposeVersion{State: "SUSPENDED"},
	}
	if got := derivePurposeState(p); got != "SUSPENDED" {
		t.Errorf("expected SUSPENDED, got %s", got)
	}
}

func TestCanDeletePurpose_Draft(t *testing.T) {
	if !canDeletePurpose("DRAFT") {
		t.Error("expected DRAFT to be deletable")
	}
}

func TestCanDeletePurpose_WaitingForApproval(t *testing.T) {
	if !canDeletePurpose("WAITING_FOR_APPROVAL") {
		t.Error("expected WAITING_FOR_APPROVAL to be deletable")
	}
}

func TestCanDeletePurpose_Active(t *testing.T) {
	if canDeletePurpose("ACTIVE") {
		t.Error("expected ACTIVE to not be deletable")
	}
}

func TestCanDeletePurpose_Suspended(t *testing.T) {
	if canDeletePurpose("SUSPENDED") {
		t.Error("expected SUSPENDED to not be deletable")
	}
}

func TestCanDeletePurpose_Archived(t *testing.T) {
	if canDeletePurpose("ARCHIVED") {
		t.Error("expected ARCHIVED to not be deletable")
	}
}

func TestInferPurposeDesiredState_Active(t *testing.T) {
	state, err := inferPurposeDesiredState("ACTIVE")
	if err != nil || state != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s (err: %v)", state, err)
	}
}

func TestInferPurposeDesiredState_Suspended(t *testing.T) {
	state, err := inferPurposeDesiredState("SUSPENDED")
	if err != nil || state != "SUSPENDED" {
		t.Errorf("expected SUSPENDED, got %s (err: %v)", state, err)
	}
}

func TestInferPurposeDesiredState_Draft(t *testing.T) {
	state, err := inferPurposeDesiredState("DRAFT")
	if err != nil || state != "DRAFT" {
		t.Errorf("expected DRAFT, got %s (err: %v)", state, err)
	}
}

func TestInferPurposeDesiredState_Archived(t *testing.T) {
	_, err := inferPurposeDesiredState("ARCHIVED")
	if err == nil {
		t.Error("expected error for ARCHIVED")
	}
}

func TestInferPurposeDesiredState_Rejected(t *testing.T) {
	_, err := inferPurposeDesiredState("REJECTED")
	if err == nil {
		t.Error("expected error for REJECTED")
	}
}

// Suppress unused import warning
var _ = uuid.New
