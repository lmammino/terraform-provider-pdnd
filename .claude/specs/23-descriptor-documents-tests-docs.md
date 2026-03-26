# Spec 23: Tests and Documentation for Descriptor Documents/Interface

## Objective

Add acceptance tests for both `pdnd_eservice_descriptor_document` and `pdnd_eservice_descriptor_interface` resources, plus documentation templates and examples. When done, `TF_ACC=1 go test ./internal/resources/... -v -run TestAccDescriptor(Document|Interface)` must pass, and `make docs` must generate docs for both new resources.

## Working Directory

`/Users/luciano/repos/pdnd-tf-provider`

## Existing Code to Read First

1. `internal/resources/descriptor_certified_attributes_resource_acc_test.go` — acceptance test pattern for descriptor sub-resources: `seedEServiceWithDraftDescriptor`, `testAccProviderConfig`, `testAccProtoV6ProviderFactories`, fake server setup
2. `internal/resources/agreement_resource_acc_test.go` — `testAccProviderConfig` definition (line 41), `testPrivateKeyPEM` (init function line 23-33)
3. `internal/resources/purpose_resource_acc_test.go` — tracker pattern for cleanup, CheckDestroy pattern
4. `internal/testing/fakepdnd/server.go` — SeedDocument, GetDocuments, SeedInterface, GetInterface methods (added in Spec 21)
5. `internal/resources/descriptor_document_resource.go` — the document resource (created in Spec 22)
6. `internal/resources/descriptor_interface_resource.go` — the interface resource (created in Spec 22)
7. `templates/resources/eservice_descriptor_certified_attributes.md.tmpl` — template format reference
8. `examples/resources/pdnd_eservice_descriptor_certified_attributes/` — example directory structure (resource.tf, import.sh)
9. `README.md` — resources table to update

## Files to Create/Modify

### 1. Create `internal/resources/descriptor_document_resource_acc_test.go`

**Setup:** Use `seedEServiceWithDraftDescriptor` helper (already exists from the attributes tests).

**Temp file creation pattern:** Each test creates a temporary file:
```go
tmpDir := t.TempDir()
filePath := filepath.Join(tmpDir, "test-doc.yaml")
os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644)
```

**Tests:**

| Test | Description |
|------|-------------|
| `TestAccDescriptorDocument_Create` | Upload a document, verify document_id, name, pretty_name, content_type, created_at, file_hash are set |
| `TestAccDescriptorDocument_Delete` | Create then destroy (remove config), verify `fake.GetDocuments()` returns empty or no match |
| `TestAccDescriptorDocument_Import` | Create, then import via `eservice_id/descriptor_id/document_id`. Use ImportStateIdFunc to extract document_id from state. ImportStateVerifyIgnore: `["file_path", "file_hash"]` (not recoverable from API) |

**HCL config template:**
```hcl
resource "pdnd_eservice_descriptor_document" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "Test Document"
  file_path     = %q
  content_type  = "application/yaml"
}
```

**Import test detail:** The import test needs the document_id which is server-assigned. Use `ImportStateIdFunc`:
```go
ImportStateIdFunc: func(s *terraform.State) (string, error) {
    for _, rs := range s.RootModule().Resources {
        if rs.Type == "pdnd_eservice_descriptor_document" {
            return fmt.Sprintf("%s/%s/%s",
                rs.Primary.Attributes["eservice_id"],
                rs.Primary.Attributes["descriptor_id"],
                rs.Primary.Attributes["document_id"]), nil
        }
    }
    return "", fmt.Errorf("resource not found")
},
ImportStateVerifyIgnore: []string{"file_path", "file_hash"},
```

### 2. Create `internal/resources/descriptor_interface_resource_acc_test.go`

**Tests:**

| Test | Description |
|------|-------------|
| `TestAccDescriptorInterface_Create` | Upload interface file, verify computed attributes |
| `TestAccDescriptorInterface_Delete` | Create then destroy, verify `fake.GetInterface()` returns nil |
| `TestAccDescriptorInterface_Import` | Import via `eservice_id/descriptor_id`, ImportStateVerifyIgnore: `["file_path", "file_hash"]` |

**HCL config template:**
```hcl
resource "pdnd_eservice_descriptor_interface" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "OpenAPI Spec"
  file_path     = %q
  content_type  = "application/yaml"
}
```

### 3. Create `templates/resources/eservice_descriptor_document.md.tmpl`

