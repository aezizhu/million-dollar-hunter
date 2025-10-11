package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	PortfolioConn   *grpc.ClientConn
	MarketDataConn  *grpc.ClientConn
}

func NewGRPCClients(portfolioAddr, marketDataAddr string) (*GRPCClients, error) {
	clients := &GRPCClients{}
	
	if portfolioAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		conn, err := grpc.DialContext(ctx, portfolioAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to portfolio service at %s: %w", portfolioAddr, err)
		}
		clients.PortfolioConn = conn
	}
	
	if marketDataAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		conn, err := grpc.DialContext(ctx, marketDataAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			if clients.PortfolioConn != nil {
				clients.PortfolioConn.Close()
			}
			return nil, fmt.Errorf("failed to connect to market data service at %s: %w", marketDataAddr, err)
		}
		clients.MarketDataConn = conn
	}
	
	return clients, nil
}

func (c *GRPCClients) Close() {
	if c.PortfolioConn != nil {
		c.PortfolioConn.Close()
	}
	if c.MarketDataConn != nil {
		c.MarketDataConn.Close()
	}
}
