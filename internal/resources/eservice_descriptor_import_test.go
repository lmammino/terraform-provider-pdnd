package resources

import (
	"testing"
)

func TestParseDescriptorCompositeID_Valid(t *testing.T) {
	eserviceID, descriptorID, err := parseDescriptorCompositeID("550e8400-e29b-41d4-a716-446655440000/660e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if eserviceID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected eservice_id 550e8400-e29b-41d4-a716-446655440000, got %s", eserviceID)
	}
	if descriptorID != "660e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected descriptor_id 660e8400-e29b-41d4-a716-446655440000, got %s", descriptorID)
	}
}

func TestParseDescriptorCompositeID_InvalidNoSlash(t *testing.T) {
	_, _, err := parseDescriptorCompositeID("550e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Fatal("expected error for ID without slash")
	}
}

func TestParseDescriptorCompositeID_InvalidTooManySlashes(t *testing.T) {
	_, _, err := parseDescriptorCompositeID("a/b/c")
	if err == nil {
		t.Fatal("expected error for ID with too many slashes")
	}
}

func TestParseDescriptorCompositeID_InvalidUUIDs(t *testing.T) {
	_, _, err := parseDescriptorCompositeID("not-uuid/not-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUIDs")
	}
}

func TestParseDescriptorCompositeID_InvalidFirstUUID(t *testing.T) {
	_, _, err := parseDescriptorCompositeID("not-uuid/550e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Fatal("expected error for invalid first UUID")
	}
}

func TestParseDescriptorCompositeID_InvalidSecondUUID(t *testing.T) {
	_, _, err := parseDescriptorCompositeID("550e8400-e29b-41d4-a716-446655440000/not-uuid")
	if err == nil {
		t.Fatal("expected error for invalid second UUID")
	}
}
