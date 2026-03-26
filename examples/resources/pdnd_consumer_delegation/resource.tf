resource "pdnd_consumer_delegation" "example" {
  eservice_id = pdnd_eservice.example.id
  delegate_id = "550e8400-e29b-41d4-a716-446655440000"
}
