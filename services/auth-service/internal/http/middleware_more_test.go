package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func TestWithAuthWrongAudience(t *testing.T) {
	m := jwtmgr.New("iss","aud", time.Minute, time.Hour, []byte("k"))
	token, _, _ := m.GenerateToken("u","e@x.com", time.Minute)
	h := WithAuth(tv{m}, "other-aud", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong audience")
	}
}

func TestWithAuthMissingHeader(t *testing.T) {
	m := jwtmgr.New("iss","aud", time.Minute, time.Hour, []byte("k"))
	h := WithAuth(tv{m}, "aud", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without Authorization header")
	}
}

func TestClaimsFromContextNil(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if c := ClaimsFromContext(req); c != nil {
		t.Fatalf("expected nil claims from plain context")
	}
}
