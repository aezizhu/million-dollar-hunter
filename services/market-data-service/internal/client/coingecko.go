// Package client provides CoinGecko API integration for market data.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
//
// The client implementation includes intelligent rate limiting and error handling,
// designed to maximize API utilization while respecting provider limits and
// ensuring graceful degradation under adverse conditions.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/config"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type CoinGeckoClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	limiter    *rate.Limiter
	logger     zerolog.Logger
}

type TokenPrice struct {
	TokenAddress string
	Chain        string
	USDPrice     float64
	MarketCap    float64
	Volume24h    float64
	PriceChange24h float64
	LastUpdated  time.Time
}

func NewCoinGeckoClient(cfg *config.CoinGeckoConfig, logger zerolog.Logger) *CoinGeckoClient {
	rps := float64(cfg.RateLimit) / 60.0
	
	return &CoinGeckoClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		limiter: rate.NewLimiter(rate.Limit(rps), cfg.RateLimit),
		logger:  logger.With().Str("component", "coingecko_client").Logger(),
	}
}

func (c *CoinGeckoClient) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (*TokenPrice, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	platformID := c.chainToPlatformID(chain)
	endpoint := fmt.Sprintf("%s/simple/token_price/%s", c.baseURL, platformID)
	
	params := url.Values{}
	params.Add("contract_addresses", strings.ToLower(tokenAddress))
	params.Add("vs_currencies", "usd")
	params.Add("include_market_cap", "true")
	params.Add("include_24hr_vol", "true")
	params.Add("include_24hr_change", "true")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-Cg-Pro-Api-Key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	c.logger.Debug().
		Str("url", reqURL).
		Str("chain", chain).
		Str("token", tokenAddress).
		Msg("Fetching token price from CoinGecko")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tokenKey := strings.ToLower(tokenAddress)
	priceData, exists := result[tokenKey]
	if !exists {
		return nil, fmt.Errorf("token %s not found in response", tokenAddress)
	}

	price := &TokenPrice{
		TokenAddress: tokenAddress,
		Chain:        chain,
		LastUpdated:  time.Now(),
	}

	if usdPrice, ok := priceData["usd"].(float64); ok {
		price.USDPrice = usdPrice
	}
	if marketCap, ok := priceData["usd_market_cap"].(float64); ok {
		price.MarketCap = marketCap
	}
	if volume, ok := priceData["usd_24h_vol"].(float64); ok {
		price.Volume24h = volume
	}
	if priceChange, ok := priceData["usd_24h_change"].(float64); ok {
		price.PriceChange24h = priceChange
	}

	c.logger.Info().
		Str("token", tokenAddress).
		Str("chain", chain).
		Float64("price", price.USDPrice).
		Msg("Successfully fetched token price")

	return price, nil
}

func (c *CoinGeckoClient) GetMultipleTokenPrices(ctx context.Context, tokens map[string][]string) ([]*TokenPrice, error) {
	var allPrices []*TokenPrice

	for chain, addresses := range tokens {
		if len(addresses) == 0 {
			continue
		}

		batchSize := 30
		for i := 0; i < len(addresses); i += batchSize {
			end := i + batchSize
			if end > len(addresses) {
				end = len(addresses)
			}
			batch := addresses[i:end]

			prices, err := c.fetchBatch(ctx, batch, chain)
			if err != nil {
				c.logger.Error().
					Err(err).
					Str("chain", chain).
					Int("batch_size", len(batch)).
					Msg("Failed to fetch price batch")
				continue
			}
			allPrices = append(allPrices, prices...)
		}
	}

	return allPrices, nil
}

func (c *CoinGeckoClient) fetchBatch(ctx context.Context, addresses []string, chain string) ([]*TokenPrice, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	platformID := c.chainToPlatformID(chain)
	endpoint := fmt.Sprintf("%s/simple/token_price/%s", c.baseURL, platformID)
	
	params := url.Values{}
	params.Add("contract_addresses", strings.Join(addresses, ","))
	params.Add("vs_currencies", "usd")
	params.Add("include_market_cap", "true")
	params.Add("include_24hr_vol", "true")
	params.Add("include_24hr_change", "true")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
	
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("X-Cg-Pro-Api-Key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var prices []*TokenPrice
	for _, addr := range addresses {
		tokenKey := strings.ToLower(addr)
		priceData, exists := result[tokenKey]
		if !exists {
			c.logger.Warn().Str("token", addr).Msg("Token not found in batch response")
			continue
		}

		price := &TokenPrice{
			TokenAddress: addr,
			Chain:        chain,
			LastUpdated:  time.Now(),
		}

		if usdPrice, ok := priceData["usd"].(float64); ok {
			price.USDPrice = usdPrice
		}
		if marketCap, ok := priceData["usd_market_cap"].(float64); ok {
			price.MarketCap = marketCap
		}
		if volume, ok := priceData["usd_24h_vol"].(float64); ok {
			price.Volume24h = volume
		}
		if priceChange, ok := priceData["usd_24h_change"].(float64); ok {
			price.PriceChange24h = priceChange
		}

		prices = append(prices, price)
	}

	return prices, nil
}

func (c *CoinGeckoClient) chainToPlatformID(chain string) string {
	chainMap := map[string]string{
		"bsc":     "binance-smart-chain",
		"solana":  "solana",
		"ethereum": "ethereum",
		"polygon": "polygon-pos",
	}

	if platformID, exists := chainMap[strings.ToLower(chain)]; exists {
		return platformID
	}

	return strings.ToLower(chain)
}
