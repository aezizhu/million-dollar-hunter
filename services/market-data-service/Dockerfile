# syntax = docker/dockerfile:1.6
# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git protobuf-dev
WORKDIR /workspace
COPY go.work go.work.sum* ./
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
COPY ingestion-service/ ./ingestion-service/
WORKDIR /workspace/services/market-data-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/market-data-service ./cmd/market-data-service

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /app/market-data-service .
COPY --from=builder /workspace/services/market-data-service/db ./db
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 50051
CMD ["./market-data-service"]
