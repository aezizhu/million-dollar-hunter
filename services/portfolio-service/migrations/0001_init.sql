CREATE TABLE IF NOT EXISTS wallets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID,
  address TEXT NOT NULL,
  chain TEXT NOT NULL,
  nickname TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS assets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  token_address TEXT NOT NULL,
  symbol TEXT,
  name TEXT,
  current_balance NUMERIC(38, 18),
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS asset_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  ts TIMESTAMPTZ NOT NULL,
  balance NUMERIC(38, 18),
  usd_value NUMERIC(38, 18)
);

CREATE TABLE IF NOT EXISTS transactions_view (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  ts TIMESTAMPTZ NOT NULL,
  type TEXT,
  from_addr TEXT,
  to_addr TEXT,
  asset_symbol TEXT,
  amount NUMERIC(38, 18),
  usd_value NUMERIC(38, 18),
  tx_hash TEXT
);

CREATE INDEX IF NOT EXISTS idx_assets_wallet_id ON assets(wallet_id);
CREATE INDEX IF NOT EXISTS idx_asset_snapshots_asset_id_ts ON asset_snapshots(asset_id, ts);
CREATE INDEX IF NOT EXISTS idx_tx_view_wallet_id_ts ON transactions_view(wallet_id, ts);
CREATE INDEX IF NOT EXISTS idx_assets_updated_at ON assets(updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_address_chain ON wallets(address, chain);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tx_view_hash ON transactions_view(tx_hash);
CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_wallet_token ON assets(wallet_id, token_address);
