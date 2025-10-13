# Million Dollar Hunter - Crypto Dashboard

AI agent briefing for the Million Dollar Hunter cryptocurrency portfolio tracking platform. A microservices architecture built with Go backends, Next.js frontend, PostgreSQL, Redis, and Kafka.

## Build & Test Commands

### Repository Root (Go Workspace)

The project uses a **Go workspace** (`go.work`) that includes all Go services. When working in the root:

```bash
# Sync all module dependencies
go work sync

# Run all tests across workspace
go test ./...
```

### API Gateway

```bash
cd api-gateway
make build              # Build binary
make test               # Run tests
make validate-openapi   # Validate OpenAPI spec compliance
make k6                 # Run load tests
```

### Auth Service

```bash
cd services/auth-service
make check              # Format, lint, and run short tests (recommended pre-commit)
make fmt                # Format code with gofumpt
make lint               # Run golangci-lint
make test               # Run all tests (requires Docker for integration tests)
make test-short         # Run unit tests only (skip integration)
make build              # Build binary to bin/auth-service
```

**Integration tests require Docker daemon running** (uses Testcontainers for PostgreSQL).

### Portfolio Service

```bash
cd services/portfolio-service
make build              # Build binary to bin/portfolio-service
make proto              # Regenerate gRPC code from proto files
make migrate-up         # Run database migrations (requires DATABASE_URL)
make migrate-down       # Rollback migrations
make run                # Run service locally
```

### Market Data Service

```bash
cd services/market-data-service
make build              # Build binary
go test ./internal/... -v                      # Unit tests
docker-compose up -d postgres redis            # Start dependencies
go test ./tests -v -run Integration            # Integration tests
go test ./tests -v -run Load                   # Load tests
go test ./tests -bench=. -benchmem             # Benchmarks
```

### Ingestion Service

```bash
cd ingestion-service
make build              # Build binary
make test               # Run tests (requires docker-compose up)
make lint               # Run golangci-lint
make bench              # Run performance benchmarks (≥100 tx/s target)
make up                 # Start docker-compose dependencies (postgres, redis, wiremock, kafka)
make down               # Stop and remove containers
```

### Frontend (Next.js)

```bash
cd frontend
npm run dev             # Start development server (port 3000)
npm run build           # Production build
npm start               # Start production server
npm run lint            # Run ESLint
npm test                # Run Jest tests
npm run test:watch      # Run tests in watch mode
npm run test:coverage   # Generate coverage report
```

**TypeScript strict mode is enabled**. Path alias: `@/*` maps to `./src/*`.

## Project Layout

```
million-dollar-hunter/
├── api-gateway/                # Public-facing API gateway (Go, Fiber)
│   ├── cmd/api-gateway/        # Main entry point
│   ├── tests/k6/               # Load tests for rate limiting
│   └── Makefile
├── services/
│   ├── auth-service/           # JWT authentication (Go, gRPC)
│   │   ├── cmd/auth-service/
│   │   ├── internal/           # Service logic, store, handlers
│   │   ├── tests/              # Integration tests with Testcontainers
│   │   ├── db/migrations/      # PostgreSQL migrations
│   │   └── .golangci.yml       # Linter config
│   ├── portfolio-service/      # CQRS read model (Go, gRPC)
│   │   ├── cmd/server/
│   │   ├── internal/           # Service, repository, Kafka consumer
│   │   └── migrations/         # Database migrations
│   ├── market-data-service/    # CoinGecko integration (Go, gRPC)
│   │   ├── cmd/market-data-service/
│   │   ├── internal/           # Cache, client, handler, repository, worker
│   │   └── docker-compose.yml  # Local dev stack
│   └── ingestion-service/      # CQRS write model (Go, HTTP)
│       ├── cmd/ingestion-service/
│       ├── internal/           # Alchemy/Moralis clients, rate limiting
│       └── docker-compose.yml  # WireMock, Postgres, Redis, Kafka
├── frontend/                   # Next.js 15 App Router
│   ├── src/
│   │   ├── app/                # App Router pages
│   │   ├── components/         # React components
│   │   ├── lib/                # API client, utils
│   │   └── context/            # React context (auth, etc.)
│   ├── tsconfig.json           # TypeScript config (strict mode)
│   └── eslint.config.mjs       # ESLint flat config
├── proto/                      # Shared protobuf definitions
├── docs/                       # Architecture, ADRs, deployment guides
├── ops/                        # Docker compose orchestration files
└── go.work                     # Go workspace definition
```

