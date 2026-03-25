data "pdnd_verified_attributes" "by_name" {
  name = "PEC"
}

output "verified_attribute_count" {
  value = length(data.pdnd_verified_attributes.by_name.attributes)
}
