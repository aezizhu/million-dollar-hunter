package service

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/logging"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/models"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
)

func BenchmarkIngestionThroughput(b *testing.B) {
	_ = os.Setenv("ALCHEMY_BASE_URL", "http://localhost:8080/alchemy")
	_ = os.Setenv("MORALIS_BASE_URL", "http://localhost:8080/moralis")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")
	_ = os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable")

	ctx := context.Background()
	cfg, _ := config.Load()
	logger := logging.New(cfg)
	db, err := repository.NewPostgres(ctx, cfg, logger)
	if err != nil {
		b.Skip("postgres not available")
	}
	defer db.Close()
	if err := repository.RunMigrations(ctx, cfg, logger); err != nil {
		b.Fatal(err)
	}
	svc, err := New(ctx, cfg, logger, db)
	if err != nil {
		b.Fatal(err)
	}
	go svc.StartWorkers(ctx)
	for i := 0; i < 10; i++ {
		svc.Enqueue(models.IngestionJob{Wallet: "0xabc", Chain: "bsc"})
	}
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	var wg sync.WaitGroup
	count := 200
	start := time.Now()
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.Enqueue(models.IngestionJob{Wallet: "0xabc", Chain: "bsc"})
		}()
	}
	wg.Wait()
	elapsed := time.Since(start).Seconds()
	tps := float64(count) / elapsed
	b.ReportMetric(tps, "tx/s")
}
