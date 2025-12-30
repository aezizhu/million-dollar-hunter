# syntax = docker/dockerfile:1.6
FROM golang:1.24 as builder
WORKDIR /workspace
COPY go.work go.work.sum* ./
COPY pkg/ ./pkg/
COPY ingestion-service/ ./ingestion-service/
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
WORKDIR /workspace/ingestion-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/ingestion ./cmd/ingestion-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
RUN adduser -D appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/ingestion /usr/local/bin/ingestion
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/ingestion"]
