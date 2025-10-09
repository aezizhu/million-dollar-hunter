# Quick Reference - Million Dollar Hunter

**Last Updated**: 2025-10-09  
**Purpose**: Fast project overview for rapid context restoration

⏱️ **Reading Time**: 2 minutes

---

## What Is This Project?

**Million Dollar Hunter** is a single-user cryptocurrency portfolio analytics dashboard that enables real-time monitoring and analysis of blockchain wallet activity across BSC and Solana networks.

**Key Capabilities**:
- Track multiple wallets across chains (BSC, Solana, Ethereum, Polygon, Arbitrum, Optimism)
- View token holdings, balances, and historical portfolio value
- Monitor whale activity and top holder rankings
- Analyze transaction history with detailed on-chain data
- Export portfolio data (CSV/JSON)
- Real-time price tracking from CoinGecko

**User**: Single owner/operator (MVP hardcoded credentials: `aezi` / `Aa@123456789`)

---

## Architecture in 60 Seconds

### System Design: Microservices

```
┌─────────────┐
│   Next.js   │  Agent F (Frontend)
│   Frontend  │
└──────┬──────┘
       │ HTTPS/REST
       ↓
┌─────────────┐
│ API Gateway │  Agent B (Routing, Rate Limiting, Auth)
└──────┬──────┘
       │ gRPC
       ↓
┌──────────────────────────────────────────┐
│  Backend Microservices (Go)              │
├──────────────┬───────────────┬───────────┤
│ auth-service │ portfolio-    │ market-   │
│ (Agent A)    │ service       │ data-     │
│              │ (Agent C)     │ service   │
│              │               │ (Agent E) │
│              ├───────────────┤           │
│              │ ingestion-    │           │
│              │ service       │           │
│              │ (Agent D)     │           │
└──────────────┴───────────────┴───────────┘
       │              │              │
       ↓              ↓              ↓
  PostgreSQL     Apache Kafka    Redis
  (per service)  (Agent C)       (Agent E)
                      │
                      ↓
              External APIs
         (Alchemy, Moralis, CoinGecko)
```

### Core Services

1. **auth-service** (Agent A) - JWT authentication, user management
2. **api-gateway** (Agent B) - Public REST API, rate limiting, observability
3. **portfolio-service** (Agent C) - Wallet tracking, CQRS read model, Kafka broker
4. **ingestion-service** (Agent D) - External blockchain data, CQRS write model
5. **market-data-service** (Agent E) - Token prices, CoinGecko integration, Redis cache

### Key Patterns

- **Microservices**: Each service has own database, communicates via APIs
- **CQRS**: Write (Agent D) → Kafka → Read (Agent C)
- **Event-Driven**: Kafka event streaming for async workflows
- **Database-Per-Service**: PostgreSQL 15+ per backend service
- **API Gateway**: Single public entry point for all client requests

---

## Technology Stack

### Backend (Agents A, B, C, D, E)
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP routing)
- **RPC**: gRPC with Protobuf
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Messaging**: Apache Kafka
- **Auth**: golang-jwt/jwt v5+
- **Logging**: rs/zerolog
- **Monitoring**: Prometheus + Grafana
- **Tracing**: OpenTelemetry
- **Migrations**: golang-migrate

### Frontend (Agent F)
- **Framework**: Next.js 15+ (App Router)
- **Language**: TypeScript
- **UI Library**: Material-UI (MUI)
- **Charts**: TradingView Lightweight Charts
- **State**: TanStack Query (server state), React Context (UI state)

### Infrastructure
- **Containers**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions

### External APIs
- **Blockchain Data**: Alchemy (primary), Moralis (fallback + Solana)
- **Market Data**: CoinGecko

---

## Agent Team Structure (A-F)

