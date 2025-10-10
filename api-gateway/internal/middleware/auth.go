package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"api-gateway/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	cfg    *config.Config
	logger zerolog.Logger
}

func NewAuthMiddleware(cfg *config.Config, logger zerolog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.respondUnauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			a.respondUnauthorized(c, "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.cfg.JWTSecret), nil
		})

		if err != nil {
			a.logger.Debug().Err(err).Msg("jwt parse failed")
			a.respondUnauthorized(c, "invalid token")
			return
		}

		if !token.Valid {
			a.respondUnauthorized(c, "token not valid")
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			a.respondUnauthorized(c, "token expired")
			return
		}

		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)

		c.Next()
	}
}

func (a *AuthMiddleware) ValidateMVPCredentials(username, password string) bool {
	return username == a.cfg.MVPUsername && password == a.cfg.MVPPassword
}

func (a *AuthMiddleware) GenerateToken(userID, email, username string) (string, error) {
	expiresAt := time.Now().Add(time.Duration(a.cfg.JWTExpiryMinutes) * time.Minute)
	
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "million-hunter-api-gateway",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.JWTSecret))
}

func (a *AuthMiddleware) GenerateRefreshToken(userID string) (string, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "million-hunter-api-gateway-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.JWTSecret))
}

func (a *AuthMiddleware) respondUnauthorized(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", "Bearer")
	c.JSON(http.StatusUnauthorized, gin.H{
		"error":   "unauthorized",
		"message": message,
	})
	c.Abort()
}

func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
