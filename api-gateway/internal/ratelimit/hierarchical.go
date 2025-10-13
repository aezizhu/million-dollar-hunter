package ratelimit

import (
	"context"
	"encoding/json"
	"net"
	"time"
)

type DenialDimension string

const (
	DimNone  DenialDimension = ""
	DimIP    DenialDimension = "ip"
	DimUser  DenialDimension = "user"
	DimRoute DenialDimension = "route"
)

type Decision struct {
	Allowed    bool
	Dimension  DenialDimension
	Limit      int
	Remaining  int
	Reset      time.Time
	RetryAfter time.Duration
}

type HierarchicalLimiter struct {
	ipLimiter    SimpleLimiter
	userLimiter  SimpleLimiter
	routeLimiter SimpleLimiter
	allowCIDRs   []*net.IPNet
}

func NewHierarchicalLimiter(ipLim, userLim, routeLim SimpleLimiter, allowlistJSON string) *HierarchicalLimiter {
	var entries []string
	if allowlistJSON != "" {
		_ = json.Unmarshal([]byte(allowlistJSON), &entries)
	}
	cidrs := parseCIDRs(entries)
	return &HierarchicalLimiter{ipLimiter: ipLim, userLimiter: userLim, routeLimiter: routeLim, allowCIDRs: cidrs}
}

func parseCIDRs(entries []string) []*net.IPNet {
	var out []*net.IPNet
	for _, e := range entries {
		if _, nw, err := net.ParseCIDR(e); err == nil && nw != nil {
			out = append(out, nw)
			continue
		}
		if ip := net.ParseIP(e); ip != nil {
			var cidr string
			if ip.To4() != nil {
				cidr = ip.String() + "/32"
			} else {
				cidr = ip.String() + "/128"
			}
			if _, nw, err := net.ParseCIDR(cidr); err == nil && nw != nil {
				out = append(out, nw)
			}
		}
	}
	return out
}

func (h *HierarchicalLimiter) inAllowlist(ipStr string) bool {
	if ipStr == "" {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, nw := range h.allowCIDRs {
		if nw.Contains(ip) {
			return true
		}
	}
	return false
}

func (h *HierarchicalLimiter) InAllowlist(ipStr string) bool {
	return h.inAllowlist(ipStr)
}

func (h *HierarchicalLimiter) Allow(ctx context.Context, ip, userID, route string, bypass bool) Decision {
	if bypass || h.inAllowlist(ip) {
		return Decision{Allowed: true, Dimension: DimNone, Limit: 0, Remaining: 0, Reset: time.Now(), RetryAfter: 0}
	}
	ipKey := "ratelimit:ip:" + ip
	ok, lim, rem, reset, retry := h.ipLimiter.Allow(ctx, ipKey)
	if !ok {
		return Decision{Allowed: false, Dimension: DimIP, Limit: lim, Remaining: rem, Reset: reset, RetryAfter: retry}
	}
	if userID != "" {
		uKey := "ratelimit:user:" + userID
		ok, lim, rem, reset, retry = h.userLimiter.Allow(ctx, uKey)
		if !ok {
			return Decision{Allowed: false, Dimension: DimUser, Limit: lim, Remaining: rem, Reset: reset, RetryAfter: retry}
		}
	}
	rKey := "ratelimit:route:" + route
	if userID != "" {
		rKey += ":" + userID
	}
	ok, lim, rem, reset, retry = h.routeLimiter.Allow(ctx, rKey)
	if !ok {
		return Decision{Allowed: false, Dimension: DimRoute, Limit: lim, Remaining: rem, Reset: reset, RetryAfter: retry}
	}
	return Decision{Allowed: true, Dimension: DimNone, Limit: lim, Remaining: rem, Reset: reset, RetryAfter: 0}
}
