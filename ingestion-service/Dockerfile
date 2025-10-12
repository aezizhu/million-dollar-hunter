FROM golang:1.24 as builder
WORKDIR /workspace
# Copy go workspace files
COPY go.work go.work.sum* ./
# Copy all service directories needed for workspace
COPY ingestion-service/ ./ingestion-service/
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
# Build from ingestion-service directory
WORKDIR /workspace/ingestion-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/ingestion ./cmd/ingestion-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
COPY --from=builder /bin/ingestion /ingestion
EXPOSE 8090
ENTRYPOINT ["/ingestion"]
