data "pdnd_eservice" "example" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}

output "eservice_name" {
  value = data.pdnd_eservice.example.name
}
