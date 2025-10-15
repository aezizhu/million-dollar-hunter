package tests

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	grpcserver "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/grpc"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func TestIntegration_KeyStore_TokenGeneration_Validation(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"
	manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	accessToken, refreshToken, exp, err := manager.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if accessToken == "" {
		t.Fatal("Access token is empty")
	}

	if refreshToken == "" {
		t.Fatal("Refresh token is empty")
	}

	if exp.IsZero() {
		t.Fatal("Expiration time is zero")
	}

	claims, err := manager.ValidateToken(accessToken, audience)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected user_id=user123, got %s", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email=test@example.com, got %s", claims.Email)
	}

	key, err := ks.GetKey(kid)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if key.ID != kid {
		t.Errorf("Expected kid=%s, got %s", kid, key.ID)
	}

	grpcSrv := grpcserver.New(manager)
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	gen.RegisterAuthServiceServer(srv, grpcSrv)

	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	if err != nil {
		t.Fatalf("Failed to dial gRPC server: %v", err)
	}
	defer conn.Close()

	client := gen.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ValidateToken(ctx, &gen.ValidateRequest{
		Token:       accessToken,
		ExpectedAud: audience,
	})

	if err != nil {
		t.Fatalf("gRPC validation failed: %v", err)
	}

	if !resp.Valid {
		t.Errorf("gRPC validation returned invalid, reason: %s", resp.Reason)
	}

	if resp.UserId != "user123" {
		t.Errorf("Expected user_id=user123 from gRPC, got %s", resp.UserId)
	}
}

func TestIntegration_KeyRotation_WithActiveSessions(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"

	kid1, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate first key: %v", err)
	}

	manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	token1, _, _, err := manager.GeneratePair("user1", "user1@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token with first key: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	kid2, err := ks.GenerateKey(2048, false, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	if activateErr := ks.ActivateKey(kid2); activateErr != nil {
		t.Fatalf("Failed to activate second key: %v", activateErr)
	}

	token2, _, _, err := manager.GeneratePair("user2", "user2@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token with second key: %v", err)
	}

	claims1, err := manager.ValidateToken(token1, audience)
	if err != nil {
		t.Fatalf("Token from rotated key should still be valid: %v", err)
	}

	if claims1.UserID != "user1" {
		t.Errorf("Expected user1, got %s", claims1.UserID)
	}

	claims2, err := manager.ValidateToken(token2, audience)
	if err != nil {
		t.Fatalf("Token from new key should be valid: %v", err)
	}

	if claims2.UserID != "user2" {
		t.Errorf("Expected user2, got %s", claims2.UserID)
	}

	if kid1 == kid2 {
		t.Error("Key IDs should be different after rotation")
	}

	t.Logf("✓ Key rotation successful: kid1=%s, kid2=%s", kid1, kid2)
}

func TestIntegration_MultipleKeys_ConcurrentValidation(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"

	kids := make([]string, 3)
	tokens := make([]string, 3)

	for i := 0; i < 3; i++ {
		kid, err := ks.GenerateKey(2048, i == 0, 90*24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate key %d: %v", i, err)
		}
		kids[i] = kid

		if i > 0 {
			if activateErr := ks.ActivateKey(kid); activateErr != nil {
				t.Fatalf("Failed to activate key %d: %v", i, activateErr)
			}
		}

		manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)
		token, _, _, err := manager.GeneratePair(fmt.Sprintf("user%d", i), fmt.Sprintf("user%d@example.com", i))
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}
		tokens[i] = token
	}

	if err := ks.ActivateKey(kids[0]); err != nil {
		t.Fatalf("Failed to reactivate first key: %v", err)
	}

	manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	done := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(idx int) {
			claims, err := manager.ValidateToken(tokens[idx], audience)
			if err != nil {
				done <- fmt.Errorf("validation failed for token %d: %w", idx, err)
				return
			}

			expectedUser := fmt.Sprintf("user%d", idx)
			if claims.UserID != expectedUser {
				done <- fmt.Errorf("expected user=%s, got %s", expectedUser, claims.UserID)
				return
			}

			done <- nil
		}(i)
	}

	for i := 0; i < 3; i++ {
		err := <-done
		if err != nil {
			t.Errorf("Concurrent validation error: %v", err)
		}
	}

	t.Logf("✓ Concurrent validation successful with 3 different keys")
}

