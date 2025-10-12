.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## Format all code with gofumpt
	@echo "Formatting code..."
	@command -v gofumpt >/dev/null 2>&1 || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }
	@gofumpt -l -w .
	@echo "Done!"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linters..."
	@golangci-lint cache clean
	@golangci-lint run
	@echo "Done!"

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...
	@echo "Done!"

.PHONY: test-short
test-short: ## Run short tests (skip integration tests)
	@echo "Running short tests..."
	@go test -short -v ./...
	@echo "Done!"

.PHONY: build
build: ## Build the auth-service binary
	@echo "Building..."
	@go build -o bin/auth-service ./cmd/auth-service
	@echo "Done! Binary at bin/auth-service"

.PHONY: check
check: fmt lint test-short ## Format, lint, and test (pre-commit checks)
	@echo "All checks passed!"

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "Cleaning..."
	@rm -rf bin/
	@golangci-lint cache clean
	@go clean -cache
	@echo "Done!"
