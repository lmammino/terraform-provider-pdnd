resource "pdnd_tenant_verified_attribute" "example" {
  tenant_id    = data.pdnd_tenant.example.id
  attribute_id = data.pdnd_verified_attribute.example.id
  agreement_id = pdnd_agreement.example.id
}
