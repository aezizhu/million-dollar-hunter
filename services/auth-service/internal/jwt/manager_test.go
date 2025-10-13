package jwtmgr

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestExpiredToken(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claims := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Minute)),
			ID:        "exp",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s, "aud"); err == nil {
		t.Fatalf("expected expiration error")
	}
}

func TestIssuerMismatch(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claims := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bad-issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "iss",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s, "aud"); err == nil {
		t.Fatalf("expected issuer mismatch error")
	}
}

func TestAudienceArrayAndString(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claimsArr := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "aud1",
		},
	}
	tok1 := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsArr)
	s1, err := tok1.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s1, "aud"); err != nil {
		t.Fatalf("unexpected error for aud array: %v", err)
	}

	c2 := jwt.MapClaims{
		"iss":   "issuer",
		"sub":   "u",
		"aud":   "aud",
		"iat":   now.Unix(),
		"exp":   now.Add(1 * time.Minute).Unix(),
		"nbf":   now.Unix(),
		"jti":   "aud2",
		"uid":   "u",
		"email": "e@x.com",
	}
	var s2 string

	tok2 := jwt.NewWithClaims(jwt.SigningMethodHS256, c2)
	s2, err = tok2.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s2, "aud"); err != nil {
		t.Fatalf("unexpected error for aud string: %v", err)
	}
}

func TestIssuedAtInFuture(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claims := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now.Add(2 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(3 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			ID:        "iatf",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s, "aud"); err == nil {
		t.Fatalf("expected iat-in-future error")
	}
}

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
func TestNotBeforeInFuture(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claims := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			ID:        "1",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte("key"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := m.ValidateToken(s, "aud"); err == nil {
		t.Fatalf("expected not-before error")
	}
}

func TestNoneAlgorithmAttackRejected(t *testing.T) {
	m := New("issuer", "aud", 1*time.Minute, 1*time.Hour, []byte("key"))
	now := time.Now().UTC()
	claims := Claims{
		UserID: "u",
		Email:  "e@x.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "issuer",
			Subject:   "u",
			Audience:  jwt.ClaimStrings{"aud"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "2",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	s, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none err: %v", err)
	}
	if _, err := m.ValidateToken(s, "aud"); err == nil {
		t.Fatalf("expected validation to reject alg none")
	}
}
