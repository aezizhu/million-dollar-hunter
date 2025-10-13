package jwtmgr

import (
	"testing"
	"time"
)

func TestManagerWithKeyStoreGeneration(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	mgr := NewWithKeyStore("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, ks)

	access, refresh, exp, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if access == "" {
		t.Error("Access token is empty")
	}

	if refresh == "" {
		t.Error("Refresh token is empty")
	}

	if exp.IsZero() {
		t.Error("Expiration time is zero")
	}
}

func TestManagerWithKeyStoreValidation(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	mgr := NewWithKeyStore("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, ks)

	access, _, _, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	claims, err := mgr.ValidateToken(access, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Subject != "user123" {
		t.Errorf("Subject mismatch: got %s, want user123", claims.Subject)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Email mismatch: got %s, want test@example.com", claims.Email)
	}
}

func TestManagerKeyRotation(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid1, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate first key: %v", err)
	}

	mgr := NewWithKeyStore("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, ks)

	token1, _, _, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token with first key: %v", err)
	}

	kid2, err := ks.GenerateKey(2048, false, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	if actErr := ks.ActivateKey(kid2); actErr != nil {
		t.Fatalf("Failed to activate second key: %v", actErr)
	}

	token2, _, _, err := mgr.GeneratePair("user456", "test2@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token with second key: %v", err)
	}

	claims1, err := mgr.ValidateToken(token1, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate token from first key: %v", err)
	}

	if claims1.Subject != "user123" {
		t.Errorf("Subject mismatch for first token: got %s, want user123", claims1.Subject)
	}

	claims2, err := mgr.ValidateToken(token2, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate token from second key: %v", err)
	}

	if claims2.Subject != "user456" {
		t.Errorf("Subject mismatch for second token: got %s, want user456", claims2.Subject)
	}

	if kid1 == kid2 {
		t.Error("Key IDs should be different")
	}
}

func TestManagerGracePeriodValidation(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid1, err := ks.GenerateKey(2048, true, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate first key: %v", err)
	}

	mgr := NewWithKeyStore("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, ks)

	token1, _, _, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	_, _, _, err = mgr.GeneratePair("user456", "test2@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token with new key: %v", err)
	}

	_, err = mgr.ValidateToken(token1, "test-audience")
	if err != nil {
		t.Errorf("Token should still validate during grace period: %v", err)
	}

	if err := ks.ActivateKey(kid1); err != nil {
		t.Errorf("Should be able to activate key during grace period: %v", err)
	}
}

func TestManagerLegacyMode(t *testing.T) {
	legacyKey := []byte("test-secret-key-that-is-long-enough-for-testing")
	mgr := New("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, legacyKey)

	access, _, _, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token in legacy mode: %v", err)
	}

	claims, err := mgr.ValidateToken(access, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate token in legacy mode: %v", err)
	}

	if claims.Subject != "user123" {
		t.Errorf("Subject mismatch: got %s, want user123", claims.Subject)
	}
}

func TestManagerBackwardCompatibility(t *testing.T) {
	legacyKey := []byte("test-secret-key-that-is-long-enough-for-testing")
	mgrLegacy := New("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, legacyKey)

	legacyToken, _, _, err := mgrLegacy.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate legacy token: %v", err)
	}

	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	mgrNew := &Manager{
		issuer:     "test-issuer",
		audience:   "test-audience",
		accessTTL:  15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
		signingKey: legacyKey,
		keyStore:   ks,
	}

	_, err = mgrNew.ValidateToken(legacyToken, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate legacy token with new manager: %v", err)
	}

	newToken, _, _, err := mgrNew.GeneratePair("user456", "test2@example.com")
	if err != nil {
		t.Fatalf("Failed to generate new token: %v", err)
	}

	_, err = mgrNew.ValidateToken(newToken, "test-audience")
	if err != nil {
		t.Fatalf("Failed to validate new token: %v", err)
	}
}

func TestManagerInvalidAudience(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	mgr := NewWithKeyStore("test-issuer", "test-audience", 15*time.Minute, 7*24*time.Hour, ks)

	access, _, _, err := mgr.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = mgr.ValidateToken(access, "wrong-audience")
	if err == nil {
		t.Error("Expected error when validating token with wrong audience")
	}
}
