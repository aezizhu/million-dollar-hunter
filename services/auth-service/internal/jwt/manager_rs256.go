package jwtmgr

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore"
)

type RSManager struct {
	issuer      string
	audience    string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	ks          keystore.KeyStore
	legacyKey   []byte
	allowLegacy bool
}

func NewWithKeyStore(issuer, audience string, accessTTL, refreshTTL time.Duration, ks keystore.KeyStore, legacyHSKey []byte, allowLegacy bool) *RSManager {
	return &RSManager{
		issuer:      issuer,
		audience:    audience,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		ks:          ks,
		legacyKey:   legacyHSKey,
		allowLegacy: allowLegacy,
	}
}

func (m *RSManager) GenerateToken(userID, email string, ttl time.Duration) (string, time.Time, error) {
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
	priv, kid, err := m.ks.GetActivePrivate()
	if err != nil {
		return "", time.Time{}, err
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = kid
	s, signErr := t.SignedString(priv)
	return s, exp, signErr
}

func (m *RSManager) GeneratePair(userID, email string) (string, string, time.Time, error) {
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

func (m *RSManager) ValidateToken(tokenStr string, expectedAud string) (*Claims, error) {
	parserRS := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	var lastErr error
	token, err := parserRS.ParseWithClaims(tokenStr, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
		if kidVal, ok := tok.Header["kid"].(string); ok && kidVal != "" {
			if pub, found := m.ks.GetByKID(kidVal); found {
				return pub, nil
			}
		}
		lastErr = errors.New("kid not found")
		return nil, lastErr
	})
	if err == nil && token != nil && token.Valid {
		if claims, ok := token.Claims.(*Claims); ok {
			if claims.Issuer != m.issuer {
				return nil, errors.New("invalid issuer")
			}
			if !audHas(claims.Audience, m.audience) {
				return nil, errors.New("token missing required audience")
			}
			if expectedAud != "" && expectedAud != m.audience && !audHas(claims.Audience, expectedAud) {
				return nil, fmt.Errorf("token missing expected audience: %s", expectedAud)
			}
			return claims, nil
		}
	}
	for _, jwk := range m.ks.ListPublic() {
		pub, ok := m.ks.GetByKID(jwk.Kid)
		if !ok {
			continue
		}
		tok2, err2 := parserRS.ParseWithClaims(tokenStr, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
			return pub, nil
		})
		if err2 == nil && tok2 != nil && tok2.Valid {
			claims, okC := tok2.Claims.(*Claims)
			if !okC {
				continue
			}
			if claims.Issuer != m.issuer {
				continue
			}
			if !audHas(claims.Audience, m.audience) {
				continue
			}
			if expectedAud != "" && expectedAud != m.audience && !audHas(claims.Audience, expectedAud) {
				continue
			}
			return claims, nil
		}
	}
	if m.allowLegacy && len(m.legacyKey) > 0 {
		parserHS := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		tok, vErr := parserHS.ParseWithClaims(tokenStr, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
			return m.legacyKey, nil
		})
		if vErr == nil && tok != nil && tok.Valid {
			claims, ok := tok.Claims.(*Claims)
			if ok && claims.Issuer == m.issuer && audHas(claims.Audience, m.audience) {
				if expectedAud != "" && expectedAud != m.audience && !audHas(claims.Audience, expectedAud) {
					return nil, fmt.Errorf("token missing expected audience: %s", expectedAud)
				}
				return claims, nil
			}
		}
	}
	return nil, errors.New("invalid token")
}

func audHas(auds jwt.ClaimStrings, val string) bool {
	for _, a := range auds {
		if a == val {
			return true
		}
	}
	return false
}
