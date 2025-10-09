.PHONY: help proto build test test-integration test-load run clean docker-build docker-up docker-down migrate-up migrate-down

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate gRPC code from proto files
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/market_data.proto

build: proto ## Build the service binary
	go build -o bin/market-data-service ./cmd/market-data-service

test: ## Run unit tests
	go test ./internal/... -v -cover

test-integration: ## Run integration tests (requires running services)
	go test ./tests -v -run Integration

test-load: ## Run load tests (requires running services)
	go test ./tests -v -run Load -timeout 30m

bench: ## Run benchmarks
	go test ./tests -bench=. -benchmem

run: ## Run the service locally
	go run cmd/market-data-service/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf pkg/pb/*.go

docker-build: ## Build Docker image
	docker build -t market-data-service:latest .

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## View service logs
	docker-compose logs -f market-data-service

migrate-up: ## Run database migrations
	docker-compose up migrate

migrate-down: ## Rollback database migrations
	docker run --rm -v $(PWD)/db/migrations:/migrations \
	  --network host migrate/migrate \
	  -path=/migrations \
	  -database "postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable" \
	  down

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

deps: ## Download dependencies
	go mod download
	go mod tidy

all: clean proto build test ## Build and test everything