| Agent | Role | Service | Responsibilities |
|-------|------|---------|------------------|
| **A** | Authentication & Security | auth-service | JWT, user management, security hardening |
| **B** | API Gateway & Orchestration | api-gateway | REST API, rate limiting, observability |
| **C** | Portfolio & Aggregation | portfolio-service | Wallet tracking, CQRS read, Kafka broker |
| **D** | Data Ingestion | ingestion-service | Alchemy/Moralis, CQRS write, blockchain data |
| **E** | Market Data | market-data-service | CoinGecko, price caching, Redis |
| **F** | Frontend & Visualization | Next.js app | UI, Material-UI, TradingView charts |

### Dependencies
- A → B: JWT validation interface
- B → F: REST API surface
- C ← D: Kafka events (transaction data)
- C ← E: Price data (gRPC)
- D → C: Event publishing
- E → C: Price enrichment

---

## Project Phases (16 Weeks)

### Phase 0: Planning & Documentation ✅ COMPLETED
- PRD, Technical Plan, Multi-Agent Plan created
- Documentation infrastructure established

### Phase 1: Foundation (Weeks 1-4) - NOT STARTED
**Active Agents**: A, B, C
- Auth service + JWT
- API Gateway skeleton  
- Kafka infrastructure
- Basic portfolio service

### Phase 2: Data Pipeline (Weeks 5-8) - NOT STARTED
**Active Agents**: D, E, C
- Ingestion service (Alchemy/Moralis)
- Market data service (CoinGecko)
- Data aggregation pipeline

### Phase 3: Frontend (Weeks 9-12) - NOT STARTED
**Active Agents**: F, B
- Next.js application
- Material-UI components
- Dashboard and wallet views

### Phase 4: Integration (Weeks 13-16) - NOT STARTED
**Active Agents**: All (A-F)
- TradingView charts
- End-to-end testing
- Performance optimization
- Production deployment

---

## Key File Locations

### Documentation (`./docs/`)
- **QUICK-REFERENCE.md** (this file) - Project overview
- **SESSION-RECOVERY.md** - Context restoration guide
- **DEVELOPMENT-STATUS.md** - Current progress and tasks
- **AGENT-ASSIGNMENTS.md** - Agent roles and responsibilities
- **API-STATUS.md** - Endpoint implementation tracking
- **TECHNICAL-DECISIONS.md** - Decision log with rationale
- **AGENT-HANDOFF.md** - Handoff protocol

### Specifications
- **openapi.yaml** - REST API specification (425 lines)
- **PRD-Million-Dollar-Hunter-Crypto-Dashboard.md** - Product requirements (293 lines)
- **Technical Development Plan.md** - Technical architecture (415 lines)
- **million-hunter-development-plan.md** - Multi-agent plan (948 lines)

### Supporting Docs
- **architecture-decisions.md** - Architecture patterns
- **database-migration-strategy.md** - Schema versioning
- **external-api-integrations.md** - API rate limits, costs
- **frontend-components.md** - UI component library
- **monitoring-alerting.md** - Observability setup
- **performance-requirements.md** - SLO targets
- **security-hardening.md** - Security practices
- **testing-strategy.md** - Test requirements

### Code Structure (To Be Created)
```
million-dollar-hunter/
├── docs/                    # All documentation
├── backend/
│   ├── auth-service/       # Agent A
│   ├── api-gateway/        # Agent B
│   ├── portfolio-service/  # Agent C
│   ├── ingestion-service/  # Agent D
│   └── market-data-service/# Agent E
├── frontend/               # Agent F (Next.js)
└── infrastructure/         # Docker, K8s configs
```

---

## REST API Endpoints (Public)

Base URL: `http://localhost:8000` (local)

### Authentication
- `POST /api/v1/auth/login` - Login (MVP: hardcoded check)
- `POST /api/v1/auth/register` - Register user (post-MVP)
- `POST /api/v1/auth/refresh` - Refresh token (post-MVP)

