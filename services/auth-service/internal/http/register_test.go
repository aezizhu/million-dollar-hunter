package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type fakeStore struct {
	create func(email, hash string) (store.User, error)
}

func (f fakeStore) Create(_ context.Context, email, passwordHash string) (store.User, error) {
	return f.create(email, passwordHash)
}

func (f fakeStore) GetByEmail(_ context.Context, email string) (store.User, error) {
	return store.User{}, nil
}

func TestRegisterBadJSON(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	s := &Server{Store: fakeStore{create: func(email, hash string) (store.User, error) {
		return store.User{}, nil
	}}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("{badjson")))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegisterFlagOff(t *testing.T) {
	_ = os.Unsetenv("ENABLE_MULTI_USER")
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(`{"email":"a@b.com","password":"ValidPass12!"}`)))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501")
	}
}

func TestRegisterSuccess(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	fs := fakeStore{
		create: func(email, hash string) (store.User, error) {
			return store.User{ID: "u-1", Email: email, PasswordHash: hash}, nil
		},
	}
	s := &Server{Store: fs}
	body, _ := json.Marshal(RegisterRequest{Email: "user@example.com", Password: "ValidPass12!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	fs := fakeStore{
		create: func(email, hash string) (store.User, error) {
			return store.User{}, nil
		},
	}
	s := &Server{Store: fs}
	body, _ := json.Marshal(RegisterRequest{Email: "invalid-email", Password: "ValidPass12!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	fs := fakeStore{
		create: func(email, hash string) (store.User, error) {
			return store.User{}, nil
		},
	}
	s := &Server{Store: fs}
	body, _ := json.Marshal(RegisterRequest{Email: "user@example.com", Password: "short1!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegisterConflict(t *testing.T) {
	_ = os.Setenv("ENABLE_MULTI_USER", "true")
	defer os.Unsetenv("ENABLE_MULTI_USER")
	fs := fakeStore{
		create: func(email, hash string) (store.User, error) {
			return store.User{}, fmt.Errorf("duplicate")
		},
	}
	s := &Server{Store: fs}
	body, _ := json.Marshal(RegisterRequest{Email: "user@example.com", Password: "ValidPass12!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Register(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}
