# Terraform Provider for PDND Interoperability

> **WARNING: This project is extremely early stage and experimental. Use it at your own risk. APIs, resource schemas, and behaviors may change without notice. This provider is not yet suitable for production use.**

A Terraform provider for managing resources on the [PDND Interoperability](https://developers.interop.pagopa.it/) platform (API v3).

> **Milestone 2**: This provider supports agreement and e-service management.

## Prerequisites

- [Go](https://golang.org/dl/) 1.22+
- [Terraform](https://www.terraform.io/downloads.html) 1.0+

## Authentication

The provider authenticates using DPoP (Demonstration of Proof-of-Possession). You need:

- An **access token** for the PDND API
- A **PEM-encoded RSA private key** for DPoP proof generation
- A **key ID** identifying your DPoP key

## Provider Configuration

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  access_token     = var.pdnd_access_token
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

| Attribute          | Type   | Required | Description                                |
|--------------------|--------|----------|--------------------------------------------|
| `base_url`         | string | yes      | Base URL of the PDND API                   |
| `access_token`     | string | yes      | Access token for API authentication        |
| `dpop_private_key` | string | yes      | PEM-encoded private key for DPoP           |
| `dpop_key_id`      | string | yes      | Key ID for the DPoP private key            |
| `request_timeout_s`| number | no       | Request timeout in seconds (default: 30)   |

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

## Available Data Sources

| Data Source                  | Description                                      |
|------------------------------|--------------------------------------------------|
| `pdnd_agreement`             | Fetches a single agreement by ID                 |
| `pdnd_agreements`            | Lists agreements with optional filters           |
| `pdnd_agreement_purposes`    | Lists purposes associated with an agreement      |
| `pdnd_eservice`              | Fetches a single e-service by ID                 |
| `pdnd_eservices`             | Lists e-services with optional filters           |
| `pdnd_eservice_descriptor`   | Fetches a single descriptor by ID                |
| `pdnd_eservice_descriptors`  | Lists descriptors for an e-service               |

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
