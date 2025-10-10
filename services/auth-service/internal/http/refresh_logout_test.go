package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/golang-jwt/jwt/v5"
)

func TestLogoutOK(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	s.Logout(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestRefreshNotImplemented(t *testing.T) {
	s := &Server{}
	body, _ := json.Marshal(RefreshRequest{RefreshToken: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Refresh(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501")
	}
}

func TestRefreshBadJSON(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte("{bad")))
	w := httptest.NewRecorder()
	s.Refresh(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}
func TestRefreshUnauthorized_InvalidJWTSignature(t *testing.T) {
	mGood := jwtmgr.New("mdh-auth", "mdh-api", time.Minute, 5*time.Minute, []byte("good-key"))
	mBad := jwtmgr.New("mdh-auth", "mdh-api", time.Minute, 5*time.Minute, []byte("bad-key"))
	_, badRefresh, _, _ := mBad.GeneratePair("u-1", "u@example.com")

	s := &Server{
		JWT:           &jwtFixed{j: mGood},
		RefreshTokens: &memRefresh{},
	}
	body, _ := json.Marshal(RefreshRequest{RefreshToken: badRefresh})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Refresh(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when JWT signature invalid, got %d", w.Code)
	}
}

type fakeAudit struct{ lastEvent string; lastUser *string }
func (f *fakeAudit) Log(ctx context.Context, userID *string, event string, ip *string, ua *string) error { f.lastEvent = event; f.lastUser = userID; return nil }
func (f *fakeAudit) CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error) { return 0, nil }

type fakeRevokeAll struct{ user string; called bool }
func (f *fakeRevokeAll) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (store.RefreshToken, error) { return store.RefreshToken{}, nil }
func (f *fakeRevokeAll) GetValidRefreshToken(ctx context.Context, token string, now time.Time) (store.RefreshToken, error) { return store.RefreshToken{}, assertErr{} }
func (f *fakeRevokeAll) RevokeRefreshToken(ctx context.Context, token string) error { return nil }
func (f *fakeRevokeAll) RevokeAllForUser(ctx context.Context, userID string) error { f.user = userID; f.called = true; return nil }

func TestLogout_WithAuditAndRevokeAll(t *testing.T) {
	claims := &jwtmgr.Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "u-abc"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxClaimsKey, claims))
	w := httptest.NewRecorder()
	a := &fakeAudit{}
	r := &fakeRevokeAll{}
	s := &Server{Audit: a, RefreshTokens: r}
	s.Logout(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
	if a.lastEvent != "logout" || a.lastUser == nil || *a.lastUser != "u-abc" {
		t.Fatalf("audit not recorded")
	}
	if !r.called || r.user != "u-abc" {
		t.Fatalf("revoke-all not called")
	}
}
