ALTER TABLE transactions_view
  ADD COLUMN IF NOT EXISTS token_address TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tx_view_hash ON transactions_view(tx_hash);
CREATE INDEX IF NOT EXISTS idx_tx_view_wallet_token_ts ON transactions_view(wallet_id, token_address, ts);

ALTER TABLE asset_snapshots
  ADD COLUMN IF NOT EXISTS wallet_id UUID,
  ADD COLUMN IF NOT EXISTS token_address TEXT;

UPDATE asset_snapshots s
SET wallet_id = a.wallet_id,
    token_address = a.token_address
FROM assets a
WHERE s.asset_id = a.id
  AND (s.wallet_id IS NULL OR s.token_address IS NULL);

CREATE INDEX IF NOT EXISTS idx_asset_snapshots_wallet_token_ts
  ON asset_snapshots(wallet_id, token_address, ts DESC);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
