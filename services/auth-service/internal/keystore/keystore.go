package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusGrace   Status = "grace"
	StatusRetired Status = "retired"
)

type PublicKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type KeyRecord struct {
	Kid        string    `json:"kid"`
	CreatedAt  time.Time `json:"created_at"`
	Status     Status    `json:"status"`
	NotAfter   time.Time `json:"not_after"`
	PrivatePEM string    `json:"private_pem"`
	PublicPEM  string    `json:"public_pem"`
}

type KeyStore interface {
	ListPublic() []PublicKey
	GetActivePrivate() (*rsa.PrivateKey, string, error)
	GetByKID(kid string) (*rsa.PublicKey, bool)
	Rotate(bits int) (string, error)
	Cleanup(now time.Time) (int, error)
	GraceWindow() time.Duration
}

func GenerateRSAKey(bits int) (*rsa.PrivateKey, string, error) {
	if bits <= 0 {
		bits = 2048
	}
	priv, genErr := rsa.GenerateKey(rand.Reader, bits)
	if genErr != nil {
		return nil, "", genErr
	}
	nB := base64.RawURLEncoding.EncodeToString(priv.PublicKey.N.Bytes())
	eInt := priv.PublicKey.E
	eB := base64.RawURLEncoding.EncodeToString([]byte{byte(eInt >> 16), byte(eInt >> 8), byte(eInt)})
	sum := sha256.Sum256([]byte(nB + "." + eB))
	kid := base64.RawURLEncoding.EncodeToString(sum[:])
	return priv, kid, nil
}
