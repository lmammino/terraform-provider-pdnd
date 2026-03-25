# Manage certified attribute requirements on a descriptor.
# Attributes within a group are OR'd (consumer needs at least one).
# Groups are AND'd (consumer must satisfy all groups).
resource "pdnd_eservice_descriptor_certified_attributes" "example" {
  eservice_id   = pdnd_eservice.example.id
  descriptor_id = pdnd_eservice_descriptor.example.id

  # Group 1: consumer must have this attribute
  group {
    attribute_ids = [data.pdnd_certified_attribute.ipa.id]
  }

  # Group 2: consumer must have at least one of these attributes
  group {
    attribute_ids = [
      data.pdnd_certified_attribute.cie.id,
      data.pdnd_certified_attribute.spid.id,
    ]
  }
}
