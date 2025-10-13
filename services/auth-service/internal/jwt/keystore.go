package jwtmgr

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SigningKey struct {
	ID         string          `json:"id"`
	Algorithm  string          `json:"algorithm"`
	PrivateKey *rsa.PrivateKey `json:"-"`
	PublicKey  *rsa.PublicKey  `json:"-"`
	PrivatePEM string          `json:"private_pem"`
	PublicPEM  string          `json:"public_pem"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  time.Time       `json:"expires_at"`
	Active     bool            `json:"active"`
}

type KeyStore struct {
	mu       sync.RWMutex
	keys     map[string]*SigningKey
	filePath string
}

func NewKeyStore(filePath string) (*KeyStore, error) {
	ks := &KeyStore{
		keys:     make(map[string]*SigningKey),
		filePath: filePath,
	}

	if filePath != "" {
		if err := ks.loadFromFile(); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load keystore: %w", err)
			}
		}
	}

	return ks, nil
}

func (ks *KeyStore) GenerateKey(bitSize int, active bool, expiresIn time.Duration) (string, error) {
	if bitSize < 2048 {
		return "", errors.New("key size must be at least 2048 bits")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	privatePEM, publicPEM, err := encodeKeyPair(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to encode key pair: %w", err)
	}

	pubKeyBytes, marshalErr := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if marshalErr != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", marshalErr)
	}
	hash := sha256.Sum256(pubKeyBytes)
	kid := hex.EncodeToString(hash[:8])

	now := time.Now().UTC()
	key := &SigningKey{
		ID:         kid,
		Algorithm:  "RS256",
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		PrivatePEM: privatePEM,
		PublicPEM:  publicPEM,
		CreatedAt:  now,
		ExpiresAt:  now.Add(expiresIn),
		Active:     active,
	}

	ks.mu.Lock()
	ks.keys[kid] = key
	ks.mu.Unlock()

	if ks.filePath != "" {
		if saveErr := ks.saveToFile(); saveErr != nil {
			return "", fmt.Errorf("failed to save keystore: %w", saveErr)
		}
	}

	return kid, nil
}

func (ks *KeyStore) GetActiveKey() (*SigningKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	now := time.Now().UTC()
	for _, key := range ks.keys {
		if key.Active && now.Before(key.ExpiresAt) {
			return key, nil
		}
	}

	return nil, errors.New("no active signing key found")
}

func (ks *KeyStore) GetKey(kid string) (*SigningKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	key, ok := ks.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	now := time.Now().UTC()
	graceExpiry := key.ExpiresAt.Add(24 * time.Hour)
	if now.After(graceExpiry) {
		return nil, fmt.Errorf("key expired (past grace period): %s", kid)
	}

	return key, nil
}

func (ks *KeyStore) ActivateKey(kid string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	key, ok := ks.keys[kid]
	if !ok {
		return fmt.Errorf("key not found: %s", kid)
	}

	for _, k := range ks.keys {
		k.Active = false
	}

	key.Active = true

	if ks.filePath != "" {
		if saveErr := ks.saveToFile(); saveErr != nil {
			return fmt.Errorf("failed to save keystore: %w", saveErr)
		}
	}

	return nil
}

func (ks *KeyStore) ListKeys() []*SigningKey {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	now := time.Now().UTC()
	keys := make([]*SigningKey, 0, len(ks.keys))

	for _, key := range ks.keys {
		graceExpiry := key.ExpiresAt.Add(24 * time.Hour)
		if now.Before(graceExpiry) {
			keys = append(keys, key)
		}
	}

	return keys
}

func (ks *KeyStore) CleanupExpiredKeys() (int, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	now := time.Now().UTC()
	removed := 0

	for kid, key := range ks.keys {
		graceExpiry := key.ExpiresAt.Add(7 * 24 * time.Hour)
		if now.After(graceExpiry) {
			delete(ks.keys, kid)
			removed++
		}
	}

	if removed > 0 && ks.filePath != "" {
		if saveErr := ks.saveToFile(); saveErr != nil {
			return removed, fmt.Errorf("failed to save keystore after cleanup: %w", saveErr)
		}
	}

	return removed, nil
}

func (ks *KeyStore) saveToFile() error {
	dir := filepath.Dir(ks.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create keystore directory: %w", err)
	}

	keys := make([]*SigningKey, 0, len(ks.keys))
	for _, key := range ks.keys {
		keys = append(keys, key)
	}

	data, err := json.MarshalIndent(map[string]interface{}{
		"keys": keys,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	tmpFile := ks.filePath + ".tmp"
	if writeErr := os.WriteFile(tmpFile, data, 0600); writeErr != nil {
		return fmt.Errorf("failed to write keystore file: %w", writeErr)
	}

	if renameErr := os.Rename(tmpFile, ks.filePath); renameErr != nil {
		return fmt.Errorf("failed to rename keystore file: %w", renameErr)
	}

	return nil
}

func (ks *KeyStore) loadFromFile() error {
	data, err := os.ReadFile(ks.filePath)
	if err != nil {
		return err
	}

	var payload struct {
		Keys []*SigningKey `json:"keys"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal keystore: %w", err)
	}

	for _, key := range payload.Keys {
		privateKey, publicKey, decodeErr := decodeKeyPair(key.PrivatePEM, key.PublicPEM)
		if decodeErr != nil {
			return fmt.Errorf("failed to decode key pair for kid %s: %w", key.ID, decodeErr)
		}

		key.PrivateKey = privateKey
		key.PublicKey = publicKey
		ks.keys[key.ID] = key
	}

	return nil
}

func encodeKeyPair(privateKey *rsa.PrivateKey) (string, string, error) {
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})

	return string(privatePEM), string(publicPEM), nil
}

func decodeKeyPair(privatePEM, publicPEM string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateBlock, _ := pem.Decode([]byte(privatePEM))
	if privateBlock == nil {
		return nil, nil, errors.New("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicBlock, _ := pem.Decode([]byte(publicPEM))
	if publicBlock == nil {
		return nil, nil, errors.New("failed to decode public key PEM")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("public key is not RSA")
	}

	return privateKey, publicKey, nil
}
