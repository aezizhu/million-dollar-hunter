package clients

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	mdpb "github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MarketDataGRPCClient struct {
	cc      *grpc.ClientConn
	client  mdpb.MarketDataServiceClient
	timeout time.Duration
}

func NewMarketDataGRPCClient(addr string, timeout time.Duration, useTLS bool) (*MarketDataGRPCClient, error) {
	var opts []grpc.DialOption
	if useTLS {
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	cc, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &MarketDataGRPCClient{
		cc:      cc,
		client:  mdpb.NewMarketDataServiceClient(cc),
		timeout: timeout,
	}, nil
}

func (c *MarketDataGRPCClient) Close() error { return c.cc.Close() }

func (c *MarketDataGRPCClient) GetTokenPrices(ctx context.Context, tokens []ports.TokenIdentifier) ([]ports.TokenPrice, error) {
	seen := make(map[string]struct{}, len(tokens))
	req := &mdpb.GetTokenPricesRequest{}
	for _, t := range tokens {
		addr := strings.ToLower(strings.TrimSpace(t.Address))
		chain := strings.ToLower(strings.TrimSpace(t.Chain))
		if addr == "" || chain == "" {
			continue
		}
		key := chain + ":" + addr
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		req.Tokens = append(req.Tokens, &mdpb.TokenIdentifier{
			TokenAddress: addr,
			Chain:        chain,
		})
	}
	if len(req.Tokens) == 0 {
		return []ports.TokenPrice{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.client.GetTokenPrices(ctx, req)
	if err != nil {
		return nil, err
	}
	out := make([]ports.TokenPrice, 0, len(resp.Prices))
	for _, p := range resp.Prices {
		out = append(out, ports.TokenPrice{
			Address:    p.TokenAddress,
			Chain:      p.Chain,
			USDPrice:   p.UsdPrice,
			LastUpdate: p.LastUpdated,
			FromCache:  p.FromCache,
		})
	}
	return out, nil
}

func (c *MarketDataGRPCClient) GetTokenPrice(ctx context.Context, t ports.TokenIdentifier) (*ports.TokenPrice, error) {
	addr := strings.ToLower(strings.TrimSpace(t.Address))
	chain := strings.ToLower(strings.TrimSpace(t.Chain))
	if addr == "" || chain == "" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.client.GetTokenPrice(ctx, &mdpb.GetTokenPriceRequest{
		TokenAddress: addr,
		Chain:        chain,
	})
	if err != nil {
		return nil, err
	}
	return &ports.TokenPrice{
		Address:    resp.TokenAddress,
		Chain:      resp.Chain,
		USDPrice:   resp.UsdPrice,
		LastUpdate: resp.LastUpdated,
		FromCache:  resp.FromCache,
	}, nil
}
