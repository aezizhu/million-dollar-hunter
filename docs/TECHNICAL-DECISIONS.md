# Technical Decisions Log - Million Dollar Hunter

**Purpose**: This document records all significant technical decisions made during the development of Million Dollar Hunter, including rationale, alternatives considered, and impact on the system.

**Last Updated**: 2025-10-09  
**Format**: Newest decisions at the top

---

## Decision Log

### TD-001: Microservices Architecture
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Adopt a microservices architecture with 5 core services (auth, api-gateway, portfolio, ingestion, market-data).

**Rationale**:
- Independent scalability: Data ingestion service has highly variable loads
- Technological flexibility: Each service can use optimal tech for its task
- Fault isolation: Failure of one service doesn't crash entire system
- Team autonomy: Small teams can own services independently
- Clear domain boundaries align with business capabilities

**Alternatives Considered**:
1. **Monolithic Architecture**: Rejected due to tight coupling, difficult to scale specific components
2. **Serverless Functions**: Rejected due to complexity of managing state and long-running processes
3. **Modular Monolith**: Rejected due to inability to scale components independently

**Impact**:
- Increased operational complexity (distributed system challenges)
- Need for service discovery, API gateway, and event streaming
- Benefits outweigh complexity for this domain

**Related Components**: All services

---

### TD-002: Go as Backend Language
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Go 1.21+ as the primary backend language for all microservices.

**Rationale**:
- High performance with compiled binaries
- First-class concurrency support via goroutines (critical for data ingestion)
- Strong standard library reduces external dependencies
- Efficient resource usage (important for cost optimization)
- Good ecosystem for microservices (gRPC, Kafka clients)
- Static typing reduces runtime errors

**Alternatives Considered**:
1. **Python**: Rejected due to performance concerns for high-throughput ingestion
2. **Node.js**: Rejected due to less mature concurrency primitives
3. **Rust**: Rejected due to steeper learning curve and longer development time
4. **Java/Kotlin**: Rejected due to higher resource consumption

**Impact**:
- All backend agents (A, B, C, D, E) must use Go
- Need golang-migrate for database migrations
- Use golang-jwt/jwt for authentication

**Related Components**: auth-service, api-gateway, portfolio-service, ingestion-service, market-data-service

---

### TD-003: Next.js with App Router for Frontend
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Next.js 15+ with App Router for the frontend application.

**Rationale**:
- Server-side rendering improves initial page load and SEO
- App Router provides modern React patterns (Server Components)
- Built-in routing reduces boilerplate
- Excellent TypeScript support
- Strong ecosystem and community
- Easy deployment options (Vercel, static hosting)

**Alternatives Considered**:
1. **Create React App**: Rejected due to lack of SSR and being deprecated
2. **Vite + React**: Rejected due to need for SSR capabilities
3. **Vue.js/Nuxt**: Rejected due to team familiarity and ecosystem size
4. **Remix**: Rejected due to less mature ecosystem

**Impact**:
- Agent F responsible for Next.js implementation
- Need to configure Material-UI for App Router
- TradingView charts require client-side only components

**Related Components**: Frontend application (Agent F)

---

### TD-004: PostgreSQL Per-Service Database Pattern
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Each microservice has its own PostgreSQL 15+ database with exclusive schema ownership. No cross-service foreign keys.

**Rationale**:
- Database-per-service is fundamental to microservices pattern
- Enforces loose coupling between services
- Allows independent scaling and optimization per service
- Enables schema evolution without cascading changes
- Supports different data models per domain (CQRS)

**Alternatives Considered**:
1. **Shared Database**: Rejected due to tight coupling and scaling issues
2. **NoSQL per Service**: Rejected due to need for relational integrity in financial data
3. **Mixed SQL/NoSQL**: Rejected to keep operational complexity manageable

**Impact**:
- 5 separate PostgreSQL databases (one per backend service)
- Services communicate via APIs, not database joins
- Need distributed transaction patterns (Saga)
- golang-migrate for schema versioning per service

**Related Components**: All backend services (separate DBs for each)

---

### TD-005: CQRS Pattern for Data Management
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Implement CQRS (Command Query Responsibility Segregation) with ingestion-service as write model and portfolio-service as read model.

**Rationale**:
- Fundamentally different workloads: write-heavy ingestion vs read-heavy queries
- Write model can optimize for throughput and idempotency (raw JSONB storage)
- Read model can optimize for query performance (denormalized views)
- Separation allows independent scaling of read and write paths
- Enables reprocessing of raw data without re-fetching from external APIs

**Alternatives Considered**:
1. **Single Service with Dual Schema**: Rejected due to coupling and operational complexity
2. **Event Sourcing**: Rejected as overkill for MVP requirements
3. **Traditional CRUD**: Rejected due to conflicting optimization requirements

