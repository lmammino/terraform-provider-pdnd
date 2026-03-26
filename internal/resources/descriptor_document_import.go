package resources

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// parseDocumentCompositeID parses a composite import ID of the form "eservice_id/descriptor_id/document_id".
func parseDocumentCompositeID(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected import ID format: eservice_id/descriptor_id/document_id, got: %s", id)
	}
	for i, label := range []string{"eservice_id", "descriptor_id", "document_id"} {
		if _, err := uuid.Parse(parts[i]); err != nil {
			return "", "", "", fmt.Errorf("invalid %s UUID: %s", label, parts[i])
		}
	}
	return parts[0], parts[1], parts[2], nil
}
