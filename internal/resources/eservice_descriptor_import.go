package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *eserviceDescriptorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	eserviceID, descriptorID, err := parseDescriptorCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eservice_id"), types.StringValue(eserviceID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(descriptorID))...)
}

// parseDescriptorCompositeID parses a composite import ID of the form "eservice_id/descriptor_id".
func parseDescriptorCompositeID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import ID format: eservice_id/descriptor_id, got: %s", id)
	}

	eserviceID := parts[0]
	descriptorID := parts[1]

	if _, err := uuid.Parse(eserviceID); err != nil {
		return "", "", fmt.Errorf("invalid eservice_id UUID: %s", eserviceID)
	}
	if _, err := uuid.Parse(descriptorID); err != nil {
		return "", "", fmt.Errorf("invalid descriptor_id UUID: %s", descriptorID)
	}

	return eserviceID, descriptorID, nil
}
