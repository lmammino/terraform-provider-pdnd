# Create an e-service with an initial descriptor
resource "pdnd_eservice" "example" {
  name        = "My API Service"
  description = "A REST API service for data exchange"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
