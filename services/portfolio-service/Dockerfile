FROM golang:1.24-alpine AS builder
WORKDIR /workspace
# Copy go workspace files
COPY go.work go.work.sum* ./
# Copy all service directories
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
COPY ingestion-service/ ./ingestion-service/
# Build from portfolio-service directory
WORKDIR /workspace/services/portfolio-service
RUN go mod download
RUN go build -o /bin/portfolio-service ./cmd/server

FROM alpine:3.20
RUN adduser -D appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/portfolio-service /usr/local/bin/portfolio-service
ENV GRPC_ADDR=":8081"
CMD ["/usr/local/bin/portfolio-service"]