**Impact**:
- Agent D owns write model (ingestion-service)
- Agent C owns read model (portfolio-service)
- Kafka event streaming connects write to read
- Raw data stored as JSONB for reprocessing capability

**Related Components**: ingestion-service (write), portfolio-service (read), Kafka

---

### TD-006: Apache Kafka for Event Streaming
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Apache Kafka as the message broker for asynchronous service communication and event-driven workflows.

**Rationale**:
- Decouples services for better resilience
- Enables event-driven architecture (CQRS, Saga patterns)
- Provides message buffering for variable load handling
- Supports replay capability for reprocessing
- Scales horizontally with partitioning
- Industry standard with strong Go client support

**Alternatives Considered**:
1. **RabbitMQ**: Rejected due to less robust horizontal scaling
2. **Redis Pub/Sub**: Rejected due to lack of persistence and replay
3. **AWS SQS/SNS**: Rejected to avoid cloud vendor lock-in for MVP
4. **Direct gRPC Calls**: Rejected due to tight coupling for async workflows

**Impact**:
- Agent C responsible for Kafka broker deployment
- Event schemas must be clearly defined
- Key events: WalletTrackingRequested, TransactionDataIngested, PortfolioUpdated
- Need consumer group management for scalability

**Related Components**: Agent C (infrastructure owner), Agents D/E (producers), Agent C (consumer)

---

### TD-007: Redis for Caching and Rate Limiting
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Redis 7+ for price caching (Agent E) and rate limiting (Agent B).

**Rationale**:
- In-memory performance critical for cache hit rate target (≥80%)
- Token bucket rate limiting requires fast atomic operations
- Simple data structures fit Redis perfectly
- Lower operational complexity than distributed cache alternatives
- Supports TTL for automatic expiration (price cache 60s)

**Alternatives Considered**:
1. **Memcached**: Rejected due to less feature-rich and no persistence
2. **In-Process Cache**: Rejected due to inability to share across instances
3. **Database Caching**: Rejected due to performance limitations

**Impact**:
- Agent E owns Redis deployment and price caching
- Agent B uses Redis for rate limit state (token bucket)
- Shared infrastructure requires key namespace coordination
- Need monitoring for cache hit rates

**Related Components**: market-data-service (Agent E primary), api-gateway (Agent B secondary)

---

### TD-008: JWT-Based Authentication (Scaffolded for MVP)
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved (Partially Disabled for MVP)

**Decision**: Implement JWT-based stateless authentication architecture, but use hardcoded credentials for MVP single-user mode.

**Rationale**:
- JWT provides stateless authentication for microservices
- No session state reduces infrastructure complexity
- golang-jwt/jwt v5+ is industry standard
- Scaffolding now enables easy multi-user upgrade later
- MVP hardcoded auth (username: `aezi`, password: `Aa@123456789`) sufficient for single user

**Alternatives Considered**:
1. **Session-Based Auth**: Rejected due to state management complexity
2. **OAuth 2.0**: Rejected as overkill for MVP
3. **No Auth**: Rejected due to security requirements

**Impact**:
- Agent A implements JWT generation/validation
- Agent B integrates JWT middleware
- MVP: Simple hardcoded credential check at gateway
- Post-MVP: Enable full JWT flow for multi-user

**Related Components**: auth-service (Agent A), api-gateway (Agent B)

---

### TD-009: Material-UI for Frontend Design System
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Material-UI (MUI) as the React component library and design system.

**Rationale**:
- Comprehensive component library reduces development time
- Consistent, professional design out of the box
- Excellent TypeScript support
- Customizable theming system
- Strong Next.js integration documentation
- Accessible components (WCAG 2.1 AA compliance)

**Alternatives Considered**:
1. **Tailwind CSS**: Rejected due to need for pre-built components
2. **Ant Design**: Rejected due to less modern React patterns
3. **Chakra UI**: Rejected due to smaller ecosystem

**Impact**:
- Agent F uses MUI for all UI components
- Need to configure AppRouterCacheProvider for Next.js
- Custom theme defined in src/theme.ts

**Related Components**: Frontend (Agent F)

---

### TD-010: TradingView Lightweight Charts for Data Visualization
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use TradingView Lightweight Charts library for financial data visualization.

**Rationale**:
- Purpose-built for financial charting
- High performance with large datasets
- Small footprint (lightweight)
- Interactive features (zoom, pan, crosshairs)
- Responsive design
- Free and open source

**Alternatives Considered**:
1. **Recharts**: Rejected due to performance issues with large datasets
2. **Chart.js**: Rejected due to lack of financial-specific features
3. **D3.js**: Rejected due to development complexity

**Impact**:
- Agent F integrates charts
- Requires client-side only rendering (dynamic import with ssr: false)
- Need to manage chart lifecycle (cleanup on unmount)

