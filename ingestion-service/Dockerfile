FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/ingestion ./cmd/ingestion-service

FROM gcr.io/distroless/base-debian11
COPY --from=builder /bin/ingestion /ingestion
EXPOSE 8090
ENTRYPOINT ["/ingestion"]
