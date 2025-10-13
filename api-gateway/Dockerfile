# syntax = docker/dockerfile:1.6
FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum* ./
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
COPY ingestion-service/ ./ingestion-service/
WORKDIR /workspace/api-gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/api-gateway"]
