# Authentication Guide

This guide explains how to obtain the credentials needed to configure the PDND Terraform provider.

## Overview

The PDND Interoperability API uses **DPoP (Demonstration of Proof-of-Possession)** authentication. Each API request requires:

1. A **DPoP access token** (voucher) issued by the PDND authorization server
2. A **DPoP proof JWT** generated per-request using your private key

The provider supports two authentication modes:

| Mode | When to Use | What You Need |
|------|-------------|---------------|
| **Auto-token** (recommended) | Normal usage | `client_id`, `purpose_id`, `dpop_private_key`, `dpop_key_id` |
| **Manual token** | Debugging or advanced use | `access_token`, `dpop_private_key`, `dpop_key_id` |

In **auto-token mode**, the provider automatically obtains and refreshes access tokens using your client credentials. You never need to deal with JWT assertions, curl commands, or token expiration.

---

## Quick Start (Recommended: Auto-Token Mode)

### Step 1: Get the API Base URL

| Environment | Base URL |
|-------------|----------|
| **Production** | `https://api.interop.pagopa.it/v3` |

Use the production URL unless PagoPA has provided you with a separate test/collaudo environment URL.

### Step 2: Create a Client in the PDND Portal

The PDND self-service portal is available at:

**https://selfcare.pagopa.it**

To create an API client:

1. Log in to the self-service portal with your organization's credentials (SPID or CIE)
2. Navigate to **PDND Interoperabilita**
3. Go to the **Client** section
4. Click **Crea nuovo client** (Create new client)
5. Fill in the client name and description
6. Associate the client with a **purpose** (finalita) -- this links the client to the e-service you want to consume

After creation, copy the **Client ID** (UUID) -- this is your `client_id`.

> For detailed instructions, see the [PDND Operational Manual -- Creating a Client](https://developer.pagopa.it/pdnd-interoperabilita/guides/manuale-operativo-pdnd-interoperabilita).

### Step 3: Get Your Purpose ID

Each purpose in the PDND portal has a unique **Purpose ID** (UUID). You can find it in the portal under the **Finalita** (Purposes) section of your organization.

Copy the purpose ID associated with the e-service you want to consume -- this is your `purpose_id`.

### Step 4: Generate Your Cryptographic Keys

PDND requires a key pair for DPoP authentication. The private key stays with you; the public key is uploaded to the portal.

#### Generate an RSA Key Pair

```bash
# Generate a 2048-bit RSA private key
openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048

# Extract the public key
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

**Key requirements:**
- Algorithm: **RSA** (2048 or 4096 bits)
- Format: **PEM** (PKCS#8 `BEGIN PRIVATE KEY` or PKCS#1 `BEGIN RSA PRIVATE KEY`)
- EC keys (P-256, P-384) are also supported by the provider but may not be accepted by all PDND environments

#### Upload the Public Key to PDND

1. In the PDND portal, go to your **Client** detail page
2. Navigate to the **Chiavi pubbliche** (Public Keys) section
3. Click **Aggiungi chiave pubblica** (Add public key)
4. Upload your `public_key.pem` file
5. After upload, PDND assigns a **Key ID (kid)** -- copy this value

The **Key ID** is the `dpop_key_id` parameter for the provider.

> For detailed instructions, see the [PDND Operational Manual -- Generating Cryptographic Keys](https://developer.pagopa.it/pdnd-interoperabilita/guides/manuale-operativo-pdnd-interoperabilita).

### Step 5: Configure the Provider

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  client_id        = var.pdnd_client_id
  purpose_id       = var.pdnd_purpose_id
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

That's it! The provider will automatically obtain and refresh tokens as needed. No manual token management required.

---

## Alternative: Manual Token Mode

For debugging or advanced use cases, you can supply a pre-obtained access token directly. This bypasses automatic token generation.

### How to Manually Obtain a Voucher

#### Step 1: Create a Client Assertion JWT

Build a JWT signed with your private key:

**Header:**
```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "<your-key-id>"
}
```

**Payload:**
```json
{
  "iss": "<your-client-id>",
  "sub": "<your-client-id>",
  "aud": "auth.interop.pagopa.it/client-assertion",
  "jti": "<unique-uuid>",
  "iat": <current-unix-timestamp>,
  "exp": <expiration-unix-timestamp>,
  "purposeId": "<your-purpose-id>"
}
```

#### Step 2: Request the Token

```bash
curl -X POST https://auth.interop.pagopa.it/token.oauth2 \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=<your-client-id>" \
  -d "client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer" \
  -d "client_assertion=<signed-jwt>"
```

The response contains your access token:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "DPoP",
  "expires_in": 600
}
```

#### Step 3: Configure the Provider

```hcl
provider "pdnd" {
  base_url         = "https://api.interop.pagopa.it/v3"
  access_token     = var.pdnd_access_token
  dpop_private_key = file(var.dpop_private_key_path)
  dpop_key_id      = var.dpop_key_id
}
```

> **Note:** Vouchers have a limited lifetime (typically 10 minutes). For long-running Terraform operations, you may need to refresh the token. Consider using auto-token mode instead.

---

## Security Best Practices

- **Never commit your private key** to version control
- **Never hardcode tokens or secrets** in `.tf` files -- use variables or a secrets manager
- **Rotate keys periodically** through the PDND portal
- **Prefer auto-token mode** -- it handles token refresh automatically and avoids storing long-lived tokens
- **Store your private key securely** -- consider using a hardware security module (HSM) or a vault service

---

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| `401 Unauthorized` | Expired or invalid access token | Use auto-token mode, or request a fresh voucher |
| `403 Forbidden` | Client not authorized for this operation | Check purpose/client associations in the portal |
| `DPoP proof invalid` | Wrong private key or key ID | Verify the key ID matches the uploaded public key |
| `PEM parse error` | Invalid key format | Ensure the key is PEM-encoded RSA (PKCS#8 or PKCS#1) |
| `token request failed` | Invalid client_id, purpose_id, or key | Check your credentials in the PDND portal |

---

## Further Reading

- [PDND Operational Manual](https://developer.pagopa.it/pdnd-interoperabilita/guides/manuale-operativo-pdnd-interoperabilita)
- [PDND API Reference](https://developer.pagopa.it/pdnd-interoperabilita/api)
- [PagoPA Self-Service Portal](https://selfcare.pagopa.it)
- [RFC 9449 -- DPoP (Demonstration of Proof-of-Possession)](https://datatracker.ietf.org/doc/html/rfc9449)
