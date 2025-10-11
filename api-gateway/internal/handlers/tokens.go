package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func TopHolders(marketDataConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenAddress := c.Param("tokenAddress")
		if tokenAddress == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "tokenAddress is required",
			})
			return
		}

		if marketDataConn == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "market data service not configured",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":   tokenAddress,
			"holders": []interface{}{},
			"message": "Top holders endpoint requires implementation in market-data-service",
		})
	}
}
