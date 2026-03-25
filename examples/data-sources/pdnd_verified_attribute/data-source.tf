data "pdnd_verified_attribute" "example" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}

output "verified_attribute_name" {
  value = data.pdnd_verified_attribute.example.name
}
