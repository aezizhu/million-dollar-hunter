# Monitoring & Alerting

## SLOs
- API Gateway: Availability 99.9%; p95 latency <= 300ms; error rate < 1%.
- Portfolio-service: p95 query <= 200ms; cache hit >= 80%.
- Ingestion-service: Ingestion lag <= 5 minutes for active wallets.
- Market-data-service: Price freshness <= 60s.

## Metrics
- RED per endpoint; DB query durations; cache hit/miss; external API call outcomes.
- Rate limit counters and rejections.

## Alerts
- Sev1: Availability < 99% over 1h; ingestion lag > 30m; DB saturation > 90%.
- Sev2: p95 latency > SLO for 10m; 5xx rate > 5% for 10m.
- Budget: External API cost forecast > 90% of monthly cap.

## Dashboards
- Gateway: traffic, error rate, latency histograms, rate limits.
- Services: per-service RED, DB panels, cache panels.
- Integrations: provider availability, error codes, backoff retries.

## On-call & Escalation
- Agent-based monitoring with automated remediation (cache TTL bump, breaker trip).
- Owner escalation if Sev1 persists beyond 15 minutes.
