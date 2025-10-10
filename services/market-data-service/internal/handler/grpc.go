package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/cache"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/client"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/worker"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedMarketDataServiceServer
	coinGecko *client.CoinGeckoClient
	cache     *cache.RedisCache
	repo      *repository.Repository
	worker    *worker.PriceRefresher
	logger    zerolog.Logger
}

func NewGRPCHandler(
	coinGecko *client.CoinGeckoClient,
	cache *cache.RedisCache,
	repo *repository.Repository,
	worker *worker.PriceRefresher,
	logger zerolog.Logger,
) *GRPCHandler {
	return &GRPCHandler{
		coinGecko: coinGecko,
		cache:     cache,
		repo:      repo,
		worker:    worker,
		logger:    logger.With().Str("component", "grpc_handler").Logger(),
	}
}

func (h *GRPCHandler) GetTokenPrice(ctx context.Context, req *pb.GetTokenPriceRequest) (*pb.GetTokenPriceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := validateTokenRequest(req.TokenAddress, req.Chain); err != nil {
		return nil, err
	}

	h.logger.Debug().
		Str("token", req.TokenAddress).
		Str("chain", req.Chain).
		Msg("GetTokenPrice request")

	cachedPrice, err := h.cache.GetPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Cache lookup error")
	}

	if cachedPrice != nil {
		return &pb.GetTokenPriceResponse{
			TokenAddress: cachedPrice.TokenAddress,
			Chain:        cachedPrice.Chain,
			UsdPrice:     cachedPrice.USDPrice,
			LastUpdated:  cachedPrice.CachedAt.Unix(),
			FromCache:    true,
		}, nil
	}

	dbPrice, err := h.repo.GetTokenPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Database lookup error")
	}

	if dbPrice != nil {
		cachedPrice := &cache.CachedPrice{
			TokenAddress:   dbPrice.TokenAddress,
			Chain:          dbPrice.Chain,
			USDPrice:       dbPrice.USDPrice,
			MarketCap:      dbPrice.MarketCap,
			Volume24h:      dbPrice.Volume24h,
			PriceChange24h: dbPrice.PriceChange24h,
		}
		if err := h.cache.SetPrice(ctx, cachedPrice); err != nil {
			h.logger.Error().Err(err).Msg("Failed to cache price from database")
		}

		return &pb.GetTokenPriceResponse{
			TokenAddress: dbPrice.TokenAddress,
			Chain:        dbPrice.Chain,
			UsdPrice:     dbPrice.USDPrice,
			LastUpdated:  dbPrice.LastUpdated.Unix(),
			FromCache:    false,
		}, nil
	}

	price, err := h.coinGecko.GetTokenPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to fetch price from CoinGecko")
		return nil, fmt.Errorf("failed to fetch price: %w", err)
	}

	dbPrice = &repository.TokenPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
		LastUpdated:    price.LastUpdated,
	}
	if err := h.repo.SaveTokenPrice(ctx, dbPrice); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save price to database")
	}

	cachedPrice = &cache.CachedPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
	}
	if err := h.cache.SetPrice(ctx, cachedPrice); err != nil {
		h.logger.Error().Err(err).Msg("Failed to cache price")
	}

	return &pb.GetTokenPriceResponse{
		TokenAddress: price.TokenAddress,
		Chain:        price.Chain,
		UsdPrice:     price.USDPrice,
		LastUpdated:  price.LastUpdated.Unix(),
		FromCache:    false,
	}, nil
}

