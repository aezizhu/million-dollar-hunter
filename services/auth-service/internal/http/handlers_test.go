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
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type fakeJWT struct{}

func (f *fakeJWT) GeneratePair(userID, email string) (string, string, time.Time, error) {
	return "access", "refresh", time.Now().Add(1 * time.Minute), nil
}

func (f *fakeJWT) ValidateToken(tokenStr string, expectedAud string) (*jwtmgr.Claims, error) {
	c := &jwtmgr.Claims{}
	return c, nil
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
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
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
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
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

func (f fakeUserStore) Create(ctx context.Context, email, passwordHash string) (store.User, error) {
	return store.User{}, nil
}

func (f fakeUserStore) GetByEmail(ctx context.Context, email string) (store.User, error) {
	return store.User{}, assertErr{}
}

type storeWithHash struct{ hash string }

func (s storeWithHash) Create(ctx context.Context, email, passwordHash string) (store.User, error) {
	return store.User{}, nil
}

func (s storeWithHash) GetByEmail(ctx context.Context, email string) (store.User, error) {
	return store.User{ID: "u-123", Email: "user@example.com", PasswordHash: s.hash}, nil
}

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

func TestLoginMultiUserSuccess(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "true")
	pw := "S3curePass!"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Fatalf("hash err: %v", err)
	}
	s := &Server{JWT: &fakeJWT{}, Store: storeWithHash{hash: hash}}
	body, _ := json.Marshal(LoginRequest{Username: "user@example.com", Password: pw})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp LoginResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.ExpiresIn <= 0 {
		t.Fatalf("missing tokens or expires")
	}
}

func TestLoginMultiUserBadPassword(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "true")
	pw := "S3curePass!"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Fatalf("hash err: %v", err)
	}
	s := &Server{JWT: &fakeJWT{}, Store: storeWithHash{hash: hash}}
	body, _ := json.Marshal(LoginRequest{Username: "user@example.com", Password: "WrongPass1!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginMVP_Misconfigured(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	os.Unsetenv("MVP_USERNAME")
	os.Unsetenv("MVP_PASSWORD")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "x", Password: "y"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

type fakeUserStoreOK struct{}

func (f *fakeUserStoreOK) Create(ctx context.Context, email, passwordHash string) (store.User, error) {
	return store.User{ID: "u-2", Email: email, PasswordHash: "$2a$10$1iKqkQ6q6o0t8W9u2pF1uOm5b2mSgT3j1bJ5o7fOAkf0vZK3Jk5e2"}, nil
}

func (f *fakeUserStoreOK) GetByEmail(ctx context.Context, email string) (store.User, error) {
	return store.User{ID: "u-2", Email: email, PasswordHash: "$2a$10$1iKqkQ6q6o0t8W9u2pF1uOm5b2mSgT3j1bJ5o7fOAkf0vZK3Jk5e2"}, nil
}

type fakeAuditManyFails struct{}

func (f *fakeAuditManyFails) Log(ctx context.Context, userID *string, event string, ip *string, ua *string) error {
	return nil
}

func (f *fakeAuditManyFails) CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error) {
	return 999, nil
}

func TestLoginMultiUser_Lockout(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	s := &Server{
		JWT:   &fakeJWT{},
		Store: &fakeUserStoreOK{},
		Audit: &fakeAuditManyFails{},
	}
	body, _ := json.Marshal(LoginRequest{Username: "user@example.com", Password: "AnyPass1!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

type fakeRefreshPersist struct {
	called bool
	user   string
	token  string
	exp    time.Time
}

func (f *fakeRefreshPersist) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (store.RefreshToken, error) {
	f.called, f.user, f.token, f.exp = true, userID, token, expiresAt
	return store.RefreshToken{UserID: userID, Token: token, ExpiresAt: expiresAt}, nil
}

func (f *fakeRefreshPersist) GetValidRefreshToken(ctx context.Context, token string, now time.Time) (store.RefreshToken, error) {
	return store.RefreshToken{}, assertErr{}
}
func (f *fakeRefreshPersist) RevokeRefreshToken(ctx context.Context, token string) error { return nil }
func (f *fakeRefreshPersist) RevokeAllForUser(ctx context.Context, userID string) error  { return nil }

func TestLoginMultiUserSuccess_PersistRefresh(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	pw := "S3curePass!"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Fatalf("hash err: %v", err)
	}
	ref := &fakeRefreshPersist{}
	s := &Server{JWT: &fakeJWT{}, Store: storeWithHash{hash: hash}, RefreshTokens: ref}
	body, _ := json.Marshal(LoginRequest{Username: "user@example.com", Password: pw})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !ref.called || ref.user == "" || ref.token == "" || ref.exp.Before(time.Now()) {
		t.Fatalf("expected refresh token persisted")
	}
}
func TestLogin_SQLInjectionEmail(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "' OR 1=1; --", Password: "anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400/401, got %d", w.Code)
	}
}

func TestLogin_XSSPayload(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "<script>alert(1)</script>", Password: "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("User-Agent", "<script>alert(1)</script>")
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 for XSS payload")
	}
}

func TestLogin_HeaderInjection(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
	s := &Server{JWT: &fakeJWT{}}
	body, _ := json.Marshal(LoginRequest{Username: "aezi", Password: "bad"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("X-Forwarded-For", "127.0.0.1\r\nX-Bad: evil")
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 for header injection")
	}
}

func TestLogin_ParameterPollution(t *testing.T) {
	os.Setenv("ENABLE_MULTI_USER", "false")
	os.Setenv("MVP_USERNAME", "aezi")
	os.Setenv("MVP_PASSWORD", "Aa@123456789")
	s := &Server{JWT: &fakeJWT{}}
	dup := []byte(`{"username":"aezi","username":"other","password":"Aa@123456789"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(dup))
	w := httptest.NewRecorder()
	s.Login(w, req)
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500 for duplicate parameter")
	}
}
