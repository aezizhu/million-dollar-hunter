package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"address": c.Param("address"),
			"chain":   c.Query("chain"),
			"assets":  []interface{}{},
			"history": []interface{}{},
		})
	}
}

func GetTransactions() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items":    []interface{}{},
			"page":     1,
			"pageSize": 50,
			"total":    0,
		})
	}
}
