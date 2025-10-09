package jwtmgr

import (
	"os"
	"testing"
	"time"
)

func TestGeneratePair_Coverage(t *testing.T) {
	_ = os.Setenv("JWT_SIGNING_KEY", "test-secret")
	_ = os.Setenv("JWT_ISSUER", "mdh-auth")
	_ = os.Setenv("JWT_AUDIENCE", "mdh-api")
	_ = os.Setenv("JWT_ACCESS_TTL", "2m")
	_ = os.Setenv("JWT_REFRESH_TTL", "5m")
	m := New("mdh-auth", "mdh-api", 2*time.Minute, 5*time.Minute, []byte("test-secret"))
	access, refresh, exp, err := m.GeneratePair("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("GeneratePair error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("missing tokens")
	}
	if time.Until(exp) <= 0 {
		t.Fatalf("expires not in future")
	}
	claims, err := m.ValidateToken(access, "mdh-api")
	if err != nil || claims == nil {
		t.Fatalf("validate access failed: %v", err)
	}
}
