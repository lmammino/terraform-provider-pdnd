resource "pdnd_tenant_declared_attribute" "example" {
  tenant_id    = data.pdnd_tenant.example.id
  attribute_id = data.pdnd_declared_attribute.example.id
}