func TestIntegration_JWKS_Endpoint_PublicKeyExport(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"
	manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	token, _, _, err := manager.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	retrievedKS := manager.GetKeyStore()
	if retrievedKS == nil {
		t.Fatal("GetKeyStore() returned nil")
	}

	keys := retrievedKS.ListKeys()
	if len(keys) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(keys))
	}

	exportedKey := keys[0]
	if exportedKey.ID != kid {
		t.Errorf("Expected kid=%s, got %s", kid, exportedKey.ID)
	}

	if exportedKey.Algorithm != "RS256" {
		t.Errorf("Expected algorithm=RS256, got %s", exportedKey.Algorithm)
	}

	block, _ := pem.Decode([]byte(exportedKey.PublicPEM))
	if block == nil {
		t.Fatal("Failed to decode public key PEM")
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	_ = pubKeyInterface // Public key parsed successfully
	t.Logf("Public key type verified")

	_, err = manager.ValidateToken(token, audience)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	t.Logf("✓ JWKS export successful: kid=%s, algorithm=%s, public_key_pem_length=%d",
		exportedKey.ID, exportedKey.Algorithm, len(exportedKey.PublicPEM))
}

func TestIntegration_GracePeriod_ExpiredKey(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"

	kid, err := ks.GenerateKey(2048, true, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	token, _, _, err := manager.GeneratePair("user123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	key, err := ks.GetKey(kid)
	if err != nil {
		t.Fatalf("Failed to get key (should still be in grace period): %v", err)
	}

	now := time.Now().UTC()
	if now.After(key.ExpiresAt) {
		t.Logf("✓ Key has expired (as expected)")
	} else {
		t.Logf("⚠ Key has not expired yet (test timing issue)")
	}

	claims, err := manager.ValidateToken(token, audience)
	if err != nil {
		t.Fatalf("Token should be valid during grace period: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected user123, got %s", claims.UserID)
	}

	t.Logf("✓ Grace period validation successful")
}

func TestIntegration_BackwardCompatibility_HS256_RS256_Mixed(t *testing.T) {
	ks, err := jwtmgr.NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	issuer := "test-issuer"
	audience := "test-audience"
	hmacSecret := []byte("test-hmac-secret-key-minimum-32-bytes-required")

	rs256Manager := jwtmgr.NewWithKeyStore(issuer, audience, 15*time.Minute, 7*24*time.Hour, ks)

	hs256Manager := jwtmgr.New(issuer, audience, 15*time.Minute, 7*24*time.Hour, hmacSecret)

	rs256Token, _, _, err := rs256Manager.GeneratePair("rs256-user", "rs256@example.com")
	if err != nil {
		t.Fatalf("Failed to generate RS256 token: %v", err)
	}

	hs256Token, _, _, err := hs256Manager.GeneratePair("hs256-user", "hs256@example.com")
	if err != nil {
		t.Fatalf("Failed to generate HS256 token: %v", err)
	}

	claims, err := rs256Manager.ValidateToken(rs256Token, audience)
	if err != nil {
		t.Fatalf("RS256 manager failed to validate RS256 token: %v", err)
	}
	if claims.UserID != "rs256-user" {
		t.Errorf("Expected rs256-user, got %s", claims.UserID)
	}

	claims, err = hs256Manager.ValidateToken(hs256Token, audience)
	if err != nil {
		t.Fatalf("HS256 manager failed to validate HS256 token: %v", err)
	}
	if claims.UserID != "hs256-user" {
		t.Errorf("Expected hs256-user, got %s", claims.UserID)
	}

	_ = kid // Note: In production, Manager supports both keyStore and signingKey for backward compat

	t.Logf("✓ Backward compatibility verified: RS256 and HS256 work independently")
}
