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

func GetWallet(portfolioConn *grpc.ClientConn) gin.HandlerFunc {
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

		client := pb.NewPortfolioServiceClient(portfolioConn)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetPortfolio(ctx, &pb.GetPortfolioRequest{
			WalletId: address,
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
			"chain":   c.Query("chain"),
			"assets":  assets,
			"history": []interface{}{},
		})
	}
}

func GetTransactions(portfolioConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items":    []interface{}{},
			"page":     1,
			"pageSize": 50,
			"total":    0,
			"message":  "Transaction history requires additional gRPC endpoint in portfolio-service",
		})
	}
}
