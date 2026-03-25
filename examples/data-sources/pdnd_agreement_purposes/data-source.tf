data "pdnd_agreement_purposes" "example" {
  agreement_id = "550e8400-e29b-41d4-a716-446655440000"
}

output "purpose_count" {
  value = length(data.pdnd_agreement_purposes.example.purposes)
}
