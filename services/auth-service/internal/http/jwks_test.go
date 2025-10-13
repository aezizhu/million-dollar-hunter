package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func TestJWKS_RS256Keys(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	mgr := jwtmgr.NewWithKeyStore("test-issuer", "test-aud", 15*time.Minute, 7*24*time.Hour, ks)
	server := &Server{JWT: mgr}

	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	server.JWKS(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("Expected Cache-Control public, max-age=3600, got %s", cacheControl)
	}

	var jwks JWKSResponse
	if err := json.NewDecoder(w.Body).Decode(&jwks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(jwks.Keys) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(jwks.Keys))
	}

	jwk := jwks.Keys[0]

	if jwk.Kty != "RSA" {
		t.Errorf("Expected kty=RSA, got %s", jwk.Kty)
	}

	if jwk.Use != "sig" {
		t.Errorf("Expected use=sig, got %s", jwk.Use)
	}

	if jwk.Kid != kid {
		t.Errorf("Expected kid=%s, got %s", kid, jwk.Kid)
	}

	if jwk.Alg != "RS256" {
		t.Errorf("Expected alg=RS256, got %s", jwk.Alg)
	}

	if jwk.N == "" {
		t.Error("Expected N (modulus) to be non-empty")
	}

	if jwk.E == "" {
		t.Error("Expected E (exponent) to be non-empty")
	}

	if len(jwk.N) < 342 {
		t.Errorf("Expected N to be at least 342 chars (base64url of 2048-bit key), got %d", len(jwk.N))
	}
}

func TestJWKS_OnlyRS256Included(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate RS256 key: %v", err)
	}

	mgr := jwtmgr.NewWithKeyStore("test-issuer", "test-aud", 15*time.Minute, 7*24*time.Hour, ks)
	server := &Server{JWT: mgr}

	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	server.JWKS(w, req)

	var jwks JWKSResponse
	if err := json.NewDecoder(w.Body).Decode(&jwks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	for _, jwk := range jwks.Keys {
		if jwk.Alg != "RS256" {
			t.Errorf("Non-RS256 key found in JWKS: %s", jwk.Alg)
		}
	}
}

func TestJWKS_LegacyModeNotSupported(t *testing.T) {
	legacyKey := []byte("test-secret-key-that-is-long-enough")
	mgr := jwtmgr.New("test-issuer", "test-aud", 15*time.Minute, 7*24*time.Hour, legacyKey)
	server := &Server{JWT: mgr}

	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	server.JWKS(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}
}

func TestJWKS_MultipleKeys(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid1, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate first key: %v", err)
	}

	kid2, err := ks.GenerateKey(2048, false, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	mgr := jwtmgr.NewWithKeyStore("test-issuer", "test-aud", 15*time.Minute, 7*24*time.Hour, ks)
	server := &Server{JWT: mgr}

	req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	server.JWKS(w, req)

	var jwks JWKSResponse
	if err := json.NewDecoder(w.Body).Decode(&jwks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(jwks.Keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(jwks.Keys))
	}

	kidMap := make(map[string]bool)
	for _, jwk := range jwks.Keys {
		kidMap[jwk.Kid] = true
	}

	if !kidMap[kid1] || !kidMap[kid2] {
		t.Error("Not all generated keys are in JWKS response")
	}
}
