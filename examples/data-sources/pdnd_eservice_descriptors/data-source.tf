data "pdnd_eservice_descriptors" "example" {
  eservice_id = "550e8400-e29b-41d4-a716-446655440000"
}

output "descriptor_count" {
  value = length(data.pdnd_eservice_descriptors.example.descriptors)
}
