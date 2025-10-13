# Security Hardening - Million Dollar Hunter

**Last Updated**: 2025-10-11  
**Owner**: Agent A (Authentication & Security Service)  
**Status**: Phase 1 - Initial Implementation

---

## Overview

This document outlines the security hardening measures implemented across the Million Dollar Hunter platform, with a focus on the authentication service (Agent A) and API Gateway (Agent B).

---

## Authentication & Authorization

### JWT Token Management

**Implementation**: `services/auth-service/internal/jwt/manager.go`, `keystore.go`

**Security Measures**:
- **Algorithm**: RS256 (RSA asymmetric signing) with HS256 legacy support
- **Key Management**: 
  - Multi-key support with Key ID (kid) in JWT header
  - File-backed keystore with private key encryption
  - Automated key rotation with zero downtime
  - 24-hour grace period for validation
  - 7-day cleanup period for expired keys
- **Token Expiration**: 
  - Access tokens: 15 minutes (configurable via `ACCESS_TTL`)
  - Refresh tokens: 7 days (configurable via `REFRESH_TTL`)
- **Claims Validation**: All tokens include issuer (`iss`), audience (`aud`), subject (`sub`), expiration (`exp`), and JWT ID (`jti`)
- **Audience Verification**: API Gateway validates expected audience to prevent token reuse across services
- **JWKS Endpoint**: `/.well-known/jwks.json` serves public keys for RS256 validation

**Configuration**:
```bash
# Legacy mode (HS256)
JWT_SIGNING_KEY=<strong-secret-key>  # REQUIRED for legacy mode

# Production mode (RS256 with keystore)
KEYSTORE_PATH=./config/keystore.json
JWT_ISSUER=auth-service
JWT_AUDIENCE=api-gateway
ACCESS_TTL=15m
REFRESH_TTL=168h
```

**Key Rotation**:
```bash
# Generate new 2048-bit key (90-day expiry)
./bin/keytool -keystore ./config/keystore.json -generate -bits 2048 -expires 2160h

# Activate new key
./bin/keytool -keystore ./config/keystore.json -activate <kid>

# Clean up expired keys
./bin/keytool -keystore ./config/keystore.json -cleanup
```

**Operational Runbook**: See `/docs/JWT-ROTATION-RUNBOOK.md` for detailed procedures

**Status**:
- [x] RS256 asymmetric signing implemented
- [x] Multi-key support with kid header
- [x] JWKS endpoint for public key distribution
- [x] Key rotation tooling (keytool CLI)
- [x] Operational runbooks documented
- [x] Backward compatibility with HS256

### Password Policy

**Implementation**: `services/auth-service/internal/auth/password.go`

**Requirements**:
- Minimum length: 12 characters
- Must contain:
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one digit
  - At least one special character from: `!@#$%^&*()-_=+[]{}|;:,.<>?`

**Hashing**: bcrypt with cost factor 10 (configurable via `BCRYPT_COST`)

**Hardening Recommendations**:
- [x] bcrypt cost factor 10 (current)
- [ ] Increase cost factor to 12 for production
- [ ] Implement password breach database check (HaveIBeenPwned API)
- [ ] Add password history (prevent reuse of last 5 passwords)

### Login Lockout Protection

**Implementation**: `services/auth-service/internal/http/login.go`

**Current Configuration**:
- **Threshold**: 3 failed attempts within 15-minute window
- **Response**: HTTP 429 Too Many Requests
- **Storage**: In-memory lockout tracking (single instance)

**Hardening Recommendations**:
- [x] Login lockout implemented (current)
- [ ] Move lockout tracking to Redis for distributed systems
- [ ] Add exponential backoff (5 min → 15 min → 1 hour)
- [ ] Implement CAPTCHA after 2 failed attempts
- [ ] Add email notification on lockout

### Refresh Token Rotation

**Implementation**: `services/auth-service/internal/http/refresh_logout.go`

**Security Measures**:
- **One-Time Use**: Each refresh operation revokes old token and issues new pair
- **Persistence**: Refresh tokens stored in PostgreSQL with user association
- **Revocation**: All user tokens revoked on logout
- **Audit Trail**: All refresh operations logged to `auth_audit` table

**Database Schema**: `services/auth-service/db/migrations/002_refresh_tokens.up.sql`

**Hardening Recommendations**:
- [x] Refresh token rotation (current)
- [ ] Add refresh token family tracking (detect reuse attacks)
- [ ] Implement refresh token TTL sliding window
- [ ] Add device fingerprinting for refresh tokens

---

## API Gateway Security

### Rate Limiting

**Implementation**: `api-gateway/internal/ratelimit/`

