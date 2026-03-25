data "pdnd_eservice_descriptor" "example" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "770e8400-e29b-41d4-a716-446655440002"
}

output "descriptor_state" {
  value = data.pdnd_eservice_descriptor.example.state
}
