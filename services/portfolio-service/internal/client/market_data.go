package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	
	marketdatapb "github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/ports"
)

var _ ports.MarketDataClient = (*MarketDataClient)(nil)

type MarketDataClient struct {
	conn          *grpc.ClientConn
	client        marketdatapb.MarketDataServiceClient
	singleTimeout time.Duration
	batchTimeout  time.Duration
}

func NewMarketDataClient(cfg config.Config) (*MarketDataClient, error) {
	var creds credentials.TransportCredentials
	
	if cfg.MarketDataTLSEnabled {
		tlsConfig := &tls.Config{
			ServerName: cfg.MarketDataServerName,
		}
		
		if cfg.MarketDataCAFile != "" {
			caCert, err := os.ReadFile(cfg.MarketDataCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA file: %w", err)
			}
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to append CA cert")
			}
			tlsConfig.RootCAs = certPool
		}
		
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}
	
	conn, err := grpc.NewClient(
		cfg.MarketDataServiceAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 3 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create market-data client: %w", err)
	}

	log.Printf("Initialized market-data-service client for %s (lazy connect)", cfg.MarketDataServiceAddr)

	return &MarketDataClient{
		conn:          conn,
		client:        marketdatapb.NewMarketDataServiceClient(conn),
		singleTimeout: cfg.MarketDataSingleTimeout,
		batchTimeout:  cfg.MarketDataBatchTimeout,
	}, nil
}

func (c *MarketDataClient) withDefaultTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}

func (c *MarketDataClient) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error) {
	ctx, cancel := c.withDefaultTimeout(ctx, c.singleTimeout)
	defer cancel()

	resp, err := c.client.GetTokenPrice(ctx, &marketdatapb.GetTokenPriceRequest{
		TokenAddress: tokenAddress,
		Chain:        chain,
	})
	if err != nil {
		return 0, fmt.Errorf("get token price: %w", err)
	}

	return resp.UsdPrice, nil
}

func (c *MarketDataClient) GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
	ctx, cancel := c.withDefaultTimeout(ctx, c.batchTimeout)
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
		return nil, fmt.Errorf("get token prices: %w", err)
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
