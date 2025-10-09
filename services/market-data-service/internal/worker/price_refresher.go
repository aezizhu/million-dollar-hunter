package worker

import (
	"context"
	"sync"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/cache"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/client"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/repository"
	"github.com/rs/zerolog"
)

type PriceRefresher struct {
	coinGecko  *client.CoinGeckoClient
	cache      *cache.RedisCache
	repo       *repository.Repository
	cfg        *config.WorkerConfig
	logger     zerolog.Logger
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewPriceRefresher(
	coinGecko *client.CoinGeckoClient,
	cache *cache.RedisCache,
	repo *repository.Repository,
	cfg *config.WorkerConfig,
	logger zerolog.Logger,
) *PriceRefresher {
	return &PriceRefresher{
		coinGecko: coinGecko,
		cache:     cache,
		repo:      repo,
		cfg:       cfg,
		logger:    logger.With().Str("component", "price_refresher").Logger(),
		stopChan:  make(chan struct{}),
	}
}

func (w *PriceRefresher) Start(ctx context.Context) {
	if !w.cfg.Enabled {
		w.logger.Info().Msg("Price refresher worker is disabled")
		return
	}

	w.logger.Info().
		Dur("interval", w.cfg.RefreshInterval).
		Int("batch_size", w.cfg.BatchSize).
		Msg("Starting price refresher worker")

	w.wg.Add(1)
	go w.run(ctx)
}

func (w *PriceRefresher) Stop() {
	w.logger.Info().Msg("Stopping price refresher worker")
	close(w.stopChan)
	w.wg.Wait()
	w.logger.Info().Msg("Price refresher worker stopped")
}

func (w *PriceRefresher) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cfg.RefreshInterval)
	defer ticker.Stop()

	w.refreshPrices(ctx)

	for {
		select {
		case <-ticker.C:
			w.refreshPrices(ctx)
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (w *PriceRefresher) refreshPrices(ctx context.Context) {
	startTime := time.Now()
	w.logger.Info().Msg("Starting price refresh cycle")

	tokens, err := w.repo.GetAllTokenAddresses(ctx)
	if err != nil {
		w.logger.Error().Err(err).Msg("Failed to get token addresses")
		return
	}

	totalTokens := 0
	for _, addresses := range tokens {
		totalTokens += len(addresses)
	}

	if totalTokens == 0 {
		w.logger.Debug().Msg("No tokens to refresh")
		return
	}

	w.logger.Info().
		Int("total_tokens", totalTokens).
		Int("chains", len(tokens)).
		Msg("Refreshing prices for tracked tokens")

	successCount := 0
	errorCount := 0

	for chain, addresses := range tokens {
		for i := 0; i < len(addresses); i += w.cfg.BatchSize {
			end := i + w.cfg.BatchSize
			if end > len(addresses) {
				end = len(addresses)
			}
			batch := addresses[i:end]

			tokenMap := map[string][]string{
				chain: batch,
			}

			prices, err := w.coinGecko.GetMultipleTokenPrices(ctx, tokenMap)
			if err != nil {
				w.logger.Error().
					Err(err).
					Str("chain", chain).
					Int("batch_size", len(batch)).
					Msg("Failed to fetch price batch")
				errorCount += len(batch)
				continue
			}

			dbPrices := make([]*repository.TokenPrice, len(prices))
			for i, p := range prices {
				dbPrices[i] = &repository.TokenPrice{
					TokenAddress:   p.TokenAddress,
					Chain:          p.Chain,
					USDPrice:       p.USDPrice,
					MarketCap:      p.MarketCap,
					Volume24h:      p.Volume24h,
					PriceChange24h: p.PriceChange24h,
					LastUpdated:    p.LastUpdated,
				}
			}

			if err := w.repo.SaveMultipleTokenPrices(ctx, dbPrices); err != nil {
				w.logger.Error().
					Err(err).
					Str("chain", chain).
					Int("count", len(prices)).
					Msg("Failed to save prices to database")
			}

			cachedPrices := make([]*cache.CachedPrice, len(prices))
			for i, p := range prices {
				cachedPrices[i] = &cache.CachedPrice{
					TokenAddress:   p.TokenAddress,
					Chain:          p.Chain,
					USDPrice:       p.USDPrice,
					MarketCap:      p.MarketCap,
					Volume24h:      p.Volume24h,
					PriceChange24h: p.PriceChange24h,
				}
			}

			if err := w.cache.SetMultiplePrices(ctx, cachedPrices); err != nil {
				w.logger.Error().
					Err(err).
					Str("chain", chain).
					Int("count", len(prices)).
					Msg("Failed to cache prices")
			}

			successCount += len(prices)

			w.logger.Debug().
				Str("chain", chain).
				Int("refreshed", len(prices)).
				Msg("Refreshed batch")
		}
	}

	duration := time.Since(startTime)
	w.logger.Info().
		Int("total", totalTokens).
		Int("success", successCount).
		Int("errors", errorCount).
		Dur("duration", duration).
		Msg("Completed price refresh cycle")
}

func (w *PriceRefresher) RefreshToken(ctx context.Context, tokenAddress, chain string) error {
	w.logger.Info().
		Str("token", tokenAddress).
		Str("chain", chain).
		Msg("Manually refreshing token price")

	price, err := w.coinGecko.GetTokenPrice(ctx, tokenAddress, chain)
	if err != nil {
		return err
	}

	dbPrice := &repository.TokenPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
		LastUpdated:    price.LastUpdated,
	}

	if err := w.repo.SaveTokenPrice(ctx, dbPrice); err != nil {
		w.logger.Error().Err(err).Msg("Failed to save price to database")
	}

	cachedPrice := &cache.CachedPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
	}

	if err := w.cache.SetPrice(ctx, cachedPrice); err != nil {
		w.logger.Error().Err(err).Msg("Failed to cache price")
	}

	w.logger.Info().
		Str("token", tokenAddress).
		Str("chain", chain).
		Float64("price", price.USDPrice).
		Msg("Successfully refreshed token price")

	return nil
}