func (h *GRPCHandler) GetTokenPrices(ctx context.Context, req *pb.GetTokenPricesRequest) (*pb.GetTokenPricesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	h.logger.Debug().
		Int("count", len(req.Tokens)).
		Msg("GetTokenPrices request")

	if len(req.Tokens) == 0 {
		return &pb.GetTokenPricesResponse{Prices: []*pb.TokenPrice{}}, nil
	}

	seen := make(map[string]bool)
	var tokens []cache.TokenIdentifier
	for _, t := range req.Tokens {
		if err := validateTokenRequest(t.TokenAddress, t.Chain); err != nil {
			h.logger.Warn().
				Str("token", t.TokenAddress).
				Str("chain", t.Chain).
				Err(err).
				Msg("Skipping invalid token in batch request")
			continue
		}

		key := t.Chain + ":" + t.TokenAddress
		if seen[key] {
			continue
		}
		seen[key] = true

		tokens = append(tokens, cache.TokenIdentifier{
			Address: t.TokenAddress,
			Chain:   t.Chain,
		})
	}

	if len(tokens) == 0 {
		return &pb.GetTokenPricesResponse{Prices: []*pb.TokenPrice{}}, nil
	}

	cachedPrices, misses, err := h.cache.GetMultiplePrices(ctx, tokens)
	if err != nil {
		h.logger.Error().Err(err).Msg("Batch cache lookup error")
		misses = tokens // Treat all as misses on error
	}

	var response []*pb.TokenPrice

	for _, cp := range cachedPrices {
		response = append(response, &pb.TokenPrice{
			TokenAddress: cp.TokenAddress,
			Chain:        cp.Chain,
			UsdPrice:     cp.USDPrice,
			LastUpdated:  cp.CachedAt.Unix(),
			FromCache:    true,
		})
	}

	if len(misses) > 0 {
		missesMap := make(map[string][]string)
		for _, miss := range misses {
			missesMap[miss.Chain] = append(missesMap[miss.Chain], miss.Address)
		}

		prices, err := h.coinGecko.GetMultipleTokenPrices(ctx, missesMap)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to fetch prices from CoinGecko")
		} else {
			dbPrices := make([]*repository.TokenPrice, len(prices))
			cachePrices := make([]*cache.CachedPrice, len(prices))

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

				cachePrices[i] = &cache.CachedPrice{
					TokenAddress:   p.TokenAddress,
					Chain:          p.Chain,
					USDPrice:       p.USDPrice,
					MarketCap:      p.MarketCap,
					Volume24h:      p.Volume24h,
					PriceChange24h: p.PriceChange24h,
				}

				response = append(response, &pb.TokenPrice{
					TokenAddress: p.TokenAddress,
					Chain:        p.Chain,
					UsdPrice:     p.USDPrice,
					LastUpdated:  p.LastUpdated.Unix(),
					FromCache:    false,
				})
			}

			if err := h.repo.SaveMultipleTokenPrices(ctx, dbPrices); err != nil {
				h.logger.Error().Err(err).Msg("Failed to save prices to database")
			}

			if err := h.cache.SetMultiplePrices(ctx, cachePrices); err != nil {
				h.logger.Error().Err(err).Msg("Failed to cache prices")
			}
		}
	}

	return &pb.GetTokenPricesResponse{Prices: response}, nil
}

