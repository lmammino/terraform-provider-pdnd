# Upload a document to an e-service descriptor.
resource "pdnd_eservice_descriptor_document" "api_spec" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id
  pretty_name   = "API Specification"
  file_path     = "${path.module}/api-spec.yaml"
  content_type  = "application/yaml"
  file_hash     = filebase64sha256("${path.module}/api-spec.yaml")
}
