package jwtmgr

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration

	mu         sync.RWMutex
	signingKey []byte
	keys       map[string][]byte
	currentKID string

	keyStore *KeyStore
}

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"eml"`
	jwt.RegisteredClaims
}

func New(issuer, audience string, accessTTL, refreshTTL time.Duration, signingKey []byte) *Manager {
	return &Manager{
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		signingKey: signingKey,
		keyStore:   nil,
	}
}

// NewWithKeyStore creates a new JWT manager with multi-key support
func NewWithKeyStore(issuer, audience string, accessTTL, refreshTTL time.Duration, ks *KeyStore) *Manager {
	return &Manager{
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		signingKey: nil,
		keyStore:   ks,
	}
}

func NewWithKeys(issuer, audience string, accessTTL, refreshTTL time.Duration, keys map[string][]byte, currentKID string) *Manager {
	cp := make(map[string][]byte, len(keys))
	for k, v := range keys {
		cp[k] = append([]byte(nil), v...)
	}
	return &Manager{
		issuer:     issuer,
		audience:   audience,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		keys:       cp,
		currentKID: currentKID,
	}
}

func (m *Manager) currentKey() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.keys != nil && m.currentKID != "" {
		if k, ok := m.keys[m.currentKID]; ok {
			return k
		}
	}
	return m.signingKey
}

func (m *Manager) keyByKID(kid string) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.keys == nil {
		return m.signingKey
	}
	if k, ok := m.keys[kid]; ok {
		return k
	}
	return m.signingKey
}

func (m *Manager) UpdateKeys(keys map[string][]byte, currentKID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make(map[string][]byte, len(keys))
	for k, v := range keys {
		cp[k] = append([]byte(nil), v...)
	}
	m.keys = cp
	m.currentKID = currentKID
}

func (m *Manager) GenerateToken(userID, email string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		},
	}

	if m.keyStore != nil {
		activeKey, err := m.keyStore.GetActiveKey()
		if err != nil {
			return "", time.Time{}, fmt.Errorf("failed to get active key: %w", err)
		}

		t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		t.Header["kid"] = activeKey.ID

		s, err := t.SignedString(activeKey.PrivateKey)
		return s, exp, err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if m.currentKID != "" {
		t.Header["kid"] = m.currentKID
	}
	s, err := t.SignedString(m.currentKey())
	return s, exp, err
}

func (m *Manager) GeneratePair(userID, email string) (string, string, time.Time, error) {
	access, exp, err := m.GenerateToken(userID, email, m.accessTTL)
	if err != nil {
		return "", "", time.Time{}, err
	}
	refresh, _, err := m.GenerateToken(userID, email, m.refreshTTL)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return access, refresh, exp, nil
}

// GetKeyStore returns the underlying KeyStore for JWKS endpoint access.
// Returns nil if running in legacy single-key mode.
func (m *Manager) GetKeyStore() *KeyStore {
	return m.keyStore
}

func (m *Manager) ValidateToken(tokenStr string, expectedAud string) (*Claims, error) {
	validMethods := []string{jwt.SigningMethodHS256.Alg(), jwt.SigningMethodRS256.Alg()}
	parser := jwt.NewParser(jwt.WithValidMethods(validMethods))

	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if m.keyStore != nil {
			kidRaw, ok := token.Header["kid"]
			if ok {
				kid, ok := kidRaw.(string)
				if !ok {
					return nil, errors.New("kid header is not a string")
				}
				key, err := m.keyStore.GetKey(kid)
				if err != nil {
					return nil, fmt.Errorf("failed to get key: %w", err)
				}
				return key.PublicKey, nil
			}
			if m.signingKey != nil {
				return m.signingKey, nil
			}
			return nil, errors.New("missing kid for RSA token")
		}

		if kid, ok := token.Header["kid"].(string); ok && kid != "" {
			return m.keyByKID(kid), nil
		}
		if k := m.currentKey(); k != nil {
			return k, nil
		}
		return nil, errors.New("no valid signing key found")
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != m.issuer {
		return nil, errors.New("invalid issuer")
	}

	hasCfg := false
	for _, aud := range claims.Audience {
		if aud == m.audience {
			hasCfg = true
			break
		}
	}
	if !hasCfg {
		return nil, errors.New("token missing required audience")
	}

	if expectedAud != "" && expectedAud != m.audience {
		hasExp := false
		for _, aud := range claims.Audience {
			if aud == expectedAud {
				hasExp = true
				break
			}
		}
		if !hasExp {
			return nil, fmt.Errorf("token missing expected audience: %s", expectedAud)
		}
	}

	if claims.Subject == "" {
		return nil, errors.New("missing subject")
	}

	return claims, nil
}
