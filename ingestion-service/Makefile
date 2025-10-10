.PHONY: build test lint bench up down

build:
	go build ./...

test:
	go test ./... -v

lint:
	golangci-lint run

bench:
	go test -bench=. -run=^$$ ./internal/service

up:
	docker compose up -d

down:
	docker compose down -v
