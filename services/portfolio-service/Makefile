PROTO=../../proto
PKG=github.com/aezizhu/million-dollar-hunter/services/portfolio-service
BIN=portfolio-service

all: build
build:
	go build -o bin/$(BIN) ./cmd/server

proto:
	protoc -I $(PROTO) --go_out=. --go-grpc_out=. $(PROTO)/portfolio/v1/portfolio.proto

migrate-up:
	migrate -path ./migrations -database $$DATABASE_URL up

migrate-down:
	migrate -path ./migrations -database $$DATABASE_URL down

run:
	go run ./cmd/server
