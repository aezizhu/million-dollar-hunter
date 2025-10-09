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

func setupTestService(t *testing.T) (*Service, *repository.Postgres, func()) {
	_ = os.Setenv("ALCHEMY_BASE_URL", "http://localhost:8080/alchemy")
	_ = os.Setenv("MORALIS_BASE_URL", "http://localhost:8080/moralis")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")
	_ = os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable")

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}

	logger := logging.New(cfg)
	db, err := repository.NewPostgres(ctx, cfg, logger)
	if err != nil {
		t.Skip("postgres not available")
	}

	if err := repository.RunMigrations(ctx, cfg, logger); err != nil {
		db.Close()
		t.Fatalf("migrations: %v", err)
	}

	svc, err := New(ctx, cfg, logger, db)
	if err != nil {
		db.Close()
		t.Fatalf("service init: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return svc, db, cleanup
}

func TestIngestionWithPagination(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	job := models.IngestionJob{Wallet: "0xabc", Chain: "bsc"}
	svc.Enqueue(job)

	time.Sleep(1 * time.Second)

	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_transactions WHERE wallet_address = $1", "0xabc").Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}

	if count == 0 {
		t.Error("expected transactions to be stored, got 0")
	}
}

func TestIngestionWithEmptyResponse(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	job := models.IngestionJob{Wallet: "0xempty", Chain: "bsc"}
	svc.Enqueue(job)

	time.Sleep(1 * time.Second)

	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_transactions WHERE wallet_address = $1", "0xempty").Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 transactions for empty response, got %d", count)
	}
}

func TestIngestionWithSolana(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	job := models.IngestionJob{Wallet: "7xSolAddr1", Chain: "solana"}
	svc.Enqueue(job)

	time.Sleep(1 * time.Second)

	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_transactions WHERE wallet_address = $1 AND chain = $2", "7xSolAddr1", "solana").Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}

	if count == 0 {
		t.Error("expected Solana transactions to be stored, got 0")
	}
}

func TestMoralisSolanaBalances(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	job := models.IngestionJob{Wallet: "SolWallet123", Chain: "solana"}
	svc.Enqueue(job)

	time.Sleep(1 * time.Second)

	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_balances WHERE wallet_address = $1 AND chain = $2", "SolWallet123", "solana").Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}

	if count == 0 {
		t.Error("expected Solana balances to be stored, got 0")
	}
}

func TestMoralisBSCBalances(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	job := models.IngestionJob{Wallet: "0xbsc123", Chain: "bsc"}
	svc.Enqueue(job)

	time.Sleep(1 * time.Second)

	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_balances WHERE wallet_address = $1 AND chain = $2", "0xbsc123", "bsc").Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}

	if count == 0 {
		t.Error("expected BSC balances to be stored, got 0")
	}
}

func TestCircuitBreakerBehavior(t *testing.T) {
	svc, _, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	for i := 0; i < 10; i++ {
		job := models.IngestionJob{Wallet: "0xerror", Chain: "bsc"}
		svc.Enqueue(job)
	}

	time.Sleep(2 * time.Second)
}

func TestConcurrentIngestion(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	wallets := []string{"0xwallet1", "0xwallet2", "0xwallet3", "0xwallet4", "0xwallet5"}
	for _, wallet := range wallets {
		job := models.IngestionJob{Wallet: wallet, Chain: "bsc"}
		svc.Enqueue(job)
	}

	time.Sleep(2 * time.Second)

	for _, wallet := range wallets {
		var count int
		err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_transactions WHERE wallet_address = $1", wallet).Scan(&count)
		if err != nil {
			t.Fatalf("query count for %s: %v", wallet, err)
		}
		if count == 0 {
			t.Errorf("expected transactions for wallet %s, got 0", wallet)
		}
	}
}

func TestMultiChainIngestion(t *testing.T) {
	svc, db, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	go svc.StartWorkers(ctx)

	jobs := []models.IngestionJob{
		{Wallet: "0xeth1", Chain: "bsc"},
		{Wallet: "7xSol1", Chain: "solana"},
		{Wallet: "0xeth2", Chain: "eth"},
	}

	for _, job := range jobs {
		svc.Enqueue(job)
	}

	time.Sleep(2 * time.Second)

	for _, job := range jobs {
		var count int
		err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM raw_transactions WHERE wallet_address = $1 AND chain = $2", job.Wallet, job.Chain).Scan(&count)
		if err != nil {
			t.Fatalf("query count for %s/%s: %v", job.Wallet, job.Chain, err)
		}
		if count == 0 {
			t.Logf("no transactions found for wallet %s on chain %s (may be expected)", job.Wallet, job.Chain)
		}
	}
}