```markdown
---
page_title: "pdnd_eservice_descriptor_document Resource - PDND"
description: |-
  Manages a document on a PDND E-Service Descriptor.
---

# pdnd_eservice_descriptor_document (Resource)

Manages a single document attached to a PDND E-Service Descriptor. To add multiple documents, create multiple resource instances.

All user-configurable attributes trigger replacement when changed — documents cannot be updated in place.

## Example Usage

{{ tffile "examples/resources/pdnd_eservice_descriptor_document/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

Descriptor documents can be imported using the composite ID `eservice_id/descriptor_id/document_id`:

{{ codefile "shell" "examples/resources/pdnd_eservice_descriptor_document/import.sh" }}
```

### 4. Create `templates/resources/eservice_descriptor_interface.md.tmpl`

```markdown
---
page_title: "pdnd_eservice_descriptor_interface Resource - PDND"
description: |-
  Manages the interface file on a PDND E-Service Descriptor.
---

# pdnd_eservice_descriptor_interface (Resource)

Manages the interface file (e.g., OpenAPI specification) on a PDND E-Service Descriptor. Each descriptor can have at most one interface.

All user-configurable attributes trigger replacement when changed — interfaces cannot be updated in place.

## Example Usage

{{ tffile "examples/resources/pdnd_eservice_descriptor_interface/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

Descriptor interfaces can be imported using the composite ID `eservice_id/descriptor_id`:

{{ codefile "shell" "examples/resources/pdnd_eservice_descriptor_interface/import.sh" }}
```

### 5. Create `examples/resources/pdnd_eservice_descriptor_document/resource.tf`

```hcl
# Upload a document to an e-service descriptor.
resource "pdnd_eservice_descriptor_document" "api_spec" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id
  pretty_name   = "API Specification"
  file_path     = "${path.module}/api-spec.yaml"
  content_type  = "application/yaml"
  file_hash     = filebase64sha256("${path.module}/api-spec.yaml")
}
```

### 6. Create `examples/resources/pdnd_eservice_descriptor_document/import.sh`

```shell
terraform import pdnd_eservice_descriptor_document.api_spec 550e8400-e29b-41d4-a716-446655440000/770e8400-e29b-41d4-a716-446655440002/dd0e8400-e29b-41d4-a716-446655440020
```

### 7. Create `examples/resources/pdnd_eservice_descriptor_interface/resource.tf`

```hcl
# Upload an interface file to an e-service descriptor.
resource "pdnd_eservice_descriptor_interface" "openapi" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id
  pretty_name   = "OpenAPI Specification"
  file_path     = "${path.module}/openapi.yaml"
  content_type  = "application/yaml"
  file_hash     = filebase64sha256("${path.module}/openapi.yaml")
}
```

### 8. Create `examples/resources/pdnd_eservice_descriptor_interface/import.sh`

```shell
terraform import pdnd_eservice_descriptor_interface.openapi 550e8400-e29b-41d4-a716-446655440000/770e8400-e29b-41d4-a716-446655440002
```

### 9. Modify `README.md` — Add rows to resources table

Add after the `pdnd_purpose` row:

```
| `pdnd_eservice_descriptor_document`  | Manages a document on a descriptor        |
| `pdnd_eservice_descriptor_interface` | Manages the interface file on a descriptor |
```

## Execution Steps

1. Read all referenced files for patterns
2. Create example directories and files (resource.tf, import.sh) for both resources
3. Create doc templates for both resources
4. Create `descriptor_document_resource_acc_test.go`
5. Create `descriptor_interface_resource_acc_test.go`
6. Update `README.md`
7. Run `TF_ACC=1 go test ./internal/resources/... -v -run TestAccDescriptorDocument`
8. Run `TF_ACC=1 go test ./internal/resources/... -v -run TestAccDescriptorInterface`
9. Run full test suite: `go test ./...`
10. Run `make lint`
11. Run `make docs`

## Verification

```shell
TF_ACC=1 go test ./internal/resources/... -v -run "TestAccDescriptor(Document|Interface)"
go test ./...
make lint
make docs
```

## Important Notes

- Do NOT break existing tests
- `testAccProviderConfig` and `testAccProtoV6ProviderFactories` are defined in `agreement_resource_acc_test.go` and available to all test files in `resources_test` package
- `seedEServiceWithDraftDescriptor` is defined in `descriptor_certified_attributes_resource_acc_test.go`
- For temp files in tests: use `t.TempDir()` which auto-cleans up after test. Write files with `os.WriteFile(path, content, 0644)`
- Import tests for document: `file_path` and `file_hash` CANNOT be recovered from the API, so add them to `ImportStateVerifyIgnore`
- Import tests for interface: same — ignore `file_path` and `file_hash`
- No cleanup tracker needed — documents/interfaces can always be deleted regardless of state
