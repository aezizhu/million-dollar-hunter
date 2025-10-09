package jwtmgr

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
	signingKey []byte
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
			ID:        "",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(m.signingKey)
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
		return m.signingKey, nil
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
	if expectedAud != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == expectedAud {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid audience")
		}
	}
	foundCfg := false
	for _, aud := range claims.Audience {
		if aud == m.audience {
			foundCfg = true
			break
		}
	}
	if !foundCfg {
		return nil, errors.New("invalid audience")
	}
	return claims, nil
}
