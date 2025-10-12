package ports

import "context"

type TokenIdentifier struct {
	Address string
	Chain   string
}

type TokenPrice struct {
	Address    string
	Chain      string
	USDPrice   float64
	LastUpdate int64
	FromCache  bool
}

type MarketDataClient interface {
	GetTokenPrices(ctx context.Context, tokens []TokenIdentifier) ([]TokenPrice, error)
	GetTokenPrice(ctx context.Context, token TokenIdentifier) (*TokenPrice, error)
}
