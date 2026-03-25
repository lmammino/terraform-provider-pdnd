data "pdnd_eservices" "rest_services" {
  technology = "REST"
}

output "rest_eservice_count" {
  value = length(data.pdnd_eservices.rest_services.eservices)
}
