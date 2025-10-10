Portfolio Service

Overview
- Go microservice implementing CQRS read model for Million Dollar Hunter
- Consumes TransactionDataIngested from Kafka; publishes PortfolioUpdated
- Exposes gRPC for GetPortfolio and Export (CSV/JSON)

Build
- make build

Proto
- make proto (requires protoc and protoc-gen-go, protoc-gen-go-grpc)

Migrations
- Set DATABASE_URL, then: make migrate-up

Run locally
- docker compose -f ../../ops/docker-compose.yml up -d --build

Env
- GRPC_ADDR, DATABASE_URL, KAFKA_BROKERS, TOPIC_TRANSACTION_INGESTED, TOPIC_PORTFOLIO_UPDATED, KAFKA_GROUP_ID, EXPORT_DIR
