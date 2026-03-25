data "pdnd_declared_attribute" "example" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}

output "declared_attribute_name" {
  value = data.pdnd_declared_attribute.example.name
}
