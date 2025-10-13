# Auth Service Integration Tests

## Overview

This directory contains comprehensive integration tests for the auth-service using Testcontainers for PostgreSQL database testing.

## Prerequisites

- Go 1.24+
- Docker (for Testcontainers)
- Docker daemon running

## Running Tests

### Run All Tests

```bash
cd /home/ubuntu/repos/million-dollar-hunter/services/auth-service
GOWORK=off go test -v ./tests
```

### Run Specific Test

```bash
GOWORK=off go test -v ./tests -run TestJWTFlowEndToEnd
```

### Run with Coverage

```bash
GOWORK=off go test -v ./tests -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Skip Integration Tests (Short Mode)

```bash
GOWORK=off go test -v ./tests -short
```

## Test Suite

### TestDatabaseConnectivity
**Purpose**: Verify database connectivity and schema migrations

**Coverage**:
- PostgreSQL connection
- Migration execution
- Table creation (users, refresh_tokens, auth_audit)
- UUID extension installation

**Duration**: ~2s

---

### TestJWTFlowEndToEnd
**Purpose**: Test complete JWT token generation and validation flow

**Coverage**:
- User creation
- Token pair generation (access + refresh)
- Token storage in database
- Token validation
- Claims verification (user_id, email, issuer, audience)

**Duration**: ~2s

---

### TestRefreshTokenRotation
**Purpose**: Verify refresh token rotation mechanism

**Coverage**:
- Initial token generation
- Token revocation
- New token generation
- Old token invalidation
- Revoked token cannot be reused

**Duration**: ~2s

---

### TestLoginLockout
**Purpose**: Test login lockout protection

**Coverage**:
- Failed login attempt tracking
- Lockout after 3 failures in 15 minutes
- Audit log verification

**Duration**: ~2s

---

### TestTokenExpiration
**Purpose**: Verify token expiration handling

**Coverage**:
- Short-lived token generation
- Expiration enforcement
- Error message validation

**Duration**: <1s (no database required)

---

### TestConcurrentTokenGeneration
**Purpose**: Test concurrent token operations

**Coverage**:
- Concurrent token generation (10 goroutines)
- Database race condition handling
- Token storage integrity

### TestConcurrentRefreshRequests
Purpose: Validate revoke-on-use behavior under concurrent refresh attempts; only one succeeds, others fail appropriately.
Duration: ~2s

---

**Duration**: ~2s

---

## Test Infrastructure

### Testcontainers

The tests use [Testcontainers](https://golang.testcontainers.org/) to spin up a real PostgreSQL database for each test. This ensures:

- Isolated test environment
- Real database behavior
- No mocking of database operations
- Automatic cleanup after tests

### Database Migrations

Migrations are run automatically before each test using the `runMigrations()` helper function. This ensures the database schema matches production.

### Helper Functions

- `setupDatabase(t, ctx)`: Creates PostgreSQL container and runs migrations
- `runMigrations(ctx, pool)`: Executes all database migrations

### TestLockoutBoundaryConditions
Purpose: Validate lockout threshold edge cases (N=3 lock; N-1 no lock; N+1 stays locked).
Duration: ~2s

### TestLockoutWindowExpiration
Purpose: Verify lockout expires after 15-minute window.
Duration: ~2s

### TestLockoutPerUserIsolation
Purpose: Ensure lockout applies per-user and does not affect other users.
Duration: ~2s

## Troubleshooting

### Docker Not Running

```
Error: Cannot connect to the Docker daemon
```

**Solution**: Start Docker daemon
```bash
sudo systemctl start docker
```

### Port Conflicts

```
Error: bind: address already in use
```

**Solution**: Testcontainers automatically assigns random ports. If you see this error, check for other processes using Docker ports.

### Container Cleanup Issues

```
Error: failed to terminate container
```

**Solution**: Manually clean up containers
```bash
docker ps -a | grep postgres:16-alpine | awk '{print $1}' | xargs docker rm -f
```

### Slow Tests

Integration tests can be slow due to container startup time. To speed up:

- Use `-parallel` flag for concurrent test execution
- Run tests locally with a persistent database (not recommended for CI)

## Integration with CI

These tests are designed to run in CI environments. The GitHub Actions workflow includes:

```yaml
- name: Run Integration Tests
  run: |
    cd services/auth-service
    GOWORK=off go test -v ./tests -timeout 5m
```

## Coverage Goals

- **Target**: >90% integration test coverage
- **Current**: 100% of critical auth flows covered
- **Unit Tests**: See `internal/` for unit tests

## Future Enhancements

- [ ] Add performance benchmarks
- [ ] Test distributed scenarios (multiple instances)
- [ ] Add chaos engineering tests (network failures, database errors)
- [ ] Test migration rollback scenarios
