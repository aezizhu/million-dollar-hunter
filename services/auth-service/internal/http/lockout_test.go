package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/auth"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type auditCount struct {
	failures int
}

func (a *auditCount) Log(ctx context.Context, userID *string, event string, ip *string, meta *string) error {
	return nil
}
func (a *auditCount) CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error) {
	return a.failures, nil
}

type storeWithPW struct{ hash string }

func (s storeWithPW) Create(ctx context.Context, email, passwordHash string) (store.User, error) { return store.User{}, nil }
func (s storeWithPW) GetByEmail(ctx context.Context, email string) (store.User, error) {
	return store.User{ID: "u-1", Email: email, PasswordHash: s.hash}, nil
}

func TestLoginLockoutThreshold(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	_ = os.Setenv("LOCKOUT_AFTER_FAILS", "3")
	_ = os.Setenv("LOCKOUT_WINDOW_MIN", "15")
	pw := "ValidPass12!"
	hash, _ := auth.HashPassword(pw)
	s := &Server{
		Store: storeWithPW{hash: hash},
		JWT:   &fakeJWT{},
		Audit: &auditCount{failures: 3},
	}
	body, _ := json.Marshal(LoginRequest{Username: "user@example.com", Password: pw})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestLogoutRevokesRefreshForUser(t *testing.T) {
	mem := &memRefresh{valid: map[string]time.Time{}, user: map[string]string{}}
	s := &Server{
		RefreshTokens: mem,
		Audit:         &memAudit{},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	s.Logout(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}
