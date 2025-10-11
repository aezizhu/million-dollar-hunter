# ingestion-service

Go microservice for blockchain data ingestion following CQRS write model.

## Features
- Alchemy asset transfers client with full pagination
- Moralis balances client with multi-chain support (BSC, Solana, ETH)
- Redis token-bucket rate limiting, circuit breaker
- Write-optimized Postgres schema: ingestion_jobs, raw_transactions, raw_balances, holder_snapshots
- Kafka producer for publishing TransactionDataIngested events
- WireMock setup for external API mocking
- Performance benchmark target: ≥100 transactions/second

## Run locally
1. Start dependencies:
   make up
2. Build and run:
   go build ./cmd/ingestion-service && ./ingestion-service
3. Health:
   curl localhost:8090/healthz

## Tests
- Ensure docker compose is up (postgres, redis, wiremock):
  make up
- Run tests:
  make test

## Benchmarks
- With mocks running:
  make bench

## Env
- DATABASE_URL=postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable
- REDIS_ADDR=localhost:6379
- ALCHEMY_BASE_URL=http://localhost:8080/alchemy
- ALCHEMY_API_KEY=test
- MORALIS_BASE_URL=http://localhost:8080/moralis
- MORALIS_API_KEY=test
- HTTP_PORT=8090
- USE_API_MOCKS=true
- KAFKA_BROKERS=localhost:9092
- KAFKA_TOPIC_TX_INGESTED=TransactionDataIngested
