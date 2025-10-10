# API Gateway Service

The API Gateway is the single public-facing entry point for the Million Hunter crypto dashboard platform. It handles authentication, rate limiting, request routing, and observability for all client requests.

## Features

- **JWT Authentication**: Secure token-based authentication with configurable expiry
- **Redis-based Rate Limiting**: Atomic token bucket implementation using Lua scripts (fixes race conditions)
- **OpenAPI Compliance**: Full implementation of the REST API specification
- **Prometheus Metrics**: RED metrics (Rate, Errors, Duration) for observability
- **Structured Logging**: Zero-allocation JSON logging with zerolog
- **Health Checks**: Comprehensive health, readiness, and liveness endpoints
- **CORS Support**: Configurable cross-origin resource sharing
- **Graceful Shutdown**: Proper signal handling and connection draining

## Architecture

The API Gateway follows the microservices architecture pattern and acts as:
- Authentication enforcer (JWT validation)
- Rate limit enforcer (Redis token bucket)
- Request router (to internal services via gRPC)
- Aggregation layer (combines responses from multiple services)
- Observability collector (metrics, logs, traces)

## Project Structure

```
api-gateway/
├── cmd/
│   └── api-gateway/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   ├── auth.go              # Authentication endpoints
│   │   ├── health.go            # Health check endpoints
│   │   ├── portfolio.go         # Portfolio endpoints
│   │   ├── wallet.go            # Wallet endpoints
│   │   ├── token.go             # Token endpoints
│   │   └── export.go            # Export endpoints
│   ├── middleware/
│   │   ├── auth.go              # JWT middleware
│   │   └── ratelimit.go         # Rate limiting middleware
│   ├── ratelimit/
│   │   └── redis_rate_limiter.go # Redis-based rate limiter with Lua script
│   ├── logging/
│   │   └── logger.go            # Structured logging
│   └── metrics/
│       └── metrics.go           # Prometheus metrics
├── tests/
│   └── k6/
│       ├── load-test.js         # Load testing scenarios
│       └── stress-test.js       # Stress testing scenarios
├── docker-compose.yml           # Local development setup
├── Dockerfile                   # Container image
├── Makefile                     # Build automation
└── README.md                    # This file
```

## Prerequisites

- Go 1.21 or higher
- Redis 7+ (for rate limiting and caching)
- Docker & Docker Compose (for containerized deployment)
- k6 (for load testing)

## Configuration

Configuration is managed via environment variables. See `.env.example` for all available options.

### Key Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `HTTP_PORT` | Server port | `8080` |
| `JWT_SECRET` | JWT signing secret | `dev-secret-change-in-production` |
| `JWT_EXPIRY_MINUTES` | Access token expiry | `15` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `RATE_LIMIT_PER_MINUTE` | Rate limit per user | `100` |
| `RATE_LIMIT_BURST` | Rate limit burst size | `200` |
| `LOG_LEVEL` | Logging level | `info` |
| `MVP_USERNAME` | MVP hardcoded username | `aezi` |
| `MVP_PASSWORD` | MVP hardcoded password | `Aa@123456789` |

## Getting Started

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Start Redis:**
   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   ```

3. **Run the service:**
   ```bash
   make run
   # or
   go run ./cmd/api-gateway/main.go
   ```

### Using Docker Compose

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## API Endpoints

### Health Endpoints (No Auth Required)

- `GET /health` - Detailed health status with dependency checks
- `GET /healthz` - Liveness probe (always returns 200 if service is running)
- `GET /ready` - Readiness probe (returns 200 if dependencies are healthy)
- `GET /metrics` - Prometheus metrics endpoint

### Authentication Endpoints (No Auth Required)

- `POST /api/v1/auth/login` - Login with email/password, returns JWT tokens
- `POST /api/v1/auth/register` - User registration (disabled in MVP)
- `POST /api/v1/auth/refresh` - Refresh access token (not yet implemented)

### Protected Endpoints (Require JWT + Rate Limiting)

All endpoints below require:
- `Authorization: Bearer <token>` header
- Subject to rate limiting (100 req/min default, 200 burst)

#### Portfolio Management
- `GET /api/v1/portfolios` - List tracked wallets
- `POST /api/v1/portfolios` - Add wallet to track

#### Wallet Analysis
- `GET /api/v1/wallets/{address}` - Get wallet overview and assets
- `GET /api/v1/wallets/{address}/transactions` - Get transaction history

#### Token Analysis
- `GET /api/v1/tokens/{tokenAddress}/holders` - Get top token holders

#### Data Export
- `GET /api/v1/export/wallet/{address}` - Export wallet data (CSV/JSON)

## Rate Limiting

The API Gateway implements Redis-based token bucket rate limiting with atomic operations via Lua scripts. This fixes the race condition identified in PR #5 comments on the ingestion service.

### Features
- **Atomic Operations**: Lua script ensures no race conditions
- **Per-User Limiting**: Authenticated users tracked by user ID, others by IP
- **Configurable Limits**: Adjust rate and burst capacity via environment
- **Standard Headers**: Returns `X-RateLimit-*` headers per RFC 6585

### Response Headers
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets
- `Retry-After`: Seconds to wait before retry (on 429 responses)

## Observability

### Metrics

Prometheus metrics are exposed at `/metrics`:

- `http_requests_total` - Total HTTP requests by method, path, status
- `http_request_duration_seconds` - Request duration histogram
- `http_request_size_bytes` - Request size histogram
- `http_response_size_bytes` - Response size histogram
- `rate_limit_exceeded_total` - Rate limit rejections
- `auth_attempts_total` - Authentication attempts by result
- `service_up` - Service health status (1 = up, 0 = down)
- `dependency_up` - Dependency health by service name

### Logging

Structured JSON logging with zerolog:
- Request/response logging with latency, status, method, path
- Error logging with stack traces
- Configurable log levels (debug, info, warn, error)

### Health Checks

- **Liveness**: Returns 200 if service is running
- **Readiness**: Checks Redis connectivity, returns 503 if dependencies unhealthy
- **Health**: Detailed status with latency for each dependency

## Testing

### Unit Tests

```bash
make test
make test-coverage
```

### Load Testing

Load tests use k6 to simulate realistic traffic patterns:

```bash
# Standard load test (10-100 concurrent users)
make test-load

