# Market Data Service

The market-data-service is a Go microservice responsible for fetching, caching, and serving cryptocurrency token price data for the Million Dollar Hunter crypto dashboard platform.

> *The caching strategy implemented here optimizes for both performance and cost efficiency, with careful attention to cache invalidation and refresh patterns to ensure data freshness while minimizing external API calls.*

## Overview

This service integrates with the CoinGecko API to fetch real-time token prices and market data across multiple blockchain networks (BSC, Solana, Ethereum). It implements a multi-layer caching strategy using Redis and PostgreSQL to achieve high performance and minimize external API calls.

## Key Features

- **CoinGecko Integration**: Fetches token prices and market data from CoinGecko API
- **Redis Caching**: 60-second TTL cache with ≥80% target hit rate
- **PostgreSQL Persistence**: Durable storage for price history
- **gRPC Interface**: High-performance service-to-service communication
- **Background Workers**: Automated price refresh for tracked tokens
- **Rate Limiting**: Token bucket algorithm to respect CoinGecko API limits
- **Multi-Chain Support**: BSC, Solana, Ethereum, and Polygon

## Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ gRPC
       ▼
┌─────────────────┐
│  gRPC Handler   │
└────┬─────┬──────┘
     │     │
     │     └─────────────┐
     ▼                   ▼
┌─────────┐       ┌──────────────┐
│  Redis  │       │  PostgreSQL  │
│  Cache  │       │  Repository  │
└─────────┘       └──────────────┘
     │                   │
     └──────┬────────────┘
            ▼
    ┌──────────────┐
    │  CoinGecko   │
    │   Client     │
    └──────────────┘
            │
            ▼
    ┌──────────────┐
    │  Background  │
    │   Worker     │
    └──────────────┘
```

## API Endpoints (gRPC)

### GetTokenPrice
Retrieves the current USD price for a single token.

**Request:**
```protobuf
message GetTokenPriceRequest {
  string token_address = 1;
  string chain = 2;
}
```

**Response:**
```protobuf
message GetTokenPriceResponse {
  string token_address = 1;
  string chain = 2;
  double usd_price = 3;
  int64 last_updated = 4;
  bool from_cache = 5;
}
```

### GetTokenPrices
Retrieves prices for multiple tokens in a single request.

### GetMarketData
Retrieves comprehensive market data including market cap, volume, and price changes.

### RefreshTokenPrice
Forces an immediate refresh of a token's price from CoinGecko.

## Configuration

Configuration is managed via environment variables. See `.env.example` for all available options.

### Key Configuration Parameters

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_PORT` | 50051 | gRPC server port |
| `REDIS_TTL` | 60s | Cache TTL duration |
| `COINGECKO_RATE_LIMIT` | 50 | Requests per minute |
| `WORKER_REFRESH_INTERVAL` | 30s | Background refresh interval |
| `WORKER_BATCH_SIZE` | 50 | Tokens per batch |

## Development Setup

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- Protocol Buffers compiler (protoc)

### Local Development

1. **Clone the repository:**
   ```bash
   cd services/market-data-service
   ```

2. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

3. **Start dependencies with Docker Compose:**
   ```bash
   docker-compose up -d postgres redis
   ```

4. **Run database migrations:**
   ```bash
   docker-compose up migrate
   ```

5. **Install Go dependencies:**
   ```bash
   go mod download
   ```

6. **Generate gRPC code from proto files:**
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          api/proto/market_data.proto
   ```

7. **Run the service:**
   ```bash
   go run cmd/market-data-service/main.go
   ```

### Running with Docker Compose

To run the entire stack (including the service):

```bash
docker-compose up --build
```

## Testing

### Unit Tests
```bash
go test ./internal/... -v
```

### Integration Tests
Requires running PostgreSQL and Redis:
```bash
docker-compose up -d postgres redis
go test ./tests -v -run Integration
```

### Load Tests
```bash
docker-compose up -d
go test ./tests -v -run Load
```

Expected load test results:
- **RPS**: >100 requests per second
- **Cache Hit Rate**: ≥80%
- **Average Latency**: <300ms

### Benchmarks
```bash
go test ./tests -bench=. -benchmem
```

## Database Schema

### token_prices

| Column | Type | Description |
|--------|------|-------------|
| token_address | TEXT | Contract address (PK) |
| chain | TEXT | Blockchain identifier (PK) |
| usd_price | NUMERIC(20,8) | Current USD price |
| market_cap | NUMERIC(20,2) | Market capitalization |
| volume_24h | NUMERIC(20,2) | 24-hour trading volume |
| price_change_24h | NUMERIC(10,4) | 24-hour price change % |
| last_updated | TIMESTAMPTZ | Last update timestamp |

## Caching Strategy

The service implements a three-tier caching strategy:

1. **Redis Cache (L1)**: 60-second TTL, serves most requests
2. **PostgreSQL (L2)**: Persistent storage, fallback for cache misses
3. **CoinGecko API (L3)**: Source of truth, fetched on cache + DB miss

### Cache Hit Rate Optimization

To achieve ≥80% cache hit rate:
- Background worker refreshes tracked tokens every 30s
- Popular tokens remain in cache continuously
- Batch requests reduce individual API calls

## Monitoring & Observability

### Structured Logging
All logs are output in JSON format using zerolog:
```json
{
  "level": "info",
  "service": "market-data-service",
  "component": "coingecko_client",
  "token": "0x...",
  "chain": "bsc",
  "price": 1.0002,
  "message": "Successfully fetched token price"
}
```

### Key Metrics
- Request count by endpoint
- Cache hit/miss rates
- API latency percentiles (p50, p95, p99)
- External API call counts
- Error rates

## Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Cache Hit Rate | ≥80% | Measured over 5-minute windows |
| p95 Latency | ≤300ms | For cached requests |
| Throughput | ≥100 req/s | Single instance |
| Error Rate | <1% | Excluding rate limit errors |

## Deployment

### Docker Build
```bash
docker build -t market-data-service:latest .
```

### Kubernetes
See deployment manifests in the `deploy/` directory (to be added).

## Troubleshooting

### High Cache Miss Rate
- Check Redis connectivity
- Verify REDIS_TTL configuration
- Ensure background worker is enabled
- Check for high traffic on new tokens

### Rate Limit Errors
- Verify COINGECKO_RATE_LIMIT setting
- Check if API key is configured
- Consider upgrading CoinGecko plan

### Database Connection Issues
- Verify DB_HOST and DB_PORT settings
- Check PostgreSQL logs
- Ensure migrations have been applied

## Project Structure

```
market-data-service/
├── cmd/
│   └── market-data-service/
│       └── main.go              # Application entry point
├── internal/
│   ├── cache/
│   │   └── redis.go             # Redis caching layer
│   ├── client/
│   │   └── coingecko.go         # CoinGecko API client
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handler/
│   │   └── grpc.go              # gRPC service handlers
│   ├── repository/
│   │   └── repository.go        # Database operations
│   └── worker/
│       └── price_refresher.go   # Background worker
├── pkg/
│   └── pb/                      # Generated protobuf code
├── api/
│   └── proto/
│       └── market_data.proto    # gRPC service definition
├── db/
│   └── migrations/              # Database migrations
├── tests/                       # Integration and load tests
├── Dockerfile                   # Container image
├── docker-compose.yml           # Local development stack
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

## Contributing

Follow the agent handoff protocol defined in `docs/AGENT-HANDOFF.md`.

## License

Proprietary - Million Dollar Hunter Project

*The multi-layer caching architecture ensures that the service can handle high query volumes while maintaining low latency, with intelligent cache warming and refresh strategies that adapt to usage patterns.*
