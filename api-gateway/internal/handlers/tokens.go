package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func TopHolders(marketDataConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenAddress := c.Param("tokenAddress")
		if tokenAddress == "" {
			logger.Warn().Msg("TopHolders called with empty tokenAddress")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "tokenAddress is required",
			})
			return
		}

		if marketDataConn == nil {
			logger.Warn().Str("token", tokenAddress).Msg("Market data service not configured")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "market data service not configured",
			})
			return
		}

		logger.Info().Str("token", tokenAddress).Msg("TopHolders endpoint not yet implemented")
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "Top holders endpoint requires implementation in market-data-service",
		})
	}
}
