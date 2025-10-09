# Architecture Decision Records (ADRs)

## ADR-001: Microservices over Monolith
- Decision: Adopt microservices for clear bounded contexts and independent scaling.
- Status: Accepted
- Rationale: Aligns with ingestion-heavy workloads and resilience patterns.

## ADR-002: Go for Backend
- Decision: Use Go for backend services.
- Status: Accepted
- Rationale: Performance, concurrency, strong standard lib, container-friendly.

## ADR-003: PostgreSQL as Primary Store
- Decision: Use PostgreSQL for durable data.
- Status: Accepted
- Rationale: ACID guarantees, rich SQL, mature tooling.

## ADR-004: Next.js App Router
- Decision: Use Next.js App Router for frontend.
- Status: Accepted
- Rationale: SSR, RSC, scalable structure, great DX.
