package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
)

type stubLimiter struct {
	allowed    bool
	limit      int
	remaining  int
	reset      time.Time
	retryAfter time.Duration
}

func (s stubLimiter) Allow(key string) (bool, int, int, time.Time, time.Duration) {
	return s.allowed, s.limit, s.remaining, s.reset, s.retryAfter
}

func testEngine(l Limiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(l))
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestRateLimitAllowed(t *testing.T) {
	lim := stubLimiter{
		allowed:    true,
		limit:      10,
		remaining:  9,
		reset:      time.Now().Add(1 * time.Second),
		retryAfter: 0,
	}
	r := testEngine(lim)

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

func TestRateLimitBlocked(t *testing.T) {
	lim := stubLimiter{
		allowed:    false,
		limit:      10,
		remaining:  0,
		reset:      time.Now().Add(500 * time.Millisecond),
		retryAfter: 500 * time.Millisecond,
	}
	r := testEngine(lim)

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
