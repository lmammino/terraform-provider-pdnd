# Auto-token mode (recommended): the provider obtains and refreshes tokens automatically.
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  client_id        = var.pdnd_client_id
  purpose_id       = var.pdnd_purpose_id
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}

# Manual token mode (alternative): supply a pre-obtained access token.
# provider "pdnd" {
#   base_url         = "https://api.interop.pagopa.it/v3"
#   access_token     = var.pdnd_access_token
#   dpop_private_key = file(var.dpop_private_key_path)
#   dpop_key_id      = var.dpop_key_id
# }
