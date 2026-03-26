package resources

import "fmt"

// PurposeTransitionType represents a type of purpose state transition.
type PurposeTransitionType string

const (
	PurposeTransitionActivate  PurposeTransitionType = "activate"
	PurposeTransitionSuspend   PurposeTransitionType = "suspend"
	PurposeTransitionUnsuspend PurposeTransitionType = "unsuspend"
)

// PurposeTransition represents a single state transition to execute.
type PurposeTransition struct {
	Type PurposeTransitionType
}

// ComputePurposeTransitions determines the API calls needed to move from current state to desired state.
func ComputePurposeTransitions(currentState, desiredState string) ([]PurposeTransition, error) {
	// Terminal/unmanageable states
	switch currentState {
	case "ARCHIVED":
		return nil, fmt.Errorf("cannot transition from ARCHIVED state")
	case "REJECTED":
		return nil, fmt.Errorf("cannot transition from REJECTED state")
	case "WAITING_FOR_APPROVAL":
		return nil, fmt.Errorf("cannot transition from WAITING_FOR_APPROVAL state")
	}

	switch {
	case currentState == "DRAFT" && desiredState == "DRAFT":
		return []PurposeTransition{}, nil
	case currentState == "DRAFT" && desiredState == "ACTIVE":
		return []PurposeTransition{{Type: PurposeTransitionActivate}}, nil
	case currentState == "DRAFT" && desiredState == "SUSPENDED":
		return nil, fmt.Errorf("cannot transition from DRAFT to SUSPENDED")
	case currentState == "ACTIVE" && desiredState == "ACTIVE":
		return []PurposeTransition{}, nil
	case currentState == "ACTIVE" && desiredState == "SUSPENDED":
		return []PurposeTransition{{Type: PurposeTransitionSuspend}}, nil
	case currentState == "ACTIVE" && desiredState == "DRAFT":
		return nil, fmt.Errorf("cannot transition from ACTIVE to DRAFT")
	case currentState == "SUSPENDED" && desiredState == "SUSPENDED":
		return []PurposeTransition{}, nil
	case currentState == "SUSPENDED" && desiredState == "ACTIVE":
		return []PurposeTransition{{Type: PurposeTransitionUnsuspend}}, nil
	case currentState == "SUSPENDED" && desiredState == "DRAFT":
		return nil, fmt.Errorf("cannot transition from SUSPENDED to DRAFT")
	default:
		return nil, fmt.Errorf("unsupported purpose transition from %s to %s", currentState, desiredState)
	}
}
