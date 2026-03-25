# Create and publish a descriptor
resource "pdnd_eservice_descriptor" "v2" {
  eservice_id               = pdnd_eservice.example.id
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.example.com"]
  daily_calls_per_consumer  = 2000
  daily_calls_total         = 20000
  voucher_lifespan          = 7200
  description               = "Version 2 with increased quotas"
}

# Create a descriptor that tolerates waiting for approval
resource "pdnd_eservice_descriptor" "delegated" {
  eservice_id                = pdnd_eservice.example.id
  desired_state              = "PUBLISHED"
  allow_waiting_for_approval = true
  agreement_approval_policy  = "MANUAL"
  audience                   = ["api.example.com"]
  daily_calls_per_consumer   = 500
  daily_calls_total          = 5000
  voucher_lifespan           = 3600
}
