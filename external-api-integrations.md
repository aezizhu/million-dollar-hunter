# External API Integrations

Providers: Alchemy, Moralis, CoinGecko.

## Rate Limiting & Quotas
- Central limiter at API Gateway; per-provider client-side limiters.
- Redis counters with sliding window; expose X-RateLimit-* headers.
- Budget dashboards per provider with daily/weekly quotas.

## Fallbacks
- Prefer primary Alchemy for transfers; fallback to Moralis on 5xx/timeouts after exponential backoff.
- Cache last known good prices for CoinGecko; serve stale with warning if provider unavailable.
- Circuit breaker at service layer to prevent cascading failures.

## Cost Estimation & Budget Tracking
- Estimate cost per call from provider plans; maintain monthly budget caps.
- Record usage metrics (requests, cache hits, retries) to Prometheus; alert at 80/90/100% of budget.

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
