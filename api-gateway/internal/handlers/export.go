package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func ExportWallet(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			logger.Warn().Msg("ExportWallet called with empty address")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "address is required",
			})
			return
		}

		format := c.DefaultQuery("format", "json")
		if format != "json" && format != "csv" {
			logger.Warn().Str("format", format).Msg("Invalid export format requested")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "format must be 'json' or 'csv'",
			})
			return
		}

		logger.Info().
			Str("address", address).
			Str("format", format).
			Msg("Export endpoint not yet fully implemented")
		
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "not_implemented",
			"message": "File export and download is not yet available. The backend can generate files but the gateway cannot serve them to clients.",
		})
	}
}
