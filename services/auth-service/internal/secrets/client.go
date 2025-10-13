package secrets

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("secret not found")

type Client interface {
	Get(ctx context.Context, name string) (string, error)
	GetJSON(ctx context.Context, name string, v any) error
	StartBackgroundRefresh(ctx context.Context)
	Health(ctx context.Context) error
}

type Config struct {
	CacheTTL       time.Duration
	RefreshInterval time.Duration
}

type cacheEntry struct {
	val       string
	expiresAt time.Time
}

type baseClient struct {
	cfg   Config
	mu    sync.RWMutex
	cache map[string]cacheEntry
	now   func() time.Time
}

func newBase(cfg Config) baseClient {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = time.Hour
	}
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = time.Minute
	}
	return baseClient{
		cfg:   cfg,
		cache: make(map[string]cacheEntry),
		now:   time.Now,
	}
}

func (b *baseClient) getCached(name string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if ce, ok := b.cache[name]; ok {
		if b.now().Before(ce.expiresAt) {
			return ce.val, true
		}
	}
	return "", false
}

func (b *baseClient) setCached(name, val string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache[name] = cacheEntry{val: val, expiresAt: b.now().Add(b.cfg.CacheTTL)}
}
