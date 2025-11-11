# Million Dollar Hunter

A personal cryptocurrency portfolio tracking platform designed for monitoring blockchain tokens and wallet activity across multiple chains. Built with a microservices architecture using CQRS pattern for high-performance data ingestion and querying.

> *"Every great hunter knows that patience and precision lead to the most valuable discoveries."* — Crafted with meticulous attention to architectural patterns and system design.

## Overview

Million Dollar Hunter is a single-user, on-chain cryptocurrency dashboard that empowers individuals to monitor, query, and analyze blockchain tokens and wallet activity in real time. The platform provides deep analytics for wallets and tokens across BSC, Solana, Ethereum, and Polygon blockchains, with customizable dashboards and comprehensive data export capabilities.

The system uses a microservices-based architecture that separates write operations (data ingestion) from read operations (portfolio queries) using Kafka event streaming, enabling independent scaling and optimization of each component.

## Quick Start

Get the application running in 5 steps:

### 1. Start Infrastructure Services

```bash
cd ops
docker-compose up -d
```

This starts PostgreSQL, Redis, and Kafka using Docker Compose.

### 2. Run Database Migrations

```bash
# Auth service
cd services/auth-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
migrate -path db/migrations -database "$DATABASE_URL" up

# Portfolio service
cd services/portfolio-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
make migrate-up

# Repeat for market-data-service and ingestion-service as needed
```

### 3. Start Backend Services

Open separate terminals for each service:

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

### 4. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

### 5. Access the Application

- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **Health Check**: `curl http://localhost:8080/healthz`

**MVP Authentication Credentials**:
- Username: `aezi`
- Password: `Aa@123456789`

## Architecture Overview

### Microservices with CQRS Pattern

The platform implements Command Query Responsibility Segregation (CQRS), separating write operations (ingestion-service) from read operations (portfolio-service) using Kafka event streaming.

### Core Components

1. **Frontend (Next.js 15)**
   - Modern React 19 application with App Router
   - Material-UI component library
   - TanStack Query for server state management
   - TypeScript with strict mode enabled

2. **API Gateway (Go/Fiber)**
   - Single public entry point for all client requests
   - JWT authentication middleware
   - Rate limiting with Redis token bucket
   - Request routing to backend gRPC services

3. **Auth Service (JWT/gRPC)**
   - Dual HTTP/gRPC interface
   - JWT token generation and validation
   - Refresh token rotation
   - Login lockout protection (3 failures in 15 minutes)

4. **Portfolio Service (CQRS Read Model)**
   - gRPC server for portfolio queries
   - Kafka consumer for transaction events
   - Aggregates wallet balances from transaction history
   - Provides portfolio summaries and export functionality

5. **Market Data Service (CoinGecko Integration)**
   - Real-time token price data from CoinGecko API
   - Redis caching with 60-second TTL
   - Background worker for price refresh
   - gRPC interface for price queries

6. **Ingestion Service (CQRS Write Model)**
   - Fetches blockchain data from Alchemy/Moralis APIs
   - Publishes transaction events to Kafka
   - Handles multi-chain wallet tracking
   - Circuit breaker pattern for API resilience

### Data Flow

1. **Ingestion Service** fetches blockchain transactions from Alchemy/Moralis APIs
2. Publishes `TransactionDataIngested` events to Kafka
3. **Portfolio Service** consumes events and updates read models
4. **Market Data Service** enriches portfolios with real-time token prices from CoinGecko
5. **API Gateway** routes authenticated requests and enforces rate limits
6. **Auth Service** issues and validates JWT tokens via gRPC

## Tech Stack

### Backend
- **Language**: Go 1.21+ (workspace mode)
- **Databases**: PostgreSQL 15 (primary), Redis (caching, rate limiting)
- **Messaging**: Apache Kafka (event streaming)
- **APIs**: gRPC (inter-service), REST (public gateway)
- **Observability**: OpenTelemetry, Prometheus, structured JSON logging (zerolog)

### Frontend
- **Framework**: Next.js 15 with App Router
- **UI Library**: React 19, TypeScript, Material-UI
- **State Management**: TanStack Query (server state), React Context (UI state)
- **Charts**: TradingView Lightweight Charts

### External APIs
- **CoinGecko**: Market data and token prices
- **Alchemy**: Blockchain transaction data (Ethereum, BSC, Polygon)
- **Moralis**: Multi-chain wallet balances (Solana, fallback)

## Prerequisites

Before starting development, ensure you have the following tools installed:

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

## Project Structure

