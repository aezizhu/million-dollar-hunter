FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api-gateway ./cmd/api-gateway

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/api-gateway"]
