provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  access_token     = var.pdnd_access_token
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
