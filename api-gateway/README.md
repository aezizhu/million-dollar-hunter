JWT Validation Modes

- AUTH_VALIDATE_MODE=local (default): JWT validated in gateway using JWT_SECRET
- AUTH_VALIDATE_MODE=grpc: Gateway delegates validation to auth-service via gRPC
  - Required: AUTH_GRPC_ADDR (e.g., localhost:9090), JWT_AUDIENCE
  - Behavior: Requests with invalid/expired tokens receive 401; failures to reach auth-service also return 401

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
| `REDIS_URL` | Redis connection URL (format: `redis://host:port` for TCP, `localhost:6379` for default, `rediss://host:port` for TLS) | `localhost:6379` |
| `AUTH_SERVICE_URL` | Auth service base URL for login/refresh | (required) |
| `FRONTEND_URL` | CORS allowed origin (comma-separated list; **NEVER use `*` with credentials**) | (required in production) |
| `OPENAPI_PATH` | Path to OpenAPI spec for validation | `../docs/openapi.yaml` |
| `STRICT_OPENAPI_VALIDATION` | Fail startup if OpenAPI spec has syntax errors or cannot be loaded (**recommended `true` in production to catch config issues early**) | `false` |
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

Shell-safe example (single quotes, no escaping needed):
```bash
ROUTE_LIMITS='{"/api/v1/portfolios":{"rps":5,"burst":10},"/api/v1/wallets/:address":{"rps":2,"burst":4}}'
```

Plain JSON (validate with `jq`):
```json
{
  "/api/v1/portfolios": {"rps": 5, "burst": 10},
  "/api/v1/wallets/:address": {"rps": 2, "burst": 4}
}
```

**Validation**:
```bash
# Validate JSON before setting
echo "$ROUTE_LIMITS" | jq . || echo "Invalid JSON"
```

**Behavior**:
- Unknown routes fall back to default limits
- Invalid JSON format will cause startup failure with error logged
- Route limit key includes route + user_id (from JWT) or client IP as fallback

Supports Redis (distributed) or in-memory backends.

## Health & Observability

### Health Endpoints

**Health Check**: `GET /healthz`
- Returns 200 if service process is running
- When `REDIS_URL` is set, also performs shallow Redis connectivity check
- **Use Case**: Kubernetes liveness and readiness probes
- **Note**: In production, consider separating liveness (process alive) from readiness (dependencies healthy) if deep dependency checks are needed

**Metrics**: `GET /metrics`
- Prometheus exposition format
- Includes RED metrics (Rate, Errors, Duration) for all endpoints
- Rate limiting metrics: `rate_limit_allowed_total`, `rate_limit_blocked_total`

**Response Headers** (on all API responses):
- `X-RateLimit-Limit` - Maximum requests allowed per window
- `X-RateLimit-Remaining` - Remaining requests in current window  
- `X-RateLimit-Reset` - Unix timestamp when limit resets
- `Retry-After` - Seconds to wait (on 429 responses)

### Tracing

OpenTelemetry distributed tracing:
- **Development**: Stdout exporter (default)
- **Production**: Set `OTEL_EXPORTER_OTLP_ENDPOINT` for OTLP export
  - Supports both HTTP (`http://collector:4318`) and gRPC (`collector:4317`)
- Automatic span creation for all HTTP requests
- Trace context propagation to downstream services

**Example Prometheus scrape config**:
```yaml
scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['gateway:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

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

### CORS Configuration

**⚠️ CRITICAL**: The gateway sets `Access-Control-Allow-Credentials: true`. Using `FRONTEND_URL=*` **violates the CORS specification** and will cause browser errors.

**Single origin** (production):
```bash
FRONTEND_URL=https://app.million-hunter.com
```

**Multiple origins** (comma-separated):
```bash
FRONTEND_URL=https://app.million-hunter.com,https://dashboard.million-hunter.com
```

**Development only** (insecure):
```bash
FRONTEND_URL=http://localhost:3000
```

**❌ NEVER in production**:
```bash
FRONTEND_URL=*  # INVALID with credentials - will fail
```

### OpenAPI Validation

**Development** (warns on mismatch):
```bash
STRICT_OPENAPI_VALIDATION=false  # Default - logs warnings
make validate-openapi             # Manual validation
```

**Production** (fails fast on schema drift):
```bash
STRICT_OPENAPI_VALIDATION=true   # Recommended - fails startup on errors
```

**CI/CD Integration**:
```yaml
# Example GitHub Actions
- name: Validate OpenAPI Compliance
  run: |
    cd api-gateway
    make validate-openapi
```

### Production Deployment Checklist

Before deploying to production:

1. **Secrets & Authentication**
   - [ ] Set cryptographically secure `JWT_SECRET` (min 32 bytes, random)
   - [ ] Ensure `JWT_SECRET` matches auth-service configuration
   - [ ] Rotate secrets regularly (at least quarterly)

2. **CORS & Security Headers**
   - [ ] Configure specific `FRONTEND_URL` origins (NO wildcards)
   - [ ] Verify browser can successfully make credentialed requests
   - [ ] Enable HTTPS via load balancer/reverse proxy
   - [ ] Set secure headers (HSTS, CSP, X-Frame-Options)

3. **Rate Limiting**
   - [ ] Adjust `RATE_DEFAULT_RPS` and `RATE_DEFAULT_BURST` for expected traffic
   - [ ] Configure `ROUTE_LIMITS` for high-traffic endpoints
   - [ ] Use Redis (`REDIS_URL`) for multi-instance deployments
   - [ ] Test rate limiting under load (use k6 tests)

4. **Observability**
   - [ ] Set `OTEL_EXPORTER_OTLP_ENDPOINT` for distributed tracing
   - [ ] Configure Prometheus scraping of `/metrics`
   - [ ] Set up alerting on rate limit blocked requests
   - [ ] Enable structured logging with appropriate levels

5. **Validation & Testing**
   - [ ] Set `STRICT_OPENAPI_VALIDATION=true` to catch API drift
   - [ ] Run load tests (k6) against staging environment
   - [ ] Verify health checks (`/healthz`) work with orchestrator
   - [ ] Test graceful shutdown behavior

6. **Configuration**
   - [ ] Verify `REDIS_URL` format (`redis://host:port` or `rediss://` for TLS)
   - [ ] Set `AUTH_SERVICE_URL` with proper timeout and retry configuration
   - [ ] Document all environment variables in deployment docs

---

**Part of the Million Hunter Crypto Dashboard microservices architecture.**