**Configuration**:
- **Algorithm**: Token bucket
- **Default**: 10 RPS with burst of 20
- **Storage**: Redis-backed (distributed) with local fallback
- **Scope**: Per-route + per-user/IP

**Route Overrides** (via `ROUTE_LIMITS` environment variable):
```json
{
  "/api/v1/portfolios": {"rps": 5, "burst": 10},
  "/api/v1/wallets/:address/transactions": {"rps": 3, "burst": 5}
}
```

**Hardening Recommendations**:
- [x] Redis-backed rate limiting (current)
- [ ] Add geolocation-based rate limits
- [ ] Implement adaptive rate limiting based on load
- [ ] Add rate limit exemptions for premium users

### CORS Configuration

**Implementation**: `api-gateway/internal/server/router.go`

**Current Settings**:
- **Allowed Origins**: Configured via `FRONTEND_URL` (single origin)
- **Allowed Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Allowed Headers**: Authorization, Content-Type, X-Request-ID
- **Credentials**: Enabled
- **Max Age**: 12 hours

**Hardening Recommendations**:
- [x] Strict origin validation (current)
- [ ] Add support for multiple allowed origins (comma-separated)
- [ ] Implement origin validation against database for dynamic clients
- [ ] Add Content-Security-Policy headers

### Request Tracing

**Implementation**: `api-gateway/internal/middleware/logging.go`, `api-gateway/internal/middleware/tracing.go`

**Security Benefits**:
- **Trace ID**: Unique identifier for request correlation across services
- **Audit Trail**: All requests logged with method, path, status, latency, client IP
- **Error Tracking**: Failed authentication attempts logged with trace ID

**Hardening Recommendations**:
- [x] Request ID generation (current)
- [x] Request/response logging (current)
- [ ] Add user ID to all log entries (post-authentication)
- [ ] Implement log aggregation (ELK/Loki)
- [ ] Add anomaly detection on log patterns

---

## Database Security

### PostgreSQL Configuration

**Services**:
- Auth Service: `postgres://localhost:5432/auth`
- Portfolio Service: `postgres://localhost:5433/portfolio`
- Ingestion Service: `postgres://localhost:5434/ingestion`

**Connection Pooling**: `services/auth-service/internal/store/postgres.go`
```go
MaxConns:          20
MinConns:          5
MaxConnLifetime:   1 hour
MaxConnIdleTime:   30 minutes
HealthCheckPeriod: 1 minute
```

**Hardening Recommendations**:
- [ ] Enable SSL/TLS for database connections (`sslmode=require`)
- [ ] Use separate read-only database user for queries
- [ ] Implement database audit logging (pg_audit extension)
- [ ] Add prepared statement caching
- [ ] Enable connection encryption at rest

### Secrets Management

**Current Approach**: Environment variables loaded at runtime

**Hardening Recommendations**:
- [ ] Migrate to HashiCorp Vault or AWS Secrets Manager
- [ ] Implement secret rotation automation
- [ ] Add secret encryption at rest
- [ ] Use service accounts with IAM roles (cloud deployments)

---

## Static Application Security Testing (SAST)

### Tools Configured

**1. Gosec** (`github.com/securego/gosec/v2`)

**CI Integration**: `.github/workflows/auth-service-ci.yml`

**Checks**:
- SQL injection vulnerabilities
- Hardcoded credentials
- Weak crypto algorithms
- File path traversal
- Command injection
- Buffer overflows

**Configuration**: Non-blocking (informational only)

**2. govulncheck** (`golang.org/x/vuln/cmd/govulncheck`)

**CI Integration**: `.github/workflows/auth-service-ci.yml`

**Checks**:
- Known vulnerabilities in Go standard library
- Vulnerable dependencies in go.mod
- CVE database lookup

**Configuration**: Non-blocking (TODO: upgrade Go 1.21 to 1.22+ to fix stdlib vulns)

### SAST Workflow

```bash
# Run locally
cd services/auth-service
gosec -fmt=text ./...
govulncheck ./...
```

**Hardening Recommendations**:
- [ ] Make SAST checks blocking in CI for high-severity findings
- [ ] Add Snyk or Dependabot for automated dependency updates
- [ ] Implement SonarQube for code quality + security
- [ ] Add OWASP Dependency-Check for transitive vulnerabilities

---

## Deployment Security

### Environment Separation

**Environments**:
- Development: Local Docker Compose
- Staging: TBD
- Production: TBD

**Hardening Recommendations**:
- [ ] Separate database instances per environment
- [ ] Use different JWT signing keys per environment
- [ ] Implement network segmentation (VPC/subnets)
- [ ] Add WAF (Web Application Firewall) for production

