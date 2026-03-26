resource "pdnd_client_key" "api_key" {
  client_id = data.pdnd_client.example.id
  key       = filebase64(var.public_key_path)
  use       = "SIG"
  alg       = "RS256"
  name      = "Production API Key"
}