**Related Components**: Frontend (Agent F)

---

### TD-011: Alchemy as Primary Blockchain Data Provider
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use Alchemy as the primary blockchain data provider with Moralis as fallback.

**Rationale**:
- Comprehensive transaction history via alchemy_getAssetTransfers
- Single endpoint for multiple transfer types (ERC20, ERC721, ERC1155)
- Efficient pagination with pageKey
- Multi-chain support (Ethereum, BSC, Polygon, Arbitrum, Optimism)
- Reliable infrastructure and uptime

**Alternatives Considered**:
1. **Self-Hosted Node**: Rejected due to operational complexity and cost
2. **Moralis Primary**: Considered equivalent but Alchemy has better pagination
3. **Multiple Providers**: Will implement as fallback strategy

**Impact**:
- Agent D integrates Alchemy API
- Moralis provides redundancy and Solana support
- Need rate limit management and cost tracking
- Fallback logic: Alchemy → Moralis on failures

**Related Components**: ingestion-service (Agent D)

---

### TD-012: OpenAPI (Swagger) for API Specification
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Define all public REST API endpoints in openapi.yaml specification.

**Rationale**:
- Single source of truth for API contracts
- Enables contract testing
- Auto-generates documentation
- Enables code generation for clients
- Industry standard format

**Alternatives Considered**:
1. **No Formal Spec**: Rejected due to coordination issues
2. **Protocol Buffers Only**: Rejected as gRPC is internal only
3. **API Blueprint**: Rejected due to smaller ecosystem

**Impact**:
- Agent B owns openapi.yaml
- All agents reference spec for API contracts
- Frontend (Agent F) uses spec for API integration
- Contract tests validate implementation matches spec

**Related Components**: api-gateway (Agent B owns), all agents reference

---

### TD-013: golang-migrate for Database Migrations
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Initial Architecture Team  
**Status**: Approved

**Decision**: Use golang-migrate for database schema versioning and migrations.

**Rationale**:
- Widely adopted in Go ecosystems
- Simple CLI and library usage
- Supports PostgreSQL fully
- Reversible migrations (up/down)
- Works well with containers and CI/CD
- Idempotent execution prevents duplicate runs

**Alternatives Considered**:
1. **Goose**: Rejected due to less active maintenance
2. **Manual SQL Scripts**: Rejected due to lack of version tracking
3. **ORM Migrations**: Rejected to avoid ORM dependency

**Impact**:
- All backend agents use golang-migrate
- Migrations stored per service in db/migrations/
- Numbered files for ordering (e.g., 001_create_users.up.sql)
- Applied on container startup and in CI

**Related Components**: All backend services

---

### TD-014: Single-User MVP with Hardcoded Credentials
**Date**: 2025-10-09 (Planning Phase)  
**Decided By**: Product Team  
**Status**: Approved (MVP Only)

**Decision**: MVP operates in single-user mode with hardcoded credentials (username: `aezi`, password: `Aa@123456789`).

**Rationale**:
- Reduces MVP complexity (no user management, registration, profiles)
- Single owner/operator use case doesn't require multi-user
- Faster time to market
- JWT architecture prepared for future multi-user upgrade
- No GDPR compliance burden for MVP

**Alternatives Considered**:
1. **Full Multi-User from Start**: Rejected as unnecessary for MVP
2. **No Authentication**: Rejected due to security requirements
3. **Environment Variable Password**: Rejected for MVP simplicity

**Impact**:
- All agents implement/support hardcoded auth check
- Agent A scaffolds JWT but disabled in MVP production
- Post-MVP: Enable JWT flow for multi-user
- Agent F implements simple login form (hardcoded validation)

**Related Components**: All services (authentication flow)

---

## Decision Template

Use this template when adding new decisions:

```markdown
### TD-XXX: [Decision Title]
**Date**: YYYY-MM-DD  
**Decided By**: [Agent X / Team / Role]  
**Status**: [Proposed / Approved / Rejected / Superseded]

**Decision**: [Clear statement of what was decided]

**Rationale**:
- [Key reason 1]
- [Key reason 2]
- [Key reason 3]

**Alternatives Considered**:
1. **[Alternative 1]**: [Why rejected]
2. **[Alternative 2]**: [Why rejected]

**Impact**:
- [Impact on architecture]
- [Impact on agents/teams]
- [Impact on timeline]

**Related Components**: [List of affected services/agents]
```

---

## Update Instructions

1. Assign next sequential TD number (TD-XXX)
2. Add new decision at **top** of Decision Log section
3. Use decision template above
4. Update "Last Updated" date
5. Reference decision in DEVELOPMENT-STATUS.md if relevant
6. Commit with message: "docs: Add technical decision TD-XXX: [title]"

---

*This log captures the "why" behind technical choices to maintain context across sessions and agent transitions.*
