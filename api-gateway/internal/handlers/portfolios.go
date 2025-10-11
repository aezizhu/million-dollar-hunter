package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func ListPortfolios(portfolioConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items":    []interface{}{},
			"page":     1,
			"pageSize": 20,
			"total":    0,
			"message":  "Portfolio listing requires user-wallet association implementation in portfolio-service",
		})
	}
}

type AddWalletRequest struct {
	Address  string `json:"address" binding:"required"`
	Chain    string `json:"chain" binding:"required"`
	Nickname string `json:"nickname"`
}

func AddWallet(portfolioConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddWalletRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"jobId":   "00000000-0000-0000-0000-000000000000",
			"status":  "queued",
			"message": "Wallet tracking requires ingestion service integration",
		})
	}
}
