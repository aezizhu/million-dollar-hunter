package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
