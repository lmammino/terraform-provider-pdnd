package resources

import "fmt"

// TransitionType represents a type of agreement state transition.
type TransitionType string

const (
	TransitionSubmit    TransitionType = "submit"
	TransitionSuspend   TransitionType = "suspend"
	TransitionUnsuspend TransitionType = "unsuspend"
	TransitionUpgrade   TransitionType = "upgrade"
)

// Transition represents a single state transition to execute.
type Transition struct {
	Type TransitionType
}

// ComputeTransitions determines the API calls needed to move from current state to desired state.
func ComputeTransitions(currentState, desiredState string, descriptorChanged bool) ([]Transition, error) {
	// Terminal/unmanageable states
	switch currentState {
	case "PENDING":
		return nil, fmt.Errorf("cannot transition from PENDING state")
	case "REJECTED":
		return nil, fmt.Errorf("cannot transition from REJECTED state")
	case "ARCHIVED":
		return nil, fmt.Errorf("cannot transition from ARCHIVED state")
	case "MISSING_CERTIFIED_ATTRIBUTES":
		return nil, fmt.Errorf("cannot transition from MISSING_CERTIFIED_ATTRIBUTES state")
	}

	switch {
	case currentState == "DRAFT" && desiredState == "DRAFT":
		return []Transition{}, nil
	case currentState == "DRAFT" && desiredState == "ACTIVE":
		return []Transition{{Type: TransitionSubmit}}, nil
	case currentState == "DRAFT" && desiredState == "SUSPENDED":
		return nil, fmt.Errorf("cannot transition from DRAFT to SUSPENDED")
	case currentState == "ACTIVE" && desiredState == "ACTIVE" && !descriptorChanged:
		return []Transition{}, nil
	case currentState == "ACTIVE" && desiredState == "ACTIVE" && descriptorChanged:
		return []Transition{{Type: TransitionUpgrade}}, nil
	case currentState == "ACTIVE" && desiredState == "SUSPENDED":
		return []Transition{{Type: TransitionSuspend}}, nil
	case currentState == "ACTIVE" && desiredState == "DRAFT":
		return nil, fmt.Errorf("cannot transition from ACTIVE to DRAFT")
	case currentState == "SUSPENDED" && desiredState == "SUSPENDED" && !descriptorChanged:
		return []Transition{}, nil
	case currentState == "SUSPENDED" && desiredState == "SUSPENDED" && descriptorChanged:
		return []Transition{{Type: TransitionUpgrade}}, nil
	case currentState == "SUSPENDED" && desiredState == "ACTIVE":
		return []Transition{{Type: TransitionUnsuspend}}, nil
	case currentState == "SUSPENDED" && desiredState == "DRAFT":
		return nil, fmt.Errorf("cannot transition from SUSPENDED to DRAFT")
	default:
		return nil, fmt.Errorf("unsupported transition from %s to %s", currentState, desiredState)
	}
}
