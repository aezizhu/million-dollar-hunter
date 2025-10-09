package repository

import (
	"context"
	"os"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*Postgres, error) {
	conf, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}

func RunMigrations(ctx context.Context, cfg *config.Config, logger zerolog.Logger) error {
	up := `
CREATE TABLE IF NOT EXISTS ingestion_jobs (
 id UUID PRIMARY KEY,
 wallet_address TEXT NOT NULL,
 chain TEXT NOT NULL,
 status TEXT NOT NULL,
 last_run_timestamp TIMESTAMPTZ,
 cursor TEXT,
 created_at TIMESTAMPTZ DEFAULT NOW(),
 updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS raw_transactions (
 id BIGSERIAL PRIMARY KEY,
 source_api TEXT NOT NULL,
 wallet_address TEXT NOT NULL,
 chain TEXT NOT NULL,
 data JSONB NOT NULL,
 ingested_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_raw_tx_wallet ON raw_transactions (wallet_address, chain, ingested_at DESC);
CREATE TABLE IF NOT EXISTS raw_balances (
 id BIGSERIAL PRIMARY KEY,
 wallet_address TEXT NOT NULL,
 chain TEXT NOT NULL,
 data JSONB NOT NULL,
 ingested_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_raw_bal_wallet ON raw_balances (wallet_address, chain, ingested_at DESC);
CREATE TABLE IF NOT EXISTS holder_snapshots (
 id BIGSERIAL PRIMARY KEY,
 token_address TEXT,
 holder_address TEXT,
 balance NUMERIC,
 rank INT,
 timestamp TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_holder_snapshots_token_ts ON holder_snapshots (token_address, timestamp);
CREATE INDEX IF NOT EXISTS idx_holder_snapshots_rank ON holder_snapshots (token_address, rank, timestamp);
`
	conn, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Exec(ctx, up)
	return err
}

func mustRead(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}
