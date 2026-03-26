# List all active purposes for a specific e-service
data "pdnd_purposes" "active" {
  eservice_ids = [pdnd_eservice.example.id]
  states       = ["ACTIVE"]
}

output "active_purpose_count" {
  value = length(data.pdnd_purposes.active.purposes)
}
