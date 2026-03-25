package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ImportState implements resource.ResourceWithImportState.
func (r *eserviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, idPath, req, resp)
}
