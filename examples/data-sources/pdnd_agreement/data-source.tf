data "pdnd_agreement" "example" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}

output "agreement_state" {
  value = data.pdnd_agreement.example.state
}