# Stress test (100-300 concurrent users)
make test-stress

# Custom k6 test
k6 run -e BASE_URL=http://localhost:8080 tests/k6/load-test.js
```

### SLO Validation

The load tests validate:
- **p95 latency ≤ 300ms** (per monitoring-alerting.md requirements)
- **Error rate < 1%** (5xx responses)
- **Rate limiting works correctly** (429 responses under high load)

## Security

### MVP Authentication

For the MVP phase, the API Gateway uses hardcoded credentials:
- Username: `aezi`
- Password: `Aa@123456789`

The login endpoint validates these credentials and issues JWT tokens.

### Production Considerations

Before production deployment:
1. **Change JWT_SECRET**: Use a cryptographically secure random string
2. **Enable HTTPS**: Terminate TLS at load balancer or use reverse proxy
3. **Rotate Secrets**: Implement secret rotation for JWT signing keys
4. **Implement Refresh Token Flow**: Complete the token refresh endpoint
5. **Add Request Signing**: Consider HMAC signatures for sensitive operations
6. **Enable Auth Service Integration**: Replace hardcoded auth with auth-service gRPC calls

## Dependencies

### Internal Services (Future Integration)
- `auth-service` - User authentication and authorization
- `portfolio-service` - Wallet and portfolio data aggregation
- `market-data-service` - Token prices and market data

### External Dependencies
- **Redis 7+** - Rate limiting, caching, session storage
- **PostgreSQL 15+** (via services) - Persistent data storage
- **Kafka** (via services) - Event streaming for async operations

## Performance

### Benchmarks

Target performance metrics (per PRD and monitoring requirements):
- **Throughput**: ≥ 100 requests/second per instance
- **Latency (p95)**: ≤ 300ms for authenticated endpoints
- **Latency (p99)**: ≤ 500ms
- **Error Rate**: < 1% (5xx responses)
- **Uptime**: 99.9% availability

### Optimization

- Efficient connection pooling for Redis
- Gin framework for high-performance routing
- Zero-allocation logging with zerolog
- Atomic rate limiting to minimize Redis round-trips
- Minimal middleware chain overhead

## Troubleshooting

### Common Issues

**Service won't start:**
- Check Redis connectivity: `redis-cli -h localhost -p 6379 ping`
- Verify port 8080 is not in use: `lsof -i :8080`
- Check logs for configuration errors

**Authentication fails:**
- Verify JWT_SECRET is consistent across restarts
- Check token hasn't expired (default 15 minutes)
- Ensure credentials match MVP_USERNAME and MVP_PASSWORD

**Rate limiting not working:**
- Verify Redis is accessible
- Check Redis logs for Lua script errors
- Confirm RATE_LIMIT_PER_MINUTE and RATE_LIMIT_BURST are set correctly

**High latency:**
- Check Redis latency: `redis-cli --latency -h localhost -p 6379`
- Monitor Prometheus metrics at `/metrics`
- Review slow query logs if database-backed services are integrated

## Development

### Adding New Endpoints

1. Define handler in `internal/handlers/`
2. Register route in `cmd/api-gateway/main.go`
3. Add tests in `tests/k6/`
4. Update OpenAPI spec in project docs
5. Document in this README

### Contributing

1. Follow Go best practices and project conventions
2. Add tests for new functionality
3. Update documentation
4. Run linter: `make lint`
5. Ensure all tests pass: `make test`

## License

Copyright © 2025 Million Hunter Project

---

**Agent B Deliverable** - Part of the Million Hunter Crypto Dashboard microservices architecture.
