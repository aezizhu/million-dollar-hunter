package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

type tv struct{ m *jwtmgr.Manager }

func (t tv) ValidateToken(tokenStr string, expectedAud string) (*jwtmgr.Claims, error) {
	return t.m.ValidateToken(tokenStr, expectedAud)
}

func TestWithAuthSuccess(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("k"))
	token, _, _ := m.GenerateToken("u", "e@x.com", time.Minute)
	called := false
	h := WithAuth(tv{m}, "aud", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		cl := ClaimsFromContext(r)
		if cl == nil || cl.UserID != "u" {
			t.Fatalf("claims missing")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if !called || w.Code != http.StatusOK {
		t.Fatalf("auth middleware failed")
	}
}

func TestWithAuthUnauthorized(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("k"))
	h := WithAuth(tv{m}, "aud", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401")
	}
}
