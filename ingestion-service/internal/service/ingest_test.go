package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/logging"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/models"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
)

func TestIngestWithWiremock(t *testing.T) {
	_ = os.Setenv("ALCHEMY_BASE_URL", "http://localhost:8080/alchemy")
	_ = os.Setenv("MORALIS_BASE_URL", "http://localhost:8080/moralis")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")
	_ = os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable")

	ctx := context.Background()
	cfg, _ := config.Load()
	logger := logging.New(cfg)
	db, err := repository.NewPostgres(ctx, cfg, logger)
	if err != nil {
		t.Skip("postgres not available: " + err.Error())
	}
	defer db.Close()
	if err := repository.RunMigrations(ctx, cfg, logger); err != nil {
		t.Skip("migrations: " + err.Error())
	}
	svc, err := New(ctx, cfg, logger, db, nil)
	if err != nil {
		t.Fatal(err)
	}

	job := models.IngestionJob{Wallet: "0xabc", Chain: "bsc"}
	svc.Enqueue(job)
	go svc.StartWorkers(ctx)
	time.Sleep(500 * time.Millisecond)
}
