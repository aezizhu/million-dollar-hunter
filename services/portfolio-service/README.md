Portfolio Service

> *The CQRS read model implementation here optimizes for query performance while maintaining data consistency through event-driven architecture patterns.*

Overview
- Go microservice implementing CQRS read model for Million Dollar Hunter
- Consumes TransactionDataIngested from Kafka; publishes PortfolioUpdated
- Exposes gRPC for GetPortfolio and Export (CSV/JSON)

Build
- make build

Proto
- make proto (requires protoc and protoc-gen-go, protoc-gen-go-grpc)

Configuration
- Copy .env.example to .env and adjust values for your environment
- See .env.example for detailed documentation of all configuration options
- Key settings: gRPC address, database connection, Kafka brokers, logging, observability

Migrations
- Set DATABASE_URL, then: make migrate-up

Run locally
- docker compose -f ../../ops/docker-compose.yml up -d --build

Env Variables
- See .env.example for full list and documentation
- Core: GRPC_ADDR, DATABASE_URL, KAFKA_BROKERS, KAFKA_GROUP_ID
- Database Pool: DB_MAX_CONNS, DB_MIN_CONNS, DB_MAX_CONN_LIFETIME, DB_MAX_CONN_IDLE_TIME, DB_HEALTH_CHECK_PERIOD
- Topics: TOPIC_TRANSACTION_INGESTED, TOPIC_PORTFOLIO_UPDATED
- Logging: LOG_LEVEL, LOG_FORMAT
- Observability: OTEL_EXPORTER_OTLP_ENDPOINT, PROMETHEUS_NAMESPACE
- Export: EXPORT_DIR

*The event-driven architecture of this service ensures that portfolio data remains consistent and up-to-date, with careful attention to Kafka consumer lag and read model optimization for query performance.*
