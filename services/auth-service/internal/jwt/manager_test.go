package jwtmgr

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	access, _, err := m.GenerateToken("uid", "e@x.com", 1*time.Minute)
	if err != nil || access == "" {
		t.Fatalf("expected token, got err=%v", err)
	}
	cl, err := m.ValidateToken(access, "aud")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if cl.UserID != "uid" || cl.Email != "e@x.com" {
		t.Fatalf("claims mismatch")
	}
}

func TestAudienceIssuerChecks(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	access, _, _ := m.GenerateToken("uid", "e@x.com", 1*time.Minute)
	if _, err := m.ValidateToken(access, "wrong-aud"); err == nil {
		t.Fatalf("expected audience error")
	}
	m2 := New("wrong-issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	if _, err := m2.ValidateToken(access, "aud"); err == nil {
		t.Fatalf("expected issuer error")
	}
}

func TestExpiration(t *testing.T) {
	m := New("issuer", "aud", 1*time.Millisecond, 1*time.Millisecond, []byte("key"))
	access, _, _ := m.GenerateToken("uid", "e@x.com", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if _, err := m.ValidateToken(access, "aud"); err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestGeneratePair(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 2*time.Minute, []byte("key"))
	a, r, exp, err := m.GeneratePair("u", "e@x.com")
	if err != nil || a == "" || r == "" || exp.Before(time.Now()) {
		t.Fatalf("invalid pair result")
	}
}

func TestInvalidSignature(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	token, _, _ := m.GenerateToken("u", "e@x.com", 1*time.Minute)
	other := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("other"))
	if _, err := other.ValidateToken(token, "aud"); err == nil {
		t.Fatalf("expected signature error")
	}
}

func TestMalformedToken(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	if _, err := m.ValidateToken("not-a-jwt", "aud"); err == nil {
		t.Fatalf("expected parse error")
	}
	parts := strings.Split("a.b.", ".")
	_ = parts
}
