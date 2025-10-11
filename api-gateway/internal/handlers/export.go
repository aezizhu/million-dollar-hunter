package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

		if portfolioConn == nil {
			logger.Warn().Str("address", address).Msg("Portfolio service not configured for export")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "portfolio service not configured",
			})
			return
		}

		var exportFormat pb.ExportFormat
		if format == "csv" {
			exportFormat = pb.ExportFormat_EXPORT_FORMAT_CSV
		} else {
			exportFormat = pb.ExportFormat_EXPORT_FORMAT_JSON
		}

		client := pb.NewPortfolioServiceClient(portfolioConn)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		resp, err := client.Export(ctx, &pb.ExportRequest{
			WalletId: address,
			Format:   exportFormat,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.NotFound:
					logger.Info().Str("address", address).Msg("Wallet not found for export")
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "not_found",
						"message": "wallet not found",
					})
					return
				case codes.InvalidArgument:
					logger.Warn().Str("address", address).Err(err).Msg("Invalid argument for export")
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "validation_error",
						"message": st.Message(),
					})
					return
				case codes.DeadlineExceeded:
					logger.Error().Str("address", address).Err(err).Msg("Export timeout")
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"error":   "timeout",
						"message": "export request timed out",
					})
					return
				case codes.Unavailable:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service unavailable for export")
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "service_unavailable",
						"message": "portfolio service is temporarily unavailable",
					})
					return
				}
			}
			logger.Error().Str("address", address).Err(err).Msg("Failed to export wallet data")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to export wallet data",
			})
			return
		}

		logger.Info().Str("address", address).Str("format", format).Str("path", resp.Path).Msg("Export completed successfully")
		
		c.JSON(http.StatusOK, gin.H{
			"path":    resp.Path,
			"format":  format,
			"message": "Export file generated successfully. Note: File path returned, streaming not yet implemented.",
		})
	}
}
