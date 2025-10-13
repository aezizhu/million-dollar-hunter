package handlers

import (
	"context"
	"net/http"
	"time"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetWallet(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			logger.Warn().Msg("GetWallet called with empty address")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "address is required",
			})
			return
		}

		if portfolioConn == nil {
			logger.Warn().Str("address", address).Msg("Portfolio service not configured")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "portfolio service not configured",
			})
			return
		}

		client := pb.NewPortfolioServiceClient(portfolioConn)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetPortfolio(ctx, &pb.GetPortfolioRequest{
			WalletId: address,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.NotFound:
					logger.Info().Str("address", address).Msg("Wallet not found")
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "not_found",
						"message": "wallet not found",
					})
					return
				case codes.InvalidArgument:
					logger.Warn().Str("address", address).Err(err).Msg("Invalid argument")
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "validation_error",
						"message": st.Message(),
					})
					return
				case codes.DeadlineExceeded:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service timeout")
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"error":   "timeout",
						"message": "request to portfolio service timed out",
					})
					return
				case codes.Unavailable:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service unavailable")
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "service_unavailable",
						"message": "portfolio service is temporarily unavailable",
					})
					return
				}
			}
			logger.Error().Str("address", address).Err(err).Msg("Failed to fetch wallet details")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to fetch wallet details",
			})
			return
		}

		assets := make([]gin.H, 0, len(resp.Assets))
		for _, asset := range resp.Assets {
			assets = append(assets, gin.H{
				"tokenAddress": asset.TokenAddress,
				"symbol":       asset.Symbol,
				"balance":      asset.Amount,
				"usdValue":     asset.UsdValue,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"address": address,
			"assets":  assets,
		})
	}
}

func GetTransactions(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info().Msg("GetTransactions endpoint not yet implemented")
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "Transaction history endpoint requires additional gRPC implementation in portfolio-service",
		})
	}
}
