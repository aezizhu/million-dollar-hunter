// Package alchemy provides client integration with Alchemy blockchain APIs.
// The client implements pagination handling, rate limiting, and circuit breaker
// patterns to ensure reliable data ingestion while respecting API constraints.
package alchemy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	userAgent string
}

func New(baseURL, apiKey string, br *breaker.Breaker, rl *ratelimit.TokenBucket) *Client {
	return &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		http:      httpclient.New(),
		breaker:   br,
		limiter:   rl,
		userAgent: "mdh-ingestion/1.0",
	}
}

type getTransfersParams struct {
	FromBlock       string   `json:"fromBlock,omitempty"`
	ToBlock         string   `json:"toBlock,omitempty"`
	FromAddress     string   `json:"fromAddress,omitempty"`
	ToAddress       string   `json:"toAddress,omitempty"`
	Category        []string `json:"category,omitempty"`
	WithMetadata    bool     `json:"withMetadata,omitempty"`
	MaxCount        string   `json:"maxCount,omitempty"`
	PageKey         string   `json:"pageKey,omitempty"`
}

type rpcReq struct {
	Jsonrpc string            `json:"jsonrpc"`
	ID      int               `json:"id"`
	Method  string            `json:"method"`
	Params  []getTransfersParams `json:"params"`
}

type rpcResp struct {
	Result struct {
		Transfers []map[string]any `json:"transfers"`
		PageKey   string           `json:"pageKey"`
	} `json:"result"`
	Error any `json:"error"`
}

func (c *Client) GetAssetTransfersAll(ctx context.Context, p getTransfersParams) ([]map[string]any, error) {
	var out []map[string]any
	page := ""
	for {
		ok, wait, err := c.limiter.Allow(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		pp := p
		pp.PageKey = page

		body, _ := json.Marshal(rpcReq{
			Jsonrpc: "2.0",
			ID:      1,
			Method:  "alchemy_getAssetTransfers",
			Params:  []getTransfersParams{pp},
		})

		url := fmt.Sprintf("%s/%s", c.baseURL, c.apiKey)
		resI, err := c.breaker.Do(ctx, func() (interface{}, error) {
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", c.userAgent)
			resp, err := c.http.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 500 {
				return nil, fmt.Errorf("server error: %d", resp.StatusCode)
			}
			var rr rpcResp
			if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
				return nil, err
			}
			return rr, nil
		})
		if err != nil {
			return nil, err
		}
		rr := resI.(rpcResp)
		out = append(out, rr.Result.Transfers...)
		if rr.Result.PageKey == "" {
			break
		}
		page = rr.Result.PageKey
	}
	return out, nil
}
