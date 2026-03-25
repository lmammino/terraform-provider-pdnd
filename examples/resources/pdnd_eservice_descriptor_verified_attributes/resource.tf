# Manage verified attribute requirements on a descriptor.
resource "pdnd_eservice_descriptor_verified_attributes" "example" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id

  group {
    attribute_ids = [data.pdnd_verified_attribute.example.id]
  }
}
