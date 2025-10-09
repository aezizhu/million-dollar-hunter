# api-gateway

Run:
- env PORT=8080 AUTH_MODE=mvp-gate ADMIN_USER=aezi ADMIN_PASS=Aa@123456789 go run ./cmd/api-gateway

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

This service follows docs/openapi.yaml contracts and includes basic rate limiting, logging, and metrics.
