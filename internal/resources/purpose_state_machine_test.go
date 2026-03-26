package resources

import "testing"

func TestComputePurposeTransitions_DraftToActive(t *testing.T) {
	transitions, err := ComputePurposeTransitions("DRAFT", "ACTIVE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 1 || transitions[0].Type != PurposeTransitionActivate {
		t.Errorf("expected [activate], got %v", transitions)
	}
}

func TestComputePurposeTransitions_DraftToDraft(t *testing.T) {
	transitions, err := ComputePurposeTransitions("DRAFT", "DRAFT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 0 {
		t.Errorf("expected no transitions, got %v", transitions)
	}
}

func TestComputePurposeTransitions_DraftToSuspended(t *testing.T) {
	_, err := ComputePurposeTransitions("DRAFT", "SUSPENDED")
	if err == nil {
		t.Error("expected error for DRAFT -> SUSPENDED")
	}
}

func TestComputePurposeTransitions_ActiveToSuspended(t *testing.T) {
	transitions, err := ComputePurposeTransitions("ACTIVE", "SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 1 || transitions[0].Type != PurposeTransitionSuspend {
		t.Errorf("expected [suspend], got %v", transitions)
	}
}

func TestComputePurposeTransitions_ActiveToActive(t *testing.T) {
	transitions, err := ComputePurposeTransitions("ACTIVE", "ACTIVE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 0 {
		t.Errorf("expected no transitions, got %v", transitions)
	}
}

func TestComputePurposeTransitions_ActiveToDraft(t *testing.T) {
	_, err := ComputePurposeTransitions("ACTIVE", "DRAFT")
	if err == nil {
		t.Error("expected error for ACTIVE -> DRAFT")
	}
}

func TestComputePurposeTransitions_SuspendedToActive(t *testing.T) {
	transitions, err := ComputePurposeTransitions("SUSPENDED", "ACTIVE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 1 || transitions[0].Type != PurposeTransitionUnsuspend {
		t.Errorf("expected [unsuspend], got %v", transitions)
	}
}

func TestComputePurposeTransitions_SuspendedToSuspended(t *testing.T) {
	transitions, err := ComputePurposeTransitions("SUSPENDED", "SUSPENDED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 0 {
		t.Errorf("expected no transitions, got %v", transitions)
	}
}

func TestComputePurposeTransitions_SuspendedToDraft(t *testing.T) {
	_, err := ComputePurposeTransitions("SUSPENDED", "DRAFT")
	if err == nil {
		t.Error("expected error for SUSPENDED -> DRAFT")
	}
}

func TestComputePurposeTransitions_ArchivedIsTerminal(t *testing.T) {
	_, err := ComputePurposeTransitions("ARCHIVED", "ACTIVE")
	if err == nil {
		t.Error("expected error for ARCHIVED state")
	}
}

func TestComputePurposeTransitions_RejectedIsTerminal(t *testing.T) {
	_, err := ComputePurposeTransitions("REJECTED", "ACTIVE")
	if err == nil {
		t.Error("expected error for REJECTED state")
	}
}

func TestComputePurposeTransitions_WaitingForApprovalIsTerminal(t *testing.T) {
	_, err := ComputePurposeTransitions("WAITING_FOR_APPROVAL", "ACTIVE")
	if err == nil {
		t.Error("expected error for WAITING_FOR_APPROVAL state")
	}
}
