package solana

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
	baseURL string
	apiKey  string
	http    *http.Client
	breaker *breaker.Breaker
	limiter *ratelimit.TokenBucket
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

type GetTransactionsParams struct {
	Address   string `json:"address"`
	Limit     int    `json:"limit,omitempty"`
	Before    string `json:"before,omitempty"`
	Until     string `json:"until,omitempty"`
}

func (c *Client) GetTransactions(ctx context.Context, params GetTransactionsParams) ([]map[string]any, error) {
	ok, _, err := c.limiter.Allow(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("rate limited")
	}

	url := fmt.Sprintf("%s/%s/account/%s/transactions", c.baseURL, "api", params.Address)
	if params.Limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", url, params.Limit)
	}

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

func (c *Client) GetTokenBalances(ctx context.Context, address string) ([]map[string]any, error) {
	ok, _, err := c.limiter.Allow(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("rate limited")
	}

	url := fmt.Sprintf("%s/%s/account/%s/tokens", c.baseURL, "api", address)

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
