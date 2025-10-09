# Agent Assignments - Quick Reference Guide

## Overview
This project uses **6 parallel AI engineer agents (A-F)** to develop the Million Dollar Hunter cryptocurrency dashboard platform simultaneously.

---

## Agent Roster

### **Agent A - Authentication & Security Engineer**
**Responsibility:** auth-service, JWT implementation, security hardening  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Goals, Security, Technical Considerations)
- `Technical Development Plan.md` (Section II.E: Security Architecture)
- `security-hardening.md` (Complete)
- `database-migration-strategy.md` (Schema migrations)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- auth-service (Go microservice)
- JWT middleware
- PostgreSQL users schema
- Security documentation
- Unit/integration tests (>90% coverage)

---

### **Agent B - API Gateway & Orchestration Engineer**
**Responsibility:** api-gateway service, REST API contracts, rate limiting  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Complete)
- `Technical Development Plan.md` (Section II: Backend Development)
- `openapi.yaml` (PRIMARY REFERENCE - Complete REST API spec)
- `external-api-integrations.md` (Rate limiting algorithms)
- `monitoring-alerting.md` (Observability requirements)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- api-gateway service (Go)
- OpenAPI specification
- Rate limiting implementation
- Logging/tracing infrastructure
- Prometheus metrics
- Load tests (k6/Locust)

---

### **Agent C - Portfolio & Aggregation Engineer**
**Responsibility:** portfolio-service, CQRS read models, Kafka infrastructure  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Token Analytics, Wallet Monitoring, Top Holders)
- `Technical Development Plan.md` (Section II.B: Microservices, Section II.D: Database Schemas)
- `database-migration-strategy.md` (Schema versioning)
- `architecture-decisions.md` (CQRS and Saga patterns)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- portfolio-service (Go microservice)
- Read-optimized PostgreSQL schemas
- Kafka broker setup
- Event consumer implementation
- Export functionality (CSV/JSON)
- gRPC service definitions

---

### **Agent D - Data Ingestion Engineer**
**Responsibility:** ingestion-service, Alchemy/Moralis integration, external APIs  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Functional Requirements, Cross-Chain Support)
- `Technical Development Plan.md` (Section II.B: ingestion-service, Section IV.A: External APIs)
- `external-api-integrations.md` (CRITICAL - Complete document for rate limits, fallbacks, cost tracking)
- `database-migration-strategy.md` (Write-optimized schemas)
- `architecture-decisions.md` (CQRS write model)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- ingestion-service (Go microservice)
- Alchemy API client (full pagination)
- Moralis API client (multi-chain)
- Rate limiting and circuit breaker
- WireMock testing setup
- Performance tests (≥100 tx/s throughput)

---

### **Agent E - Market Data Engineer**
**Responsibility:** market-data-service, CoinGecko integration, price caching  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Price Tracking, Data Export)
- `Technical Development Plan.md` (Section II.B: market-data-service, Section IV.A: CoinGecko)
- `external-api-integrations.md` (Data freshness, caching strategies)
- `database-migration-strategy.md` (Token prices schema)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- market-data-service (Go microservice)
- CoinGecko client
- Redis caching (60s TTL, ≥80% hit rate)
- gRPC service interface
- Background price refresh workers
- Load tests for concurrent requests

---

### **Agent F - Frontend & Visualization Engineer**
**Responsibility:** Next.js application, Material-UI, TradingView charts, all UI  
**Must Read Docs:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` (Complete - user stories, UX requirements)
- `Technical Development Plan.md` (Section III: Frontend Development - Complete)
- `frontend-components.md` (Complete - component library, props, design system)
- `openapi.yaml` (REST API contracts for frontend calls)
- `AGENT-HANDOFF.md` (Handoff protocol)

**Key Deliverables:**
- Complete Next.js application
- Material-UI design system
- TradingView chart integration
- Authentication flow (login/register)
- Dashboard and wallet analysis views
- TanStack Query setup
- Component tests (>70% coverage)
- Performance compliance (TTFB ≤300ms, LCP ≤2.5s, bundle ≤250KB)

---

## Development Timeline

### Phase 1 (Weeks 1-4): Foundation
**Active:** Agents A, B, C  
- Auth service + JWT
- API Gateway skeleton
- Kafka infrastructure
- Basic portfolio service

### Phase 2 (Weeks 5-8): Data Pipeline
**Active:** Agents D, E, C  
- Ingestion service (Alchemy/Moralis)
- Market data service (CoinGecko)
- Data aggregation pipeline
- Price enrichment

### Phase 3 (Weeks 9-12): Frontend
**Active:** Agents F, B  
- Next.js application
- Material-UI components
- Authentication pages
- Dashboard views
- API endpoint finalization

### Phase 4 (Weeks 13-16): Integration
**Active:** All Agents (A-F)  
- TradingView charts
- End-to-end testing
- Performance optimization
- Security hardening
- Production deployment

---

## Agent Coordination Rules

### Daily Practices:
1. **Morning Status**: Post what you completed, today's goals, and blockers
2. **Handoff Protocol**: Follow `AGENT-HANDOFF.md` when completing dependencies
3. **Documentation**: Update relevant specs when making architectural changes
4. **Testing**: Run full test suite before marking features complete

### Key Dependencies:
- **Agent A → Agent B**: JWT validation interface
- **Agent B → Agent F**: REST API surface
- **Agent C → Agent B**: gRPC service interfaces
- **Agent D → Agent C**: Kafka events (TransactionDataIngested)
- **Agent E → Agent C**: Price data via gRPC
- **Agent E → Agent B**: Market data endpoints

### Communication Channels:
- **Async**: GitHub PRs, comments, issues
- **Sync**: Scheduled coordination sessions for complex integrations
- **Escalation**: Owner/operator notified for blocking decisions

---

## Testing Standards

### Coverage Targets:
- **Backend (A, B, C, D, E)**: ≥80% line coverage (≥90% for auth/security)
- **Frontend (F)**: ≥70% statements/branches

### Test Types:
- Unit tests (all agents)
- Integration tests with Testcontainers (backend agents)
- Contract tests (Agent B for OpenAPI)
- E2E tests (Agent F with backend)
- Performance tests (Agent B: load, Agent D: throughput)

---

## MVP Credentials (Single-User Mode)
- **Username**: `aezi`
- **Password**: `Aa@123456789`
- All agents must implement/support this hardcoded authentication for MVP

---

## Success Criteria
- ✅ All services passing health checks (99.9% uptime)
- ✅ API p95 latency ≤ 300ms
- ✅ Frontend LCP ≤ 2.5s on 3G Fast
- ✅ Cache hit rate ≥ 80%
- ✅ Zero critical security vulnerabilities
- ✅ Test coverage targets met
- ✅ All documentation complete

---

## Quick Links
- **Main Plan**: `million-hunter-development-plan.md`
- **API Spec**: `openapi.yaml`
- **PRD**: `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md`
- **Tech Plan**: `Technical Development Plan.md`
- **Handoff Protocol**: `AGENT-HANDOFF.md`

---

*Last Updated: 2025-10-09*
