package main

import (
	"log"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/server"
)

func main() {
	cfg := config.Load()
	if err := server.ValidateOpenAPI(cfg.OpenAPIPath); err != nil {
		log.Fatalf("OpenAPI validation failed: %v", err)
	}
}
