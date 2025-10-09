# Performance Requirements

## API Latency Targets
- Gateway p50 <= 100ms, p95 <= 300ms, p99 <= 600ms
- Portfolio endpoints p95 <= 200ms with warmed cache
- Transactions listing p95 <= 400ms (paginated)

## Frontend Budgets
- TTFB <= 300ms (SSR), LCP <= 2.5s on 3G Fast, CLS < 0.1
- JS bundle <= 250KB gzipped on dashboard route

## Database
- Queries p95 <= 50ms for read models; long scans avoided
- Index coverage for common filters; no sequential scans in hot paths

## Ingestion
- Data lag <= 5 minutes under normal load
- Backfill throughput >= 100 tx/s with batching

## Capacity & Throughput
- Support 1 active user with headroom for spikes; scale ingestion independently
- Cache hit rate >= 80% for price reads

## Error Budgets
- Availability 99.9%; monthly error budget <= 43m
