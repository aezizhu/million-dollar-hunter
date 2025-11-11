// Package config handles environment variable configuration for the auth service.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
//
// The configuration system provides type-safe access to environment variables
// with validation and sensible defaults, ensuring the service can operate
// correctly across different deployment environments.
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
	JWTCurrentKID   string
	JWTKeys         map[string][]byte
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

	currentKID := os.Getenv("JWT_CURRENT_KID")
	currentKey := os.Getenv("JWT_CURRENT_KEY")
	prevKID := os.Getenv("JWT_PREVIOUS_KID")
	prevKey := os.Getenv("JWT_PREVIOUS_KEY")
	if currentKID != "" && currentKey != "" {
		cfg.JWTCurrentKID = currentKID
		cfg.JWTKeys = map[string][]byte{currentKID: []byte(currentKey)}
		if prevKID != "" && prevKey != "" {
			cfg.JWTKeys[prevKID] = []byte(prevKey)
		}
	}

	return cfg, nil
}
