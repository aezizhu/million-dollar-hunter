package moralis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/breaker"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/httpclient"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/ratelimit"
)

type Client struct {
	baseURL   string
	apiKey    string
	http      *http.Client
	breaker   *breaker.Breaker
	limiter   *ratelimit.TokenBucket
}

func New(baseURL, apiKey string, br *breaker.Breaker, rl *ratelimit.TokenBucket) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    httpclient.New(),
		breaker: br,
		limiter: rl,
	}
}

func chainParam(chain string) string {
	switch chain {
	case "eth", "ethereum":
		return "eth"
	case "bsc", "binance-smart-chain":
		return "bsc"
	case "sol", "solana":
		return "solana"
	default:
		return chain
	}
}

func (c *Client) GetWalletTokenBalancesPrice(ctx context.Context, address, chain string) ([]map[string]any, error) {
	ok, _, err := c.limiter.Allow(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("rate limited")
	}
	url := fmt.Sprintf("%s/%s/wallets/%s/balances?chain=%s", c.baseURL, "api", address, chainParam(chain))
	resI, err := c.breaker.Do(ctx, func() (interface{}, error) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		req.Header.Set("X-API-Key", c.apiKey)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("server error: %d", resp.StatusCode)
		}
		var data []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, err
		}
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return resI.([]map[string]any), nil
}
