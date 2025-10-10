DROP TABLE IF EXISTS raw_balances;
DROP TABLE IF EXISTS raw_transactions;
DROP TABLE IF EXISTS ingestion_jobs;
</create_file_file>
<create_file path="/home/ubuntu/repos/million-dollar-hunter/ingestion-service/db/migrations/002_holder_snapshots.up.sql">CREATE TABLE IF NOT EXISTS holder_snapshots (
 id BIGSERIAL PRIMARY KEY,
 token_address TEXT,
 holder_address TEXT,
 balance NUMERIC,
 rank INT,
 timestamp TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_holder_snapshots_token_ts ON holder_snapshots (token_address, timestamp);
CREATE INDEX IF NOT EXISTS idx_holder_snapshots_rank ON holder_snapshots (token_address, rank, timestamp);
