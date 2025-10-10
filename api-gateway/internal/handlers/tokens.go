package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TopHolders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"token":   c.Param("tokenAddress"),
			"holders": []interface{}{},
		})
	}
}
