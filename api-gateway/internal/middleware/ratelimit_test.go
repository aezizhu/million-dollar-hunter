package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
)

type simpleStub struct {
	ok        bool
	limit     int
	remaining int
	reset     time.Time
	retry     time.Duration
}

func (s simpleStub) Allow(_ context.Context, _ string) (bool, int, int, time.Time, time.Duration) {
	return s.ok, s.limit, s.remaining, s.reset, s.retry
}

func newHierStub(ipOK, userOK, routeOK bool, lim, rem int, retry time.Duration) *ratelimit.HierarchicalLimiter {
	now := time.Now().Add(1 * time.Second)
	ip := simpleStub{ok: ipOK, limit: lim, remaining: rem, reset: now, retry: retry}
	user := simpleStub{ok: userOK, limit: lim, remaining: rem, reset: now, retry: retry}
	route := simpleStub{ok: routeOK, limit: lim, remaining: rem, reset: now, retry: retry}
	return ratelimit.NewHierarchicalLimiter(ip, user, route, "")
}

func testEngineHier(h *ratelimit.HierarchicalLimiter, cfg config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitHier(h, cfg))
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestRateLimitHier_Allowed(t *testing.T) {
	h := newHierStub(true, true, true, 10, 9, 0)
	r := testEngineHier(h, config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get(headers.RateLimit); got != "10" {
		t.Fatalf("X-RateLimit-Limit expected 10, got %s", got)
	}
	if got := w.Header().Get(headers.RateRemaining); got != "9" {
		t.Fatalf("X-RateLimit-Remaining expected 9, got %s", got)
	}
	if got := w.Header().Get(headers.RateReset); got == "" {
		t.Fatalf("X-RateLimit-Reset expected set")
	}
	if got := w.Header().Get(headers.RetryAfter); got != "" {
		t.Fatalf("Retry-After should be empty when allowed, got %s", got)
	}
}

func TestRateLimitHier_Blocked(t *testing.T) {
	h := newHierStub(true, true, false, 10, 0, 500*time.Millisecond)
	r := testEngineHier(h, config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if got := w.Header().Get(headers.RateLimit); got != "10" {
		t.Fatalf("X-RateLimit-Limit expected 10, got %s", got)
	}
	if got := w.Header().Get(headers.RateRemaining); got != "0" {
		t.Fatalf("X-RateLimit-Remaining expected 0, got %s", got)
	}
	if got := w.Header().Get(headers.RateReset); got == "" {
		t.Fatalf("X-RateLimit-Reset expected set")
	}
	if got := w.Header().Get(headers.RetryAfter); got != "0" && got != "1" {
		t.Fatalf("Retry-After expected small integer seconds, got %s", got)
	}
}
