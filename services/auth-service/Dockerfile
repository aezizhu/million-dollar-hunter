FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go workspace files
COPY go.work go.work.sum* ./

# Copy all service directories
COPY pkg/ ./pkg/
COPY services/ ./services/
COPY api-gateway/ ./api-gateway/
COPY ingestion-service/ ./ingestion-service/

# Build from auth-service directory
WORKDIR /workspace/services/auth-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/auth-service ./cmd/auth-service

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth-service .

# Copy migrations if they exist
COPY --from=builder /workspace/services/auth-service/db ./db

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose HTTP and gRPC ports
EXPOSE 8081 50051

# Run the application
CMD ["./auth-service"]
