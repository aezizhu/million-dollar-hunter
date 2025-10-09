# Operational Runbook

## Health Checks
- Liveness: process up; Readiness: DB/Redis/connectivity checks.
- Verify /metrics and critical endpoints return expected status.

## Troubleshooting
- Elevated 5xx: Inspect API Gateway logs (trace_id), check downstream circuit breakers.
- Latency spikes: Review Prometheus RED metrics, DB slow query logs, cache hit rates.
- External API failures: Check provider status, switch to fallback, increase cache TTL temporarily.

## Scaling
- Scale ingestion-service first during catch-up backfills.
- Add Redis memory/bandwidth when cache hit rate < 80%.
- Portfolio-service HPA on p95 latency > SLO for 10m.

## Backups & DR
- Nightly PostgreSQL snapshots; 7-day retention; weekly offsite copy.
- Redis persistence optional; reconstructable from DB and providers.
- DR test quarterly; recovery runbook validated.

## Incident Response
- Sev1: Data loss or prolonged outage; page immediately; rollback recent deploy; fail open with cached data.
- Sev2: Elevated error rate or latency; throttle requests; enable circuit breakers.
- Comms: Owner notified via preferred channel; incident doc with timeline and actions.

## Change Management
- Rolling deploys; feature flags for risky changes.
- Pre-deploy smoke tests; post-deploy watch for 30 minutes.
