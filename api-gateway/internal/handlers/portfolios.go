package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func ListPortfolios(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info().Msg("ListPortfolios endpoint not yet implemented")
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "Portfolio listing requires user-wallet association implementation in portfolio-service",
		})
	}
}

type AddWalletRequest struct {
	Address  string `json:"address" binding:"required"`
	Chain    string `json:"chain" binding:"required,oneof=ethereum bsc polygon arbitrum optimism solana"`
	Nickname string `json:"nickname"`
}

func AddWallet(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddWalletRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid AddWallet request")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": err.Error(),
			})
			return
		}

		logger.Info().Str("address", req.Address).Str("chain", req.Chain).Msg("AddWallet endpoint not yet implemented")
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "Wallet tracking requires ingestion service integration",
		})
	}
}
