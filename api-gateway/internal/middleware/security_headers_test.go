package middleware_test

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

func TestSecurityHeaders_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.EnableHSTS = false
	cfg.CSPPolicy = ""
	cfg.FrontendURL = "http://example.com"
	cfg.AuthValidateMode = "local"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get(headers.XContentTypeOptions); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
	if got := w.Header().Get(headers.ReferrerPolicy); got != "strict-origin-when-cross-origin" {
		t.Fatalf("expected Referrer-Policy strict-origin-when-cross-origin, got %q", got)
	}
	if got := w.Header().Get(headers.XFrameOptions); got != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", got)
	}
	if got := w.Header().Get(headers.StrictTransportSecurity); got != "" {
		t.Fatalf("expected no HSTS by default, got %q", got)
	}
	if got := w.Header().Get(headers.ContentSecurityPolicy); got != "" {
		t.Fatalf("expected no CSP by default, got %q", got)
	}
}

func TestSecurityHeaders_HSTSAndCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.HSTSEnabled = true
	cfg.HSTSMaxAge = 15552000
	cfg.CSPPolicy = "default-src 'self'"
	cfg.FrontendURL = "http://example.com"
	cfg.AuthValidateMode = "local"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get(headers.StrictTransportSecurity); got == "" {
		t.Fatalf("expected HSTS header when HTTPS, got %q", got)
	}
	if got := w.Header().Get(headers.ContentSecurityPolicy); got != cfg.CSPPolicy {
		t.Fatalf("expected CSP %q, got %q", cfg.CSPPolicy, got)
	}
}
func TestSecurityHeaders_HSTSGatedByHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.HSTSEnabled = true
	cfg.HSTSMaxAge = 15552000
	cfg.CSPPolicy = ""
	cfg.FrontendURL = "http://example.com"
	cfg.AuthValidateMode = "local"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get(headers.StrictTransportSecurity); got != "" {
		t.Fatalf("expected no HSTS without HTTPS signal, got %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req2.Header.Set("X-Forwarded-Proto", "https")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get(headers.StrictTransportSecurity); got == "" {
		t.Fatalf("expected HSTS when HTTPS signal present")
	}
}

func TestSecurityHeaders_ErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Load()
	cfg.HSTSEnabled = false
	cfg.CSPPolicy = ""
	cfg.FrontendURL = "http://example.com"
	cfg.AuthValidateMode = "local"
	reg := prometheus.NewRegistry()
	logger := zerolog.Nop()
	server.Register(r, cfg, logger, reg)

	r.GET("/abort401", func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) })
	r.GET("/abort500", func(c *gin.Context) { c.AbortWithStatus(http.StatusInternalServerError) })

	for _, path := range []string{"/abort401", "/abort500"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Header().Get(headers.XContentTypeOptions) == "" ||
			w.Header().Get(headers.ReferrerPolicy) == "" ||
			w.Header().Get(headers.XFrameOptions) == "" {
			t.Fatalf("expected security headers on error response for %s", path)
		}
	}
}
