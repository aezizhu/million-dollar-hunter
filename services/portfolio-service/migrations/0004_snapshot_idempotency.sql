ALTER TABLE asset_snapshots
  ADD COLUMN IF NOT EXISTS ts_bucket TIMESTAMPTZ;

UPDATE asset_snapshots
SET ts_bucket = date_trunc('minute', ts)
WHERE ts_bucket IS NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND indexname = 'idx_asset_snapshots_idem'
  ) THEN
    EXECUTE 'CREATE UNIQUE INDEX idx_asset_snapshots_idem ON asset_snapshots(wallet_id, token_address, ts_bucket)';
  END IF;
END $$;
