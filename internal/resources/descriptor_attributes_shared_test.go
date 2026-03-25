package resources

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
)

func TestReconstructGroups(t *testing.T) {
	id1 := uuid.MustParse("aa000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("bb000000-0000-0000-0000-000000000002")
	id3 := uuid.MustParse("cc000000-0000-0000-0000-000000000003")

	tests := []struct {
		name     string
		entries  []api.DescriptorAttributeEntry
		expected int // number of groups
	}{
		{
			name:     "empty",
			entries:  nil,
			expected: 0,
		},
		{
			name: "single group with one attribute",
			entries: []api.DescriptorAttributeEntry{
				{AttributeID: id1, GroupIndex: 0},
			},
			expected: 1,
		},
		{
			name: "single group with two attributes",
			entries: []api.DescriptorAttributeEntry{
				{AttributeID: id1, GroupIndex: 0},
				{AttributeID: id2, GroupIndex: 0},
			},
			expected: 1,
		},
		{
			name: "two groups",
			entries: []api.DescriptorAttributeEntry{
				{AttributeID: id1, GroupIndex: 0},
				{AttributeID: id2, GroupIndex: 1},
			},
			expected: 2,
		},
		{
			name: "non-contiguous indices",
			entries: []api.DescriptorAttributeEntry{
				{AttributeID: id1, GroupIndex: 0},
				{AttributeID: id2, GroupIndex: 5},
				{AttributeID: id3, GroupIndex: 5},
			},
			expected: 2,
		},
		{
			name: "entries out of order",
			entries: []api.DescriptorAttributeEntry{
				{AttributeID: id3, GroupIndex: 2},
				{AttributeID: id1, GroupIndex: 0},
				{AttributeID: id2, GroupIndex: 1},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconstructGroups(tt.entries)
			if len(result) != tt.expected {
				t.Errorf("expected %d groups, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestReconstructGroups_Ordering(t *testing.T) {
	id1 := uuid.MustParse("aa000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("bb000000-0000-0000-0000-000000000002")
	id3 := uuid.MustParse("cc000000-0000-0000-0000-000000000003")

	entries := []api.DescriptorAttributeEntry{
		{AttributeID: id3, GroupIndex: 2},
		{AttributeID: id1, GroupIndex: 0},
		{AttributeID: id2, GroupIndex: 1},
	}

	result := reconstructGroups(entries)

	if result[0][0] != id1 {
		t.Errorf("expected first group first attribute to be %s, got %s", id1, result[0][0])
	}
	if result[1][0] != id2 {
		t.Errorf("expected second group first attribute to be %s, got %s", id2, result[1][0])
	}
	if result[2][0] != id3 {
		t.Errorf("expected third group first attribute to be %s, got %s", id3, result[2][0])
	}
}

func TestReconstructGroups_MultipleAttributesInGroup(t *testing.T) {
	id1 := uuid.MustParse("aa000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("bb000000-0000-0000-0000-000000000002")

	entries := []api.DescriptorAttributeEntry{
		{AttributeID: id1, GroupIndex: 0},
		{AttributeID: id2, GroupIndex: 0},
	}

	result := reconstructGroups(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 group, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Errorf("expected 2 attributes in group, got %d", len(result[0]))
	}
}
