# API Gateway Service

The API Gateway is the single public-facing entry point for the Million Hunter crypto dashboard platform. It handles authentication, rate limiting, request routing, and observability for all client requests.

## Features

- **JWT Authentication**: Secure token-based authentication with JWT validation
- **OpenTelemetry Tracing**: Distributed tracing for request observability
- **OpenAPI Validation**: Startup validation against OpenAPI specification
- **Per-Route Rate Limiting**: Configurable rate limits with Redis or in-memory backends
- **Prometheus Metrics**: RED metrics (Rate, Errors, Duration) for observability
- **Structured Logging**: JSON logging with request tracing
- **Health Checks**: Comprehensive liveness and readiness endpoints
- **CORS Support**: Configurable cross-origin resource sharing
- **Graceful Shutdown**: Proper signal handling and connection draining

## Quick Start

**Run locally:**
```bash
env PORT=8080 OPENAPI_PATH=../docs/openapi.yaml JWT_SECRET=devsecret \
    AUTH_SERVICE_URL=http://localhost:9000 FRONTEND_URL=http://localhost:3000 \
    go run ./cmd/api-gateway
```

**Available Commands:**
- `make validate-openapi` - Validate OpenAPI specification compliance
- `make build` - Build the binary
- `make test` - Run tests

## Configuration

All configuration via environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `JWT_SECRET` | JWT signing secret for validation | (required) |
| `REDIS_URL` | Redis connection URL | `localhost:6379` |
| `AUTH_SERVICE_URL` | Auth service base URL for login/refresh | (required) |
| `FRONTEND_URL` | CORS allowed origin (use * only for dev) | `*` |
| `OPENAPI_PATH` | Path to OpenAPI spec for validation | `../docs/openapi.yaml` |
| `STRICT_OPENAPI_VALIDATION` | Fail startup on validation errors | `false` |
| `RATE_DEFAULT_RPS` | Default rate limit (requests/second) | `10` |
| `RATE_DEFAULT_BURST` | Default burst capacity | `20` |
| `ROUTE_LIMITS` | Per-route rate limit overrides (JSON) | (optional) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | (optional) |

## Authentication

- **Token Validation**: Gateway validates JWTs (HS256) using `JWT_SECRET` and extracts `user_id` from `sub` claim
- **Proxying**: Login/refresh requests are proxied to `AUTH_SERVICE_URL`
- **No Token Issuance**: Gateway does NOT issue tokens (handled by auth-service)

## Rate Limiting

**Default limits** (apply to all routes):
```bash
RATE_DEFAULT_RPS=10      # 10 requests per second
RATE_DEFAULT_BURST=20    # Burst capacity of 20 requests
```

**Per-route overrides** (JSON format):
```bash
ROUTE_LIMITS='{"\/api\/v1\/portfolios":{"rps":5,"burst":10},"\/api\/v1\/wallets\/:address":{"rps":2,"burst":4}}'
```

Rate limit key includes route + user_id (from JWT) or client IP as fallback. Supports Redis (distributed) or in-memory backends.

## Health & Observability

**Health Endpoints:**
- `GET /healthz` - Liveness probe (checks Redis connectivity when REDIS_URL is set)
- `GET /metrics` - Prometheus metrics (RED metrics + rate-limit counters)

**Tracing:**
- OpenTelemetry tracing enabled (stdout by default)
- Set `OTEL_EXPORTER_OTLP_ENDPOINT` for production OTLP export

## API Endpoints

### Public (No Auth)
- `POST /api/v1/auth/login` - Login (proxied to auth-service)
- `POST /api/v1/auth/refresh` - Refresh token (proxied to auth-service)
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics

### Protected (Require JWT + Rate Limiting)
- `GET /api/v1/portfolios` - List tracked wallets
- `POST /api/v1/portfolios` - Add wallet to track
- `GET /api/v1/wallets/:address` - Get wallet overview
- `GET /api/v1/wallets/:address/transactions` - Get transaction history
- `GET /api/v1/tokens/:tokenAddress/holders` - Get top token holders
- `GET /api/v1/export/wallet/:address` - Export wallet data (CSV/JSON)

## Troubleshooting

**Service won't start:**
- Check `JWT_SECRET` and `AUTH_SERVICE_URL` are set
- Verify port 8080 is available: `lsof -i :8080`
- Review logs for configuration errors

**Authentication fails:**
- Verify JWT_SECRET matches auth-service secret
- Check token hasn't expired
- Ensure `Authorization: Bearer <token>` header format

**Rate limiting issues:**
- Check Redis connectivity: `redis-cli -h localhost -p 6379 ping`
- Review rate limit headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`
- Check metrics at `/metrics`

**High latency:**
- Monitor Prometheus metrics at `/metrics`
- Review OpenTelemetry traces
- Check auth-service response times

## Security

**CORS Configuration:**
```bash
FRONTEND_URL=https://app.million-hunter.com  # Production
FRONTEND_URL=*                               # Development only
```

**⚠️ Important**: Never use `FRONTEND_URL=*` in production.

**Production Checklist:**
1. Set cryptographically secure `JWT_SECRET`
2. Configure specific `FRONTEND_URL` (no wildcards)
3. Enable HTTPS via load balancer/reverse proxy
4. Adjust rate limits for production traffic
5. Configure `OTEL_EXPORTER_OTLP_ENDPOINT` for tracing
6. Use Redis for multi-instance rate limiting

---

**Part of the Million Hunter Crypto Dashboard microservices architecture.**