## Architecture Overview

**Microservices with CQRS**: The platform separates write operations (ingestion-service) from read operations (portfolio-service) using Kafka event streaming.

**Core Data Flow**:
1. **Ingestion Service** fetches blockchain transactions from Alchemy/Moralis APIs
2. Publishes `TransactionDataIngested` events to Kafka
3. **Portfolio Service** consumes events and updates read models
4. **Market Data Service** enriches portfolios with real-time token prices from CoinGecko
5. **API Gateway** routes authenticated requests and enforces rate limits
6. **Auth Service** issues and validates JWT tokens via gRPC

**Tech Stack**:
- **Backend**: Go 1.21+ (workspace mode)
- **Frontend**: Next.js 15 (App Router), React 19, TypeScript, Material-UI
- **Storage**: PostgreSQL 15 (primary), Redis (caching, rate limiting)
- **Messaging**: Kafka (event streaming)
- **APIs**: gRPC (inter-service), REST (public gateway)
- **Observability**: OpenTelemetry, Prometheus, structured JSON logging

## Conventions & Patterns

### Go Services

**Code Style**:
- Use `gofumpt` for formatting (stricter than `gofmt`)
- Line length limit: 140 characters
- Linter: `golangci-lint` with `.golangci.yml` (errcheck, govet, staticcheck, gosec enabled)
- Test files: `*_test.go` with `_test` suffix for packages

**Project Structure** (per service):
```
cmd/              # Main entry points
internal/         # Private application code (not importable by other services)
  ├── config/     # Configuration loading
  ├── handler/    # HTTP/gRPC handlers
  ├── service/    # Business logic
  ├── repository/ # Data access layer
  └── store/      # Database operations
api/              # API definitions (protobuf, OpenAPI)
tests/            # Integration and load tests
db/migrations/    # SQL migrations (numbered: 001_*.sql)
```

**Error Handling**:
- Return errors explicitly; avoid panics in production code
- Use `fmt.Errorf` with `%w` for error wrapping
- Log errors with structured fields (zerolog): `log.Error().Err(err).Str("component", "...").Msg("...")`

**Testing**:
- Coverage target: ≥80% for services (≥90% for critical packages)
- Integration tests use **Testcontainers** for PostgreSQL (requires Docker)
- Test short flag: `go test -short` skips integration tests
- Mock external APIs with WireMock (ingestion-service) or httptest

**Database Migrations**:
- Use `golang-migrate/migrate` CLI tool
- Naming: `001_initial_schema.up.sql` / `001_initial_schema.down.sql`
- Never edit applied migrations; create new ones for schema changes

**gRPC**:
- Protobuf definitions in `proto/` (shared) or `api/proto/` (service-specific)
- Generate code: `protoc --go_out=. --go-grpc_out=. <proto_file>`
- Use `insecure.NewCredentials()` for local dev (TLS required in production)

### Frontend (Next.js)

**Code Style**:
- TypeScript strict mode enabled
- ESLint config: `next/core-web-vitals` + `next/typescript`
- No semicolons, single quotes (enforced by ESLint)
- Path imports: `@/` resolves to `src/`

**Component Structure**:
```
src/
├── app/                    # App Router pages (file-based routing)
│   ├── layout.tsx          # Root layout
│   ├── page.tsx            # Home page
│   └── [route]/page.tsx    # Dynamic routes
├── components/             # Reusable React components
│   ├── common/             # Shared UI components
│   └── features/           # Feature-specific components
├── lib/                    # Non-React utilities
│   ├── api.ts              # API client (axios)
│   └── utils.ts            # Helper functions
└── context/                # React Context providers (auth, theme)
```

**State Management**:
- Use React Query (`@tanstack/react-query`) for server state
- React Context for global UI state (auth, theme)
- Avoid prop drilling; prefer composition

**Testing**:
- Use Jest + React Testing Library
- Coverage target: ≥70% statements/branches
- Focus on critical user flows (login, portfolio view, export)

## Security

### Authentication

