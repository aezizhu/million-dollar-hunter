package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ExportWallet(portfolioConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "address is required",
			})
			return
		}

		if portfolioConn == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "portfolio service not configured",
			})
			return
		}

		format := c.DefaultQuery("format", "json")
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
			if ok && st.Code() == codes.NotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "not_found",
					"message": "wallet not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to export wallet data",
			})
			return
		}

		if format == "csv" {
			c.Header("Content-Type", "text/csv")
			c.Header("Content-Disposition", "attachment; filename="+resp.Path)
		}
		
		c.JSON(http.StatusOK, gin.H{
			"path":    resp.Path,
			"message": "Export file generated successfully",
		})
	}
}
