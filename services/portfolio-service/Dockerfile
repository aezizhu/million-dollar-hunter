# syntax = docker/dockerfile:1.6
FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum* ./
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
COPY ingestion-service/ ./ingestion-service/
WORKDIR /workspace/services/portfolio-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/portfolio-service ./cmd/server

FROM alpine:3.20
RUN adduser -D appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/portfolio-service /usr/local/bin/portfolio-service
ENV GRPC_ADDR=":8081"
CMD ["/usr/local/bin/portfolio-service"]