**JWT Tokens** (auth-service):
- HS256 signing with `JWT_SIGNING_KEY` (min 32 bytes)
- Access token TTL: 15 minutes
- Refresh token TTL: 7 days (stored in DB, single-use rotation)
- Claims: `user_id` (sub), `email`, `iss` (issuer), `aud` (audience)

**API Gateway Validation**:
- `AUTH_VALIDATE_MODE=local`: Gateway validates JWT using `JWT_SECRET`
- `AUTH_VALIDATE_MODE=grpc`: Gateway delegates validation to auth-service
- Invalid tokens return 401 Unauthorized

**Login Protection**:
- Brute force lockout: 3 failed attempts in 15 minutes
- Audit log: Failed login attempts recorded in `auth_audit` table

### CORS

**CRITICAL**: API Gateway sets `Access-Control-Allow-Credentials: true`. NEVER use `FRONTEND_URL=*` (violates CORS spec).

**Correct configuration**:
```bash
# Production (single origin)
FRONTEND_URL=https://app.million-hunter.com

# Multiple origins (comma-separated)
FRONTEND_URL=https://app.million-hunter.com,https://dashboard.million-hunter.com

# Development only
FRONTEND_URL=http://localhost:3000
```

### Secrets Management

**Required secrets** (set via environment variables):
- `JWT_SECRET` / `JWT_SIGNING_KEY`: Min 32 random bytes, rotate quarterly
- `DATABASE_URL`: PostgreSQL connection string (never commit)
- `ALCHEMY_API_KEY`, `MORALIS_API_KEY`: External API keys (see ingestion-service)
- `COINGECKO_API_KEY`: CoinGecko API key (optional, increases rate limits)

**Never commit**:
- `.env` files (use `.env.example` for templates)
- Credentials, API keys, or private keys
- `credentials.json`, `.pem`, `.key` files

## Git Workflows

### Branching Strategy

- **main**: Production-ready code (protected)
- **develop**: Integration branch for features (optional, not currently used)
- **Feature branches**: `feature/<slug>` or `devin/<timestamp>-<slug>`
- **Bugfix branches**: `bugfix/<slug>`

**Force push rules**:
- ❌ NEVER force push to `main`
- ✅ Allowed on feature branches with `git push --force-with-lease`
- Prefer `git merge` over `git rebase` to preserve history

### Commit Conventions

Use conventional commits for clarity:
- `feat:` New feature
- `fix:` Bug fix
- `test:` Add or update tests
- `refactor:` Code refactoring without functional changes
- `docs:` Documentation changes
- `chore:` Maintenance tasks (dependencies, CI config)

Example: `feat(auth): add refresh token rotation`

### Pre-Commit Checks

**Before committing**:
1. Run `make check` (auth-service) or `make lint` + `make test` (other services)
2. Ensure tests pass: `make test` or `go test ./...`
3. Frontend: `npm run lint` and `npm test`
4. Verify no uncommitted secrets or sensitive data

### Pull Request Requirements

**Every PR must include**:
1. ✅ All tests passing (`make test` or `npm test`)
2. ✅ Linting passes (`make lint` or `npm run lint`)
3. ✅ Type checking passes (TypeScript: `npm run build`)
4. ✅ Diff confined to relevant files (avoid unrelated changes)
5. ✅ **Proof artifact**:
   - Bug fix: Add failing test first, then fix (test now passes)
   - Feature: New tests or manual test evidence (screenshots, logs)
6. ✅ One-paragraph description: What changed, why, and any gotchas
7. ✅ No drop in coverage (check CI reports)
8. ✅ No unexplained new runtime dependencies

**CI/CD checks** (GitHub Actions):
- Build verification (all services)
- Unit and integration tests
- Linting and security scanning (gosec, govulncheck)
- OpenAPI validation (API Gateway)
- Coverage upload to Codecov

## External Services

### API Integrations

**CoinGecko** (market-data-service):
- API: `https://api.coingecko.com/api/v3/`
- Rate limit: 50 req/min (free tier), higher with API key
- Chains: BSC, Solana, Ethereum, Polygon
- Redis cache: 60s TTL, ≥80% hit rate target

**Alchemy** (ingestion-service):
- API: Asset transfers endpoint with pagination
- Rate limit: Token bucket with Redis
- Purpose: Fetch blockchain transactions (Ethereum, Polygon)

