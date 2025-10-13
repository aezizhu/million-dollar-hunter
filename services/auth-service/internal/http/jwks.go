package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

// JWKSResponse represents the JSON Web Key Set response format per RFC 7517.
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key with RSA public key components.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

func (s *Server) JWKS(w http.ResponseWriter, r *http.Request) {
	type keystoreProvider interface {
		GetKeyStore() *jwtmgr.KeyStore
	}
	kp, ok := any(s.JWT).(keystoreProvider)
	if !ok {
		http.Error(w, "JWKS not supported in legacy mode", http.StatusNotImplemented)
		return
	}
	keyStore := kp.GetKeyStore()
	if keyStore == nil {
		http.Error(w, "JWKS not supported in legacy mode", http.StatusNotImplemented)
		return
	}

	keys := keyStore.ListKeys()
	jwks := JWKSResponse{
		Keys: make([]JWK, 0, len(keys)),
	}

	for _, key := range keys {
		if key.Algorithm != "RS256" {
			continue
		}

		n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())

		jwks.Keys = append(jwks.Keys, JWK{
			Kty: "RSA",
			Use: "sig",
			Kid: key.ID,
			Alg: "RS256",
			N:   n,
			E:   e,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_ = json.NewEncoder(w).Encode(jwks)
}
