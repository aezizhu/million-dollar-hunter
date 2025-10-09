package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	logger zerolog.Logger
}

type CachedPrice struct {
	TokenAddress   string    `json:"token_address"`
	Chain          string    `json:"chain"`
	USDPrice       float64   `json:"usd_price"`
	MarketCap      float64   `json:"market_cap"`
	Volume24h      float64   `json:"volume_24h"`
	PriceChange24h float64   `json:"price_change_24h"`
	CachedAt       time.Time `json:"cached_at"`
}

func NewRedisCache(addr string, password string, db int, ttl time.Duration, logger zerolog.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().
		Str("addr", addr).
		Int("db", db).
		Dur("ttl", ttl).
		Msg("Connected to Redis cache")

	return &RedisCache{
		client: client,
		ttl:    ttl,
		logger: logger.With().Str("component", "redis_cache").Logger(),
	}, nil
}

func (c *RedisCache) GetPrice(ctx context.Context, tokenAddress, chain string) (*CachedPrice, error) {
	key := c.priceKey(tokenAddress, chain)
	
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var price CachedPrice
	if err := json.Unmarshal([]byte(val), &price); err != nil {
		c.logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal cached price")
		return nil, fmt.Errorf("failed to unmarshal cached price: %w", err)
	}

	c.logger.Debug().
		Str("token", tokenAddress).
		Str("chain", chain).
		Float64("price", price.USDPrice).
		Msg("Cache hit")

	return &price, nil
}

func (c *RedisCache) SetPrice(ctx context.Context, price *CachedPrice) error {
	key := c.priceKey(price.TokenAddress, price.Chain)
	
	price.CachedAt = time.Now()
	
	data, err := json.Marshal(price)
	if err != nil {
		return fmt.Errorf("failed to marshal price: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug().
		Str("token", price.TokenAddress).
		Str("chain", price.Chain).
		Float64("price", price.USDPrice).
		Dur("ttl", c.ttl).
		Msg("Cached price")

	return nil
}

func (c *RedisCache) GetMultiplePrices(ctx context.Context, tokens []TokenIdentifier) ([]*CachedPrice, []TokenIdentifier, error) {
	if len(tokens) == 0 {
		return nil, nil, nil
	}

	keys := make([]string, len(tokens))
	for i, token := range tokens {
		keys[i] = c.priceKey(token.Address, token.Chain)
	}

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get multiple from cache: %w", err)
	}

	var cachedPrices []*CachedPrice
	var misses []TokenIdentifier

	for i, val := range values {
		if val == nil {
			misses = append(misses, tokens[i])
			continue
		}

		var price CachedPrice
		if err := json.Unmarshal([]byte(val.(string)), &price); err != nil {
			c.logger.Error().Err(err).Str("key", keys[i]).Msg("Failed to unmarshal cached price")
			misses = append(misses, tokens[i])
			continue
		}

		cachedPrices = append(cachedPrices, &price)
	}

	hitRate := float64(len(cachedPrices)) / float64(len(tokens)) * 100
	c.logger.Debug().
		Int("total", len(tokens)).
		Int("hits", len(cachedPrices)).
		Int("misses", len(misses)).
		Float64("hit_rate", hitRate).
		Msg("Batch cache lookup")

	return cachedPrices, misses, nil
}

func (c *RedisCache) SetMultiplePrices(ctx context.Context, prices []*CachedPrice) error {
	if len(prices) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	now := time.Now()

	for _, price := range prices {
		key := c.priceKey(price.TokenAddress, price.Chain)
		price.CachedAt = now

		data, err := json.Marshal(price)
		if err != nil {
			c.logger.Error().Err(err).Str("token", price.TokenAddress).Msg("Failed to marshal price")
			continue
		}

		pipe.Set(ctx, key, data, c.ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	c.logger.Debug().
		Int("count", len(prices)).
		Dur("ttl", c.ttl).
		Msg("Cached multiple prices")

	return nil
}

func (c *RedisCache) InvalidatePrice(ctx context.Context, tokenAddress, chain string) error {
	key := c.priceKey(tokenAddress, chain)
	
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	c.logger.Debug().
		Str("token", tokenAddress).
		Str("chain", chain).
		Msg("Invalidated cache entry")

	return nil
}

func (c *RedisCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	dbSize, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get db size: %w", err)
	}

	stats := map[string]interface{}{
		"db_size": dbSize,
		"info":    info,
	}

	return stats, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) priceKey(tokenAddress, chain string) string {
	return fmt.Sprintf("price:%s:%s", chain, tokenAddress)
}

type TokenIdentifier struct {
	Address string
	Chain   string
}
