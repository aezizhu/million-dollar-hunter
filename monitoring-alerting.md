# Monitoring & Alerting

## SLOs
- API p95 latency: 300ms (MVP), 200ms (stretch)
- Error rate: < 1% 5xx
- Uptime: 99.5%

## Dashboards
- RED metrics per service (Requests, Errors, Duration)
- External API usage and estimated cost per provider
- Cache hit rate, DB slow queries

## Alerts
- p95 latency > SLO for 10m (warning), 20m (critical)
- 5xx > 5% for 5m (warning), >10% for 5m (critical)
- External API failures: sustained 5xx/timeouts from provider
- Budget tracking:
  - Free APIs: No cap; alert only on provider outages or anomalous spikes.
  - Paid APIs: Caps are user-managed via env (e.g., BUDGET_CAP_ALCHEMY_USD). Alert at 80/90/100% if cap set; otherwise alert on forecast thresholds (BUDGET_WARN_USD, BUDGET_CRIT_USD).

## On-call & Escalation
- Notify owner via configured channel
- Page on Sev1: prolonged outage, data loss, or budget overrun (if cap set)



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
## External API Cost Alerts
- Metric: external_api_estimated_cost by provider
- Thresholds: 80% (warning), 90% (critical), 100% (page)
- Dashboard: monthly budget burn-down with projected end-of-month forecast
