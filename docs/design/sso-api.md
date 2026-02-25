# SSO OIDC API Overview

> Reference: [Design doc](../plans/2025-02-25-sso-oidc-design.md)

## Endpoints

### OIDC Discovery

**GET** `/.well-known/openid-configuration`

Returns the OIDC discovery document with issuer, authorize, token, userinfo, and JWKS URLs.

### OAuth2/OIDC Flow

| Endpoint  | Method | Purpose                                      |
|-----------|--------|----------------------------------------------|
| /authorize| GET    | Initiate auth code flow; redirects to /login if unauthenticated |
| /token    | POST   | Exchange authorization code or refresh_token for access_token, id_token |
| /userinfo | GET    | Return user claims (Authorization: Bearer &lt;access_token&gt;) |

### Authentication UI

| Endpoint        | Method | Purpose                       |
|-----------------|--------|-------------------------------|
| /login          | GET    | Login page (HTML)             |
| /login          | POST   | Submit credentials            |
| /register       | GET    | Registration page             |
| /register       | POST   | Create account                |
| /account/delete | GET    | Account deletion confirmation (requires login) |
| /account/delete | POST   | Delete account (requires login, confirm with "yes") |

### Federation (Upstream IdP)

| Endpoint                        | Method | Purpose                           |
|---------------------------------|--------|-----------------------------------|
| /auth/federation/:connector_id  | GET    | Redirect to upstream IdP          |
| /auth/callback/:connector_id   | GET    | OAuth callback; create session   |

### Health

| Endpoint | Method | Purpose              |
|----------|--------|----------------------|
| /healthz | GET    | Liveness (no deps)   |
| /ready   | GET    | Readiness (incl. DB) |

## Error Response

```json
{
  "code": "string",
  "message": "string",
  "request_id": "string (optional)"
}
```

## Design Reference

See [docs/plans/2025-02-25-sso-oidc-design.md](../plans/2025-02-25-sso-oidc-design.md) for architecture, data flow, and implementation details.