### TLS/SSL

**Current Status**: HTTP only (local development)

**Hardening Recommendations**:
- [ ] Enable TLS 1.3 for all external endpoints
- [ ] Use Let's Encrypt for certificate management
- [ ] Implement certificate pinning for gRPC services
- [ ] Add HSTS (HTTP Strict Transport Security) headers

### Container Security

**Current Status**: Docker images not yet hardened

**Hardening Recommendations**:
- [ ] Use distroless base images (e.g., `gcr.io/distroless/static`)
- [ ] Run containers as non-root user
- [ ] Add Trivy/Clair image scanning in CI
- [ ] Implement image signing with Cosign/Notary
- [ ] Use multi-stage builds to minimize attack surface

---

## Monitoring & Incident Response

### Prometheus Metrics

**Implementation**: `api-gateway/internal/observability/metrics.go`

**Security Metrics**:
- `auth_grpc_validation_total{outcome="success|invalid|error"}` - Authentication attempts
- `rate_limit_blocked_total{route}` - Rate limit violations
- `http_requests_total{method,route,status}` - All API requests

**Hardening Recommendations**:
- [ ] Add alerting for authentication failure spikes (> 100/min)
- [ ] Monitor rate limit violations by IP
- [ ] Add metrics for JWT expiration times
- [ ] Implement error budget tracking

### Audit Logging

**Implementation**: `services/auth-service/db/migrations/003_audit_logs.up.sql`

**Events Logged**:
- User login (success/failure)
- Token refresh (with old/new token IDs)
- Logout (token revocation)
- Login lockout triggers

**Hardening Recommendations**:
- [ ] Add audit logs for sensitive operations (password reset, email change)
- [ ] Implement tamper-proof logging (append-only, signed)
- [ ] Export audit logs to SIEM (Splunk/ELK)
- [ ] Add retention policy (7 years for compliance)

---

## Compliance & Standards

### OWASP Top 10 Coverage

| Risk | Status | Implementation |
|------|--------|----------------|
| A01:2021 - Broken Access Control | ✅ Mitigated | JWT validation, rate limiting |
| A02:2021 - Cryptographic Failures | ⚠️ Partial | bcrypt hashing, HS256 JWTs (upgrade to RS256) |
| A03:2021 - Injection | ✅ Mitigated | Parameterized SQL queries, input validation |
| A04:2021 - Insecure Design | ⚠️ Partial | CQRS pattern, audit logging (add threat modeling) |
| A05:2021 - Security Misconfiguration | ⚠️ Partial | CORS, rate limiting (add CSP, HSTS) |
| A06:2021 - Vulnerable Components | ⚠️ Partial | govulncheck in CI (add dependency scanning) |
| A07:2021 - Authentication Failures | ✅ Mitigated | JWT, password policy, login lockout |
| A08:2021 - Software/Data Integrity | ❌ Not Implemented | Add code signing, SBOM |
| A09:2021 - Security Logging Failures | ⚠️ Partial | Audit logs, request tracing (add SIEM) |
| A10:2021 - Server-Side Request Forgery | ✅ Mitigated | No user-controlled URLs in backend |

---

## Security Testing Checklist

### Penetration Testing

- [ ] Automated fuzzing with go-fuzz
- [ ] OWASP ZAP automated scan
- [ ] Manual penetration testing by security team
- [ ] Bug bounty program (post-production)

### Integration Tests

- [x] JWT validation tests (current)
- [x] Password policy tests (current)
- [x] Login lockout tests (current)
- [ ] Refresh token rotation tests (Testcontainers)
- [ ] Rate limiting tests (Redis-backed)
- [ ] CORS violation tests

---

## Incident Response Plan

### Security Incident Severity Levels

**Critical**: JWT signing key compromise, database breach
- Response Time: Immediate (< 15 minutes)
- Actions: Rotate keys, revoke all tokens, notify users

**High**: Authentication bypass, rate limit bypass
- Response Time: < 1 hour
- Actions: Deploy hotfix, audit logs, incident report

**Medium**: SAST finding, dependency vulnerability
- Response Time: < 24 hours
- Actions: Update dependency, regression testing

**Low**: Configuration drift, informational finding
- Response Time: < 7 days
- Actions: Update documentation, backlog ticket

### Contact Information

- Security Lead: [TBD]
- On-Call Engineer: [TBD]
- Incident Email: security@million-dollar-hunter.com (TBD)

---

## References

- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [Go Secure Coding Practices](https://golang.org/doc/security/)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)

---

*This document is a living artifact. Update after implementing security enhancements or discovering new threats.*
