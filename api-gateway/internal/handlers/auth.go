package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lr loginReq
		if err := c.ShouldBindJSON(&lr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "invalid request"})
			return
		}
		if cfg.AuthMode == "mvp-gate" {
			if lr.Email == cfg.AdminUser && lr.Password == cfg.AdminPass {
				c.JSON(http.StatusOK, gin.H{"accessToken": "dev-access", "refreshToken": "dev-refresh"})
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid credentials"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"accessToken": "dev-access", "refreshToken": "dev-refresh"})
	}
}

func Refresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"accessToken": "dev-access"})
	}
}
