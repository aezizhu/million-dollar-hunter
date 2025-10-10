# api-gateway

Run:
- env PORT=8080 OPENAPI_PATH=../docs/openapi.yaml JWT_SECRET=devsecret AUTH_SERVICE_URL=http://localhost:9000 FRONTEND_URL=http://localhost:3000 go run ./cmd/api-gateway

OpenAPI validation:
- make validate-openapi (uses OPENAPI_PATH)
- STRICT_OPENAPI_VALIDATION=true to fail startup on validation errors; false logs a warning

Authentication:
- Gateway validates JWTs (HS256 by default) when JWT_SECRET is set and sets user_id from sub
- Login/refresh are proxied to AUTH_SERVICE_URL; gateway does not issue tokens
- In MVP fallback mode without JWT_SECRET, Bearer tokens are NOT accepted unless minimum length; for real environments set JWT_SECRET

Rate limiting:
- Default token-bucket limits via env RATE_DEFAULT_RPS/RATE_DEFAULT_BURST
- Optional per-route overrides via ROUTE_LIMITS JSON (by route template), supports Redis or in-memory
  Example:
  ROUTE_LIMITS='{"\/api\/v1\/portfolios":{"rps":5,"burst":10},"\/api\/v1\/wallets\/:address":{"rps":2,"burst":4}}'
- Rate limit key includes route + user_id (from JWT) or client IP as fallback

CORS:
- Configure FRONTEND_URL (e.g., http://localhost:3000); use * only for dev

Health:
- GET /healthz checks Redis connectivity when REDIS_URL is set

Metrics and tracing:
- Prometheus at /metrics with RED metrics and rate-limit counters
- OpenTelemetry tracing (stdout by default; OTLP if OTEL_EXPORTER_OTLP_ENDPOINT set)

Endpoints:
- GET /healthz
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- GET/POST /api/v1/portfolios
- GET /api/v1/wallets/:address
- GET /api/v1/wallets/:address/transactions
- GET /api/v1/tokens/:tokenAddress/holders
- GET /api/v1/export/wallet/:address
- GET /metrics
