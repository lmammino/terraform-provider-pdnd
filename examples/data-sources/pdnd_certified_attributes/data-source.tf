data "pdnd_certified_attributes" "by_name" {
  name = "Comune"
}

output "certified_attribute_count" {
  value = length(data.pdnd_certified_attributes.by_name.attributes)
}
