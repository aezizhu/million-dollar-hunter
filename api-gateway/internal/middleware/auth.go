package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.AuthMode == "mvp-gate" {
			h := c.GetHeader("Authorization")
			if !strings.HasPrefix(h, "Bearer ") || len(h) < 8 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing token"})
				return
			}
			c.Next()
			return
		}
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing token"})
			return
		}
		c.Next()
	}
}