func (h *GRPCHandler) GetMarketData(ctx context.Context, req *pb.GetMarketDataRequest) (*pb.GetMarketDataResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := validateTokenRequest(req.TokenAddress, req.Chain); err != nil {
		return nil, err
	}

	h.logger.Debug().
		Str("token", req.TokenAddress).
		Str("chain", req.Chain).
		Msg("GetMarketData request")

	cachedPrice, err := h.cache.GetPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Cache lookup error")
	}

	if cachedPrice != nil {
		return &pb.GetMarketDataResponse{
			TokenAddress:   cachedPrice.TokenAddress,
			Chain:          cachedPrice.Chain,
			UsdPrice:       cachedPrice.USDPrice,
			MarketCap:      cachedPrice.MarketCap,
			Volume_24H:     cachedPrice.Volume24h,
			PriceChange_24H: cachedPrice.PriceChange24h,
			LastUpdated:    cachedPrice.CachedAt.Unix(),
			FromCache:      true,
		}, nil
	}

	dbPrice, err := h.repo.GetTokenPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Database lookup error")
	}

	if dbPrice != nil {
		cachedPrice := &cache.CachedPrice{
			TokenAddress:   dbPrice.TokenAddress,
			Chain:          dbPrice.Chain,
			USDPrice:       dbPrice.USDPrice,
			MarketCap:      dbPrice.MarketCap,
			Volume24h:      dbPrice.Volume24h,
			PriceChange24h: dbPrice.PriceChange24h,
		}
		if err := h.cache.SetPrice(ctx, cachedPrice); err != nil {
			h.logger.Error().Err(err).Msg("Failed to cache market data from database")
		}

		return &pb.GetMarketDataResponse{
			TokenAddress:   dbPrice.TokenAddress,
			Chain:          dbPrice.Chain,
			UsdPrice:       dbPrice.USDPrice,
			MarketCap:      dbPrice.MarketCap,
			Volume_24H:     dbPrice.Volume24h,
			PriceChange_24H: dbPrice.PriceChange24h,
			LastUpdated:    dbPrice.LastUpdated.Unix(),
			FromCache:      false,
		}, nil
	}

	price, err := h.coinGecko.GetTokenPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to fetch market data from CoinGecko")
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}

	dbPrice = &repository.TokenPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
		LastUpdated:    price.LastUpdated,
	}
	if err := h.repo.SaveTokenPrice(ctx, dbPrice); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save market data to database")
	}

	cachedPrice = &cache.CachedPrice{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		USDPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume24h:      price.Volume24h,
		PriceChange24h: price.PriceChange24h,
	}
	if err := h.cache.SetPrice(ctx, cachedPrice); err != nil {
		h.logger.Error().Err(err).Msg("Failed to cache market data")
	}

	return &pb.GetMarketDataResponse{
		TokenAddress:   price.TokenAddress,
		Chain:          price.Chain,
		UsdPrice:       price.USDPrice,
		MarketCap:      price.MarketCap,
		Volume_24H:     price.Volume24h,
		PriceChange_24H: price.PriceChange24h,
		LastUpdated:    price.LastUpdated.Unix(),
		FromCache:      false,
	}, nil
}

func (h *GRPCHandler) RefreshTokenPrice(ctx context.Context, req *pb.RefreshTokenPriceRequest) (*pb.RefreshTokenPriceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := validateTokenRequest(req.TokenAddress, req.Chain); err != nil {
		return &pb.RefreshTokenPriceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	h.logger.Info().
		Str("token", req.TokenAddress).
		Str("chain", req.Chain).
		Msg("RefreshTokenPrice request")

	err := h.worker.RefreshToken(ctx, req.TokenAddress, req.Chain)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to refresh token price")
		return &pb.RefreshTokenPriceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	cachedPrice, err := h.cache.GetPrice(ctx, req.TokenAddress, req.Chain)
	if err != nil || cachedPrice == nil {
		return &pb.RefreshTokenPriceResponse{
			Success: true,
			Message: "Price refreshed successfully",
		}, nil
	}

	return &pb.RefreshTokenPriceResponse{
		Success:     true,
		Message:     "Price refreshed successfully",
		UsdPrice:    cachedPrice.USDPrice,
		LastUpdated: cachedPrice.CachedAt.Unix(),
	}, nil
}

func validateTokenRequest(tokenAddress, chain string) error {
	if strings.TrimSpace(tokenAddress) == "" {
		return status.Error(codes.InvalidArgument, "token address cannot be empty")
	}

	if strings.TrimSpace(chain) == "" {
		return status.Error(codes.InvalidArgument, "chain cannot be empty")
	}

	validChains := map[string]bool{
		"bsc":      true,
		"solana":   true,
		"ethereum": true,
		"polygon":  true,
	}

	if !validChains[strings.ToLower(chain)] {
		return status.Errorf(codes.InvalidArgument, "invalid chain: %s (supported: bsc, solana, ethereum, polygon)", chain)
	}

	if len(tokenAddress) < 10 {
		return status.Error(codes.InvalidArgument, "token address appears to be malformed (too short)")
	}

	return nil
}
