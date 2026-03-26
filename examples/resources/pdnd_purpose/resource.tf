# Create a purpose in DRAFT state
resource "pdnd_purpose" "draft" {
  eservice_id       = pdnd_eservice.example.id
  title             = "My Purpose"
  description       = "Access to example e-service"
  daily_calls       = 1000
  is_free_of_charge = false
  desired_state     = "DRAFT"
}

# Create a purpose and activate it
resource "pdnd_purpose" "active" {
  eservice_id       = pdnd_eservice.example.id
  title             = "My Active Purpose"
  description       = "Active access to example e-service"
  daily_calls       = 5000
  is_free_of_charge = true
  desired_state     = "ACTIVE"
}

# Create a purpose with high daily calls that may need approval
resource "pdnd_purpose" "high_calls" {
  eservice_id                = pdnd_eservice.example.id
  title                      = "High Volume Purpose"
  description                = "High volume access to example e-service"
  daily_calls                = 100000
  is_free_of_charge          = false
  desired_state              = "ACTIVE"
  allow_waiting_for_approval = true
}
