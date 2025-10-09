# External API Integrations

Providers: Alchemy, Moralis, CoinGecko.

## Rate Limiting & Quotas
- Central limiter at API Gateway; per-provider client-side limiters.
- Token Bucket via Redis. Expose X-RateLimit-* headers to clients.
- Budget dashboards per provider with daily/weekly quotas.

### Rate Limiting Algorithm (Token Bucket)
- Redis keys: ratelimit:{provider}:{bucket} with fields tokens (int), last_refill_ts (epoch seconds).
- Refill: provider-specific rate (e.g., Alchemy 20 rps) with max capacity = burst size.
- Decrement tokens per request; when tokens exhausted, respond 429 with Retry-After.
- Separate buckets by critical route if needed (e.g., transactions vs holders).

## Fallbacks
- Prefer primary Alchemy for transfers; fallback to Moralis on 5xx/timeouts after exponential backoff.
- Cache last known good prices for CoinGecko; serve stale with warning if provider unavailable.
- Circuit breaker at service layer to prevent cascading failures.

## Cost Estimation & Budget Tracking
- Estimate cost per call from provider plans; maintain monthly budget caps.
- Record usage metrics (requests, cache hits, retries) to Prometheus; alert at 80/90/100% of budget.

### Monthly Budget Limits (MVP)
- Free APIs: No cap (e.g., CoinGecko free tier).
- Paid APIs (e.g., Alchemy, Moralis): Caps are user-managed via environment variables:
  - BUDGET_CAP_ALCHEMY_USD, BUDGET_CAP_MORALIS_USD (optional). If unset, alerts use forecast-only, no enforced cap.

### Cost Tracking Implementation
- Redis counter: api:budget:{provider}:{YYYY-MM}
- Prometheus metric: external_api_estimated_cost{provider="alchemy"}
- Alerts: thresholds at 80%, 90%, 100% of monthly cap (if cap provided); otherwise alert on forecast crossing $ thresholds configured via env (e.g., BUDGET_WARN_USD, BUDGET_CRIT_USD)

## Data Freshness
- Prices TTL: 60s; token holders refresh: 5–10 minutes configurable.
- Mark responses with lastUpdated timestamps; UI surfaces freshness indicators.

## Mocking for Development
- Provide WireMock mappings for key endpoints; toggle via ENV (USE_API_MOCKS=true).
- Seed canned responses and error cases for reproducible tests.
- Contract tests ensure mock parity with prod schemas.

## Security & Keys
- Store keys via sops + age and/or CI secrets; never commit secrets.
- Scope keys to least privilege; rotate quarterly or on leak suspicion.
