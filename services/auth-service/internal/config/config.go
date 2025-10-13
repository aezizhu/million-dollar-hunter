// Package config handles environment variable configuration for the auth service.
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort        string
	GRPCPort        string
	DatabaseURL     string
	JWTIssuer       string
	JWTAudience     string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	JWTSigningKey   []byte
	EnableMultiUser bool

	JWTKeyStoreFile string
	JWTGraceHours   int
	JWTDisableLegacy bool
	JWTRotateAdminToken string
	JWKSCacheTTLSeconds int
	JWKSRotationTTLSeconds int
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func Parse() (Config, error) {
	var cfg Config
	cfg.HTTPPort = getenv("PORT", "8080")
	cfg.GRPCPort = getenv("GRPC_PORT", "9090")
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.JWTIssuer = getenv("JWT_ISSUER", "million-hunter")
	cfg.JWTAudience = getenv("JWT_AUDIENCE", "million-hunter-client")
	cfg.JWTSigningKey = []byte(getenv("JWT_SIGNING_KEY", "dev-insecure-change-me"))
	accessTTL := getenv("JWT_ACCESS_TTL_MINUTES", "15")
	refreshTTL := getenv("JWT_REFRESH_TTL_HOURS", "168")

	attlMin, err := strconv.Atoi(accessTTL)
	if err != nil || attlMin <= 0 {
		attlMin = 15
	}
	rttlH, err := strconv.Atoi(refreshTTL)
	if err != nil || rttlH <= 0 {
		rttlH = 168
	}

	cfg.AccessTTL = time.Duration(attlMin) * time.Minute
	cfg.RefreshTTL = time.Duration(rttlH) * time.Hour
	cfg.EnableMultiUser = getenv("ENABLE_MULTI_USER", "false") == "true"

	cfg.JWTKeyStoreFile = getenv("JWT_KEYSTORE_FILE", "")
	if v := getenv("JWT_GRACE_HOURS", "24"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.JWTGraceHours = n
		} else {
			cfg.JWTGraceHours = 24
		}
	} else {
		cfg.JWTGraceHours = 24
	}
	cfg.JWTDisableLegacy = getenv("JWT_DISABLE_LEGACY", "false") == "true"
	cfg.JWTRotateAdminToken = getenv("JWT_ROTATE_ADMIN_TOKEN", "")
	if v := getenv("JWKS_CACHE_TTL_SECONDS", "300"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.JWKSCacheTTLSeconds = n
		}
	}
	if v := getenv("JWKS_ROTATION_TTL_SECONDS", "60"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.JWKSRotationTTLSeconds = n
		}
	}
	return cfg, nil
}
