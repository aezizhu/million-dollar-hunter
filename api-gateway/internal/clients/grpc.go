package clients

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	PortfolioConn  *grpc.ClientConn
	MarketDataConn *grpc.ClientConn
}

func NewGRPCClients(portfolioAddr, marketDataAddr string, logger zerolog.Logger) *GRPCClients {
	clients := &GRPCClients{}

	if portfolioAddr != "" {
		conn, err := grpc.NewClient(portfolioAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Warn().Err(err).Str("addr", portfolioAddr).Msg("Failed to create portfolio service client")
		} else {
			clients.PortfolioConn = conn
			logger.Info().Str("addr", portfolioAddr).Msg("Portfolio service client initialized (non-blocking)")
			
			go func() {
				state := conn.GetState()
				conn.Connect()
				logger.Debug().Str("state", state.String()).Msg("Portfolio service connection state")
			}()
		}
	}

	if marketDataAddr != "" {
		conn, err := grpc.NewClient(marketDataAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Warn().Err(err).Str("addr", marketDataAddr).Msg("Failed to create market data service client")
		} else {
			clients.MarketDataConn = conn
			logger.Info().Str("addr", marketDataAddr).Msg("Market data service client initialized (non-blocking)")
			
			go func() {
				state := conn.GetState()
				conn.Connect()
				logger.Debug().Str("state", state.String()).Msg("Market data service connection state")
			}()
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
}
