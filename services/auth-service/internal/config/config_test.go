package config

import (
	"os"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != "8080" || cfg.GRPCPort != "9090" {
		t.Fatalf("default ports mismatch: %s %s", cfg.HTTPPort, cfg.GRPCPort)
	}
	if string(cfg.JWTSigningKey) == "" {
		t.Fatalf("expected default signing key")
	}
	if cfg.EnableMultiUser {
		t.Fatalf("expected multi-user disabled by default")
	}
}

func TestParseEnvOverrides(t *testing.T) {
	os.Setenv("PORT", "18080")
	os.Setenv("GRPC_PORT", "19090")
	os.Setenv("DATABASE_URL", "postgres://x")
	os.Setenv("JWT_ISSUER", "iss")
	os.Setenv("JWT_AUDIENCE", "aud")
	os.Setenv("JWT_SIGNING_KEY", "secret")
	os.Setenv("JWT_ACCESS_TTL_MINUTES", "7")
	os.Setenv("JWT_REFRESH_TTL_HOURS", "24")
	os.Setenv("ENABLE_MULTI_USER", "true")
	cfg, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != "18080" || cfg.GRPCPort != "19090" {
		t.Fatalf("env ports mismatch")
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("db url mismatch")
	}
	if cfg.JWTIssuer != "iss" || cfg.JWTAudience != "aud" {
		t.Fatalf("jwt iss/aud mismatch")
	}
	if string(cfg.JWTSigningKey) != "secret" {
		t.Fatalf("signing key mismatch")
	}
	if cfg.AccessTTL.Minutes() != 7 {
		t.Fatalf("access ttl mismatch")
	}
	if int(cfg.RefreshTTL.Hours()) != 24 {
		t.Fatalf("refresh ttl mismatch")
	}
	if !cfg.EnableMultiUser {
		t.Fatalf("expected multi-user enabled")
	}
}
func TestParseInvalidTTLsFallback(t *testing.T) {
	os.Setenv("JWT_ACCESS_TTL_MINUTES", "not-a-number")
	os.Setenv("JWT_REFRESH_TTL_HOURS", "-5")
	cfg, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(cfg.AccessTTL.Minutes()) != 15 {
		t.Fatalf("expected fallback access ttl 15m, got %v", cfg.AccessTTL)
	}
	if int(cfg.RefreshTTL.Hours()) != 168 {
		t.Fatalf("expected fallback refresh ttl 168h, got %v", cfg.RefreshTTL)
	}
}
