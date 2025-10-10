package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort       string
	GRPCPort       string
	DatabaseURL    string
	JWTIssuer      string
	JWTAudience    string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	JWTSigningKey  []byte
	EnableMultiUser bool
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
	return cfg, nil
}