### Portfolio
- `GET /api/v1/portfolios` - List tracked wallets
- `POST /api/v1/portfolios/wallets` - Add wallet to track
- `GET /api/v1/portfolios/wallets/{address}` - Get wallet details
- `GET /api/v1/portfolios/wallets/{address}/transactions` - Get transactions

### Market Data
- `GET /api/v1/market/prices` - Get token prices

### Export
- `GET /api/v1/export/wallet/{address}` - Export wallet data

See `openapi.yaml` for complete specification.

---

## Business Requirements → Technical Components

| PRD Requirement | Technical Component | Owner |
|-----------------|---------------------|-------|
| Token Analytics | portfolio-service + market-data-service | C + E |
| Wallet Tracking | portfolio-service | C |
| Transaction Monitoring | ingestion-service → portfolio-service | D → C |
| Top Holder Analysis | portfolio-service (holder_snapshots table) | C |
| Price Tracking | market-data-service + Redis cache | E |
| Cross-Chain Support | ingestion-service (multi-chain APIs) | D |
| Single-User Auth | auth-service (hardcoded for MVP) | A |
| Data Export | portfolio-service (CSV/JSON) | C |
| Dashboard UI | Next.js frontend | F |
| Real-time Updates | Kafka events + WebSocket (future) | C |

---

## Performance Targets

| Metric | Target | Status |
|--------|--------|--------|
| API p95 latency | ≤ 300ms | Not Measured |
| Frontend LCP | ≤ 2.5s (3G Fast) | Not Measured |
| Cache hit rate | ≥ 80% | Not Measured |
| Ingestion throughput | ≥ 100 tx/s | Not Measured |
| Bundle size | ≤ 250KB gzipped | Not Measured |
| Uptime | 99.9% | Not Deployed |

See `performance-requirements.md` for complete SLOs.

---

## Database Schemas (Per Service)

### auth-service (Agent A)
- `users`: id, email, password_hash, created_at, updated_at

### portfolio-service (Agent C) - Read-Optimized
- `wallets`: id, user_id, address, chain, nickname
- `assets`: id, wallet_id, token_address, symbol, current_balance
- `asset_snapshots`: id, asset_id, timestamp, balance, usd_value
- `transactions_view`: Denormalized transaction history
- `holder_snapshots`: Historical top holder data

### ingestion-service (Agent D) - Write-Optimized
- `ingestion_jobs`: id, wallet_address, chain, status, cursor
- `raw_transactions`: id, source_api, wallet_address, data (JSONB)
- `raw_balances`: id, wallet_address, chain, data (JSONB)

### market-data-service (Agent E)
- `token_prices`: token_address, chain, usd_price, last_updated
- `market_data`: id, token_address, chain, metadata (JSONB)

---

## Kafka Event Schemas

### WalletTrackingRequested
**Publisher**: Agent B or C  
**Consumer**: Agent D  
**Purpose**: Trigger data fetch for new wallet

### TransactionDataIngested
**Publisher**: Agent D  
**Consumer**: Agent C  
**Purpose**: Signal new data ready for aggregation

### PortfolioUpdated
**Publisher**: Agent C  
**Consumer**: Future notification service  
**Purpose**: Notify of portfolio changes

---

## Common Commands (Future)

### Backend Services
```bash
# Start a service
go run cmd/SERVICE_NAME/main.go

# Run tests
go test ./...

# Run migrations
migrate -path db/migrations -database "postgres://..." up

# Build Docker image
docker build -t SERVICE_NAME .
```

### Frontend
```bash
# Start dev server
npm run dev

# Build production
npm run build

# Run tests
npm test
```

### Infrastructure
```bash
# Start local services
docker-compose up

# Deploy to Kubernetes
kubectl apply -f k8s/
```

---

## Quick Troubleshooting