```
million-dollar-hunter/
├── api-gateway/                # Public-facing API gateway (Go, Fiber)
│   ├── cmd/api-gateway/        # Main entry point
│   ├── internal/               # Private application code
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

## Environment Configuration

### API Gateway

```bash
PORT=8080
JWT_SECRET=devsecret                    # Min 32 bytes in production
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
PORTFOLIO_SERVICE_URL=localhost:8081    # gRPC
MARKET_DATA_SERVICE_URL=localhost:50051 # gRPC
AUTH_VALIDATE_MODE=grpc                 # or "local"
AUTH_GRPC_ADDR=localhost:50051
FRONTEND_URL=http://localhost:3000      # CORS configuration
RATE_DEFAULT_RPS=10
RATE_DEFAULT_BURST=20
OPENAPI_PATH=../docs/openapi.yaml
```

**CORS Security Note**: NEVER use `FRONTEND_URL=*` with credentialed requests. Always specify exact origins (comma-separated for multiple origins).

### Auth Service

```bash
PORT=8081
GRPC_PORT=50051
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_ISSUER=million-hunter
JWT_AUDIENCE=million-hunter-client
JWT_SIGNING_KEY=devsecret               # Same as JWT_SECRET
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168               # 7 days
ENABLE_MULTI_USER=false                 # MVP mode
PASSWORD_MIN_LENGTH=12
LOCKOUT_AFTER_FAILS=3
LOCKOUT_WINDOW_MIN=15
```

### Portfolio Service

```bash
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested
EXPORT_DIR=/tmp/exports
EXPORT_CLEANUP_TTL=1h
```

### Market Data Service

```bash
GRPC_PORT=50051
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_API_KEY=                      # Optional, increases rate limits
COINGECKO_BASE_URL=https://api.coingecko.com/api/v3
COINGECKO_RATE_LIMIT=50                 # Requests per minute
WORKER_ENABLED=true
WORKER_REFRESH_INTERVAL=30s
```

### Ingestion Service

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable
REDIS_ADDR=localhost:6379
ALCHEMY_BASE_URL=https://eth-mainnet.g.alchemy.com/v2/
ALCHEMY_API_KEY=your_key_here
MORALIS_BASE_URL=https://deep-index.moralis.io/api/v2/
MORALIS_API_KEY=your_key_here
HTTP_PORT=8090
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_TX_INGESTED=TransactionDataIngested
```

