package resources

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// parseTenantAttributeCompositeID parses a composite import ID of the form "tenant_id/attribute_id".
func parseTenantAttributeCompositeID(id string) (tenantID, attributeID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import ID format: tenant_id/attribute_id, got: %s", id)
	}

	tenantID = parts[0]
	attributeID = parts[1]

	if _, err := uuid.Parse(tenantID); err != nil {
		return "", "", fmt.Errorf("invalid tenant_id UUID: %s", tenantID)
	}
	if _, err := uuid.Parse(attributeID); err != nil {
		return "", "", fmt.Errorf("invalid attribute_id UUID: %s", attributeID)
	}

	return tenantID, attributeID, nil
}
