package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExportWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "json")
		if format == "csv" {
			c.Header("Content-Type", "text/csv")
			c.String(http.StatusOK, "address,balance,usdValue\n")
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
	}
}
