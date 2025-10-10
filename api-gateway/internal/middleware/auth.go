package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		if cfg.JWTSecret != "" {
			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !tok.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
				return
			}
			if claims, ok := tok.Claims.(jwt.MapClaims); ok {
				if exp, ok := claims["exp"].(float64); ok {
					if time.Unix(int64(exp), 0).Before(time.Now()) {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_expired"})
						return
					}
				}
				if sub, ok := claims["sub"].(string); ok && sub != "" {
					c.Set("user_id", sub)
				}
			}
		} else if cfg.AuthMode == "mvp-gate" {
			if len(tokenStr) < 8 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing jwt secret"})
			return
		}
		c.Next()
	}
}
