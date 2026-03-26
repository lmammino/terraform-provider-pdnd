# Upload an interface file to an e-service descriptor.
resource "pdnd_eservice_descriptor_interface" "openapi" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id
  pretty_name   = "OpenAPI Specification"
  file_path     = "${path.module}/openapi.yaml"
  content_type  = "application/yaml"
  file_hash     = filebase64sha256("${path.module}/openapi.yaml")
}
