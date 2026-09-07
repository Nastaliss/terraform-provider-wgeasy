# Terraform Provider for wg-easy

A Terraform/OpenTofu provider for managing WireGuard clients via the [wg-easy](https://github.com/wg-easy/wg-easy) API.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 or [OpenTofu](https://opentofu.org/) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- A running [wg-easy](https://github.com/wg-easy/wg-easy) instance

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    wgeasy = {
      source = "Nastaliss/wgeasy"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/Nastaliss/terraform-provider-wgeasy.git
cd terraform-provider-wgeasy
make install
```

## Provider Configuration

```hcl
provider "wgeasy" {
  endpoint = "http://localhost:51821"
  username = "admin"
  password = "secret"

  # Only if the account has 2FA enabled (wg-easy >= 15.4.0). See the
  # "Two-factor authentication" note below before using this - it weakens 2FA.
  # totp_secret = "JBSWY3DPEHPK3PXP"
}
```

### Environment Variables

All provider arguments can be set via environment variables:

| Argument      | Environment Variable  |
|---------------|----------------------|
| `endpoint`    | `WGEASY_ENDPOINT`    |
| `username`    | `WGEASY_USERNAME`    |
| `password`    | `WGEASY_PASSWORD`    |
| `totp_secret` | `WGEASY_TOTP_SECRET` |

## Authentication notes (wg-easy ≥ 15.4.0)

wg-easy 15.4.0 rewrote its API (Nuxt v4). Password login moved from
`POST /api/session` to `POST /api/auth/password`, with an optional 2FA step at
`POST /api/auth/verify-2fa`. This provider targets the new endpoints and falls
back to the legacy `/api/session` for wg-easy < 15.4.0, so it works against both.

### There is no service account / API token

wg-easy has **no API-token or service-account mechanism**, and it is effectively
**single-admin**:

- The first user created during setup is the only `ADMIN`; any subsequent
  password-created user is a `CLIENT`, and there is no UI or API to create
  additional users. (OAuth auto-registration can mint more admins, but OAuth has
  no non-interactive flow usable by Terraform.)

As a result, Terraform must authenticate as the **same admin account** a human
uses. There is no way to isolate automation behind a dedicated credential.

**Recommended setup:** protect the wg-easy API at the network layer (restrict it
to the CI runner's IP / an internal network / behind a VPN) and treat that
network boundary as the real second factor for automation. Store `password` in
your CI secret store, not in `.tf` files.

### Two-factor authentication (`totp_secret`)

If the admin account has 2FA enabled, set `totp_secret` to the base32 seed
generated during 2FA setup. The provider derives the current TOTP code itself and
completes the login; without it, a 2FA-protected account cannot authenticate.

> ⚠️ **This weakens 2FA.** Because there is only one admin account (see above),
> the TOTP secret has to live next to the password in the same CI secret store.
> Both "factors" then share one location, so anyone who compromises that store
> has both — this is effectively single-factor for the automation path. Only use
> `totp_secret` when an organizational policy forces 2FA on every account. When
> you control the account, prefer leaving 2FA off on the admin used by Terraform
> and relying on the network restriction above.

## Resources

### wgeasy_client

Manages a WireGuard client in wg-easy.

```hcl
# Minimal example - uses server defaults
resource "wgeasy_client" "example" {
  name = "my-laptop"
}

# Full example with custom settings
resource "wgeasy_client" "custom" {
  name = "custom-client"

  allowed_ips        = ["0.0.0.0/0"]
  server_allowed_ips = ["10.8.0.0/24"]
  dns                = ["1.1.1.1", "8.8.8.8"]
  mtu                = 1420
  persistent_keepalive = 25
  enabled            = true
}
```

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Client name |
| `enabled` | bool | No | Whether the client is enabled (default: true) |
| `expires_at` | string | No | Expiration date (ISO 8601 format) |
| `allowed_ips` | list(string) | No | Client-side allowed IPs |
| `server_allowed_ips` | list(string) | No | Server-side allowed IPs |
| `dns` | list(string) | No | DNS servers |
| `mtu` | number | No | MTU value |
| `persistent_keepalive` | number | No | Keepalive interval in seconds |
| `server_endpoint` | string | No | Custom server endpoint |
| `pre_up` | string | No | Pre-up script |
| `post_up` | string | No | Post-up script |
| `pre_down` | string | No | Pre-down script |
| `post_down` | string | No | Post-down script |

#### Attributes (Read-Only)

| Name | Type | Description |
|------|------|-------------|
| `id` | string | Client ID |
| `ipv4_address` | string | Assigned IPv4 address |
| `ipv6_address` | string | Assigned IPv6 address |
| `public_key` | string | WireGuard public key |
| `private_key` | string | WireGuard private key (sensitive) |
| `preshared_key` | string | WireGuard preshared key (sensitive) |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Last update timestamp |

#### Import

Clients can be imported using their numeric ID:

```bash
terraform import wgeasy_client.example 1
```

## Data Sources

### wgeasy_client

Fetch a single client by ID.

```hcl
data "wgeasy_client" "example" {
  id = 1
}

output "client_name" {
  value = data.wgeasy_client.example.name
}
```

### wgeasy_clients

Fetch all clients.

```hcl
data "wgeasy_clients" "all" {}

output "client_names" {
  value = [for c in data.wgeasy_clients.all.clients : c.name]
}
```

## License

MIT
