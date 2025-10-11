package clients

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	PortfolioConn  *grpc.ClientConn
	MarketDataConn *grpc.ClientConn
	AuthConn       *grpc.ClientConn
}

//
func NewGRPCClients(portfolioAddr, marketDataAddr, authAddr string, logger zerolog.Logger) *GRPCClients {
	clients := &GRPCClients{}

	if portfolioAddr != "" {
		conn, _ := grpc.NewClient(portfolioAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		clients.PortfolioConn = conn
		logger.Info().
			Str("addr", portfolioAddr).
			Msg("Portfolio service client created (connection will be established on first use)")
	}

	if marketDataAddr != "" {
		conn, _ := grpc.NewClient(marketDataAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		clients.MarketDataConn = conn
		logger.Info().
			Str("addr", marketDataAddr).
			Msg("Market data service client created (connection will be established on first use)")
	}

	if authAddr != "" {
		conn, _ := grpc.NewClient(authAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		clients.AuthConn = conn
		logger.Info().
			Str("addr", authAddr).
			Msg("Auth service client created (connection will be established on first use)")
	}

	return clients
}

func (c *GRPCClients) Close() {
	if c.PortfolioConn != nil {
		c.PortfolioConn.Close()
	}
	if c.MarketDataConn != nil {
		c.MarketDataConn.Close()
	}
	if c.AuthConn != nil {
		c.AuthConn.Close()
	}
}
