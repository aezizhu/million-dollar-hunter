FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/portfolio-service/go.mod services/portfolio-service/go.sum* ./ 
RUN go mod download
COPY services/portfolio-service/ ./ 
COPY proto/ ./proto/
RUN go build -o /bin/portfolio-service ./cmd/server

FROM alpine:3.20
RUN adduser -D appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /bin/portfolio-service /usr/local/bin/portfolio-service
ENV GRPC_ADDR=":8081"
CMD ["/usr/local/bin/portfolio-service"]
