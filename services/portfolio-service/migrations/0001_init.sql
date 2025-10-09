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
  wallet_id UUID NOT NULL,
  token_address TEXT NOT NULL,
  symbol TEXT,
  name TEXT,
  current_balance NUMERIC,
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS asset_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  asset_id UUID NOT NULL,
  ts TIMESTAMPTZ NOT NULL,
  balance NUMERIC,
  usd_value NUMERIC
);

CREATE TABLE IF NOT EXISTS transactions_view (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id UUID NOT NULL,
  ts TIMESTAMPTZ NOT NULL,
  type TEXT,
  from_addr TEXT,
  to_addr TEXT,
  asset_symbol TEXT,
  amount NUMERIC,
  usd_value NUMERIC,
  tx_hash TEXT
);

CREATE INDEX IF NOT EXISTS idx_assets_wallet_id ON assets(wallet_id);
CREATE INDEX IF NOT EXISTS idx_asset_snapshots_asset_id_ts ON asset_snapshots(asset_id, ts);
CREATE INDEX IF NOT EXISTS idx_tx_view_wallet_id_ts ON transactions_view(wallet_id, ts);
