.PHONY: help build run test test-load test-stress docker-build docker-up docker-down clean lint

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/api-gateway ./cmd/api-gateway

run: ## Run the application
	go run ./cmd/api-gateway/main.go

test: ## Run unit tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Generate test coverage report
	go tool cover -html=coverage.out -o coverage.html

test-load: ## Run k6 load tests
	k6 run tests/k6/load-test.js

test-stress: ## Run k6 stress tests
	k6 run tests/k6/stress-test.js

docker-build: ## Build Docker image
	docker build -t million-hunter/api-gateway:latest .

docker-up: ## Start services with docker-compose
	docker-compose up -d

docker-down: ## Stop services with docker-compose
	docker-compose down

docker-logs: ## Show docker-compose logs
	docker-compose logs -f

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

lint: ## Run linter
	golangci-lint run ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
