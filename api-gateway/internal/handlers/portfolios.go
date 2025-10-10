package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListPortfolios() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items":    []interface{}{},
			"page":     1,
			"pageSize": 20,
			"total":    0,
		})
	}
}

func AddWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"jobId":  "00000000-0000-0000-0000-000000000000",
			"status": "queued",
		})
	}
}
