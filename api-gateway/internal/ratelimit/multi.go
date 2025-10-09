package ratelimit

import (
	"context"
	"encoding/json"
	"time"
)

type SimpleLimiter interface {
	Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration)
}

type MultiLimiter struct {
	byKey map[string]SimpleLimiter
	def   SimpleLimiter
}

func NewMultiLimiter(defaultLimiter SimpleLimiter, byKey map[string]SimpleLimiter) *MultiLimiter {
	return &MultiLimiter{def: defaultLimiter, byKey: byKey}
}

func (m *MultiLimiter) Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration) {
	if l, ok := m.byKey[key]; ok {
		return l.Allow(ctx, key)
	}
	return m.def.Allow(ctx, key)
}

type RouteLimit struct {
	RPS   int `json:"rps"`
	Burst int `json:"burst"`
}

func ParseRouteLimitsJSON(s string) (map[string]RouteLimit, error) {
	if s == "" {
		return map[string]RouteLimit{}, nil
	}
	var m map[string]RouteLimit
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type LocalAdapter struct {
	Inner interface {
		Allow(key string) (bool, int, int, time.Time, time.Duration)
	}
}

func (l LocalAdapter) Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration) {
	return l.Inner.Allow(key)
}
