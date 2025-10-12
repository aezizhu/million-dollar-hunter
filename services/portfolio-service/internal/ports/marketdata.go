package ports

import "context"

type MarketDataClient interface {
	GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error)

	GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error)

	Close() error
}
