# auth-service

## Overview
- Go microservice providing JWT issuance and validation.
- For MVP, `/api/v1/auth/login` accepts hardcoded credentials (username: aezi, password: Aa@123456789) and returns access/refresh tokens. Multi-user is feature-gated.

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

## HTTP Endpoints (MVP)
- POST `/api/v1/auth/login` — returns access_token, refresh_token, expires_in (hardcoded single-user)
- POST `/api/v1/auth/logout` — stateless logout, returns ok
- POST `/api/v1/auth/refresh` — returns 501 Not Implemented for MVP (enabled post multi-user)

## Middleware
- `WithAuth` JWT middleware validates Authorization: Bearer <token> with issuer/audience checks and injects claims into request context via `ClaimsFromContext`.

## Configuration (env)
- PORT, GRPC_PORT
- DATABASE_URL
- JWT_ISSUER, JWT_AUDIENCE
- JWT_SIGNING_KEY
- JWT_ACCESS_TTL_MINUTES, JWT_REFRESH_TTL_HOURS
- ENABLE_MULTI_USER

## Security
- JWT with HS256, short-lived access tokens; issuer/audience checks.
- Env-only secrets; default dev key must be overridden.
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
