# Documentation Index

- [PRD - Crypto Dashboard](./PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)
- [Technical Development Plan](./Technical Development Plan.md)
- [Security Hardening](./security-hardening.md)
- [Database Migration Strategy](./database-migration-strategy.md)
- [Auth Service](./auth-service.md)

# Documentation Index

Purpose: Quick navigation and guidance for when to use each document.

## Index
- PRD-Million-Dollar-Hunter-Crypto-Dashboard.md — Product scope, goals, timelines, non-goals.
- Technical Development Plan.md — Architecture, services, data, security, delivery.
- openapi.yaml — Public REST API specification.
- database-migration-strategy.md — Migrations tool, versioning, rollback, zero-downtime.
- testing-strategy.md — Coverage, test types, scenarios, CI, performance.
- external-api-integrations.md — Providers, rate limits, fallbacks, costs, freshness, mocking.
- operational-runbook.md — Health checks, troubleshooting, scaling, DR, incidents, change mgmt.
- monitoring-alerting.md — SLOs, alerts, dashboards, escalation.
- dev-environment-setup.md — Local dev with Docker Compose, IDE, local k8s, mocks.
- architecture-decisions.md — Key ADRs (microservices, Go, PostgreSQL, Next.js App Router).
- security-hardening.md — Secrets, network, container/dependency scanning, encryption.
- performance-requirements.md — API and frontend budgets, DB/query thresholds.
- data-privacy-retention.md — Stored data, retention, export formats, deletion.
- frontend-components.md — Component usage and props.
- requirements-traceability.md — Map PRD requirements to Tech Plan, docs, and API/specs.
- AGENT-HANDOFF.md — Protocol for handoffs: required artifacts, validations, and notifications.

## Decision Tree
- Need system behavior or scope? Read PRD.
- Need implementation details or schemas? Read Technical Development Plan.
- Need an API contract? Read openapi.yaml.
- Changing DB schema? Read database-migration-strategy.md.
- Troubleshooting production issues? Read operational-runbook.md and monitoring-alerting.md.
- Estimating or testing performance? Read performance-requirements.md and testing-strategy.md.
- Integrating providers or handling rate limits/costs? Read external-api-integrations.md.
- Tracing a requirement to implementation/spec? Read requirements-traceability.md.
- Completing or reviewing an agent task? Read AGENT-HANDOFF.md.

## Versioning
- Each document includes a last-updated timestamp in Git history. When making substantial changes, add a “Changelog” section noting the update in the file.
