SHELL := /bin/bash

.PHONY: tidy build run test k6 validate-openapi openapi

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
	OPENAPI_PATH?=../docs/openapi.yaml ; \
	OPENAPI_PATH=$$OPENAPI_PATH go run ./cmd/openapi-validate

openapi:
	@echo "Using OPENAPI_PATH=$${OPENAPI_PATH:-../../docs/openapi.yaml}"
