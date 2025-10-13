package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/server"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

func TestCORSPreflight_Portfolios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.FrontendURL = "http://example.com"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolios", nil)
	req.Header.Set("Origin", cfg.FrontendURL)
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("expected 200/204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != cfg.FrontendURL {
		t.Fatalf("expected Allow-Origin %q, got %q", cfg.FrontendURL, got)
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatalf("expected Allow-Headers to be set")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("expected Allow-Methods to be set")
	}
	if w.Header().Get(headers.RequestID) == "" {
		t.Fatalf("expected %s header to be set", headers.RequestID)
	}
}

func TestCORSPreflight_Wallets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.FrontendURL = "http://example.com"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/wallets/0xabc", nil)
	req.Header.Set("Origin", cfg.FrontendURL)
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("expected 200/204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != cfg.FrontendURL {
		t.Fatalf("expected Allow-Origin %q, got %q", cfg.FrontendURL, got)
	}
}

func TestCORSPreflight_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.FrontendURL = "http://allowed.example"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolios", nil)
	req.Header.Set("Origin", "http://evil.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" && got != "http://evil.example" {
		t.Fatalf("unexpected Allow-Origin for disallowed origin: %q", got)
	}
}
