data "pdnd_purpose" "example" {
  id = "990e8400-e29b-41d4-a716-446655440000"
}

output "purpose_title" {
  value = data.pdnd_purpose.example.title
}
