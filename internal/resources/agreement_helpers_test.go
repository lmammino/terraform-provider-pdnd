package resources

import (
	"testing"
)

func TestCanDelete_Draft(t *testing.T) {
	if !canDelete("DRAFT") {
		t.Fatal("expected canDelete(DRAFT) to be true")
	}
}

func TestCanDelete_Pending(t *testing.T) {
	if !canDelete("PENDING") {
		t.Fatal("expected canDelete(PENDING) to be true")
	}
}

func TestCanDelete_MissingCertifiedAttributes(t *testing.T) {
	if !canDelete("MISSING_CERTIFIED_ATTRIBUTES") {
		t.Fatal("expected canDelete(MISSING_CERTIFIED_ATTRIBUTES) to be true")
	}
}

func TestCanDelete_Active(t *testing.T) {
	if canDelete("ACTIVE") {
		t.Fatal("expected canDelete(ACTIVE) to be false")
	}
}

func TestCanDelete_Suspended(t *testing.T) {
	if canDelete("SUSPENDED") {
		t.Fatal("expected canDelete(SUSPENDED) to be false")
	}
}

func TestCanDelete_Archived(t *testing.T) {
	if canDelete("ARCHIVED") {
		t.Fatal("expected canDelete(ARCHIVED) to be false")
	}
}

func TestCanDelete_Rejected(t *testing.T) {
	if canDelete("REJECTED") {
		t.Fatal("expected canDelete(REJECTED) to be false")
	}
}
