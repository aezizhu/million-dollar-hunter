package ratelimit

import (
	"context"
	"testing"
	"time"
)

type stubLimiter struct {
	allow bool
	lim   int
	rem   int
	reset time.Time
	retry time.Duration
}

func (s stubLimiter) Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration) {
	return s.allow, s.lim, s.rem, s.reset, s.retry
}

func TestParseCIDRs_IPv4(t *testing.T) {
	nets := parseCIDRs([]string{"192.168.1.0/24", "10.0.0.5"})
	if len(nets) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nets))
	}
}

func TestParseCIDRs_IPv6(t *testing.T) {
	nets := parseCIDRs([]string{"2001:db8::/32", "2001:db8::1"})
	if len(nets) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nets))
	}
}

func TestParseCIDRs_MalformedInput(t *testing.T) {
	nets := parseCIDRs([]string{"not-a-cidr", "999.999.999.999/99"})
	if len(nets) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(nets))
	}
}

func TestInAllowlist_EmptyIP(t *testing.T) {
	h := &HierarchicalLimiter{}
	if h.inAllowlist("") {
		t.Fatal("expected false for empty ip")
	}
}

func TestInAllowlist_Match(t *testing.T) {
	nets := parseCIDRs([]string{"192.168.1.0/24"})
	h := &HierarchicalLimiter{allowCIDRs: nets}
	if !h.inAllowlist("192.168.1.100") {
		t.Fatal("expected match")
	}
}

func TestInAllowlist_NoMatch(t *testing.T) {
	nets := parseCIDRs([]string{"10.0.0.0/8"})
	h := &HierarchicalLimiter{allowCIDRs: nets}
	if h.inAllowlist("192.168.1.1") {
		t.Fatal("expected no match")
	}
}

func TestAllow_BypassHeader(t *testing.T) {
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: false},
		userLimiter:  stubLimiter{allow: false},
		routeLimiter: stubLimiter{allow: false},
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", true)
	if !dec.Allowed {
		t.Fatal("expected allowed when bypass")
	}
}

func TestAllow_IPAllowlist(t *testing.T) {
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: false},
		userLimiter:  stubLimiter{allow: false},
		routeLimiter: stubLimiter{allow: false},
		allowCIDRs:   parseCIDRs([]string{"1.2.3.4/32"}),
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", false)
	if !dec.Allowed {
		t.Fatal("expected allowed for allowlisted ip")
	}
}

func TestAllow_IPBlocked(t *testing.T) {
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: false, lim: 10, rem: 0, reset: time.Now().Add(10 * time.Second), retry: 10 * time.Second},
		userLimiter:  stubLimiter{allow: true},
		routeLimiter: stubLimiter{allow: true},
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", false)
	if dec.Allowed || dec.Dimension != DimIP {
		t.Fatalf("expected ip denial, got %+v", dec)
	}
	if dec.Limit == 0 || dec.Reset.IsZero() || dec.RetryAfter == 0 {
		t.Fatalf("expected populated decision fields, got %+v", dec)
	}
}

func TestAllow_UserBlocked(t *testing.T) {
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: true},
		userLimiter:  stubLimiter{allow: false, lim: 5, rem: 0, reset: time.Now().Add(5 * time.Second), retry: 5 * time.Second},
		routeLimiter: stubLimiter{allow: true},
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", false)
	if dec.Allowed || dec.Dimension != DimUser {
		t.Fatalf("expected user denial, got %+v", dec)
	}
	if dec.Limit == 0 || dec.Reset.IsZero() || dec.RetryAfter == 0 {
		t.Fatalf("expected populated decision fields, got %+v", dec)
	}
}

func TestAllow_RouteBlocked(t *testing.T) {
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: true},
		userLimiter:  stubLimiter{allow: true},
		routeLimiter: stubLimiter{allow: false, lim: 3, rem: 0, reset: time.Now().Add(3 * time.Second), retry: 3 * time.Second},
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", false)
	if dec.Allowed || dec.Dimension != DimRoute {
		t.Fatalf("expected route denial, got %+v", dec)
	}
	if dec.Limit == 0 || dec.Reset.IsZero() || dec.RetryAfter == 0 {
		t.Fatalf("expected populated decision fields, got %+v", dec)
	}
}

func TestAllow_AllPassed(t *testing.T) {
	now := time.Now()
	h := &HierarchicalLimiter{
		ipLimiter:    stubLimiter{allow: true, lim: 100, rem: 99, reset: now.Add(time.Second)},
		userLimiter:  stubLimiter{allow: true, lim: 50, rem: 49, reset: now.Add(2 * time.Second)},
		routeLimiter: stubLimiter{allow: true, lim: 10, rem: 9, reset: now.Add(3 * time.Second)},
	}
	dec := h.Allow(context.Background(), "1.2.3.4", "u1", "/r", false)
	if !dec.Allowed {
		t.Fatalf("expected allowed, got %+v", dec)
	}
	if dec.Limit == 0 || dec.Reset.IsZero() {
		t.Fatalf("expected headers from route limiter, got %+v", dec)
	}
}
