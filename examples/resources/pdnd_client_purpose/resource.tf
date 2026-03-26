resource "pdnd_client_purpose" "link" {
  client_id  = data.pdnd_client.example.id
  purpose_id = pdnd_purpose.example.id
}
