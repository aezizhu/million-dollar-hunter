package jwtmgr

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKeyStoreGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "keystore.json")

	ks, err := NewKeyStore(keystorePath)
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if kid == "" {
		t.Fatal("Generated key ID is empty")
	}

	key, err := ks.GetKey(kid)
	if err != nil {
		t.Fatalf("Failed to get generated key: %v", err)
	}

	if key.ID != kid {
		t.Errorf("Key ID mismatch: got %s, want %s", key.ID, kid)
	}

	if key.Algorithm != "RS256" {
		t.Errorf("Algorithm mismatch: got %s, want RS256", key.Algorithm)
	}

	if !key.Active {
		t.Error("Key should be active")
	}
}

func TestKeyStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "keystore.json")

	ks1, err := NewKeyStore(keystorePath)
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks1.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	ks2, err := NewKeyStore(keystorePath)
	if err != nil {
		t.Fatalf("Failed to load keystore: %v", err)
	}

	key, err := ks2.GetKey(kid)
	if err != nil {
		t.Fatalf("Failed to get key from reloaded keystore: %v", err)
	}

	if key.ID != kid {
		t.Errorf("Key ID mismatch after reload: got %s, want %s", key.ID, kid)
	}

	if key.PrivateKey == nil {
		t.Error("Private key not loaded")
	}

	if key.PublicKey == nil {
		t.Error("Public key not loaded")
	}
}

func TestKeyStoreActivation(t *testing.T) {
	ks, err := NewKeyStore("")
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

	activeKey, err := ks.GetActiveKey()
	if err != nil {
		t.Fatalf("Failed to get active key: %v", err)
	}

	if activeKey.ID != kid1 {
		t.Errorf("Active key mismatch: got %s, want %s", activeKey.ID, kid1)
	}

	if actErr := ks.ActivateKey(kid2); actErr != nil {
		t.Fatalf("Failed to activate second key: %v", actErr)
	}

	activeKey, err = ks.GetActiveKey()
	if err != nil {
		t.Fatalf("Failed to get active key after activation: %v", err)
	}

	if activeKey.ID != kid2 {
		t.Errorf("Active key mismatch after activation: got %s, want %s", activeKey.ID, kid2)
	}

	key1, err := ks.GetKey(kid1)
	if err != nil {
		t.Fatalf("Failed to get first key: %v", err)
	}

	if key1.Active {
		t.Error("First key should no longer be active")
	}
}

func TestKeyStoreGracePeriod(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	kid, err := ks.GenerateKey(2048, true, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = ks.GetActiveKey()
	if err == nil {
		t.Error("Expected error when getting active expired key")
	}

	_, err = ks.GetKey(kid)
	if err != nil {
		t.Errorf("Key should still be accessible during grace period: %v", err)
	}
}

func TestKeyStoreCleanup(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	removed, err := ks.CleanupExpiredKeys()
	if err != nil {
		t.Fatalf("Failed to cleanup keys: %v", err)
	}

	if removed != 0 {
		t.Errorf("Should not remove keys still in grace period, removed %d", removed)
	}
}

func TestKeyStoreMinimumKeySize(t *testing.T) {
	ks, err := NewKeyStore("")
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(1024, true, 90*24*time.Hour)
	if err == nil {
		t.Error("Expected error when generating key with size < 2048")
	}
}

func TestKeyStoreListKeys(t *testing.T) {
	ks, err := NewKeyStore("")
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

	keys := ks.ListKeys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	kidMap := make(map[string]bool)
	for _, key := range keys {
		kidMap[key.ID] = true
	}

	if !kidMap[kid1] || !kidMap[kid2] {
		t.Error("Not all generated keys are in the list")
	}
}

func TestKeyStoreFilePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "keystore.json")

	ks, err := NewKeyStore(keystorePath)
	if err != nil {
		t.Fatalf("Failed to create keystore: %v", err)
	}

	_, err = ks.GenerateKey(2048, true, 90*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if _, statErr := os.Stat(keystorePath); os.IsNotExist(statErr) {
		t.Error("Keystore file was not created")
	}

	data, err := os.ReadFile(keystorePath)
	if err != nil {
		t.Fatalf("Failed to read keystore file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Keystore file is empty")
	}
}
