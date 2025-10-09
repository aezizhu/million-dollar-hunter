CREATE TABLE IF NOT EXISTS token_prices (
    token_address TEXT NOT NULL,
    chain TEXT NOT NULL,
    usd_price NUMERIC(20, 8) NOT NULL DEFAULT 0,
    market_cap NUMERIC(20, 2),
    volume_24h NUMERIC(20, 2),
    price_change_24h NUMERIC(10, 4),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (token_address, chain)
);

CREATE INDEX idx_token_prices_last_updated ON token_prices(last_updated);

CREATE INDEX idx_token_prices_chain ON token_prices(chain);

COMMENT ON TABLE token_prices IS 'Persistent cache of cryptocurrency token prices from CoinGecko';
COMMENT ON COLUMN token_prices.token_address IS 'Contract address of the token';
COMMENT ON COLUMN token_prices.chain IS 'Blockchain identifier (e.g., bsc, solana, ethereum)';
COMMENT ON COLUMN token_prices.usd_price IS 'Current USD price of the token';
COMMENT ON COLUMN token_prices.market_cap IS 'Market capitalization in USD';
COMMENT ON COLUMN token_prices.volume_24h IS '24-hour trading volume in USD';
COMMENT ON COLUMN token_prices.price_change_24h IS '24-hour price change percentage';
COMMENT ON COLUMN token_prices.last_updated IS 'Timestamp of last price update';
