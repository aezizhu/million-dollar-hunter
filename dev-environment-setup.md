# Development Environment Setup

## Prereqs
- Docker & Docker Compose
- Node 18+, pnpm/yarn; Go 1.21+
- mkcert (for local TLS optional)

## Local Services (Docker Compose)
- API Gateway, portfolio, ingestion, market-data
- PostgreSQL, Redis
- WireMock for external APIs (optional)

## Steps
1) Copy .env.example to .env per service; fill values.
2) docker-compose up -d
3) Access API at http://localhost:8080 and UI at http://localhost:3000

## IDE & Plugins
- VS Code: Go, Docker, ESLint, Prettier
- GolangCI-Lint integration; EditorConfig enabled

## Local Kubernetes (Optional)
- Use kind or minikube; apply k8s manifests; enable ingress for API/UI.

## Mock External APIs
- Start WireMock container with mappings under mocks/.
- Toggle USE_API_MOCKS=true to point services to mock.

## Debugging
- Go: delve with Docker remote debugging.
- Frontend: Next.js dev server with HMR.
- Distributed tracing: run Jaeger locally; export OTLP traces.
