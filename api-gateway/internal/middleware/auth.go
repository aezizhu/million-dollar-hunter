package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
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
			var authMetrics *observability.AuthGRPCMetrics
			if v, ok := c.Get("auth_grpc_metrics"); ok {
				if m, ok := v.(*observability.AuthGRPCMetrics); ok {
					authMetrics = m
				}
			}
			start := time.Now()
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()
			resp, err := cli.ValidateToken(ctx, &gen.ValidateRequest{
				Token:       tokenStr,
				ExpectedAud: cfg.JWTAudience,
			})
			if err != nil {
				if authMetrics != nil {
					authMetrics.Inc("error")
					authMetrics.Time(start)
				}
				rid, _ := c.Get("request_id")
				_ = c.Error(fmt.Errorf("grpc_auth_error rid=%v err=%v", rid, err))
				if !(cfg.AuthGRPCFallbackToLocal && (cfg.JWTSecret != "" || cfg.AuthMode == "mvp-gate")) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
					return
				}
			} else {
				if !resp.GetValid() {
					if authMetrics != nil {
						authMetrics.Inc("invalid")
						authMetrics.Time(start)
					}
					rid, _ := c.Get("request_id")
					_ = c.Error(fmt.Errorf("grpc_auth_invalid rid=%v reason=%v", rid, resp.GetReason()))
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
					return
				}
				if authMetrics != nil {
					authMetrics.Inc("success")
					authMetrics.Time(start)
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