**Moralis** (ingestion-service):
- API: Multi-chain wallet balances
- Chains: BSC, Solana, Ethereum
- Rate limit: Circuit breaker pattern

### Environment Variables

**Essential for local development**:

```bash
# API Gateway
PORT=8080
JWT_SECRET=devsecret  # Min 32 bytes in production
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
FRONTEND_URL=http://localhost:3000
OPENAPI_PATH=../docs/openapi.yaml

# Auth Service
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_SIGNING_KEY=devsecret  # Same as JWT_SECRET
HTTP_PORT=9000
GRPC_PORT=9090

# Portfolio Service
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested

# Market Data Service
GRPC_PORT=50051
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_RATE_LIMIT=50  # Requests per minute

# Ingestion Service
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable
REDIS_ADDR=localhost:6379
ALCHEMY_BASE_URL=https://eth-mainnet.g.alchemy.com/v2/
ALCHEMY_API_KEY=your_key_here
MORALIS_BASE_URL=https://deep-index.moralis.io/api/v2/
MORALIS_API_KEY=your_key_here
HTTP_PORT=8090
KAFKA_BROKERS=localhost:9092

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080  # Points to API Gateway
```

See `.env.example` files in each service directory for complete documentation.

## Development Setup

### Prerequisites

- **Go 1.21+** (workspace mode support)
- **Node.js 20+** and npm
- **Docker & Docker Compose** (for dependencies and integration tests)
- **Protocol Buffers compiler** (`protoc`) with Go plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **golang-migrate** CLI:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint** (for linting):
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **gofumpt** (for formatting):
  ```bash
  go install mvdan.cc/gofumpt@latest
  ```
- **k6** (optional, for load testing):
  ```bash
  # macOS: brew install k6
  # Linux: See https://k6.io/docs/get-started/installation/
  ```

### Quick Start (Full Stack)

1. **Start infrastructure** (Postgres, Redis, Kafka):
   ```bash
   cd ops
   docker-compose up -d
   ```

2. **Run database migrations** (each service):
   ```bash
   # Auth service
   cd services/auth-service
   export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
   migrate -path db/migrations -database "$DATABASE_URL" up
   
   # Portfolio service
   cd services/portfolio-service
   export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
   make migrate-up
   
   # Repeat for market-data-service and ingestion-service
   ```

3. **Start backend services** (separate terminals):
   ```bash
   # Terminal 1: Auth service
   cd services/auth-service && make run
   
   # Terminal 2: Portfolio service
   cd services/portfolio-service && make run
   
   # Terminal 3: Market data service
   cd services/market-data-service && go run cmd/market-data-service/main.go
   
   # Terminal 4: Ingestion service
   cd ingestion-service && go run cmd/ingestion-service/main.go
   
   # Terminal 5: API Gateway
   cd api-gateway && make run
   ```

4. **Start frontend**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. **Access application**:
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080
   - Health checks: `curl http://localhost:8080/healthz`

### Docker Compose Development

For integrated local development with all dependencies:

```bash
# Start full stack (from ops/ directory)
docker-compose -f ops/docker-compose.yml up -d --build

# View logs
docker-compose -f ops/docker-compose.yml logs -f

# Stop services
docker-compose -f ops/docker-compose.yml down
```

## Gotchas & Common Issues

### Go Workspace Issues

**Problem**: `go.sum` out of sync, CI fails  
**Solution**: Run `go work sync` from repository root to sync all module dependencies

**Problem**: Import errors between services  
**Solution**: Check `go.work` includes all service directories; run `go mod tidy` in each service

### Integration Test Failures

**Problem**: Tests fail with "Cannot connect to Docker daemon"  
**Solution**: Ensure Docker is running; Testcontainers requires Docker daemon access

**Problem**: Port conflicts (e.g., PostgreSQL 5432 already in use)  
**Solution**: Stop conflicting services or adjust ports in `docker-compose.yml`

### API Gateway CORS Errors

**Problem**: Browser blocks requests with "CORS policy" error  
**Solution**: Verify `FRONTEND_URL` matches frontend origin exactly (including protocol and port)

**Problem**: "Credentials mode is 'include'" error  
**Solution**: Never use `FRONTEND_URL=*` with credentialed requests; specify exact origins

