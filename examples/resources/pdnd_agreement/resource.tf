# Create a draft agreement
resource "pdnd_agreement" "draft" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "660e8400-e29b-41d4-a716-446655440001"
  desired_state = "DRAFT"
}

# Create and activate an agreement
resource "pdnd_agreement" "active" {
  eservice_id    = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id  = "660e8400-e29b-41d4-a716-446655440001"
  desired_state  = "ACTIVE"
  consumer_notes = "Requesting access to the service"
}

# Create an agreement that tolerates pending approval
resource "pdnd_agreement" "with_pending" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "660e8400-e29b-41d4-a716-446655440001"
  desired_state = "ACTIVE"
  allow_pending = true
}
