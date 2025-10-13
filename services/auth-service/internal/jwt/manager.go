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

	mu        sync.RWMutex
	signingKey []byte
	keys       map[string][]byte
	currentKID string
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

func (m *Manager) ValidateToken(tokenStr string, expectedAud string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if kid, ok := token.Header["kid"].(string); ok && kid != "" {
			return m.keyByKID(kid), nil
		}
		return m.currentKey(), nil
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
	return claims, nil
}
