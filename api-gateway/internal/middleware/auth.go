package middleware

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	secrets "github.com/aezizhu/million-dollar-hunter/pkg/secrets"
)

type jwtSecret struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

type localJWTKeys struct {
	currentKID string
	hsKeys     map[string][]byte
	rsKeys     map[string]*rsa.PublicKey
}

func getSecretsClient() secrets.Client {
	switch strings.ToLower(os.Getenv("SECRETS_PROVIDER")) {
	case "aws":
		region := os.Getenv("AWS_REGION")
		cli, err := secrets.NewAWS(context.Background(), secrets.AWSConfig{
			Config: secrets.Config{
				CacheTTL:        time.Hour,
				RefreshInterval: time.Minute,
			},
			Region: region,
		})
		if err != nil {
			return secrets.NewEnv(secrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
		}
		return cli
	default:
		return secrets.NewEnv(secrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
	}
}

func loadLocalJWTKeys(ctx context.Context, sec secrets.Client) localJWTKeys {
	prefix := os.Getenv("SECRETS_PREFIX")
	if prefix == "" {
		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}
		prefix = "mdh/" + env + "/auth/jwt"
	}
	lk := localJWTKeys{
		hsKeys: make(map[string][]byte),
		rsKeys: make(map[string]*rsa.PublicKey),
	}
	parseAndStore := func(js jwtSecret) {
		if js.KID == "" || js.Key == "" {
			return
		}
		if blk, _ := pem.Decode([]byte(js.Key)); blk != nil {
			if pub, err := x509.ParsePKIXPublicKey(blk.Bytes); err == nil {
				if rp, ok := pub.(*rsa.PublicKey); ok {
					lk.rsKeys[js.KID] = rp
					if lk.currentKID == "" {
						lk.currentKID = js.KID
					}
					return
				}
			}
		}
		lk.hsKeys[js.KID] = []byte(js.Key)
		if lk.currentKID == "" {
			lk.currentKID = js.KID
		}
	}

	var cur jwtSecret
	if err := sec.GetJSON(ctx, prefix+"/current", &cur); err == nil {
		parseAndStore(cur)
	}
	var prev jwtSecret
	if err := sec.GetJSON(ctx, prefix+"/previous", &prev); err == nil {
		parseAndStore(prev)
	}
	return lk
}

func localValidate(tokenStr string, aud string, lk localJWTKeys, envSecret string, issuer string) (string, error) {
	leeway := time.Second * 60

	checkStandard := func(cl *jwt.RegisteredClaims) error {
		now := time.Now()
		if cl.ExpiresAt != nil && now.After(cl.ExpiresAt.Time.Add(leeway)) {
			return fmt.Errorf("expired")
		}
		if cl.NotBefore != nil && now.Before(cl.NotBefore.Time.Add(-leeway)) {
			return fmt.Errorf("not_before")
		}
		if cl.IssuedAt != nil && now.Before(cl.IssuedAt.Time.Add(-leeway)) {
			return fmt.Errorf("issued_at_in_future")
		}
		if issuer != "" && cl.Issuer != issuer {
			return fmt.Errorf("iss_mismatch")
		}
		if aud != "" {
			if len(cl.Audience) == 0 {
				return fmt.Errorf("aud_missing")
			}
			found := false
			for _, a := range cl.Audience {
				if a == aud {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("aud_mismatch")
			}
		}
		return nil
	}

	if envSecret != "" {
		var rc jwt.RegisteredClaims
		parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		tok, err := parser.ParseWithClaims(tokenStr, &rc, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(envSecret), nil
		})
		if err != nil || !tok.Valid {
			return "", fmt.Errorf("invalid")
		}
		if err := checkStandard(&rc); err != nil {
			return "", err
		}
		if rc.Subject == "" {
			return "", fmt.Errorf("sub_missing")
		}
		return rc.Subject, nil
	}

	var rc jwt.RegisteredClaims
	parser := jwt.NewParser(jwt.WithValidMethods([]string{
		jwt.SigningMethodHS256.Alg(),
		jwt.SigningMethodRS256.Alg(),
	}))
	tok, err := parser.ParseWithClaims(tokenStr, &rc, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		switch t.Method.Alg() {
		case jwt.SigningMethodHS256.Alg():
			if kid != "" {
				if k, ok := lk.hsKeys[kid]; ok {
					return k, nil
				}
			}
			if lk.currentKID != "" {
				if k, ok := lk.hsKeys[lk.currentKID]; ok {
					return k, nil
				}
			}
		case jwt.SigningMethodRS256.Alg():
			if kid != "" {
				if k, ok := lk.rsKeys[kid]; ok {
					return k, nil
				}
			}
			if lk.currentKID != "" {
				if k, ok := lk.rsKeys[lk.currentKID]; ok {
					return k, nil
				}
			}
		default:
			return nil, jwt.ErrSignatureInvalid
		}
		return nil, fmt.Errorf("no_key")
	})
	if err != nil || !tok.Valid {
		return "", fmt.Errorf("invalid")
	}
	if err := checkStandard(&rc); err != nil {
		return "", err
	}
	if rc.Subject == "" {
		return "", fmt.Errorf("sub_missing")
	}
	return rc.Subject, nil
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
					if len(nlk.hsKeys) > 0 || len(nlk.rsKeys) > 0 {
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

		uid, err := localValidate(tokenStr, cfg.JWTAudience, lk, cfg.JWTSecret, cfg.JWTIssuer)
		if err != nil {
			reason := err.Error()
			switch reason {
			case "expired":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_expired"})
			case "not_before":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_not_yet_valid"})
			case "issued_at_in_future":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_issued_at_in_future"})
			case "iss_mismatch":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "issuer_mismatch"})
			case "aud_missing", "aud_mismatch":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "audience_invalid"})
			case "sub_missing":
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "subject_missing"})
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			}
			return
		}
		c.Set("user_id", uid)
		c.Next()
	}
}
