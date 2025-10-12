package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	marketdatapb "github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
)

type MarketDataClient struct {
	conn   *grpc.ClientConn
	client marketdatapb.MarketDataServiceClient
}

func NewMarketDataClient(addr string) (*MarketDataClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create market-data-service client for %s: %w", addr, err)
	}

	log.Printf("Created market-data-service client for %s (connection will be established on first RPC)", addr)

	return &MarketDataClient{
		conn:   conn,
		client: marketdatapb.NewMarketDataServiceClient(conn),
	}, nil
}

func (c *MarketDataClient) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetTokenPrice(ctx, &marketdatapb.GetTokenPriceRequest{
		TokenAddress: tokenAddress,
		Chain:        chain,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get token price: %w", err)
	}

	return resp.UsdPrice, nil
}

func (c *MarketDataClient) GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var pbTokens []*marketdatapb.TokenIdentifier
	for chain, addresses := range tokens {
		for _, addr := range addresses {
			pbTokens = append(pbTokens, &marketdatapb.TokenIdentifier{
				TokenAddress: addr,
				Chain:        chain,
			})
		}
	}

	resp, err := c.client.GetTokenPrices(ctx, &marketdatapb.GetTokenPricesRequest{
		Tokens: pbTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token prices: %w", err)
	}

	prices := make(map[string]map[string]float64)
	for _, price := range resp.Prices {
		if _, exists := prices[price.Chain]; !exists {
			prices[price.Chain] = make(map[string]float64)
		}
		prices[price.Chain][price.TokenAddress] = price.UsdPrice
	}

	return prices, nil
}

func (c *MarketDataClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
