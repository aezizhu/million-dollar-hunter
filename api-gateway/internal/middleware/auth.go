package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

func Auth(cfg config.Config, authConn *grpc.ClientConn) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")

	if cfg.AuthValidateMode == "grpc" && authConn != nil {
			cli := gen.NewAuthServiceClient(authConn)
			timeout := time.Duration(cfg.AuthGRPCTimeoutMs) * time.Millisecond
			if timeout <= 0 {
				timeout = 2 * time.Second
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()
			resp, err := cli.ValidateToken(ctx, &gen.ValidateRequest{
				Token:       tokenStr,
				ExpectedAud: cfg.JWTAudience,
			})
			if err != nil {
				if !(cfg.AuthGRPCFallbackToLocal && (cfg.JWTSecret != "" || cfg.AuthMode == "mvp-gate")) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
					return
				}
			} else {
				if !resp.GetValid() {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
					return
				}
				if uid := resp.GetUserId(); uid != "" {
					c.Set("user_id", uid)
				}
				c.Next()
				return
			}
		}

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
