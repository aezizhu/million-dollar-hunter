ALTER TABLE transactions_view ADD COLUMN IF NOT EXISTS token_address TEXT;

CREATE INDEX IF NOT EXISTS idx_tx_view_wallet_token ON transactions_view(wallet_id, token_address);

ALTER TABLE asset_snapshots ADD COLUMN IF NOT EXISTS wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE;
ALTER TABLE asset_snapshots ADD COLUMN IF NOT EXISTS token_address TEXT;

CREATE INDEX IF NOT EXISTS idx_asset_snapshots_wallet_token ON asset_snapshots(wallet_id, token_address, ts DESC);
