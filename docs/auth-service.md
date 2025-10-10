# auth-service

## Overview
- Go microservice providing JWT issuance and validation.
- For MVP, `/api/v1/auth/login` uses env-configured credentials and returns access/refresh tokens. Multi-user is feature-gated and enables DB-backed auth flows.

## Structure
- services/auth-service/
  - cmd/auth-service/main.go
  - internal/jwt/manager.go
  - internal/http/handlers.go
  - internal/http/middleware.go
  - internal/http/refresh_logout.go
  - internal/store/postgres.go
  - api/auth.proto
  - db/migrations/
- Uses pgxpool for PostgreSQL connection pooling


## HTTP Endpoints (MVP)
- POST `/api/v1/auth/login` — returns access_token, refresh_token, expires_in
- POST `/api/v1/auth/logout` — stateless logout (MVP); revokes refresh tokens in multi-user mode
- POST `/api/v1/auth/refresh` — 501 in MVP; enabled with rotation and persistence when `ENABLE_MULTI_USER=true`

## Middleware
- `WithAuth` JWT middleware validates Authorization: Bearer <token> with issuer/audience checks and injects claims into request context via `ClaimsFromContext`.

## Configuration (env)
- PORT, GRPC_PORT
- DATABASE_URL
- JWT_ISSUER, JWT_AUDIENCE
- MVP_USERNAME
- MVP_PASSWORD_BCRYPT

- JWT_SIGNING_KEY
- JWT_ACCESS_TTL_MINUTES, JWT_REFRESH_TTL_HOURS
- ENABLE_MULTI_USER

## Security
- JWT with HS256, short-lived access tokens; issuer/audience checks; unique JTI for rotation.
- Env-only secrets; default dev key must be overridden; MVP creds via env using bcrypt hash.
- Timing-attack mitigation on login: bcrypt check even when user not found.
- Structured JSON logs; avoid sensitive payloads in logs.

## Migrations
- golang-migrate style under services/auth-service/db/migrations
- 001_create_users_table
- 002_create_refresh_tokens_table
- 003_create_auth_audit_table

## Tests
- Unit tests for jwt, handlers, middleware, and config; coverage ≥ 90%.
- Run:
  - cd services/auth-service
  - go test ./... -race -cover -coverprofile=coverage.out
  - go tool cover -func=coverage.out
