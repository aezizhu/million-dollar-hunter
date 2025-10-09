SHELL := /bin/bash

.PHONY: tidy build run test k6 validate-openapi

tidy:
	go mod tidy

build:
	go build ./...

run:
	go run ./cmd/api-gateway

test:
	go test ./...

k6:
	k6 run ./load-tests/gateway-smoke.js

validate-openapi:
	echo "validate docs/openapi.yaml"
