# Requirements Traceability

Links PRD requirements to Technical Plan sections, supporting docs, and API/specs.

| PRD Requirement | Tech Plan Section | Supporting Doc | API/Schema |
|---|---|---|---|
| Single-user access gate | MVP vs Future State: Authentication | security-hardening.md | Gateway hardcoded check (no API) |
| Export CSV/JSON | Implementation Roadmap + Frontend | frontend-components.md | /api/v1/export/wallet/{address} |
| Top holders history | Ingestion Responsibilities | database-migration-strategy.md | holder_snapshots schema |
| Rate limits and budget | System Observability/Integrations | external-api-integrations.md | X-RateLimit-* headers, 429 |
| Performance SLOs | Observability + Perf | performance-requirements.md, testing-strategy.md | n/a |
| Secrets management | Security | security-hardening.md | n/a |
| Monitoring & alerts | Observability | monitoring-alerting.md | /metrics (internal) |
