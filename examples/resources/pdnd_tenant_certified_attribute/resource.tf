resource "pdnd_tenant_certified_attribute" "ipa" {
  tenant_id    = data.pdnd_tenant.example.id
  attribute_id = data.pdnd_certified_attribute.ipa.id
}
