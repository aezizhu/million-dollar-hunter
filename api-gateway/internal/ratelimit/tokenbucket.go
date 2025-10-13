package ratelimit

import (
	"sync"
	"time"
)

type LocalTokenBucket struct {
	mu       sync.Mutex
	rate     int
	burst    int
	interval time.Duration
	state    map[string]*bucket
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

func NewLocalTokenBucket(rate, burst int, interval time.Duration) *LocalTokenBucket {
	return &LocalTokenBucket{
		rate:     rate,
		burst:    burst,
		interval: interval,
		state:    make(map[string]*bucket),
	}
}

func (l *LocalTokenBucket) refill(b *bucket, now time.Time) {
	if b.lastRefill.IsZero() {
		b.lastRefill = now
		b.tokens = l.burst
		return
	}
	elapsed := now.Sub(b.lastRefill)
	if elapsed <= 0 {
		return
	}
	tokensToAdd := float64(elapsed) / float64(l.interval) * float64(l.rate)
	add := int(tokensToAdd)
	if add > 0 {
		b.tokens += add
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		advance := time.Duration(float64(add) / float64(l.rate) * float64(l.interval))
		b.lastRefill = b.lastRefill.Add(advance)
	}
}

func (l *LocalTokenBucket) Allow(key string) (bool, int, int, time.Time, time.Duration) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.state[key]
	if !ok {
		b = &bucket{}
		l.state[key] = b
	}
	l.refill(b, now)
	if b.tokens <= 0 {
		reset := b.lastRefill.Add(l.interval)
		retry := reset.Sub(now)
		if retry < time.Second {
			retry = time.Second
		}
		return false, l.rate, 0, reset, retry
	}
	b.tokens--
	remaining := b.tokens
	reset := b.lastRefill.Add(l.interval)
	return true, l.rate, remaining, reset, 0
}
