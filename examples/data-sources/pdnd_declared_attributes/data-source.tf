data "pdnd_declared_attributes" "by_name" {
  name = "SPID"
}

output "declared_attribute_count" {
  value = length(data.pdnd_declared_attributes.by_name.attributes)
}
