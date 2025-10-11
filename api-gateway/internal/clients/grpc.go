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

func NewGRPCClients(portfolioAddr, marketDataAddr, authAddr string, logger zerolog.Logger) *GRPCClients {
	clients := &GRPCClients{}

	if portfolioAddr != "" {
		conn, err := grpc.NewClient(portfolioAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Error().Err(err).Str("addr", portfolioAddr).Msg("failed to create portfolio gRPC client")
		} else {
			clients.PortfolioConn = conn
			logger.Info().Str("addr", portfolioAddr).Msg("portfolio gRPC client created")
		}
	}

	if marketDataAddr != "" {
		conn, err := grpc.NewClient(marketDataAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Error().Err(err).Str("addr", marketDataAddr).Msg("failed to create market-data gRPC client")
		} else {
			clients.MarketDataConn = conn
			logger.Info().Str("addr", marketDataAddr).Msg("market-data gRPC client created")
		}
	}

	if authAddr != "" {
		conn, err := grpc.NewClient(authAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Error().Err(err).Str("addr", authAddr).Msg("failed to create auth gRPC client")
		} else {
			clients.AuthConn = conn
			logger.Info().Str("addr", authAddr).Msg("auth gRPC client created")
		}
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
