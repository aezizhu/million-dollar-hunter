# auth-service

## Overview
- Go microservice providing JWT issuance and validation.
- For MVP, `/api/v1/auth/login` accepts hardcoded credential (username: aezi, password: Aa@123456789) and returns access/refresh tokens. Multi-user is feature-gated.

## Structure
- services/auth-service/
  - cmd/auth-service/main.go
  - internal/jwt/manager.go
  - internal/http/handlers.go
  - internal/store/postgres.go
  - api/auth.proto
  - db/migrations/

## Configuration (env)
- PORT, GRPC_PORT
- DATABASE_URL
- JWT_ISSUER, JWT_AUDIENCE
- JWT_SIGNING_KEY
- JWT_ACCESS_TTL_MINUTES, JWT_REFRESH_TTL_HOURS
- ENABLE_MULTI_USER

## Security
- JWT with HS256, short-lived access tokens; issuer/audience checks.
- No secrets in repo; use env only.
- Structured JSON logs; avoid sensitive payloads in logs.

## Migrations
- golang-migrate style under services/auth-service/db/migrations
- users table: id (UUID), email (unique), password_hash, timestamps

## Tests
- Unit tests for jwt and handlers.
- Run:
  - cd services/auth-service
  - go test ./... -race -cover -coverprofile=coverage.out
  - go tool cover -func=coverage.out
