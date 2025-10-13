package keystore

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore/pemutil"
)

type MemoryStore struct {
	mu        sync.RWMutex
	keys      []KeyRecord
	grace     time.Duration
	activeKID string
}

func NewMemoryStore(grace time.Duration, withInitial bool) (*MemoryStore, error) {
	if grace <= 0 {
		grace = 24 * time.Hour
	}
	ms := &MemoryStore{grace: grace}
	if withInitial {
		if _, rotErr := ms.Rotate(2048); rotErr != nil {
			return nil, rotErr
		}
	}
	return ms, nil
}

func (m *MemoryStore) ListPublic() []PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]PublicKey, 0, len(m.keys))
	for _, kr := range m.keys {
		if kr.Status != StatusActive && kr.Status != StatusGrace {
			continue
		}
		pub, pubErr := pemutil.ParseRSAPublicKeyFromPEM([]byte(kr.PublicPEM))
		if pubErr != nil {
			continue
		}
		n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString([]byte{byte(pub.E >> 16), byte(pub.E >> 8), byte(pub.E)})
		out = append(out, PublicKey{
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
			Kid: kr.Kid,
			N:   n,
			E:   e,
		})
	}
	return out
}

func (m *MemoryStore) GetActivePrivate() (*rsa.PrivateKey, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, kr := range m.keys {
		if kr.Kid == m.activeKID && kr.Status == StatusActive {
			priv, err := pemutil.ParseRSAPrivateKeyFromPEM([]byte(kr.PrivatePEM))
			return priv, kr.Kid, err
		}
	}
	return nil, "", errors.New("no active key")
}

func (m *MemoryStore) GetByKID(kid string) (*rsa.PublicKey, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, kr := range m.keys {
		if kr.Kid == kid && (kr.Status == StatusActive || kr.Status == StatusGrace) {
			pub, err := pemutil.ParseRSAPublicKeyFromPEM([]byte(kr.PublicPEM))
			if err != nil {
				return nil, false
			}
			return pub, true
		}
	}
	return nil, false
}

func (m *MemoryStore) Rotate(bits int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	priv, kid, err := GenerateRSAKey(bits)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	for i := range m.keys {
		if m.keys[i].Kid == m.activeKID && m.keys[i].Status == StatusActive {
			m.keys[i].Status = StatusGrace
			m.keys[i].NotAfter = now.Add(m.grace)
		}
	}
	privPEM := pemutil.EncodeRSAPrivateKeyToPEM(priv)
	pubPEM := pemutil.EncodeRSAPublicKeyToPEM(&priv.PublicKey)
	m.keys = append(m.keys, KeyRecord{
		Kid:        kid,
		CreatedAt:  now,
		Status:     StatusActive,
		NotAfter:   time.Time{},
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
	})
	m.activeKID = kid
	return kid, nil
}

func (m *MemoryStore) Cleanup(now time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	keep := m.keys[:0]
	removed := 0
	for _, kr := range m.keys {
		if kr.Status == StatusGrace && !kr.NotAfter.IsZero() && now.After(kr.NotAfter) {
			removed++
			continue
		}
		if kr.Status == StatusRetired && !kr.NotAfter.IsZero() && now.After(kr.NotAfter) {
			removed++
			continue
		}
		keep = append(keep, kr)
	}
	m.keys = keep
	return removed, nil
}

func (m *MemoryStore) GraceWindow() time.Duration {
	return m.grace
}
