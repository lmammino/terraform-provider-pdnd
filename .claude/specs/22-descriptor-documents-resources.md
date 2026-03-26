# Spec 22: Terraform Resources (Document + Interface)

## Objective

Implement two new Terraform resources: `pdnd_eservice_descriptor_document` and `pdnd_eservice_descriptor_interface`. When done, `go build ./...` must succeed and the provider must register both resources.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/descriptor_certified_attributes_resource.go` — thin resource wrapper pattern (struct, Metadata, Schema, Configure, CRUD, ImportState)
2. `internal/resources/descriptor_attributes_shared.go` — schema with RequiresReplace, Configure from providerdata, UUID validator (`uuidRegex`), composite ID pattern
3. `internal/resources/eservice_descriptor_import.go` — `parseDescriptorCompositeID` for 2-part composite ID import
4. `internal/models/descriptor_attributes.go` — model pattern with tfsdk tags
5. `internal/providerdata/providerdata.go` — ProviderData struct to extend
6. `internal/provider/provider.go` — Configure method (line ~193), Resources list (line ~211)
7. `internal/client/api/descriptor_documents.go` — DescriptorDocumentsAPI interface (created in Spec 20)
8. `internal/resources/eservice_descriptor_resource.go` — reference for schema patterns, UseStateForUnknown, RequiresReplace

## Files to Create/Modify

### 1. Create `internal/models/descriptor_document.go`

```go
package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// DescriptorDocumentResourceModel is the Terraform state for pdnd_eservice_descriptor_document.
type DescriptorDocumentResourceModel struct {
    ID           types.String `tfsdk:"id"`             // Composite: eservice_id/descriptor_id/document_id
    EServiceID   types.String `tfsdk:"eservice_id"`
    DescriptorID types.String `tfsdk:"descriptor_id"`
    PrettyName   types.String `tfsdk:"pretty_name"`
    FilePath     types.String `tfsdk:"file_path"`
    ContentType  types.String `tfsdk:"content_type"`
    DocumentID   types.String `tfsdk:"document_id"`    // Server-assigned UUID
    Name         types.String `tfsdk:"name"`            // Filename from file_path
    CreatedAt    types.String `tfsdk:"created_at"`
    FileHash     types.String `tfsdk:"file_hash"`       // SHA256 for change detection
}

