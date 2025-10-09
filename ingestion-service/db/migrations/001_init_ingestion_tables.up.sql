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