### Frontend

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080  # Points to API Gateway
```

See `.env.example` files in each service directory for complete documentation.

## Build & Test Commands

### Repository Root (Go Workspace)

The project uses a Go workspace (`go.work`) that includes all Go services:

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

**Note**: Integration tests require Docker daemon running (uses Testcontainers for PostgreSQL).

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
make up                 # Start docker-compose dependencies
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

**Note**: TypeScript strict mode is enabled. Path alias `@/*` maps to `./src/*`.

## External API Integrations

### CoinGecko (Market Data Service)

- **API**: `https://api.coingecko.com/api/v3/`
- **Rate Limit**: 50 req/min (free tier), higher with API key
- **Supported Chains**: BSC, Solana, Ethereum, Polygon
- **Caching**: Redis with 60-second TTL, targeting ≥80% cache hit rate

### Alchemy (Ingestion Service)

- **API**: Asset transfers endpoint with pagination
- **Purpose**: Fetch blockchain transactions (Ethereum, BSC, Polygon, Arbitrum, Optimism)
- **Rate Limiting**: Token bucket algorithm with Redis
- **Features**: Single endpoint for ERC20, ERC721, ERC1155 transfers

### Moralis (Ingestion Service)

- **API**: Multi-chain wallet balances
- **Supported Chains**: BSC, Solana, Ethereum
- **Purpose**: Fallback provider and Solana support
- **Rate Limiting**: Circuit breaker pattern for resilience

## Docker Compose Development

For integrated local development with all dependencies and services:

```bash
# Start full stack (from repository root)
docker-compose -f ops/docker-compose.yml up -d --build

# View logs
docker-compose -f ops/docker-compose.yml logs -f

# Stop services
docker-compose -f ops/docker-compose.yml down

# Stop and remove volumes
docker-compose -f ops/docker-compose.yml down -v
```

The Docker Compose setup includes:
- All backend services (auth, portfolio, market-data, ingestion, api-gateway)
- Frontend application
- Infrastructure (PostgreSQL, Redis, Kafka, Zookeeper)
- Health checks and automatic restarts

Configuration is managed via environment variables in `ops/docker-compose.yml`.

## Important Notes

### Go Workspace Mode

This project uses Go workspace mode (`go.work`) to manage multiple Go modules in a monorepo:

- Always run `go work sync` from the repository root after pulling changes
- Each service has its own `go.mod` file
- Workspace mode enables cross-service imports during development
- CI/CD builds each service independently with `GOWORK=off`

### Integration Tests

Integration tests require Docker for Testcontainers:

- Tests automatically spin up PostgreSQL containers
- Use `make test-short` or `go test -short` to skip integration tests
- Ensure Docker daemon is running before running full test suite
- Port conflicts may occur if services are already running

### CORS Security

The API Gateway sets `Access-Control-Allow-Credentials: true`:

- **Production**: Use specific origin (e.g., `https://app.million-hunter.com`)
- **Multiple Origins**: Comma-separated list
- **Development**: `http://localhost:3000`
- **NEVER**: Use `FRONTEND_URL=*` (violates CORS spec with credentials)

### CQRS Pattern

The system implements CQRS for optimal performance:

- **Write Model** (Ingestion Service): Optimized for high-throughput data ingestion
- **Read Model** (Portfolio Service): Optimized for fast queries and aggregations
- **Event Streaming**: Kafka connects write and read models
- **Raw Data Storage**: JSONB storage enables reprocessing without re-fetching

### Single-User MVP Design

The current implementation is designed for single-user operation:

- Hardcoded authentication credentials (username: `aezi`, password: `Aa@123456789`)
- JWT architecture is scaffolded but simplified for MVP
- Full multi-user support can be enabled by setting `ENABLE_MULTI_USER=true`
- No GDPR compliance burden for MVP

## Additional Documentation

For more detailed information, refer to these documentation files:

- **[AGENTS.md](AGENTS.md)** - Comprehensive developer reference and AI agent briefing
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Deployment instructions and production setup
- **[docs/TECHNICAL-DECISIONS.md](docs/TECHNICAL-DECISIONS.md)** - Architecture decisions and rationale
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - Testing approach and coverage targets
- **[TESTING-PHASE1-SUMMARY.md](TESTING-PHASE1-SUMMARY.md)** - Test coverage summary
- **[docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)** - Product requirements document
- **[docs/external-api-integrations.md](docs/external-api-integrations.md)** - External API integration details

## Performance Targets

### API Gateway
- Throughput: ≥100 req/s (single instance)
- p95 latency: ≤300ms (cached requests)
- Rate limit accuracy: <1% error rate

### Market Data Service
- Cache hit rate: ≥80% (5-minute windows)
- p95 latency: ≤300ms (cached), ≤2s (cache miss + CoinGecko)
- Throughput: ≥100 req/s

### Ingestion Service
- Transaction throughput: ≥100 tx/s (write model)
- WireMock latency: <100ms (local mocks)
- External API circuit breaker: 50% failure threshold

### Portfolio Service
- Read model latency: <200ms p95
- Kafka consumer lag: <10s under normal load

### Frontend
- Time to Interactive (TTI): <3s (cached)
- First Contentful Paint (FCP): <1.5s
- Test coverage: ≥70% statements/branches

## Contributing

### Pre-Commit Checks

Before committing changes:

1. Run `make check` (auth-service) or `make lint` + `make test` (other services)
2. Ensure tests pass: `make test` or `go test ./...`
3. Frontend: `npm run lint` and `npm test`
4. Verify no uncommitted secrets or sensitive data

### Commit Conventions

Use conventional commits for clarity:

- `feat:` New feature
- `fix:` Bug fix
- `test:` Add or update tests
- `refactor:` Code refactoring without functional changes
- `docs:` Documentation changes
- `chore:` Maintenance tasks (dependencies, CI config)

Example: `feat(auth): add refresh token rotation`

### Pull Request Requirements

Every PR must include:

1. ✅ All tests passing (`make test` or `npm test`)
2. ✅ Linting passes (`make lint` or `npm run lint`)
3. ✅ Type checking passes (TypeScript: `npm run build`)
4. ✅ Diff confined to relevant files (avoid unrelated changes)
5. ✅ Proof artifact (tests or manual test evidence)
6. ✅ One-paragraph description: What changed, why, and any gotchas
7. ✅ No drop in coverage (check CI reports)

## License

Dual License Agreement (Personal Use / Commercial Use)

See [LICENSE](LICENSE) for full terms.

## Project Status

**Status**: Active development  
**Primary Language**: Go (backend), TypeScript (frontend)  
**Deployment**: Docker Compose (local/staging), Kubernetes (production - planned)

---

*This platform represents a comprehensive approach to on-chain analytics, where every component has been carefully architected to balance performance, maintainability, and scalability. The design philosophy emphasizes clean separation of concerns, robust error handling, and observability-first development practices.*
