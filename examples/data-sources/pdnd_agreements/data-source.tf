data "pdnd_agreements" "active" {
  states = ["ACTIVE"]
}

output "active_agreement_count" {
  value = length(data.pdnd_agreements.active.agreements)
}