### Rate Limiting Issues

**Problem**: Getting 429 responses in development  
**Solution**: Adjust `RATE_DEFAULT_RPS` and `RATE_DEFAULT_BURST` in API Gateway `.env`

**Problem**: Rate limits not working (all requests pass)  
**Solution**: Check Redis connectivity; rate limiting falls back to in-memory if Redis unavailable

### Database Migration Errors

**Problem**: Migration fails with "dirty database version"  
**Solution**: Fix manually:
```bash
migrate -path db/migrations -database "$DATABASE_URL" force <version>
migrate -path db/migrations -database "$DATABASE_URL" up
```

**Problem**: Migration version mismatch  
**Solution**: Never edit applied migrations; create new migration files for changes

### Frontend Build Errors

**Problem**: TypeScript errors on `npm run build`  
**Solution**: Run `npm run lint` to see all errors; fix type issues (strict mode enabled)

**Problem**: "Module not found" with `@/` imports  
**Solution**: Check `tsconfig.json` paths configuration; `@/*` should map to `./src/*`

### Performance Issues

**Problem**: High API latency (>300ms p95)  
**Solution**: 
1. Check Redis cache hit rate (target ≥80%): View Prometheus metrics at `/metrics`
2. Verify background workers are running (market-data-service, portfolio-service)
3. Check external API rate limits (CoinGecko, Alchemy, Moralis)

**Problem**: Low cache hit rate  
**Solution**: 
- Increase Redis TTL (`REDIS_TTL` in market-data-service)
- Reduce background worker refresh interval (`WORKER_REFRESH_INTERVAL`)
- Add more tracked tokens to worker batch

## Links to Key Documentation

- **OpenAPI Spec**: `docs/openapi.yaml` (API contract)
- **Architecture Decisions**: `docs/architecture-decisions.md` (ADRs)
- **Testing Strategy**: `docs/testing-strategy.md` (coverage targets, test types)
- **Security Audit**: `services/auth-service/SECURITY-AUDIT-CHECKLIST.md`
- **Deployment Guide**: `docs/DEPLOYMENT.md`
- **Agent Handoff**: `docs/AGENT-HANDOFF.md` (for multi-agent workflows)
- **Performance Requirements**: `docs/performance-requirements.md` (SLOs)

## Performance Targets

**API Gateway**:
- Throughput: ≥100 req/s (single instance)
- p95 latency: ≤300ms (cached requests)
- Rate limit accuracy: <1% error rate

**Market Data Service**:
- Cache hit rate: ≥80% (5-minute windows)
- p95 latency: ≤300ms (cached), ≤2s (cache miss + CoinGecko)
- Throughput: ≥100 req/s

**Ingestion Service**:
- Transaction throughput: ≥100 tx/s (write model)
- WireMock latency: <100ms (local mocks)
- External API circuit breaker: 50% failure threshold

**Portfolio Service**:
- Read model latency: <200ms p95
- Kafka consumer lag: <10s under normal load

**Frontend**:
- Time to Interactive (TTI): <3s (cached)
- First Contentful Paint (FCP): <1.5s
- Test coverage: ≥70% statements/branches

## Testing Requirements

**Before merging**:
1. Run service-specific tests: `make test` (Go) or `npm test` (frontend)
2. Integration tests must pass (requires Docker)
3. Lint checks clean: `make lint` or `npm run lint`
4. Coverage maintained or improved (check CI reports)
5. Load tests pass for API Gateway: `cd api-gateway && make k6`

**Test evidence formats**:
- Automated: Include test output in PR description
- Manual: Screenshots, curl outputs, or browser DevTools network logs
- Performance: k6 results or Prometheus metrics snapshots

**Critical test scenarios**:
- Auth flow: Login, token refresh, expired token handling
- Portfolio CRUD: Add wallet, list portfolios, pagination
- Price enrichment: Verify CoinGecko integration, cache hits
- Rate limiting: Test 429 responses, retry-after headers
- Export: CSV and JSON formats for wallet data

---

**Project Status**: Active development  
**Primary Language**: Go (backend), TypeScript (frontend)  
**Deployment**: Docker Compose (local/staging), Kubernetes (production - planned)  
**License**: Proprietary
