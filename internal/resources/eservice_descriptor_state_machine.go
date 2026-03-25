package resources

import "fmt"

// DescriptorTransitionType represents a type of descriptor state transition.
type DescriptorTransitionType string

const (
	DescriptorTransitionPublish   DescriptorTransitionType = "publish"
	DescriptorTransitionSuspend   DescriptorTransitionType = "suspend"
	DescriptorTransitionUnsuspend DescriptorTransitionType = "unsuspend"
)

// DescriptorTransition represents a single descriptor state transition to execute.
type DescriptorTransition struct {
	Type DescriptorTransitionType
}

// ComputeDescriptorTransitions determines the API calls needed to move a descriptor
// from currentState to desiredState.
func ComputeDescriptorTransitions(currentState, desiredState string) ([]DescriptorTransition, error) {
	// Terminal/unmanageable states
	switch currentState {
	case "DEPRECATED":
		return nil, fmt.Errorf("cannot transition from DEPRECATED state")
	case "ARCHIVED":
		return nil, fmt.Errorf("cannot transition from ARCHIVED state")
	case "WAITING_FOR_APPROVAL":
		return nil, fmt.Errorf("cannot transition from WAITING_FOR_APPROVAL state")
	}

	switch {
	case currentState == "DRAFT" && desiredState == "DRAFT":
		return []DescriptorTransition{}, nil
	case currentState == "DRAFT" && desiredState == "PUBLISHED":
		return []DescriptorTransition{{Type: DescriptorTransitionPublish}}, nil
	case currentState == "DRAFT" && desiredState == "SUSPENDED":
		return nil, fmt.Errorf("cannot transition from DRAFT to SUSPENDED")
	case currentState == "PUBLISHED" && desiredState == "PUBLISHED":
		return []DescriptorTransition{}, nil
	case currentState == "PUBLISHED" && desiredState == "SUSPENDED":
		return []DescriptorTransition{{Type: DescriptorTransitionSuspend}}, nil
	case currentState == "SUSPENDED" && desiredState == "SUSPENDED":
		return []DescriptorTransition{}, nil
	case currentState == "SUSPENDED" && desiredState == "PUBLISHED":
		return []DescriptorTransition{{Type: DescriptorTransitionUnsuspend}}, nil
	default:
		return nil, fmt.Errorf("unsupported descriptor transition from %s to %s", currentState, desiredState)
	}
}
