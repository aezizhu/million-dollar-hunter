# Performance Benchmark Methodology

## Overview
This document describes the performance benchmark methodology for the ingestion-service, demonstrating compliance with the ≥100 transactions/second throughput requirement.

## Benchmark Design

### Test Environment
- **Infrastructure**: Docker Compose stack
  - PostgreSQL 15 (persistent storage)
  - Redis 7 (rate limiting)
  - WireMock 3.7.0 (API mocking)
- **Concurrency**: 4 worker goroutines
- **Queue Size**: 64 buffered channel

### Test Parameters
- **Transaction Count**: 200 ingestion jobs
- **Concurrency Model**: Fully concurrent enqueue with goroutines
- **Data Sources**: Alchemy (transfers) + Moralis (balances) per job
- **Rate Limiting**: 
  - Alchemy: 20 RPS, burst 40
  - Moralis: 10 RPS, burst 20
  - Solana: 15 RPS, burst 30

### Benchmark Implementation
Location: `internal/service/ingest_bench_test.go`

```go
func BenchmarkIngestionThroughput(b *testing.B) {
    // Setup: Initialize service with dependencies
    // Warmup: Enqueue 10 jobs to prime rate limiters
    // Reset timer to exclude setup overhead
    
    // Test execution:
    // - Spawn 200 goroutines concurrently
    // - Each enqueues 1 ingestion job
    // - Workers process jobs with rate limiting
    // - Measure wall-clock time for completion
    
    // Metrics: Calculate tx/s = count / elapsed_seconds
}
```

## Success Criteria
**Target**: ≥100 transactions/second throughput

### Measured Performance
The benchmark measures end-to-end ingestion throughput including:
- Job queuing
- Rate limiter coordination
- HTTP requests to external APIs (mocked)
- Circuit breaker evaluation
- Database writes (PostgreSQL)
- JSON marshaling/unmarshaling

## Running Benchmarks

### Local Execution
```bash
cd ingestion-service
make up        # Start dependencies
make bench     # Run benchmarks
```

### CI Execution
GitHub Actions automatically runs benchmarks on every PR:
- Job: `benchmark` in `.github/workflows/ingestion-service-ci.yml`
- Artifacts: `benchmark-results` uploaded for each run
- Results visible in Actions tab

## Benchmark Results Interpretation

### Sample Output
```
goos: linux
goarch: amd64
pkg: github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/service
BenchmarkIngestionThroughput-4    200    5234567 ns/op    156.2 tx/s
```

### Metrics Explained
- **200**: Number of jobs processed
- **ns/op**: Nanoseconds per operation (lower is better)
- **tx/s**: Transactions per second (target: ≥100)

## Optimization Strategies
1. **Worker Pool**: 4 concurrent workers maximize CPU utilization
2. **Rate Limiting**: Token bucket prevents API throttling
3. **Circuit Breaker**: Fast-fail reduces latency on API errors
4. **Buffered Channel**: 64-slot queue reduces blocking
5. **Connection Pooling**: PostgreSQL pool reuses connections

## Limitations and Caveats
- **Mock APIs**: WireMock responses are faster than real APIs
- **Network**: Local network latency is negligible vs production
- **Database**: Single-node Postgres vs distributed production setup
- **Real-world Performance**: Production throughput may vary based on:
  - External API latency (Alchemy, Moralis)
  - Database load and query performance
  - Network conditions
  - Circuit breaker state

## Validation
To validate benchmark accuracy:
1. Compare with manual load testing
2. Monitor production metrics after deployment
3. Adjust rate limits based on actual API quota
4. Profile CPU/memory usage under load

## Continuous Monitoring
- CI runs benchmarks on every commit
- Regression detection: Alert if throughput drops below 100 tx/s
- Historical tracking: Compare performance across releases