| Issue | Solution |
|-------|----------|
| "What phase are we in?" | Check `DEVELOPMENT-STATUS.md` |
| "What should I work on?" | Check `DEVELOPMENT-STATUS.md` for your agent |
| "Where's the API spec?" | See `openapi.yaml` for REST, `api/*.proto` for gRPC |
| "Why was X decided?" | Search `TECHNICAL-DECISIONS.md` for TD-XXX |
| "How do I hand off?" | Follow `AGENT-HANDOFF.md` protocol |
| "Need full context?" | Read `SESSION-RECOVERY.md` |
| "What's been completed?" | Check `DEVELOPMENT-STATUS.md` "Completed" section |
| "API implementation status?" | Check `API-STATUS.md` |

---

## Critical Information

### MVP Authentication
- **Username**: `aezi`
- **Password**: `Aa@123456789`
- **Approach**: Hardcoded credential check at API Gateway
- **Note**: JWT architecture scaffolded but disabled for MVP

### Test Coverage Targets
- Backend services: ≥80% (≥90% for auth)
- Frontend: ≥70%

### External API Rate Limits
- See `external-api-integrations.md` for details
- Token bucket algorithm via Redis
- Cost tracking and budget monitoring required

### Data Flow (Simplified)
1. User searches wallet → Frontend (F)
2. Request → API Gateway (B)
3. Gateway → portfolio-service (C) via gRPC
4. Portfolio service enriches with prices from market-data-service (E)
5. Response → Frontend
6. Background: ingestion-service (D) fetches blockchain data → Kafka → portfolio-service (C)

---

## Next Steps (New Agent Onboarding)

1. **Read this file** (you're done! ✅)
2. **Read SESSION-RECOVERY.md** for full context restoration process
3. **Check DEVELOPMENT-STATUS.md** for current progress
4. **Review AGENT-ASSIGNMENTS.md** for your role
5. **Check API-STATUS.md** for endpoint status
6. **Set up dev environment** (see `dev-environment-setup.md`)
7. **Start coding!** 🚀

---

## Documentation Links

**Essential**:
- [SESSION-RECOVERY.md](./SESSION-RECOVERY.md) - How to restore context
- [DEVELOPMENT-STATUS.md](./DEVELOPMENT-STATUS.md) - Current progress
- [AGENT-ASSIGNMENTS.md](./AGENT-ASSIGNMENTS.md) - Team structure
- [API-STATUS.md](./API-STATUS.md) - Implementation tracking
- [TECHNICAL-DECISIONS.md](./TECHNICAL-DECISIONS.md) - Decision log

**Product & Architecture**:
- [PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](./PRD-Million-Dollar-Hunter-Crypto-Dashboard.md) - Product requirements
- [Technical Development Plan.md](./Technical%20Development%20Plan.md) - Technical architecture
- [million-hunter-development-plan.md](./million-hunter-development-plan.md) - Multi-agent plan

**Specifications**:
- [openapi.yaml](./openapi.yaml) - REST API spec
- [architecture-decisions.md](./architecture-decisions.md) - Architecture patterns
- [database-migration-strategy.md](./database-migration-strategy.md) - Schema management

**Operations**:
- [AGENT-HANDOFF.md](./AGENT-HANDOFF.md) - Handoff protocol
- [monitoring-alerting.md](./monitoring-alerting.md) - Observability
- [performance-requirements.md](./performance-requirements.md) - SLOs
- [testing-strategy.md](./testing-strategy.md) - Test requirements

---

## Success Metrics

### Technical Robustness
- All core services functional
- Minimal outages
- SLO targets met

### Maintainability
- Dashboard reconfigurable in <30 minutes
- No code lock-in
- Modular architecture

### Deployment Speed
- New features deployable in real-time
- Autonomous agent delivery
- Minimal manual intervention

---

**Remember**: This is a living document. Update when project structure or key information changes.

**Last Updated**: 2025-10-09  
**Maintained By**: All agents

---

*Goal: Get any agent up to speed in 2 minutes with this document.*
