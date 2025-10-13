package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	sharedsecrets "github.com/aezizhu/million-dollar-hunter/api-gateway/internal/secrets"
)

type jwtSecret struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

type localJWTKeys struct {
	currentKID string
	keys       map[string][]byte
}

func getSecretsClient() sharedsecrets.Client {
	switch strings.ToLower(os.Getenv("SECRETS_PROVIDER")) {
	case "aws":
		region := os.Getenv("AWS_REGION")
		cli, err := sharedsecrets.NewAWS(context.Background(), sharedsecrets.AWSConfig{
			Config: sharedsecrets.Config{
				CacheTTL:        time.Hour,
				RefreshInterval: time.Minute,
			},
			Region: region,
		})
		if err != nil {
			return sharedsecrets.NewEnv(sharedsecrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
		}
		return cli
	default:
		return sharedsecrets.NewEnv(sharedsecrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
	}
}

func loadLocalJWTKeys(ctx context.Context, sec sharedsecrets.Client) localJWTKeys {
	prefix := os.Getenv("SECRETS_PREFIX")
	if prefix == "" {
		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}
		prefix = "mdh/" + env + "/auth/jwt"
	}
	lk := localJWTKeys{keys: map[string][]byte{}}
	var cur jwtSecret
	if err := sec.GetJSON(ctx, prefix+"/current", &cur); err == nil && cur.KID != "" && cur.Key != "" {
		lk.keys[cur.KID] = []byte(cur.Key)
		lk.currentKID = cur.KID
	}
	var prev jwtSecret
	if err := sec.GetJSON(ctx, prefix+"/previous", &prev); err == nil && prev.KID != "" && prev.Key != "" {
		lk.keys[prev.KID] = []byte(prev.Key)
	}
	return lk
}

func localValidate(tokenStr string, aud string, lk localJWTKeys, envSecret string) (string, error) {
	if envSecret != "" {
		tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(envSecret), nil
		})
		if err != nil || !tok.Valid {
			return "", fmt.Errorf("env jwt invalid")
		}
		if claims, ok := tok.Claims.(jwt.MapClaims); ok {
			if sub, ok := claims["sub"].(string); ok && sub != "" {
				return sub, nil
			}
		}
		return "", fmt.Errorf("missing sub")
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	tok, err := parser.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if kid, ok := t.Header["kid"].(string); ok && kid != "" {
			if k, ok := lk.keys[kid]; ok {
				return k, nil
			}
		}
		if lk.currentKID != "" {
			if k, ok := lk.keys[lk.currentKID]; ok {
				return k, nil
			}
		}
		return nil, fmt.Errorf("no key")
	})
	if err != nil || !tok.Valid {
		return "", fmt.Errorf("invalid token")
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok {
		if a, ok := claims["aud"]; ok {
			switch v := a.(type) {
			case string:
				if aud != "" && v != aud {
					return "", fmt.Errorf("aud mismatch")
				}
			case []any:
				if aud != "" {
					found := false
					for _, x := range v {
						if s, ok := x.(string); ok && s == aud {
							found = true
							break
						}
					}
					if !found {
						return "", fmt.Errorf("aud missing")
					}
				}
			}
		}
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			return sub, nil
		}
	}
	return "", fmt.Errorf("missing sub")
}

func Auth(cfg config.Config, authConn *grpc.ClientConn) gin.HandlerFunc {
	sec := getSecretsClient()
	if sec != nil {
		sec.StartBackgroundRefresh(context.Background())
	}
	lk := loadLocalJWTKeys(context.Background(), sec)

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
			if err != nil || !resp.GetValid() {
				if authMetrics != nil {
					if err != nil {
						authMetrics.Inc("error")
					} else {
						authMetrics.Inc("invalid")
					}
					authMetrics.Time(start)
				}
				if !(cfg.AuthGRPCFallbackToLocal) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
					return
				}
				if sec != nil {
					nlk := loadLocalJWTKeys(c.Request.Context(), sec)
					if len(nlk.keys) > 0 {
						lk = nlk
					}
				}
			} else {
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

		if cfg.AuthMode == "mvp-gate" {
			if len(tokenStr) < 8 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
				return
			}
			c.Next()
			return
		}

		uid, err := localValidate(tokenStr, cfg.JWTAudience, lk, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		c.Set("user_id", uid)
		c.Next()
	}
}
