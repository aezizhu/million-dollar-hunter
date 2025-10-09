# Testing Strategy

## Coverage Targets
- Backend services: >= 80% line coverage; critical packages >= 90%.
- Frontend: >= 70% statements/branches, with emphasis on critical flows.

## Test Types
- Unit Tests: Pure logic, handlers, mappers.
- Integration Tests: DB interactions using Testcontainers; gRPC/REST boundaries.
- Contract Tests: API Gateway OpenAPI-based validation for request/response schemas.
- End-to-End (E2E): Login gate, search token, view wallet, export CSV.
- Performance Tests: API p50/p95/p99 latency and throughput; cache hit rates.

## Scenarios (Integration/E2E)
- Auth Gate: Valid vs invalid credential; unauthorized access returns 401.
- Portfolio List: Pagination, empty state, large datasets.
- Wallet View: Price enrichment; fallback to cached prices.
- Transactions: Pagination, filtering; API rate limit handling (429).
- Top Holders: Data freshness thresholds and error handling when external APIs fail.

## Load & Perf
- Tools: k6 or Locust; backend profiling with pprof.
- Benchmarks: See performance-requirements.md for SLOs and targets.
- Warm caches before measurements; report cold vs warm results.

## Test Data
- Use factories and fixtures; seed minimal realistic datasets.
- Mock external APIs via WireMock/httpmock; provide recorded cassettes for determinism.
- Anonymize any captured samples.

## CI Integration
- Lint, unit, integration, and contract tests on PR.
- E2E against docker-compose in CI.
- Upload coverage artifacts and fail build if below thresholds.
