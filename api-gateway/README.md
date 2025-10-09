# api-gateway

Run:
- env PORT=8080 AUTH_MODE=mvp-gate ADMIN_USER=aezi ADMIN_PASS=Aa@123456789 OPENAPI_PATH=../docs/openapi.yaml go run ./cmd/api-gateway

OpenAPI validation:
- make validate-openapi (uses OPENAPI_PATH, defaults to ../docs/openapi.yaml)

Rate limiting:
- Default token-bucket limits via env RateDefaultRPS/RateDefaultBurst (compiled defaults 10/20)
- Optional per-route overrides via ROUTE_LIMITS JSON (by route template), supports Redis or in-memory
  Example:
  ROUTE_LIMITS='{"\/api\/v1\/portfolios":{"rps":5,"burst":10},"\/api\/v1\/wallets\/:address":{"rps":2,"burst":4}}'

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
