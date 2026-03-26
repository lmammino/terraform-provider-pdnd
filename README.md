# Terraform Provider for PDND Interoperability

> **WARNING: This project is extremely early stage and experimental. Use it at your own risk. APIs, resource schemas, and behaviors may change without notice. This provider is not yet suitable for production use.**

A Terraform provider for managing resources on the [PDND Interoperability](https://developers.interop.pagopa.it/) platform (API v3).

> **Milestone 2**: This provider supports agreement and e-service management.

## Prerequisites

- [Go](https://golang.org/dl/) 1.22+
- [Terraform](https://www.terraform.io/downloads.html) 1.0+

## Authentication

The provider authenticates using DPoP (Demonstration of Proof-of-Possession). It supports two modes:

- **Auto-token (recommended):** Supply `client_id` and `purpose_id` -- the provider automatically obtains and refreshes access tokens.
- **Manual token:** Supply a pre-obtained `access_token` for debugging or advanced use cases.

Both modes require a **PEM-encoded private key** and a **key ID** from the PDND portal.

See **[AUTH.md](AUTH.md)** for a detailed step-by-step guide on how to obtain each credential.

## Provider Configuration

### Auto-Token Mode (Recommended)

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  client_id        = var.pdnd_client_id
  purpose_id       = var.pdnd_purpose_id
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

### Manual Token Mode

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  access_token     = var.pdnd_access_token
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

| Attribute          | Type   | Required | Description                                                              |
|--------------------|--------|----------|--------------------------------------------------------------------------|
| `base_url`         | string | yes      | Base URL of the PDND API                                                 |
| `client_id`        | string | no*      | PDND client UUID for automatic token generation                          |
| `purpose_id`       | string | no*      | PDND purpose UUID for automatic token generation                         |
| `access_token`     | string | no*      | Pre-obtained access token (manual mode)                                  |
| `token_endpoint`   | string | no       | Token endpoint URL (default: `https://auth.interop.pagopa.it/token.oauth2`) |
| `dpop_private_key` | string | yes      | PEM-encoded private key for DPoP                                         |
| `dpop_key_id`      | string | yes      | Key ID for the DPoP private key                                          |
| `request_timeout_s`| number | no       | Request timeout in seconds (default: 30)                                 |

\* Either `client_id` + `purpose_id` (auto-token) or `access_token` (manual) must be provided. They are mutually exclusive.

## Quick Example

```hcl
resource "pdnd_agreement" "example" {
  eservice_id   = "550e8400-e29b-41d4-a716-446655440000"
  descriptor_id = "660e8400-e29b-41d4-a716-446655440001"
  desired_state = "ACTIVE"
}
```

## Available Resources

| Resource                     | Description                                   |
|------------------------------|-----------------------------------------------|
| `pdnd_agreement`             | Manages a PDND agreement lifecycle            |
| `pdnd_eservice`              | Manages a PDND e-service                      |
| `pdnd_eservice_descriptor`   | Manages an e-service descriptor lifecycle     |
| `pdnd_eservice_descriptor_certified_attributes` | Manages certified attribute groups on a descriptor |
| `pdnd_eservice_descriptor_declared_attributes`  | Manages declared attribute groups on a descriptor  |
| `pdnd_eservice_descriptor_verified_attributes`  | Manages verified attribute groups on a descriptor  |
| `pdnd_purpose`                                  | Manages a PDND purpose lifecycle                   |
| `pdnd_eservice_descriptor_document`  | Manages a document on a descriptor        |
| `pdnd_eservice_descriptor_interface` | Manages the interface file on a descriptor |
| `pdnd_consumer_delegation`   | Manages a PDND consumer delegation            |
| `pdnd_producer_delegation`   | Manages a PDND producer delegation            |
| `pdnd_client_key`            | Manages a public key on a PDND client         |
| `pdnd_client_purpose`        | Links a purpose to a PDND client              |

## Available Data Sources

| Data Source                  | Description                                      |
|------------------------------|--------------------------------------------------|
| `pdnd_agreement`             | Fetches a single agreement by ID                 |
| `pdnd_agreements`            | Lists agreements with optional filters           |
| `pdnd_agreement_purposes`    | Lists purposes associated with an agreement      |
| `pdnd_purpose`               | Fetches a single purpose by ID                   |
| `pdnd_purposes`              | Lists purposes with optional filters             |
| `pdnd_eservice`              | Fetches a single e-service by ID                 |
| `pdnd_eservices`             | Lists e-services with optional filters           |
| `pdnd_eservice_descriptor`   | Fetches a single descriptor by ID                |
| `pdnd_eservice_descriptors`  | Lists descriptors for an e-service               |
| `pdnd_certified_attribute`   | Fetches a single certified attribute by ID        |
| `pdnd_certified_attributes`  | Lists certified attributes with name filter       |
| `pdnd_declared_attribute`    | Fetches a single declared attribute by ID         |
| `pdnd_declared_attributes`   | Lists declared attributes with name filter        |
| `pdnd_verified_attribute`    | Fetches a single verified attribute by ID         |
| `pdnd_verified_attributes`   | Lists verified attributes with name filter        |
| `pdnd_consumer_delegation`   | Fetches a single consumer delegation by ID    |
| `pdnd_consumer_delegations`  | Lists consumer delegations with filters       |
| `pdnd_producer_delegation`   | Fetches a single producer delegation by ID    |
| `pdnd_producer_delegations`  | Lists producer delegations with filters       |
| `pdnd_client`                | Fetches a single client by ID                 |
| `pdnd_clients`               | Lists clients with optional filters           |
| `pdnd_client_keys`           | Lists keys on a client                        |

## Development

### Build

```sh
make build
```

### Run Unit Tests

```sh
make test
```

### Run Acceptance Tests

Acceptance tests run against an in-process fake PDND server and do not require network access.

```sh
make testacc
```

### Code Generation

The provider uses [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate the API client from the OpenAPI spec.

```sh
make generate
```

### Drift Check

Verify that generated code is up to date:

```sh
make drift-check
```

### Generate Documentation

```sh
make docs
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run `make test` and `make testacc` to ensure all tests pass
4. Submit a pull request
