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

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type fakeJWT struct{}

func (f *fakeJWT) GeneratePair(userID, email string) (string, string, time.Time, error) {
	return "access", "refresh", time.Now().Add(1 * time.Minute), nil
}

func TestHealth(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	s.Health(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestLoginMVP(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "aezi", Password: "Aa@123456789"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp LoginResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("tokens missing")
	}
}

func TestLoginUnauthorized(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "bad", Password: "nope"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginBadJSON(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	s := &Server{JWT: &fakeJWT{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("{badjson")))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

type fakeUserStore struct{}
func (f fakeUserStore) Create(ctx context.Context, email, passwordHash string) (store.User, error) { return store.User{}, nil }
func (f fakeUserStore) GetByEmail(ctx context.Context, email string) (store.User, error)           { return store.User{}, assertErr{} }

type assertErr struct{}
func (assertErr) Error() string { return "not found" }

func TestLoginMultiUserUnauthorized(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "true")
	s := &Server{JWT: &fakeJWT{}, Store: fakeUserStore{}}
	body, _ := json.Marshal(LoginRequest{Username: "x", Password: "y"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
