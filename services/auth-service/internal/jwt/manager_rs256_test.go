package jwtmgr

import (
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore"
)

func TestRS256_IssuanceIncludesKIDAndValidate(t *testing.T) {
	ks, _ := keystore.NewMemoryStore(24*time.Hour, true)
	m := NewWithKeyStore("iss", "aud", time.Minute, time.Hour, ks, nil, false)
	tok, _, err := m.GenerateToken("u1", "e@example.com", time.Minute)
	if err != nil || tok == "" {
		t.Fatalf("failed to issue token: %v", err)
	}
	if _, err := m.ValidateToken(tok, ""); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

func TestRS256_LegacyFallback(t *testing.T) {
	legacy := []byte("legacy")
	ks, _ := keystore.NewMemoryStore(24*time.Hour, true)
	m := NewWithKeyStore("iss", "aud", time.Minute, time.Hour, ks, legacy, true)
	legacyMgr := New("iss", "aud", time.Minute, time.Hour, legacy)
	tok, _, err := legacyMgr.GenerateToken("u1", "e@example.com", time.Minute)
	if err != nil {
		t.Fatalf("legacy issue failed: %v", err)
	}
	if _, err := m.ValidateToken(tok, ""); err != nil {
		t.Fatalf("fallback failed: %v", err)
	}
}
