# Superpowers Demo — SSO OIDC

A standards-based SSO (Single Sign-On) system built on OIDC (OpenID Connect). Monolithic architecture that acts as both IdP (Identity Provider) and RP (Relying Party). Supports local users (database) and federation with upstream IdP.

**Tech stack:** Gin, ent, ory/fosite, go-oidc, zap, Viper, Cobra

## Run

```bash
go run . svr
```

Server starts on `http://localhost:8888` by default.

## Config

Config file: `configs/settings.yaml` (relative to working directory).

| Section   | Key     | Default              | Description                          |
|-----------|---------|----------------------|--------------------------------------|
| server    | port    | 8888                 | HTTP listen port                     |
| database  | driver  | sqlite3              | DB driver                            |
| database  | dsn     | file:./data/sso.db...| Connection string (SQLite path)      |
| log       | level   | info                 | Log level (debug/info/warn/error)    |
| log       | output  | stdout               | Log output (stdout or file path)     |
| oidc      | issuer  | http://localhost:8888| OIDC issuer URL (must match base URL)|

## OIDC Endpoints

| Method | Path                              | Description                          |
|--------|-----------------------------------|--------------------------------------|
| GET    | `/.well-known/openid-configuration` | OIDC discovery document              |
| GET    | `/authorize`                      | Authorization request (OAuth2 auth code) |
| POST   | `/token`                         | Token exchange (code or refresh_token) |
| GET    | `/userinfo`                      | User claims (Bearer token required) |
| GET    | `/login`                         | Login page (HTML)                    |
| POST   | `/login`                         | Login form submission                |
| GET    | `/register`                      | Registration page (HTML)             |
| POST   | `/register`                     | Registration form submission         |

### Dev OAuth2 Client

Seeded for development: `client_id=sso-demo`, `client_secret=secret`, `redirect_uri=http://localhost:3000/callback`.

## Links

- [Design doc](docs/plans/2025-02-25-sso-oidc-design.md)
- [API overview](docs/design/sso-api.md)