// DescriptorInterfaceResourceModel is the Terraform state for pdnd_eservice_descriptor_interface.
type DescriptorInterfaceResourceModel struct {
    ID           types.String `tfsdk:"id"`             // Composite: eservice_id/descriptor_id
    EServiceID   types.String `tfsdk:"eservice_id"`
    DescriptorID types.String `tfsdk:"descriptor_id"`
    PrettyName   types.String `tfsdk:"pretty_name"`
    FilePath     types.String `tfsdk:"file_path"`
    ContentType  types.String `tfsdk:"content_type"`
    DocumentID   types.String `tfsdk:"document_id"`    // Server-assigned UUID
    Name         types.String `tfsdk:"name"`
    CreatedAt    types.String `tfsdk:"created_at"`
    FileHash     types.String `tfsdk:"file_hash"`
}
```

### 2. Create `internal/resources/descriptor_document_import.go`

Import helper for 3-part composite IDs:

```go
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
```

### 3. Create `internal/resources/descriptor_document_resource.go`

Resource: `pdnd_eservice_descriptor_document`

**Struct:**
```go
type descriptorDocumentResource struct {
    client api.DescriptorDocumentsAPI
}
```

**TypeName:** `req.ProviderTypeName + "_eservice_descriptor_document"`

**Schema attributes:**
- `id` — Computed, UseStateForUnknown
- `eservice_id` — Required, RequiresReplace, UUID validator
- `descriptor_id` — Required, RequiresReplace, UUID validator
- `pretty_name` — Required, RequiresReplace
- `file_path` — Required, RequiresReplace
- `content_type` — Required, RequiresReplace
- `file_hash` — Optional, Computed, RequiresReplace (user can provide `filebase64sha256(...)`)
- `document_id` — Computed, UseStateForUnknown
- `name` — Computed, UseStateForUnknown
- `created_at` — Computed, UseStateForUnknown

**Configure:** Get `DescriptorDocumentsAPI` from `providerdata.ProviderData`

**Create:**
1. Read file from `file_path` using `os.ReadFile`
2. Extract filename with `filepath.Base`
3. Compute file hash: `sha256.Sum256(fileContent)` → hex string
4. Call `client.UploadDocument(ctx, eserviceID, descriptorID, filename, fileContent, prettyName)`
5. Populate state from response: document_id, name, created_at
6. Set composite ID: `eserviceId/descriptorId/documentId`
7. If user provided `file_hash`, verify it matches computed hash (or just store computed)
8. Store computed `file_hash`

**Read:**
1. Call `client.GetDocumentByID(ctx, eserviceID, descriptorID, documentID)`
2. If not found (404), call `resp.State.RemoveResource`
3. Otherwise, verify document still exists — keep all state fields as-is (they're all immutable)
4. Preserve `file_path` and `file_hash` from existing state (not available from API)

**Update:** No-op — all user attributes have RequiresReplace, so Terraform handles destroy+create.

**Delete:** Call `client.DeleteDocument(ctx, eserviceID, descriptorID, documentID)`. Handle 404 gracefully.

**ImportState:** Parse 3-part composite ID, set eservice_id, descriptor_id, document_id, id. Note: `file_path` and `file_hash` will be empty after import — user must provide in config, which will trigger a replacement on next apply.

### 4. Create `internal/resources/descriptor_interface_resource.go`

Resource: `pdnd_eservice_descriptor_interface`

Same schema as document except:
- `id` composite is `eservice_id/descriptor_id` (no document_id)
- No `document_id` attribute in schema (it's embedded in the ID) — actually, still include `document_id` as Computed for reference since the upload returns a Document with an ID

**Create:** Call `client.UploadInterface` instead. Composite ID = `eserviceId/descriptorId`.

**Read:** Call `client.CheckInterfaceExists(ctx, eserviceID, descriptorID)`. If false, remove resource. Keep state fields as-is.

**Delete:** Call `client.DeleteInterface(ctx, eserviceID, descriptorID)`.

**ImportState:** Parse 2-part composite ID using existing `parseDescriptorCompositeID`. Set eservice_id, descriptor_id, id.

### 5. Extend `internal/providerdata/providerdata.go`

Add field:
```go
DescriptorDocumentsAPI api.DescriptorDocumentsAPI
```

### 6. Extend `internal/provider/provider.go`

In `Configure` (after `descriptorAttributesAPI` creation):
```go
descriptorDocumentsAPI := api.NewDescriptorDocumentsClient(genClient)
```

Add to `pd` struct:
```go
DescriptorDocumentsAPI: descriptorDocumentsAPI,
```

In `Resources` list:
```go
resources.NewDescriptorDocumentResource,
resources.NewDescriptorInterfaceResource,
```

## Execution Steps

1. Read all referenced files
2. Create `models/descriptor_document.go`
3. Create `resources/descriptor_document_import.go`
4. Create `resources/descriptor_document_resource.go`
5. Create `resources/descriptor_interface_resource.go`
6. Extend `providerdata.go`
7. Extend `provider.go`
8. Run `go build ./...`
9. Fix any compilation errors

## Verification

```shell
go build ./...
```

## Important Notes

- Do NOT break existing resources or tests
- `os.ReadFile` requires `import "os"`, `filepath.Base` requires `import "path/filepath"`, `sha256` requires `import "crypto/sha256"` and `import "encoding/hex"`
- For `file_hash`, compute as: `hex.EncodeToString(sha256.Sum256(fileContent)[:])`
- The `uuidRegex` variable is already defined in `agreement_resource.go` and available package-wide
- Both resources use `RequiresReplace()` on all user-facing attributes — there is no Update logic needed
- For the Update method implementation, just read the plan into state (Terraform only calls Update when non-RequiresReplace attributes change, which is only `file_hash` when user doesn't provide it — handle this edge case by computing hash and setting state)
