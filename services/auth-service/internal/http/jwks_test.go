package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore"
)

func TestJWKS_OK(t *testing.T) {
	ks, _ := keystore.NewMemoryStore(24*time.Hour, true)
	j := jwtmgr.NewWithKeyStore("iss", "aud", time.Minute, time.Hour, ks, nil, false)
	s := &Server{JWT: j}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()
	s.JWKS(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatalf("empty body")
	}
}
